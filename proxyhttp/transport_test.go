package proxyhttp

import (
	"net/http"
	"testing"
	"time"
)

func TestIsWebSocketRequest(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Upgrade", "websocket")
	if !IsWebSocketRequest(req) {
		t.Fatal("expected websocket request")
	}

	req2, _ := http.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Sec-WebSocket-Key", "abc")
	if !IsWebSocketRequest(req2) {
		t.Fatal("expected websocket request from key header")
	}

	req3, _ := http.NewRequest(http.MethodGet, "/", nil)
	if IsWebSocketRequest(req3) {
		t.Fatal("unexpected websocket request detection")
	}
}

func TestNewTransport_DefaultSettings(t *testing.T) {
	tr := NewTransport(TransportConfig{})
	if tr == nil {
		t.Fatal("NewTransport returned nil")
	}
	if tr.TLSHandshakeTimeout != 10*time.Second {
		t.Errorf("TLSHandshakeTimeout = %v, want 10s", tr.TLSHandshakeTimeout)
	}
	if tr.ResponseHeaderTimeout != 25*time.Second {
		t.Errorf("ResponseHeaderTimeout = %v, want 25s", tr.ResponseHeaderTimeout)
	}
	if tr.IdleConnTimeout != 90*time.Second {
		t.Errorf("IdleConnTimeout = %v, want 90s", tr.IdleConnTimeout)
	}
	if tr.ExpectContinueTimeout != 1*time.Second {
		t.Errorf("ExpectContinueTimeout = %v, want 1s", tr.ExpectContinueTimeout)
	}
	if tr.MaxIdleConns != 100 {
		t.Errorf("MaxIdleConns = %d, want 100", tr.MaxIdleConns)
	}
	if tr.TLSClientConfig != nil && tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("transport should not skip verify by default")
	}
	if tr.Proxy == nil {
		t.Fatal("transport should use ProxyFromEnvironment")
	}
}

func TestNewTransport_InsecureTLS(t *testing.T) {
	tr := NewTransport(TransportConfig{InsecureTLS: true})
	if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("transport should skip verify in insecure mode")
	}
}

func TestRemoveHopHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Connection", "keep-alive")
	h.Set("Proxy-Connection", "keep-alive")
	h.Set("Keep-Alive", "1")
	h.Set("Proxy-Authenticate", "x")
	h.Set("Proxy-Authorization", "x")
	h.Set("Te", "trailers")
	h.Set("Trailer", "X")
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Upgrade", "websocket")
	h.Set("X", "ok")
	RemoveHopHeaders(h)
	if h.Get("X") != "ok" {
		t.Fatal("non-hop header should remain")
	}
	if h.Get("Connection")+h.Get("Proxy-Connection")+h.Get("Keep-Alive")+h.Get("Proxy-Authenticate")+h.Get("Proxy-Authorization")+h.Get("Te")+h.Get("Trailer")+h.Get("Transfer-Encoding")+h.Get("Upgrade") != "" {
		t.Fatalf("hop headers must be removed: %+v", h)
	}
}

func TestApplyForwardedHeaders(t *testing.T) {
	h := http.Header{}
	ApplyForwardedHeaders(h, ForwardedHeaderConfig{
		ClientIP: "127.0.0.1",
		Proto:    "https",
		Via:      "proxykit",
	})
	if h.Get("X-Forwarded-For") != "127.0.0.1" {
		t.Fatalf("X-Forwarded-For = %q, want %q", h.Get("X-Forwarded-For"), "127.0.0.1")
	}
	if h.Get("X-Forwarded-Proto") != "https" {
		t.Fatalf("X-Forwarded-Proto = %q, want %q", h.Get("X-Forwarded-Proto"), "https")
	}
	if h.Get("Via") != "proxykit" {
		t.Fatalf("Via = %q, want %q", h.Get("Via"), "proxykit")
	}
}

func TestIsAbsoluteURL(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"ws://localhost:8080", true},
		{"/path/to/resource", false},
		{"", false},
		{"http://", false},
		{"example.com", false},
		{"ht tp://bad url", false},
	}

	for _, tt := range tests {
		if got := IsAbsoluteURL(tt.in); got != tt.want {
			t.Fatalf("IsAbsoluteURL(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
