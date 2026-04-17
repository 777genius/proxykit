package wsproxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cherrypick-agency/proxykit/observe"
	"github.com/gorilla/websocket"
)

var (
	ErrMissingTarget = errors.New("missing target")
	ErrInvalidTarget = errors.New("invalid target")
)

type Direction = observe.Direction

const (
	DirectionClientToUpstream = observe.DirectionClientToUpstream
	DirectionUpstreamToClient = observe.DirectionUpstreamToClient
)

type MessageType = observe.WSMessageType

const (
	MessageText   = observe.WSMessageText
	MessageBinary = observe.WSMessageBinary
	MessagePing   = observe.WSMessagePing
	MessagePong   = observe.WSMessagePong
	MessageClose  = observe.WSMessageClose
	MessageOther  = observe.WSMessageOther
)

type Session = observe.Session
type Frame = observe.WSFrame

type ForwardDecision struct {
	Drop  bool
	Delay time.Duration
}

type ErrorInfo = observe.ErrorInfo
type CloseInfo = observe.CloseInfo

type Hooks struct {
	OnSessionOpen      func(context.Context, Session) error
	OnSessionConnected func(context.Context, Session, *websocket.Conn, *websocket.Conn)
	OnFrame            func(context.Context, Frame) error
	OnError            func(context.Context, ErrorInfo)
	OnSessionClose     func(context.Context, CloseInfo)
}

type ResolveTargetFunc func(*http.Request) (*url.URL, error)

func (f ResolveTargetFunc) Resolve(r *http.Request) (*url.URL, error) {
	return f(r)
}

type TargetResolver interface {
	Resolve(*http.Request) (*url.URL, error)
}

type ErrorWriter func(http.ResponseWriter, *http.Request, int, error)

type QueryTargetResolver struct {
	Param           string
	DefaultTarget   string
	DropParams      []string
	NormalizeScheme bool
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
	u, err := url.Parse(target)
	if err != nil {
		return nil, ErrInvalidTarget
	}
	if r.NormalizeScheme {
		switch u.Scheme {
		case "ws", "wss":
		case "http":
			u.Scheme = "ws"
		case "https":
			u.Scheme = "wss"
		default:
			return nil, ErrInvalidTarget
		}
	}

	targetQ := u.Query()
	incomingQ := req.URL.Query()
	drop := map[string]struct{}{param: {}}
	for _, key := range r.DropParams {
		drop[key] = struct{}{}
	}
	for key := range drop {
		incomingQ.Del(key)
	}
	for k, vv := range incomingQ {
		targetQ.Del(k)
		for _, v := range vv {
			targetQ.Add(k, v)
		}
	}
	u.RawQuery = targetQ.Encode()
	return u, nil
}

type Options struct {
	Resolver               TargetResolver
	ObserveTarget          func(*url.URL) bool
	GenerateSessionID      func() string
	HandshakeTimeout       time.Duration
	InsecureTLS            bool
	AllowPlaintextFallback bool
	BeforeForward          func(Direction, MessageType, int) ForwardDecision
	ForwardHeaders         []string
	SynthesizeOrigin       bool
	WriteError             ErrorWriter
	Hooks                  Hooks
	Upgrader               websocket.Upgrader
}

type Handler struct {
	resolver               TargetResolver
	observeTarget          func(*url.URL) bool
	generateSessionID      func() string
	handshakeTimeout       time.Duration
	insecureTLS            bool
	allowPlaintextFallback bool
	beforeForward          func(Direction, MessageType, int) ForwardDecision
	forwardHeaders         []string
	synthesizeOrigin       bool
	writeError             ErrorWriter
	hooks                  Hooks
	upgrader               websocket.Upgrader
}

