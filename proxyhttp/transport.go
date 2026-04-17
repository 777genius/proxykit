package proxyhttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	http2 "golang.org/x/net/http2"
)

type TransportConfig struct {
	InsecureTLS bool
}

type ForwardedHeaderConfig struct {
	ClientIP string
	Proto    string
	Via      string
}

// NewTransport centralizes proxy-oriented http.Transport creation.
func NewTransport(cfg TransportConfig) *http.Transport {
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 25 * time.Second,
	}
	if cfg.InsecureTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	_ = http2.ConfigureTransport(tr)
	return tr
}

// RemoveHopHeaders strips hop-by-hop headers before proxying.
func RemoveHopHeaders(h http.Header) {
	hop := []string{"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade"}
	for _, k := range hop {
		h.Del(k)
	}
}

// ApplyForwardedHeaders sets standard proxy forwarding headers when values are
// available in cfg.
func ApplyForwardedHeaders(h http.Header, cfg ForwardedHeaderConfig) {
	if cfg.ClientIP != "" {
		h.Set("X-Forwarded-For", cfg.ClientIP)
	}
	if cfg.Proto != "" {
		h.Set("X-Forwarded-Proto", cfg.Proto)
	}
	if cfg.Via != "" {
		h.Set("Via", cfg.Via)
	}
}

// IsWebSocketRequest reports whether the request looks like a WebSocket upgrade.
func IsWebSocketRequest(r *http.Request) bool {
	if r.Header.Get("Upgrade") == "websocket" {
		return true
	}
	if r.Header.Get("Sec-WebSocket-Key") != "" || r.Header.Get("Sec-WebSocket-Version") != "" {
		return true
	}
	return false
}

// IsAbsoluteURL reports whether s looks like an absolute URI.
func IsAbsoluteURL(s string) bool {
	if u, err := url.Parse(s); err == nil {
		return u.Scheme != "" && u.Host != ""
	}
	return false
}
