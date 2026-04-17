package reverse

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/777genius/proxykit/observe"
)

func TestQueryTargetResolver_MergesPathAndQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/proxy/api/users?id=2&lang=uk&_resetCapture=true", nil)
	req.URL.RawQuery = "_target=" + url.QueryEscape("https://example.com/base?lang=en&id=1") + "&id=2&lang=uk&_resetCapture=true"

	resolver := QueryTargetResolver{
		MountPath:  "/proxy",
		DropParams: []string{"_resetCapture"},
	}

	got, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.String() != "https://example.com/base/api/users?id=2&lang=uk" {
		t.Fatalf("unexpected upstream URL: %s", got.String())
	}
}

func TestHandler_ForwardsRequestAndObservesRoundTrip(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "example.test" {
			t.Fatalf("unexpected upstream host: %s", r.Host)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read upstream body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"echo":"` + string(body) + `"}`))
	}))
	defer upstream.Close()

	upstreamURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	upstreamURL.Host = "example.test"

	var mu sync.Mutex
	var opened int
	var requests []observe.HTTPRequest
	var responses []observe.HTTPResponse
	var roundTrips []observe.HTTPRoundTrip
	var closed int

	handler, err := New(Options{
		Resolver: ResolveTargetFunc(func(*http.Request) (*url.URL, error) {
			u := *upstreamURL
			return &u, nil
		}),
		RoundTripper: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = strings.TrimPrefix(upstream.URL, "http://")
			return http.DefaultTransport.RoundTrip(req)
		}),
		GenerateSessionID: func() string { return "sess-1" },
		MutateRequest: func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-Test", "yes")
			return nil
		},
		Hooks: observe.Hooks{
			OnSessionOpen: func(context.Context, observe.Session) error {
				mu.Lock()
				defer mu.Unlock()
				opened++
				return nil
			},
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
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/proxy/submit", strings.NewReader("hello"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
	if body := rr.Body.String(); body != `{"echo":"hello"}` {
		t.Fatalf("unexpected response body: %q", body)
	}

	mu.Lock()
	defer mu.Unlock()
	if opened != 1 {
		t.Fatalf("opened = %d, want 1", opened)
	}
	if closed != 1 {
		t.Fatalf("closed = %d, want 1", closed)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(requests))
	}
	if requests[0].Header.Get("X-Test") != "yes" {
		t.Fatalf("request mutation was not observed")
	}
	if string(requests[0].Body.Bytes) != "hello" {
		t.Fatalf("unexpected request sample: %q", string(requests[0].Body.Bytes))
	}
	if len(responses) != 1 {
		t.Fatalf("responses = %d, want 1", len(responses))
	}
	if responses[0].StatusCode != http.StatusCreated {
		t.Fatalf("response status = %d, want %d", responses[0].StatusCode, http.StatusCreated)
	}
	if len(roundTrips) != 1 {
		t.Fatalf("roundTrips = %d, want 1", len(roundTrips))
	}
	if roundTrips[0].Session.Kind != observe.SessionKindHTTP {
		t.Fatalf("session kind = %q, want %q", roundTrips[0].Session.Kind, observe.SessionKindHTTP)
	}
	if roundTrips[0].Timings.TTFB <= 0 {
		t.Fatalf("expected positive TTFB, got %v", roundTrips[0].Timings.TTFB)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
