# proxykit

`proxykit` is a standalone Go proxy foundation extracted from `network-debugger`.

Current public packages:

- `github.com/777genius/proxykit/proxyhttp`
  - shared HTTP proxy transport helpers
  - hop-by-hop header stripping
  - WebSocket upgrade detection
  - absolute-URI detection
- `github.com/777genius/proxykit/proxyruntime`
  - forward proxy listener lifecycle
  - SOCKS5 listener lifecycle
  - runtime apply/restart semantics
  - human-readable port conflict diagnostics
- `github.com/777genius/proxykit/socketio`
  - Socket.IO packet parsing for event-style text frames
- `github.com/777genius/proxykit/observe`
  - transport-neutral observation contracts
  - shared session, HTTP, and WebSocket event types
  - hook-friendly concrete structs without storage or route coupling
- `github.com/777genius/proxykit/reverse`
  - reusable reverse HTTP proxy handler
  - query-based target resolver helper
  - redirect rewrite helper for mounted proxy paths
  - request/response mutation plus observation hooks
  - now powers the app reverse HTTP transport through an adapter layer
- `github.com/777genius/proxykit/forward`
  - reusable HTTP forward proxy handler for absolute-URI requests
  - standard forwarding header policy
  - request/response mutation plus observation hooks
  - keeps CONNECT and WS upgrades outside the core HTTP handler
- `github.com/777genius/proxykit/connect`
  - reusable HTTP CONNECT tunneling handler
  - connection hijack, upstream dial, and raw tunnel lifecycle
  - protocol event hook for tunnel establishment
  - keeps MITM outside the package
- `github.com/777genius/proxykit/mitm`
  - optional development CA authority loading and generation
  - host-based interception policy
  - leaf certificate issuance with cache-aware reuse
  - PEM encoding helper for CA export flows
- `github.com/777genius/proxykit/cookies`
  - reverse-proxy cookie rewrite helpers
  - namespace isolation for browser cookie storage
  - Set-Cookie roundtrip parsing that preserves unknown attributes
  - outbound Cookie filtering for isolated upstream forwarding
- `github.com/777genius/proxykit/wsproxy`
  - reusable WebSocket proxy handler
  - hook-driven session/frame observation
  - query-based target resolver helper
  - optional plaintext fallback for TLS-mismatch targets

Design rules:

- no Flutter or UI-specific contracts in public packages
- packages stay small and composable
- app-specific monitoring, persistence, DTOs, and REST endpoints remain outside the module
- new protocol engines should be hook-driven, not storage-driven

Examples and release hardening:

- each public transport package now has a compile-able example test that shows the intended mounting style
- tests inside `proxykit` stay self-contained and do not depend on `network-debugger` internals
- `observe` remains transport-neutral and does not encode the current app delivery protocol
- `proxyruntime` examples demonstrate listener lifecycle without app config or settings DTOs
- the module now lives in its own dedicated repository workspace and is no longer kept as a nested incubation module

This repository intentionally excludes:

- app-specific REST routes like `/httpproxy`, `/wsproxy`, `/_api/v1/proxy/config`
- admin auth, loopback security policy, and settings DTOs
- current monitor room protocol and frontend-specific realtime event names
- storage, spool, capture visibility, and session query projection rules from `network-debugger`

The intended usage model is:

1. mount one or more transport handlers from `reverse`, `forward`, `connect`, or `wsproxy`
2. attach observation and mutation hooks
3. keep persistence, monitoring delivery, and product-specific REST/API concerns in your own adapter layer
