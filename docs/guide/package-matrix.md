# Package Matrix

Use this table to choose the smallest useful set of packages for your application.

## Decision matrix

| Need | Start with | Add next | Notes |
| --- | --- | --- | --- |
| Mounted reverse proxy route | `reverse` | `cookies`, `observe` | Best when your app owns the outer router and mount path |
| Absolute-URI HTTP forward proxy | `forward` | `connect`, `proxyruntime`, `observe` | `forward` intentionally does not own CONNECT |
| HTTPS tunnels over CONNECT | `connect` | `observe`, optionally `mitm` | Keep plain tunneling separate from interception policy |
| WebSocket proxying | `wsproxy` | `observe`, `socketio` | Socket.IO stays optional |
| Runtime start and stop of listeners | `proxyruntime` | `forward`, `connect` | Good for desktop tools and admin-controlled services |
| Cookie isolation for mounted reverse proxying | `cookies` | `reverse` | Useful when browser cookies would otherwise collide at the proxy boundary |
| Neutral event hooks for capture or metrics | `observe` | any transport package | Keep storage and DTOs outside the module |
| Socket.IO event parsing | `socketio` | `wsproxy` | Only for text-frame protocol derivation |
| Development-grade TLS interception helpers | `mitm` | `connect`, your own policy layer | Intentionally optional and not a full CA control plane |

## Common bundles

### Minimal reverse proxy

- `reverse`

### Reverse proxy plus browser-safe cookies

- `reverse`
- `cookies`

### Developer proxy

- `forward`
- `connect`
- `proxyruntime`
- `observe`

### WebSocket inspector

- `wsproxy`
- `observe`
- `socketio`

### Runtime-controlled desktop proxy

- `forward`
- `connect`
- `proxyruntime`
- `observe`
- optionally `wsproxy`

## What not to do

Do not import every package by default just because they live in the same module.

If your app only needs mounted reverse proxying, start with `reverse` and stop there. Add `cookies`, `observe`, or `mitm` only when the product actually needs those responsibilities.