func New(opts Options) (*Handler, error) {
	if opts.Resolver == nil {
		return nil, errors.New("wsproxy: resolver is required")
	}
	upgrader := opts.Upgrader
	if upgrader.CheckOrigin == nil {
		upgrader.CheckOrigin = func(*http.Request) bool { return true }
	}
	return &Handler{
		resolver:               opts.Resolver,
		observeTarget:          opts.ObserveTarget,
		generateSessionID:      defaultSessionIDFunc(opts.GenerateSessionID),
		handshakeTimeout:       defaultDuration(opts.HandshakeTimeout, 10*time.Second),
		insecureTLS:            opts.InsecureTLS,
		allowPlaintextFallback: opts.AllowPlaintextFallback,
		beforeForward:          opts.BeforeForward,
		forwardHeaders:         defaultHeaders(opts.ForwardHeaders),
		synthesizeOrigin:       opts.SynthesizeOrigin,
		writeError:             defaultErrorWriter(opts.WriteError),
		hooks:                  opts.Hooks,
		upgrader:               upgrader,
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

	baseCtx := context.WithoutCancel(r.Context())
	session := Session{
		ID:         h.generateSessionID(),
		Kind:       observe.SessionKindWebSocket,
		Target:     target.String(),
		ClientAddr: r.RemoteAddr,
		StartedAt:  time.Now().UTC(),
	}

	if observeEnabled && h.hooks.OnSessionOpen != nil {
		if err := h.hooks.OnSessionOpen(baseCtx, session); err != nil {
			h.writeError(w, r, http.StatusInternalServerError, err)
			return
		}
	}

	upgrader := h.upgrader
	if len(upgrader.Subprotocols) == 0 {
		upgrader.Subprotocols = []string{r.Header.Get("Sec-WebSocket-Protocol")}
	}

	var clientConn *websocket.Conn
	var upstreamConn *websocket.Conn
	var closeOnce sync.Once

	finish := func(err error) {
		closeOnce.Do(func() {
			if clientConn != nil {
				_ = clientConn.Close()
			}
			if upstreamConn != nil {
				_ = upstreamConn.Close()
			}
			if observeEnabled && h.hooks.OnSessionClose != nil {
				h.hooks.OnSessionClose(baseCtx, CloseInfo{
					SessionID: session.ID,
					Kind:      session.Kind,
					Target:    session.Target,
					Err:       err,
					At:        time.Now().UTC(),
				})
			}
		})
	}

	notifyError := func(stage string, err error) {
		if observeEnabled && h.hooks.OnError != nil && err != nil {
			h.hooks.OnError(baseCtx, ErrorInfo{
				SessionID: session.ID,
				Kind:      session.Kind,
				Target:    session.Target,
				Stage:     stage,
				Err:       err,
				At:        time.Now().UTC(),
			})
		}
	}

	clientConn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		notifyError("upgrade_client", err)
		finish(err)
		return
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: h.handshakeTimeout,
		NetDialContext:   (&net.Dialer{Timeout: h.handshakeTimeout}).DialContext,
	}
	if h.insecureTLS && target.Scheme == "wss" {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	headers := h.buildUpstreamHeaders(r, target)
	upstreamConn, _, err = h.dialUpstream(&dialer, target, headers)
	if err != nil {
		notifyError("dial_upstream", err)
		_ = clientConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, err.Error()), time.Now().Add(2*time.Second))
		finish(err)
		return
	}

	if observeEnabled && h.hooks.OnSessionConnected != nil {
		h.hooks.OnSessionConnected(baseCtx, session, clientConn, upstreamConn)
	}

	go h.pipe(baseCtx, session, clientConn, upstreamConn, DirectionClientToUpstream, observeEnabled, notifyError, finish)
	h.pipe(baseCtx, session, upstreamConn, clientConn, DirectionUpstreamToClient, observeEnabled, notifyError, finish)
}

func (h *Handler) dialUpstream(dialer *websocket.Dialer, target *url.URL, headers http.Header) (*websocket.Conn, *http.Response, error) {
	conn, resp, err := dialer.Dial(target.String(), headers)
	if err != nil && h.allowPlaintextFallback && target.Scheme == "wss" && strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
		fallback := *target
		fallback.Scheme = "ws"
		fallbackDialer := *dialer
		fallbackDialer.TLSClientConfig = nil
		return fallbackDialer.Dial(fallback.String(), headers)
	}
	return conn, resp, err
}

func (h *Handler) buildUpstreamHeaders(r *http.Request, target *url.URL) http.Header {
	headers := http.Header{}
	for _, key := range h.forwardHeaders {
		if value := r.Header.Get(key); value != "" {
			headers.Set(key, value)
		}
	}
	if sp := r.Header.Get("Sec-WebSocket-Protocol"); sp != "" {
		headers.Set("Sec-WebSocket-Protocol", sp)
	}
	if h.synthesizeOrigin && headers.Get("Origin") == "" {
		origin := "http://" + target.Host
		if target.Scheme == "wss" {
			origin = "https://" + target.Host
		}
		headers.Set("Origin", origin)
	}
	return headers
}

func (h *Handler) pipe(
	ctx context.Context,
	session Session,
	src *websocket.Conn,
	dst *websocket.Conn,
	direction Direction,
	observe bool,
	notifyError func(string, error),
	finish func(error),
) {
	for {
		mt, data, err := src.ReadMessage()
		if err != nil {
			if !isExpectedClose(err) {
				notifyError("read_"+string(direction), err)
			}
			finish(err)
			return
		}
		if h.beforeForward != nil {
			decision := h.beforeForward(direction, messageTypeFromOpcode(mt), len(data))
			if decision.Delay > 0 {
				time.Sleep(decision.Delay)
			}
			if decision.Drop {
				continue
			}
		}
		_ = dst.SetWriteDeadline(time.Now().Add(15 * time.Second))
		if err := dst.WriteMessage(mt, data); err != nil {
			notifyError("write_"+string(direction), err)
			finish(err)
			return
		}
		if observe && h.hooks.OnFrame != nil {
			frame := Frame{
				SessionID: session.ID,
				Direction: direction,
				Type:      messageTypeFromOpcode(mt),
				Payload:   append([]byte(nil), data...),
				At:        time.Now().UTC(),
			}
			if err := h.hooks.OnFrame(ctx, frame); err != nil {
				notifyError("observe_frame", err)
			}
		}
	}
}

func defaultErrorWriter(fn ErrorWriter) ErrorWriter {
	if fn != nil {
		return fn
	}
	return func(w http.ResponseWriter, _ *http.Request, status int, err error) {
		http.Error(w, err.Error(), status)
	}
}

func defaultHeaders(in []string) []string {
	if len(in) > 0 {
		return in
	}
	return []string{"Authorization", "Cookie", "Origin", "User-Agent", "Referer"}
}

func defaultDuration(v, fallback time.Duration) time.Duration {
	if v > 0 {
		return v
	}
	return fallback
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

func isExpectedClose(err error) bool {
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return true
	}
	return errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "closed")
}

func messageTypeFromOpcode(mt int) MessageType {
	switch mt {
	case websocket.TextMessage:
		return MessageText
	case websocket.BinaryMessage:
		return MessageBinary
	case websocket.PingMessage:
		return MessagePing
	case websocket.PongMessage:
		return MessagePong
	case websocket.CloseMessage:
		return MessageClose
	default:
		return MessageOther
	}
}
