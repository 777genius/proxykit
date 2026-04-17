package forward

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/777genius/proxykit/observe"
	"github.com/777genius/proxykit/proxyhttp"
)

func TestHandler_ForwardsAbsoluteRequestAndObservesRoundTrip(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/submit" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Forwarded-For") != "127.0.0.1" {
			t.Fatalf("X-Forwarded-For = %q, want %q", r.Header.Get("X-Forwarded-For"), "127.0.0.1")
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("X-Test = %q, want %q", r.Header.Get("X-Test"), "yes")
		}
		if r.Header.Get("Connection") != "" {
			t.Fatalf("Connection hop header should be stripped, got %q", r.Header.Get("Connection"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"echo":"` + string(body) + `"}`))
	}))
	defer upstream.Close()

	var mu sync.Mutex
	var requests []observe.HTTPRequest
	var responses []observe.HTTPResponse
	var roundTrips []observe.HTTPRoundTrip
	var closed int

	handler := New(Options{
		GenerateSessionID: func() string { return "sess-forward-1" },
		ForwardedHeaders: proxyhttp.ForwardedHeaderConfig{
			ClientIP: "127.0.0.1",
			Proto:    "http",
			Via:      "proxykit",
		},
		MutateRequest: func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-Test", "yes")
			return nil
		},
		Hooks: observe.Hooks{
			OnHTTPRequest: func(_ context.Context, req observe.HTTPRequest) error {
				mu.Lock()
				defer mu.Unlock()
				requests = append(requests, req)
				return nil
			},
			OnHTTPResponse: func(_ context.Context, resp observe.HTTPResponse) error {
				mu.Lock()
				defer mu.Unlock()
				responses = append(responses, resp)
				return nil
			},
			OnHTTPRoundTrip: func(_ context.Context, rt observe.HTTPRoundTrip) error {
				mu.Lock()
				defer mu.Unlock()
				roundTrips = append(roundTrips, rt)
				return nil
			},
			OnSessionClose: func(context.Context, observe.CloseInfo) {
				mu.Lock()
				defer mu.Unlock()
				closed++
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, upstream.URL+"/submit", strings.NewReader("hello"))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusAccepted)
	}
	if rr.Body.String() != `{"echo":"hello"}` {
		t.Fatalf("unexpected response body: %q", rr.Body.String())
	}

	mu.Lock()
	defer mu.Unlock()
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if requests[0].URL != upstream.URL+"/submit" {
		t.Fatalf("request URL = %q, want %q", requests[0].URL, upstream.URL+"/submit")
	}
	if string(requests[0].Body.Bytes) != "hello" {
		t.Fatalf("request body sample = %q, want %q", string(requests[0].Body.Bytes), "hello")
	}
	if len(responses) != 1 {
		t.Fatalf("responses = %d, want 1", len(responses))
	}
	if responses[0].StatusCode != http.StatusAccepted {
		t.Fatalf("response status = %d, want %d", responses[0].StatusCode, http.StatusAccepted)
	}
	if len(roundTrips) != 1 {
		t.Fatalf("roundTrips = %d, want 1", len(roundTrips))
	}
	if roundTrips[0].Timings.Total <= 0 {
		t.Fatalf("expected positive total timing, got %v", roundTrips[0].Timings.Total)
	}
	if closed != 1 {
		t.Fatalf("closed = %d, want 1", closed)
	}
}

func TestHandler_ParsesAbsoluteRequestURIWhenURLIsNotReady(t *testing.T) {
	handler := New(Options{
		RoundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	})
	req := httptest.NewRequest(http.MethodGet, "http://example.com/path?q=1", nil)
	req.URL.Scheme = ""
	req.URL.Host = ""
	req.RequestURI = "http://example.com/path?q=1"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusBadRequest {
		t.Fatalf("expected absolute RequestURI fallback to work, got %d", rr.Code)
	}
}

func TestHandler_RejectsConnect(t *testing.T) {
	handler := New(Options{})
	req := httptest.NewRequest(http.MethodConnect, "http://example.com:443", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
