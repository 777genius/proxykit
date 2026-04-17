package observe

import "time"

// WSMessageType describes a proxied WebSocket message family.
type WSMessageType string

const (
	WSMessageText   WSMessageType = "text"
	WSMessageBinary WSMessageType = "binary"
	WSMessagePing   WSMessageType = "ping"
	WSMessagePong   WSMessageType = "pong"
	WSMessageClose  WSMessageType = "close"
	WSMessageOther  WSMessageType = "other"
)

// WSFrame describes a single proxied WebSocket message.
type WSFrame struct {
	SessionID string
	Direction Direction
	Type      WSMessageType
	Payload   []byte
	At        time.Time
}

// Clone returns a deep copy of the frame.
func (f WSFrame) Clone() WSFrame {
	out := f
	if len(f.Payload) > 0 {
		out.Payload = append([]byte(nil), f.Payload...)
	}
	return out
}

// ProtocolEvent is a transport-adjacent protocol signal derived from an
// underlying session, such as Socket.IO or similar higher-level events.
type ProtocolEvent struct {
	SessionID string
	Direction Direction
	Namespace string
	Name      string
	Payload   []byte
	At        time.Time
}

// Clone returns a deep copy of the protocol event.
func (e ProtocolEvent) Clone() ProtocolEvent {
	out := e
	if len(e.Payload) > 0 {
		out.Payload = append([]byte(nil), e.Payload...)
	}
	return out
}
