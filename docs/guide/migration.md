# Migration

This page is about adopting `proxykit` without destabilizing an existing application.

## General rule

Do not migrate your whole backend shape at once.

Migrate one transport seam at a time:

1. keep your current routes and external contracts
2. replace the transport implementation behind one route or listener
3. move capture, metrics, or storage behind hooks
4. only then remove old transport-specific glue

## From `net/http/httputil.ReverseProxy`

Typical reason to switch:

- you need mounted path plus `_target` resolution
- you want observation hooks
- you want cleaner separation between transport and adapter logic

Suggested path:

1. Replace direct `httputil.ReverseProxy` usage with `reverse.New(...)`.
2. Start with `reverse.QueryTargetResolver` if your route is query-driven.
3. Move any persistence or replay logic into `observe.Hooks`.
4. Add `cookies` only if browser cookie collisions become a real issue.

## From `goproxy`

Typical reason to switch:

- you want smaller packages instead of a more monolithic programmable proxy object
- you need reverse HTTP and WebSocket as first-class parts of the same library family

Suggested mapping:

| Existing concern | `proxykit` target |
| --- | --- |
| Forward HTTP proxy | `forward` |
| CONNECT tunnels | `connect` |
| Reverse HTTP route | `reverse` |
| WebSocket proxying | `wsproxy` |
| TLS interception helpers | `mitm` |
| Neutral capture or metrics handoff | `observe` |

Migrate the HTTP forward path first, then CONNECT, then optional MITM.

## From a product-specific proxy backend

Typical reason to switch:

- the backend transport code is too coupled to routes, DTOs, or UI protocol
- you want a reusable Go foundation that can outlive one product

Suggested path:

1. Leave current frontend and admin contracts unchanged.
2. Replace the internal transport engine under one route or listener.
3. Keep projection, storage, monitor fanout, and admin policy in your app layer.
4. Publish only the transport-safe packages.

This usually means the extracted public repo stays much smaller than the original backend. That is a success condition, not a failure.

## Safe adoption order

If you are starting from scratch inside an existing codebase, this order keeps risk low:

1. `reverse` or `forward`
2. `observe`
3. `connect`
4. `wsproxy`
5. `cookies`
6. `proxyruntime`
7. optional `mitm`

## Red flags during migration

Stop and re-scope if your public package starts to grow any of these:

- route names like `/proxy` or `/_api/v1/*`
- admin token checks
- session list DTOs
- UI preview payloads
- product-specific error categories
- monitor rooms or app event names

Those belong in the application layer above `proxykit`.
