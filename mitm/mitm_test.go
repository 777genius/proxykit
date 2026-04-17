package mitm

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestLoadAuthorityFromPEM_Valid(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("GenerateDevCA() error = %v", err)
	}

	authority, err := LoadAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadAuthorityFromPEM() error = %v", err)
	}
	if authority.RootCertificate() == nil {
		t.Fatal("RootCertificate() returned nil")
	}
	if authority.PrivateKey() == nil {
		t.Fatal("PrivateKey() returned nil")
	}
	if authority.CacheSize() != 0 {
		t.Fatalf("CacheSize() = %d, want 0", authority.CacheSize())
	}
}

func TestAuthority_IssueFor_CachesNormalizedHost(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("GenerateDevCA() error = %v", err)
	}
	authority, err := LoadAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadAuthorityFromPEM() error = %v", err)
	}

	first, err := authority.IssueFor("example.com:443")
	if err != nil {
		t.Fatalf("IssueFor() error = %v", err)
	}
	second, err := authority.IssueFor("example.com")
	if err != nil {
		t.Fatalf("IssueFor() second error = %v", err)
	}
	if len(first.Certificate) != len(second.Certificate) {
		t.Fatalf("certificate chain length mismatch: %d vs %d", len(first.Certificate), len(second.Certificate))
	}
	if !authority.HasCached("example.com") {
		t.Fatal("expected normalized host to be cached")
	}
	if authority.CacheSize() != 1 {
		t.Fatalf("CacheSize() = %d, want 1", authority.CacheSize())
	}
}

func TestPolicy_ShouldIntercept(t *testing.T) {
	certPEM, keyPEM, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("GenerateDevCA() error = %v", err)
	}
	authority, err := LoadAuthorityFromPEM(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("LoadAuthorityFromPEM() error = %v", err)
	}
	policy := Policy{
		Authority:   authority,
		AllowSuffix: []string{".example.com"},
		DenySuffix:  []string{"bad.example.com"},
	}
	if !policy.ShouldIntercept("good.example.com:443") {
		t.Fatal("expected allowed host with port to intercept")
	}
	if policy.ShouldIntercept("bad.example.com") {
		t.Fatal("expected deny list to win")
	}
	if policy.ShouldIntercept("other.com") {
		t.Fatal("expected host outside allow list to be rejected")
	}
}

func TestEncodeCertificatePEM_WritesPEM(t *testing.T) {
	certPEM, _, err := GenerateDevCA("Test CA", 1)
	if err != nil {
		t.Fatalf("GenerateDevCA() error = %v", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("pem.Decode() returned nil block")
	}
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		t.Fatalf("ParseCertificate() error = %v", err)
	}
	var buf bytes.Buffer
	if err := EncodeCertificatePEM(&buf, block.Bytes); err != nil {
		t.Fatalf("EncodeCertificatePEM() error = %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("BEGIN CERTIFICATE")) || !bytes.Contains(buf.Bytes(), []byte("END CERTIFICATE")) {
		t.Fatalf("unexpected PEM output: %q", buf.String())
	}
}
