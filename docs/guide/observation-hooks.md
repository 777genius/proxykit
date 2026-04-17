# Observation Hooks

`proxykit` is designed around a simple rule:

> transports own transport logic, your application owns product logic

The glue between those two layers is `observe`.

## The shared hook surface

```go
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
```

Not every transport uses every hook. For example:

- `reverse` and `forward` use HTTP hooks
- `wsproxy` uses session, frame, and protocol-event hooks
- `connect` uses session and protocol-event hooks

## Why this matters

Hooks let you add:

- session storage
- metrics
- request sampling
- replay indices
- live monitoring
- product-specific event translation

without putting any of that logic inside the transport packages themselves.

## Example: capture HTTP without owning transport

```go
hooks := observe.Hooks{
  OnSessionOpen: func(_ context.Context, s observe.Session) error {
    return store.OpenSession(s)
  },
  OnHTTPRoundTrip: func(_ context.Context, rt observe.HTTPRoundTrip) error {
    return store.SaveRoundTrip(rt)
  },
  OnSessionClose: func(_ context.Context, info observe.CloseInfo) {
    _ = store.CloseSession(info)
  },
}
```

## Example: derive Socket.IO events in an adapter

`proxykit` does not force Socket.IO into the WebSocket transport. You can keep it optional:

```go
hooks := wsproxy.Hooks{
  OnFrame: func(_ context.Context, frame observe.WSFrame) error {
    if frame.Type != observe.WSMessageText {
      return nil
    }
    namespace, event, argsJSON, ok := socketio.ParseEvent(string(frame.Payload))
    if !ok {
      return nil
    }
    return app.PublishProtocolEvent(frame.SessionID, namespace, event, []byte(argsJSON))
  },
}
```

## Keep hooks neutral

A good hook payload should describe what happened at the transport level.

A bad hook payload would already contain:

- UI-specific preview JSON
- pagination DTOs
- route names
- current monitor room subscriptions
- product-specific error taxonomies

Those belong in your adapter layer.
