# `observe`

`observe` is the neutral extension surface shared by the transport packages.

It models:

- sessions
- close and error info
- HTTP request and response observations
- grouped HTTP round trips
- WebSocket frames
- transport-adjacent protocol events

It deliberately avoids:

- route names
- app-specific DTOs
- preview JSON schemas
- storage ownership

## Core types

| Type | Purpose |
| --- | --- |
| `Session` | Identifies a transport exchange |
| `CloseInfo` | Terminal session outcome |
| `ErrorInfo` | Transport-stage error |
| `HTTPRequest` / `HTTPResponse` | Sampled HTTP observations |
| `HTTPRoundTrip` | Combined request/response timing group |
| `WSFrame` | Proxied WebSocket message |
| `ProtocolEvent` | Higher-level signal derived from a session |
| `Hooks` | Shared callback set |

## Why concrete structs matter

The public surface uses concrete structs and options, not wide exported interfaces.

That makes the API easier to evolve additively and keeps it easier to understand for downstream Go users.

