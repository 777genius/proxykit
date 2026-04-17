package reverse

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cherrypick-agency/proxykit/internal/httpcapture"
	"github.com/cherrypick-agency/proxykit/observe"
	"github.com/cherrypick-agency/proxykit/proxyhttp"
)

var (
	ErrMissingTarget = errors.New("missing target")
	ErrInvalidTarget = errors.New("invalid target")
)

type ResolveTargetFunc func(*http.Request) (*url.URL, error)

func (f ResolveTargetFunc) Resolve(r *http.Request) (*url.URL, error) {
	return f(r)
}

type TargetResolver interface {
	Resolve(*http.Request) (*url.URL, error)
}

// QueryTargetResolver resolves the upstream target from a query parameter or
// a configured default target, then merges path and query data from the
// incoming request.
type QueryTargetResolver struct {
	Param         string
	DefaultTarget string
	DropParams    []string
	MountPath     string
}

func (r QueryTargetResolver) Resolve(req *http.Request) (*url.URL, error) {
	param := r.Param
	if param == "" {
		param = "_target"
	}
	target := req.URL.Query().Get(param)
	if target == "" {
		target = r.DefaultTarget
	}
	if target == "" {
		return nil, ErrMissingTarget
	}

	upstream, err := url.Parse(target)
	if err != nil || (upstream.Scheme != "http" && upstream.Scheme != "https") {
		return nil, ErrInvalidTarget
	}

	out := *upstream
	suffix := req.URL.Path
	if r.MountPath != "" {
		suffix = strings.TrimPrefix(req.URL.Path, r.MountPath)
	}
	if suffix == "/" {
		suffix = ""
	}
	if suffix != "" && !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	if suffix != "" {
		out.Path = strings.TrimRight(out.Path, "/") + suffix
	}

	targetQ := out.Query()
	incomingQ := req.URL.Query()
	drop := map[string]struct{}{param: {}}
	for _, key := range r.DropParams {
		drop[key] = struct{}{}
	}
	for key := range drop {
		incomingQ.Del(key)
	}
	for key, vv := range incomingQ {
		targetQ.Del(key)
		for _, v := range vv {
			targetQ.Add(key, v)
		}
	}
	out.RawQuery = targetQ.Encode()
	if out.RawQuery == "" {
		out.ForceQuery = false
	}

	return &out, nil
}

type ErrorWriter func(http.ResponseWriter, *http.Request, int, error)

type RequestMutator func(context.Context, *http.Request) error

type ResponseMutator func(context.Context, *http.Request, *http.Response) error

type Options struct {
	Resolver                TargetResolver
	RoundTripper            http.RoundTripper
	GenerateSessionID       func() string
	ObserveTarget           func(*url.URL) bool
	WriteError              ErrorWriter
	MutateRequest           RequestMutator
	MutateResponse          ResponseMutator
	Hooks                   observe.Hooks
	PreserveHost            bool
	SampleRequestBodyBytes  int
	SampleResponseBodyBytes int
}

type Handler struct {
	resolver                TargetResolver
	roundTripper            http.RoundTripper
	generateSessionID       func() string
	observeTarget           func(*url.URL) bool
	writeError              ErrorWriter
	mutateRequest           RequestMutator
	mutateResponse          ResponseMutator
	hooks                   observe.Hooks
	preserveHost            bool
	sampleRequestBodyBytes  int
	sampleResponseBodyBytes int
}

func New(opts Options) (*Handler, error) {
	if opts.Resolver == nil {
		return nil, errors.New("reverse: resolver is required")
	}
	return &Handler{
		resolver:                opts.Resolver,
		roundTripper:            defaultRoundTripper(opts.RoundTripper),
		generateSessionID:       defaultSessionIDFunc(opts.GenerateSessionID),
		observeTarget:           opts.ObserveTarget,
		writeError:              defaultErrorWriter(opts.WriteError),
		mutateRequest:           opts.MutateRequest,
		mutateResponse:          opts.MutateResponse,
		hooks:                   opts.Hooks,
		preserveHost:            opts.PreserveHost,
		sampleRequestBodyBytes:  defaultSampleLimit(opts.SampleRequestBodyBytes),
		sampleResponseBodyBytes: defaultSampleLimit(opts.SampleResponseBodyBytes),
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := h.resolver.Resolve(r)
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrMissingTarget), errors.Is(err, ErrInvalidTarget):
			status = http.StatusBadRequest
		}
		h.writeError(w, r, status, err)
		return
	}

	observeEnabled := true
	if h.observeTarget != nil {
		observeEnabled = h.observeTarget(target)
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
	if !h.preserveHost {
		upstreamReq.Host = target.Host
	}
	proxyhttp.RemoveHopHeaders(upstreamReq.Header)

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
