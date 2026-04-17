package observe

import (
	"net/http"
	"testing"
	"time"
)

func TestCloneHeader(t *testing.T) {
	src := http.Header{
		"X-Test": []string{"a", "b"},
	}
	got := CloneHeader(src)
	got.Add("X-Test", "c")

	if len(src["X-Test"]) != 2 {
		t.Fatalf("source header mutated: %#v", src["X-Test"])
	}
	if len(got["X-Test"]) != 3 {
		t.Fatalf("unexpected clone length: %#v", got["X-Test"])
	}
}

func TestHTTPRoundTripCloneDeepCopiesBodiesAndHeaders(t *testing.T) {
	rt := HTTPRoundTrip{
		Session: Session{
			ID:        "s1",
			Kind:      SessionKindHTTP,
			Target:    "https://example.com",
			StartedAt: time.Unix(0, 0).UTC(),
		},
		Request: HTTPRequest{
			SessionID: "s1",
			Method:    http.MethodPost,
			URL:       "https://example.com/in",
			Header:    http.Header{"Content-Type": []string{"application/json"}},
			Body: InlineBody{
				Bytes:       []byte(`{"hello":"world"}`),
				TotalSize:   17,
				Truncated:   false,
				ContentType: "application/json",
			},
		},
		Response: HTTPResponse{
			SessionID:  "s1",
			StatusCode: http.StatusCreated,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: InlineBody{
				Bytes:       []byte(`{"ok":true}`),
				TotalSize:   11,
				Truncated:   false,
				ContentType: "application/json",
			},
		},
	}

	cloned := rt.Clone()
	cloned.Request.Header.Set("Content-Type", "text/plain")
	cloned.Request.Body.Bytes[0] = '!'
	cloned.Response.Body.Bytes[0] = '!'

	if rt.Request.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("request header was mutated: %q", rt.Request.Header.Get("Content-Type"))
	}
	if string(rt.Request.Body.Bytes) != `{"hello":"world"}` {
		t.Fatalf("request body was mutated: %q", string(rt.Request.Body.Bytes))
	}
	if string(rt.Response.Body.Bytes) != `{"ok":true}` {
		t.Fatalf("response body was mutated: %q", string(rt.Response.Body.Bytes))
	}
}

func TestWSFrameCloneDeepCopiesPayload(t *testing.T) {
	frame := WSFrame{
		SessionID: "s1",
		Direction: DirectionClientToUpstream,
		Type:      WSMessageText,
		Payload:   []byte("hello"),
		At:        time.Unix(0, 0).UTC(),
	}

	cloned := frame.Clone()
	cloned.Payload[0] = 'H'

	if string(frame.Payload) != "hello" {
		t.Fatalf("source payload mutated: %q", string(frame.Payload))
	}
}
