# Architecture

## Design goal

`proxykit` is designed as a **transport foundation**, not a product backend.

That distinction matters. A proxy product usually contains at least four layers:

| Layer | Responsibility | Belongs in `proxykit` |
| --- | --- | --- |
| Transport | Reverse proxy, forward proxy, CONNECT, WebSocket piping, TLS helpers | Yes |
| Observation | Neutral sessions, HTTP round trips, WS frames, protocol events | Yes |
| Application policy | Capture mode, storage, monitor fanout, throttling, replay, admin endpoints | No |
| Product UX | REST DTOs, UI payloads, rooms, tags, sessions explorer, settings pages | No |

## The boundary

Keep `proxykit` responsible for:

- target resolution
- hop-by-hop header handling
- request and response forwarding
- raw tunnel lifecycle
- WebSocket frame forwarding
- cookie rewrite helpers
- listener lifecycle management
- neutral hooks and event structs

Keep your application responsible for:

- route naming and mounting conventions
- settings APIs
- persistence and spool files
- admin authentication
- current-capture visibility rules
- monitor broadcast protocol
- UI-specific error projection

## Why the repository is intentionally smaller than an app backend

The `network-debugger` backend that originally powered this extraction is much larger than `proxykit`.

That is expected.

The product backend includes:

- session list and detail APIs
- realtime delivery protocols
- HAR import and export
- scripting orchestration
- process integration
- tags, settings, and admin endpoints
- frontend compatibility contracts

Those are application features, not reusable transport primitives.

If all of that were moved into the public module, the result would be harder to reuse, harder to version, and less idiomatic for Go consumers.

## Package shape

`proxykit` is split by responsibility:

- **Transport engines**: `reverse`, `forward`, `connect`, `wsproxy`
- **Lifecycle/runtime**: `proxyruntime`
- **Observation contracts**: `observe`
- **Supporting utilities**: `cookies`, `proxyhttp`, `socketio`, `mitm`

This keeps composition explicit. A project can use only `reverse` and `cookies`, or `forward` plus `connect`, without importing unrelated product concerns.

## A good adapter layer

A typical adapter layer above `proxykit` should do things like:

- generate app-specific session IDs
- decide which targets should be observed
- persist request and response samples
- translate neutral observations into product-specific events
- expose REST or WebSocket APIs

That adapter layer can be large. `proxykit` should stay smaller.

