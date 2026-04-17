package mitm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

// Authority encapsulates CA loading and issuing short-lived certificates for domains.
type Authority struct {
	caCert  *x509.Certificate
	caKey   *rsa.PrivateKey
	tlsCert tls.Certificate
	mu      sync.Mutex
	cache   map[string]tls.Certificate
	leafTTL time.Duration
}

func LoadAuthority(certPath, keyPath string) (*Authority, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return LoadAuthorityFromPEM(certPEM, keyPEM)
}

// LoadAuthorityFromPEM loads a CA from PEM content without temporary files.
func LoadAuthorityFromPEM(certPEM, keyPEM []byte) (*Authority, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("mitm: invalid CA certificate PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	kblk, _ := pem.Decode(keyPEM)
	if kblk == nil {
		return nil, errors.New("mitm: invalid CA key PEM")
	}

	var caKey *rsa.PrivateKey
	switch kblk.Type {
	case "RSA PRIVATE KEY":
		caKey, err = x509.ParsePKCS1PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
	case "PRIVATE KEY":
		pk, err := x509.ParsePKCS8PrivateKey(kblk.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		caKey, ok = pk.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("mitm: only RSA keys are supported for CA")
		}
	default:
		return nil, errors.New("mitm: unknown CA key PEM block type")
	}

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &Authority{
		caCert:  caCert,
		caKey:   caKey,
		tlsCert: tlsCert,
		cache:   make(map[string]tls.Certificate),
		leafTTL: 24 * time.Hour,
	}, nil
}

// GenerateDevCA generates a self-signed development root CA (RSA).
func GenerateDevCA(commonName string, yearsValid int) (certPEM, keyPEM []byte, err error) {
	if yearsValid <= 0 {
		yearsValid = 5
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().Add(-5 * time.Minute)
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             now,
		NotAfter:              now.AddDate(yearsValid, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		SubjectKeyId:          []byte{1, 2, 3, 4, 5, 6},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}

// IssueFor issues or returns from cache a certificate for host.
func (a *Authority) IssueFor(host string) (tls.Certificate, error) {
	h := normalizeHost(host)
	if h == "" {
		return tls.Certificate{}, errors.New("mitm: empty host for certificate issuance")
	}

	a.mu.Lock()
	if cert, ok := a.cache[h]; ok {
		a.mu.Unlock()
		return cert, nil
	}
	a.mu.Unlock()

	leafKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}
	now := time.Now().Add(-5 * time.Minute)
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: h},
		NotBefore:             now,
		NotAfter:              now.Add(a.leafTTL),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{h},
	}
	if ip := net.ParseIP(h); ip != nil {
		tmpl.IPAddresses = []net.IP{ip}
		tmpl.DNSNames = nil
		tmpl.Subject = pkix.Name{CommonName: ip.String()}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, a.caCert, &leafKey.PublicKey, a.caKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(leafKey)})
	leaf, err := tls.X509KeyPair(append(certPEM, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: a.caCert.Raw})...), keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}

	a.mu.Lock()
	if cached, ok := a.cache[h]; ok {
		a.mu.Unlock()
		return cached, nil
	}
	a.cache[h] = leaf
	a.mu.Unlock()
	return leaf, nil
}

// RootCertificate returns the parsed root certificate.
func (a *Authority) RootCertificate() *x509.Certificate {
	if a == nil {
		return nil
	}
	return a.caCert
}

// PrivateKey returns the parsed RSA private key used by the authority.
func (a *Authority) PrivateKey() *rsa.PrivateKey {
	if a == nil {
		return nil
	}
	return a.caKey
}

// TLSCertificate returns the loaded tls.Certificate for the root CA.
func (a *Authority) TLSCertificate() tls.Certificate {
	if a == nil {
		return tls.Certificate{}
	}
	return a.tlsCert
}

// CacheSize reports the number of cached leaf certificates.
func (a *Authority) CacheSize() int {
	if a == nil {
		return 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.cache)
}

// HasCached reports whether a normalized host is present in the cache.
func (a *Authority) HasCached(host string) bool {
	if a == nil {
		return false
	}
	h := normalizeHost(host)
	if h == "" {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.cache[h]
	return ok
}

func normalizeHost(host string) string {
	h := strings.TrimSpace(host)
	if h == "" {
		return ""
	}
	if strings.Contains(h, ":") {
		if v, _, err := net.SplitHostPort(h); err == nil {
			h = v
		}
	}
	return h
}
