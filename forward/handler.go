package forward

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/777genius/proxykit/internal/httpcapture"
	"github.com/777genius/proxykit/observe"
	"github.com/777genius/proxykit/proxyhttp"
)

var (
	ErrAbsoluteURLRequired = errors.New("forward: absolute url required")
	ErrConnectNotSupported = errors.New("forward: CONNECT is not supported")
)

type ErrorWriter func(http.ResponseWriter, *http.Request, int, error)

type RequestMutator func(context.Context, *http.Request) error

type ResponseMutator func(context.Context, *http.Request, *http.Response) error

type Options struct {
	RoundTripper            http.RoundTripper
	GenerateSessionID       func() string
	ObserveRequest          func(*http.Request) bool
	WriteError              ErrorWriter
	MutateRequest           RequestMutator
	MutateResponse          ResponseMutator
	Hooks                   observe.Hooks
	ForwardedHeaders        proxyhttp.ForwardedHeaderConfig
	SampleRequestBodyBytes  int
	SampleResponseBodyBytes int
}

type Handler struct {
	roundTripper            http.RoundTripper
	generateSessionID       func() string
	observeRequest          func(*http.Request) bool
	writeError              ErrorWriter
	mutateRequest           RequestMutator
	mutateResponse          ResponseMutator
	hooks                   observe.Hooks
	forwardedHeaders        proxyhttp.ForwardedHeaderConfig
	sampleRequestBodyBytes  int
	sampleResponseBodyBytes int
}

func New(opts Options) *Handler {
	return &Handler{
		roundTripper:            defaultRoundTripper(opts.RoundTripper),
		generateSessionID:       defaultSessionIDFunc(opts.GenerateSessionID),
		observeRequest:          opts.ObserveRequest,
		writeError:              defaultErrorWriter(opts.WriteError),
		mutateRequest:           opts.MutateRequest,
		mutateResponse:          opts.MutateResponse,
		hooks:                   opts.Hooks,
		forwardedHeaders:        opts.ForwardedHeaders,
		sampleRequestBodyBytes:  defaultSampleLimit(opts.SampleRequestBodyBytes),
		sampleResponseBodyBytes: defaultSampleLimit(opts.SampleResponseBodyBytes),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := absoluteURL(r)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrConnectNotSupported) {
			status = http.StatusMethodNotAllowed
		}
		h.writeError(w, r, status, err)
		return
	}

	observeEnabled := true
	if h.observeRequest != nil {
		observeEnabled = h.observeRequest(r)
	}

	startedAt := time.Now().UTC()
	startedWall := time.Now()
	session := observe.Session{
		ID:         h.generateSessionID(),
		Kind:       observe.SessionKindHTTP,
		Target:     target.String(),
		ClientAddr: r.RemoteAddr,
		StartedAt:  startedAt,
	}
	hookCtx := context.WithoutCancel(r.Context())

	finish := func(closeErr error) {
		if observeEnabled && h.hooks.OnSessionClose != nil {
			h.hooks.OnSessionClose(hookCtx, observe.CloseInfo{
				SessionID: session.ID,
				Kind:      session.Kind,
				Target:    session.Target,
				Err:       closeErr,
				At:        time.Now().UTC(),
			})
		}
	}

	notifyError := func(stage string, err error) {
		if observeEnabled && h.hooks.OnError != nil && err != nil {
			h.hooks.OnError(hookCtx, observe.ErrorInfo{
				SessionID: session.ID,
				Kind:      session.Kind,
				Target:    session.Target,
				Stage:     stage,
				Err:       err,
				At:        time.Now().UTC(),
			})
		}
	}

	if observeEnabled && h.hooks.OnSessionOpen != nil {
		if err := h.hooks.OnSessionOpen(hookCtx, session); err != nil {
			h.writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	upstreamReq := r.Clone(r.Context())
	upstreamReq.URL = httpcapture.CloneURL(target)
	upstreamReq.RequestURI = ""
	upstreamReq.Host = target.Host
	proxyhttp.RemoveHopHeaders(upstreamReq.Header)
	proxyhttp.ApplyForwardedHeaders(upstreamReq.Header, h.forwardedHeaders)

	if h.mutateRequest != nil {
		if err := h.mutateRequest(hookCtx, upstreamReq); err != nil {
			notifyError("mutate_request", err)
			finish(err)
			h.writeError(w, r, http.StatusBadGateway, err)
			return
		}
	}

	var observedReq observe.HTTPRequest
	if observeEnabled && (h.hooks.OnHTTPRequest != nil || h.hooks.OnHTTPRoundTrip != nil) {
		body, restored, err := httpcapture.SampleReadCloser(
			upstreamReq.Body,
			h.sampleRequestBodyBytes,
			upstreamReq.Header.Get("Content-Type"),
			upstreamReq.Header.Get("Content-Encoding"),
		)
		if err != nil {
			notifyError("sample_request_body", err)
			finish(err)
			h.writeError(w, r, http.StatusBadGateway, err)
			return
		}
		upstreamReq.Body = restored
		observedReq = observe.HTTPRequest{
			SessionID: session.ID,
			Method:    upstreamReq.Method,
			URL:       upstreamReq.URL.String(),
			Header:    observe.CloneHeader(upstreamReq.Header),
			Body:      body,
			At:        time.Now().UTC(),
		}
		if h.hooks.OnHTTPRequest != nil {
			if err := h.hooks.OnHTTPRequest(hookCtx, observedReq.Clone()); err != nil {
				notifyError("observe_request", err)
			}
		}
	}

	resp, err := h.roundTripper.RoundTrip(upstreamReq)
	if err != nil {
		notifyError("round_trip", err)
		finish(err)
		h.writeError(w, r, http.StatusBadGateway, err)
		return
	}
	defer resp.Body.Close()

	ttfb := time.Since(startedWall)

	if h.mutateResponse != nil {
		if err := h.mutateResponse(hookCtx, upstreamReq, resp); err != nil {
			notifyError("mutate_response", err)
			finish(err)
			h.writeError(w, r, http.StatusBadGateway, err)
			return
		}
	}

	var observedResp observe.HTTPResponse
	if observeEnabled && (h.hooks.OnHTTPResponse != nil || h.hooks.OnHTTPRoundTrip != nil) {
		body, restored, err := httpcapture.SampleReadCloser(
			resp.Body,
			h.sampleResponseBodyBytes,
			resp.Header.Get("Content-Type"),
			resp.Header.Get("Content-Encoding"),
		)
		if err != nil {
			notifyError("sample_response_body", err)
			finish(err)
			h.writeError(w, r, http.StatusBadGateway, err)
			return
		}
		resp.Body = restored
		observedResp = observe.HTTPResponse{
			SessionID:  session.ID,
			StatusCode: resp.StatusCode,
			Header:     observe.CloneHeader(resp.Header),
			Body:       body,
			At:         time.Now().UTC(),
		}
		if h.hooks.OnHTTPResponse != nil {
			if err := h.hooks.OnHTTPResponse(hookCtx, observedResp.Clone()); err != nil {
				notifyError("observe_response", err)
			}
		}
	}

	httpcapture.CopyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		notifyError("copy_response_body", err)
		finish(err)
		return
	}

	if observeEnabled && h.hooks.OnHTTPRoundTrip != nil {
		rt := observe.HTTPRoundTrip{
			Session:  session,
			Request:  observedReq,
			Response: observedResp,
			Timings: observe.Timings{
				TTFB:  ttfb,
				Total: time.Since(startedWall),
			},
		}
		if err := h.hooks.OnHTTPRoundTrip(hookCtx, rt.Clone()); err != nil {
			notifyError("observe_round_trip", err)
		}
	}

	finish(nil)
}

