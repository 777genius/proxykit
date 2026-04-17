package connect

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cherrypick-agency/proxykit/observe"
)

func TestHandler_TunnelsBytesAndEmitsProtocolEvent(t *testing.T) {
	clientConn, clientPeer := net.Pipe()
	upstreamConn, upstreamPeer := net.Pipe()
	defer clientPeer.Close()
	defer upstreamPeer.Close()

	var mu sync.Mutex
	var events []observe.ProtocolEvent
	var closed []observe.CloseInfo

	handler := New(Options{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return upstreamConn, nil
		},
		GenerateSessionID: func() string { return "connect-1" },
		Hooks: observe.Hooks{
			OnProtocolEvent: func(_ context.Context, event observe.ProtocolEvent) error {
				mu.Lock()
				defer mu.Unlock()
				events = append(events, event)
				return nil
			},
			OnSessionClose: func(_ context.Context, info observe.CloseInfo) {
				mu.Lock()
				defer mu.Unlock()
				closed = append(closed, info)
			},
		},
	})

	rw := newHijackResponseWriter(clientConn)
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	req.Host = "example.com:443"

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rw, req)
		close(done)
	}()

	clientReader := bufio.NewReader(clientPeer)
	statusLine, err := clientReader.ReadString('\n')
	if err != nil {
		t.Fatalf("read CONNECT response status: %v", err)
	}
	if statusLine != "HTTP/1.1 200 Connection Established\r\n" {
		t.Fatalf("status line = %q", statusLine)
	}
	for {
		line, err := clientReader.ReadString('\n')
		if err != nil {
			t.Fatalf("read CONNECT response header: %v", err)
		}
		if line == "\r\n" {
			break
		}
	}

	go func() {
		_, _ = clientPeer.Write([]byte("ping"))
	}()
	gotUpstream := make([]byte, 4)
	if _, err := io.ReadFull(upstreamPeer, gotUpstream); err != nil {
		t.Fatalf("read upstream tunneled payload: %v", err)
	}
	if string(gotUpstream) != "ping" {
		t.Fatalf("upstream payload = %q, want %q", string(gotUpstream), "ping")
	}

	go func() {
		_, _ = upstreamPeer.Write([]byte("pong"))
	}()
	gotClient := make([]byte, 4)
	if _, err := io.ReadFull(clientReader, gotClient); err != nil {
		t.Fatalf("read client tunneled payload: %v", err)
	}
	if string(gotClient) != "pong" {
		t.Fatalf("client payload = %q, want %q", string(gotClient), "pong")
	}

	_ = clientPeer.Close()
	_ = upstreamPeer.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not finish after closing tunnel")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 1 {
		t.Fatalf("events = %d, want 1", len(events))
	}
	if events[0].Name != "tunnel_established" {
		t.Fatalf("event name = %q, want %q", events[0].Name, "tunnel_established")
	}
	if len(closed) != 1 {
		t.Fatalf("closed = %d, want 1", len(closed))
	}
	if closed[0].Err != nil {
		t.Fatalf("close err = %v, want nil", closed[0].Err)
	}
}

func TestHandler_DialErrorReturnsRaw502AndCloses(t *testing.T) {
	clientConn, clientPeer := net.Pipe()
	defer clientPeer.Close()

	handler := New(Options{
		DialContext: func(context.Context, string, string) (net.Conn, error) {
			return nil, errors.New("boom")
		},
	})

	rw := newHijackResponseWriter(clientConn)
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	req.Host = "example.com:443"

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rw, req)
		close(done)
	}()

	buf := make([]byte, len(defaultDialError))
	if _, err := io.ReadFull(clientPeer, buf); err != nil {
		t.Fatalf("read dial error response: %v", err)
	}
	if string(buf) != string(defaultDialError) {
		t.Fatalf("raw dial error = %q, want %q", string(buf), string(defaultDialError))
	}
	<-done
}

func TestHandler_RejectsNonConnect(t *testing.T) {
	handler := New(Options{})
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandler_NoHijack(t *testing.T) {
	handler := New(Options{})
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	req.Host = "example.com:443"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), ErrHijackNotSupported.Error()) {
		t.Fatalf("body = %q, want message containing %q", rr.Body.String(), ErrHijackNotSupported.Error())
	}
}

type hijackResponseWriter struct {
	conn   net.Conn
	bufrw  *bufio.ReadWriter
	header http.Header
}

func newHijackResponseWriter(conn net.Conn) *hijackResponseWriter {
	return &hijackResponseWriter{
		conn:   conn,
		bufrw:  bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		header: make(http.Header),
	}
}

func (w *hijackResponseWriter) Header() http.Header {
	return w.header
}

func (w *hijackResponseWriter) Write(p []byte) (int, error) {
	return w.bufrw.Write(p)
}

func (w *hijackResponseWriter) WriteHeader(statusCode int) {}

func (w *hijackResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, w.bufrw, nil
}
