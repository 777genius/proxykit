# Packages

## Package map

| Package | Role | Use it when |
| --- | --- | --- |
| `reverse` | Mounted reverse HTTP proxy handler | You expose a path like `/proxy` and forward to a resolved upstream target |
| `forward` | Absolute-URI HTTP forward proxy handler | You need classic HTTP proxying for clients that speak forward-proxy semantics |
| `connect` | Plain CONNECT tunneling handler | You need raw tunnel establishment for HTTPS or arbitrary TCP-over-CONNECT |
| `wsproxy` | WebSocket proxy handler | You need full-duplex WS proxying with frame hooks |
| `proxyruntime` | Listener lifecycle manager | Your app enables, disables, or restarts proxy listeners dynamically |
| `observe` | Shared hooks and neutral structs | You want a stable extension surface for capture, metrics, or app adapters |
| `cookies` | Reverse-proxy cookie isolation helpers | You need mounted reverse proxies without browser-cookie collisions |
| `proxyhttp` | Transport and header helpers | You want shared transport policy, hop-header stripping, and forwarded headers |
| `socketio` | Socket.IO text packet parser | You want to decode event-style Socket.IO frames from WS traffic |
| `mitm` | Optional CA and interception helpers | You need development-grade TLS interception building blocks |

## Typical compositions

### Reverse proxy app

Use:

- `reverse`
- `cookies`
- `observe`
- optionally `proxyhttp`

### Forward proxy app

Use:

- `forward`
- `connect`
- `proxyruntime`
- `observe`

### WebSocket debugger

Use:

- `wsproxy`
- `observe`
- optionally `socketio`

### Proxy control service

Use:

- `proxyruntime`
- your own config repository and REST handlers

## What not to import first

If you are new to the module, do **not** start by importing every package.

Start with the smallest set that matches your use case. This keeps your app architecture clearer and avoids dragging in policies you do not need yet.

