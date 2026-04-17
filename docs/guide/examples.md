# Runnable Examples

`proxykit` includes runnable examples you can start with `go run`.

These are meant to reduce time-to-first-success for new users and give a clearer migration target for existing codebases.

## Reverse proxy app

Location:

- `./examples/reverse-basic`

Run:

```bash
go run ./examples/reverse-basic -target https://httpbin.org -listen :8080
```

Then send requests through the mounted route:

```bash
curl "http://127.0.0.1:8080/proxy/get?hello=world"
```

What it shows:

- mounted reverse proxy wiring
- default target configuration
- standard `http.ServeMux` integration

## Forward proxy with CONNECT support

Location:

- `./examples/forward-connect`

Run:

```bash
go run ./examples/forward-connect -listen :8080
```

Then point a client at it as an HTTP proxy.

What it shows:

- `forward` and `connect` composed on one listener
- clean separation between plain HTTP forwarding and CONNECT tunnels

## WebSocket inspector-style proxy

Location:

- `./examples/ws-observe`

Run:

```bash
go run ./examples/ws-observe -target wss://echo.websocket.events -listen :8080
```

Then connect a WebSocket client to:

```text
ws://127.0.0.1:8080/ws
```

What it shows:

- `wsproxy` mounted into a normal HTTP server
- frame logging through hooks
- optional default target resolution

## Why these examples matter

Examples in docs are useful, but runnable examples are better when you want:

- a fast local smoke test
- a migration starting point
- a reproducible bug report
- a concrete mental model of package composition

The CI workflow also builds these examples, so they stay compile-checked over time.
