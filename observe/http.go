package observe

import (
	"net/http"
	"time"
)

// InlineBody carries transport-facing inline payload bytes. Callers can treat
// it as a sampled or full body depending on Truncated and TotalSize.
type InlineBody struct {
	Bytes           []byte
	TotalSize       int64
	Truncated       bool
	ContentType     string
	ContentEncoding string
}

// Clone returns a deep copy of the inline body.
func (b InlineBody) Clone() InlineBody {
	out := b
	if len(b.Bytes) > 0 {
		out.Bytes = append([]byte(nil), b.Bytes...)
	}
	return out
}

// HTTPRequest describes an observed HTTP request after transport-level target
// resolution and header mutation.
type HTTPRequest struct {
	SessionID string
	Method    string
	URL       string
	Header    http.Header
	Body      InlineBody
	At        time.Time
}

// Clone returns a deep copy of the request.
func (r HTTPRequest) Clone() HTTPRequest {
	out := r
	out.Header = CloneHeader(r.Header)
	out.Body = r.Body.Clone()
	return out
}

// HTTPResponse describes an observed HTTP response.
type HTTPResponse struct {
	SessionID  string
	StatusCode int
	Header     http.Header
	Body       InlineBody
	At         time.Time
}

// Clone returns a deep copy of the response.
func (r HTTPResponse) Clone() HTTPResponse {
	out := r
	out.Header = CloneHeader(r.Header)
	out.Body = r.Body.Clone()
	return out
}

// Timings capture coarse-grained transport timings.
type Timings struct {
	DNS     time.Duration
	Connect time.Duration
	TLS     time.Duration
	TTFB    time.Duration
	Total   time.Duration
}

// HTTPRoundTrip groups a request/response pair under a single session.
type HTTPRoundTrip struct {
	Session  Session
	Request  HTTPRequest
	Response HTTPResponse
	Timings  Timings
}

// Clone returns a deep copy of the round trip.
func (rt HTTPRoundTrip) Clone() HTTPRoundTrip {
	out := rt
	out.Request = rt.Request.Clone()
	out.Response = rt.Response.Clone()
	return out
}

// CloneHeader returns a deep copy of h.
func CloneHeader(h http.Header) http.Header {
	if h == nil {
		return nil
	}
	out := make(http.Header, len(h))
	for k, vv := range h {
		out[k] = append([]string(nil), vv...)
	}
	return out
}
