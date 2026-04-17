package observe

import "time"

// SessionKind identifies the transport family that produced an observation.
type SessionKind string

const (
	SessionKindUnknown   SessionKind = ""
	SessionKindHTTP      SessionKind = "http"
	SessionKindWebSocket SessionKind = "ws"
	SessionKindConnect   SessionKind = "connect"
	SessionKindTCP       SessionKind = "tcp"
)

// Direction describes traffic flow relative to the proxy.
type Direction string

const (
	DirectionClientToUpstream Direction = "client->upstream"
	DirectionUpstreamToClient Direction = "upstream->client"
)

// Session identifies a single observed transport exchange.
type Session struct {
	ID         string
	Kind       SessionKind
	Target     string
	ClientAddr string
	StartedAt  time.Time
}

// ErrorInfo describes an observation error without forcing application-level
// error classification.
type ErrorInfo struct {
	SessionID string
	Kind      SessionKind
	Target    string
	Stage     string
	Err       error
	At        time.Time
}

// CloseInfo describes a terminal session outcome.
type CloseInfo struct {
	SessionID string
	Kind      SessionKind
	Target    string
	Err       error
	At        time.Time
}