func absoluteURL(r *http.Request) (*url.URL, error) {
	if r.Method == http.MethodConnect {
		return nil, ErrConnectNotSupported
	}
	if r.URL != nil && r.URL.Scheme != "" && r.URL.Host != "" {
		return httpcapture.CloneURL(r.URL), nil
	}
	if proxyhttp.IsAbsoluteURL(r.RequestURI) {
		u, err := url.Parse(r.RequestURI)
		if err == nil && u.Scheme != "" && u.Host != "" {
			return u, nil
		}
	}
	return nil, ErrAbsoluteURLRequired
}

func defaultRoundTripper(rt http.RoundTripper) http.RoundTripper {
	if rt != nil {
		return rt
	}
	return http.DefaultTransport
}

func defaultErrorWriter(fn ErrorWriter) ErrorWriter {
	if fn != nil {
		return fn
	}
	return func(w http.ResponseWriter, _ *http.Request, status int, err error) {
		http.Error(w, err.Error(), status)
	}
}

func defaultSampleLimit(limit int) int {
	if limit < 0 {
		return 0
	}
	if limit > 0 {
		return limit
	}
	return 64 << 10
}

var sessionCounter uint64

func defaultSessionIDFunc(fn func() string) func() string {
	if fn != nil {
		return fn
	}
	return func() string {
		n := atomic.AddUint64(&sessionCounter, 1)
		return strings.ToLower(strings.ReplaceAll(time.Now().UTC().Format("20060102T150405.000000000"), ".", "")) + "-" + strconvUint64(n)
	}
}

func strconvUint64(v uint64) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if v == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[v%36]
		v /= 36
	}
	return string(buf[i:])
}
