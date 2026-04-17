package wsproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestQueryTargetResolver(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/wsproxy", nil)
	req.URL.RawQuery = "_target=" + url.QueryEscape("https://example.com/socket.io?EIO=4&foo=1&foo=2&bar=1") + "&foo=9&z=2&_resetCapture=true"

	resolver := QueryTargetResolver{
		Param:           "_target",
		DefaultTarget:   "",
		DropParams:      []string{"_resetCapture"},
		NormalizeScheme: true,
	}
	u, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if u.Scheme != "wss" {
		t.Fatalf("expected wss scheme, got %s", u.Scheme)
	}
	if u.Query().Get("foo") != "9" || u.Query().Get("z") != "2" || u.Query().Get("bar") != "1" {
		t.Fatalf("unexpected merged query: %s", u.RawQuery)
	}
	if u.Query().Get("_resetCapture") != "" {
		t.Fatalf("drop param leaked into target query: %s", u.RawQuery)
	}
}

func TestHandler_ProxyTextFramesAndHooks(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade upstream: %v", err)
		}
		defer conn.Close()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, data); err != nil {
				return
			}
		}
	}))
	defer upstream.Close()

	target, _ := url.Parse(upstream.URL)
	target.Scheme = "ws"
	target.Path = "/echo"

	var mu sync.Mutex
	var opened, closed int
	var frameCount int
	closedCh := make(chan struct{}, 1)

	handler, err := New(Options{
		Resolver: ResolveTargetFunc(func(*http.Request) (*url.URL, error) {
			u := *target
			return &u, nil
		}),
		GenerateSessionID: func() string { return "sess-1" },
		Hooks: Hooks{
			OnSessionOpen: func(context.Context, Session) error {
				mu.Lock()
				opened++
				mu.Unlock()
				return nil
			},
			OnFrame: func(context.Context, Frame) error {
				mu.Lock()
				frameCount++
				mu.Unlock()
				return nil
			},
			OnSessionClose: func(context.Context, CloseInfo) {
				mu.Lock()
				closed++
				mu.Unlock()
				select {
				case closedCh <- struct{}{}:
				default:
				}
			},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	proxySrv := httptest.NewServer(handler)
	defer proxySrv.Close()

	wsURL := "ws" + strings.TrimPrefix(proxySrv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}

	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		t.Fatalf("write proxy frame: %v", err)
	}
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read proxy frame: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected echoed payload: %q", data)
	}
	_ = conn.Close()

	select {
	case <-closedCh:
	case <-time.After(2 * time.Second):
		t.Fatal("session close hook was not called")
	}

	mu.Lock()
	defer mu.Unlock()
	if opened != 1 {
		t.Fatalf("opened = %d, want 1", opened)
	}
	if closed != 1 {
		t.Fatalf("closed = %d, want 1", closed)
	}
	if frameCount < 2 {
		t.Fatalf("frameCount = %d, want at least 2", frameCount)
	}
}
