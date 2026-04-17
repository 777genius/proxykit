package observe

import "context"

// Hooks collect optional observation callbacks shared across proxykit
// transports. Packages may use only the subset relevant to their transport.
type Hooks struct {
	OnSessionOpen   func(context.Context, Session) error
	OnSessionClose  func(context.Context, CloseInfo)
	OnError         func(context.Context, ErrorInfo)
	OnHTTPRequest   func(context.Context, HTTPRequest) error
	OnHTTPResponse  func(context.Context, HTTPResponse) error
	OnHTTPRoundTrip func(context.Context, HTTPRoundTrip) error
	OnWSFrame       func(context.Context, WSFrame) error
	OnProtocolEvent func(context.Context, ProtocolEvent) error
}
