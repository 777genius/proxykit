# Utilities

These packages support the main transports but stay focused.

## `proxyhttp`

Shared HTTP transport helpers:

- `NewTransport`
- `RemoveHopHeaders`
- `ApplyForwardedHeaders`
- `IsWebSocketRequest`
- `IsAbsoluteURL`

Use it when you want the same transport and header policy across multiple handlers without duplicating boilerplate.

## `socketio`

Parses event-style Socket.IO text packets.

Use it when you want to derive protocol events from proxied WebSocket frames without baking Socket.IO into your transport engine.

## `mitm`

Optional TLS interception helpers:

- CA loading and generation
- host-based interception policy
- leaf certificate issuance
- PEM encoding helpers

This package exists because MITM is useful for some tools, but it should stay optional and separate from plain reverse, forward, CONNECT, and WebSocket transport.

