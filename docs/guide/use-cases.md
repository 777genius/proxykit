# Use Cases

This is the fastest page for deciding whether `proxykit` fits your application.

## Mounted reverse proxy inside an existing app

You already own the main HTTP server and want a route such as `/proxy/...` to forward to a selected upstream.

Start with:

- `reverse`
- optionally `cookies`
- optionally `observe`

Why `proxykit` fits:

- mounted path handling is first-class
- target resolution can stay query-driven or become fully custom
- your app keeps its own routes, auth, storage, and UI payloads

## Classic HTTP forward proxy with CONNECT support

You need a developer proxy, test proxy, or internal tool that accepts absolute-URI HTTP requests and HTTPS tunnels.

Start with:

- `forward`
- `connect`
- optionally `proxyruntime`
- optionally `observe`

Why `proxykit` fits:

- `forward` and `connect` are separate on purpose
- you can keep HTTP and tunnel policies different
- listener lifecycle can stay dynamic if your app enables or disables proxy modes at runtime

## WebSocket or Socket.IO inspector

You need to proxy WebSocket traffic, observe frames, and maybe decode Socket.IO text frames without turning all WebSocket traffic into Socket.IO-specific abstractions.

Start with:

- `wsproxy`
- `observe`
- optionally `socketio`

Why `proxykit` fits:

- WS transport stays transport-focused
- Socket.IO stays optional
- frame observation and protocol derivation happen in your adapter layer

## Desktop or developer tool with runtime-managed listeners

Your app needs to start and stop forward and SOCKS listeners based on settings, UI state, or feature flags.

Start with:

- `proxyruntime`
- your own config repository
- your own admin or UI layer

Why `proxykit` fits:

- listener lifecycle is separate from your product settings API
- forward and SOCKS can be toggled without turning the module into a control plane

## Extracting a transport core from a larger product backend

You have an existing backend with sessions, storage, replay, or admin routes and want transport code that is smaller and reusable.

Start with:

- the transport package that matches your first seam
- `observe` for neutral event handoff

Why `proxykit` fits:

- it gives you a smaller, more versionable transport core
- your product backend can keep its current contracts while migrating internally

## When `proxykit` is the wrong choice

Choose another tool if you primarily want:

- a ready-made gateway platform
- service discovery and load balancing policies out of the box
- a full programmable proxy engine with a larger built-in rule system
- a product backend with REST admin APIs already included

For that tradeoff discussion, see [Comparisons](/guide/comparisons) and [Limits and Non-Goals](/guide/limits).
