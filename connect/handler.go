package connect

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cherrypick-agency/proxykit/observe"
)

var (
	ErrConnectRequired     = errors.New("connect: CONNECT method is required")
	ErrEmptyTarget         = errors.New("connect: target is required")
	ErrHijackNotSupported  = errors.New("connect: hijacking not supported")
	defaultConnectResponse = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")
	defaultDialError       = []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n")
	defaultOpenError       = []byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")
)

type DialContextFunc func(context.Context, string, string) (net.Conn, error)

type ErrorWriter func(http.ResponseWriter, *http.Request, int, error)

type Options struct {
	DialContext       DialContextFunc
	GenerateSessionID func() string
	ObserveRequest    func(*http.Request) bool
	WriteError        ErrorWriter
	Hooks             observe.Hooks
	Network           string
}

type Handler struct {
	dialContext       DialContextFunc
	generateSessionID func() string
	observeRequest    func(*http.Request) bool
	writeError        ErrorWriter
	hooks             observe.Hooks
	network           string
}

func New(opts Options) *Handler {
	return &Handler{
		dialContext:       defaultDialContext(opts.DialContext),
		generateSessionID: defaultSessionIDFunc(opts.GenerateSessionID),
		observeRequest:    opts.ObserveRequest,
		writeError:        defaultErrorWriter(opts.WriteError),
		hooks:             opts.Hooks,
		network:           defaultNetwork(opts.Network),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	target, err := connectTarget(r)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrConnectRequired) {
			status = http.StatusMethodNotAllowed
		}
		h.writeError(w, r, status, err)
		return
	}

	observeEnabled := true
	if h.observeRequest != nil {
		observeEnabled = h.observeRequest(r)
	}

	session := observe.Session{
		ID:         h.generateSessionID(),
		Kind:       observe.SessionKindConnect,
		Target:     target,
		ClientAddr: r.RemoteAddr,
		StartedAt:  time.Now().UTC(),
	}
	hookCtx := context.WithoutCancel(r.Context())
	opened := false

	finish := func(closeErr error) {
		if opened && observeEnabled && h.hooks.OnSessionClose != nil {
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

	hj, ok := w.(http.Hijacker)
	if !ok {
		notifyError("hijack_not_supported", ErrHijackNotSupported)
		finish(ErrHijackNotSupported)
		h.writeError(w, r, http.StatusInternalServerError, ErrHijackNotSupported)
		return
	}

	clientConn, bufrw, err := hj.Hijack()
	if err != nil {
		notifyError("hijack", err)
		finish(err)
		return
	}

	upstreamConn, err := h.dialContext(r.Context(), h.network, target)
	if err != nil {
		notifyError("dial", err)
		writeRawResponse(bufrw, defaultDialError)
		_ = clientConn.Close()
		finish(err)
		return
	}

	if observeEnabled && h.hooks.OnSessionOpen != nil {
		if err := h.hooks.OnSessionOpen(hookCtx, session); err != nil {
			notifyError("session_open", err)
			writeRawResponse(bufrw, defaultOpenError)
			_ = upstreamConn.Close()
			_ = clientConn.Close()
			return
		}
	}
	opened = true

	writeRawResponse(bufrw, defaultConnectResponse)

	if observeEnabled && h.hooks.OnProtocolEvent != nil {
		if err := h.hooks.OnProtocolEvent(hookCtx, observe.ProtocolEvent{
			SessionID: session.ID,
			Namespace: "/_sys",
			Name:      "tunnel_established",
			At:        time.Now().UTC(),
		}); err != nil {
			notifyError("observe_protocol_event", err)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		_, err := io.Copy(upstreamConn, clientConn)
		_ = upstreamConn.Close()
		errCh <- normalizeCopyError(err)
	}()

	_, downstreamErr := io.Copy(clientConn, upstreamConn)
	_ = clientConn.Close()
	upstreamErr := <-errCh
	downstreamErr = normalizeCopyError(downstreamErr)

	if upstreamErr != nil {
		notifyError("copy_client_to_upstream", upstreamErr)
	}
	if downstreamErr != nil {
		notifyError("copy_upstream_to_client", downstreamErr)
	}
	finish(firstError(upstreamErr, downstreamErr))
}

func connectTarget(r *http.Request) (string, error) {
	if r.Method != http.MethodConnect {
		return "", ErrConnectRequired
	}
	target := strings.TrimSpace(r.Host)
	if target == "" && r.URL != nil {
		target = strings.TrimSpace(r.URL.Host)
	}
	if target == "" {
		return "", ErrEmptyTarget
	}
	return target, nil
}

func defaultDialContext(fn DialContextFunc) DialContextFunc {
	if fn != nil {
		return fn
	}
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	return dialer.DialContext
}

func defaultErrorWriter(fn ErrorWriter) ErrorWriter {
	if fn != nil {
		return fn
	}
	return func(w http.ResponseWriter, _ *http.Request, status int, err error) {
		http.Error(w, err.Error(), status)
	}
}

func defaultNetwork(network string) string {
	if network == "" {
		return "tcp"
	}
	return network
}

func normalizeCopyError(err error) error {
	switch {
	case err == nil,
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrClosedPipe),
		errors.Is(err, net.ErrClosed),
		strings.Contains(strings.ToLower(err.Error()), "closed network connection"):
		return nil
	default:
		return err
	}
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
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

func writeRawResponse(bufrw *bufio.ReadWriter, payload []byte) {
	if bufrw == nil {
		return
	}
	_, _ = bufrw.Write(payload)
	_ = bufrw.Flush()
}
