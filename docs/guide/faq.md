# FAQ

## Why is the repository smaller than a full proxy backend?

Because it is supposed to be.

`proxykit` is the reusable transport foundation extracted from a larger application backend. The product backend contains lots of application logic that should not live in a public Go library:

- REST routes
- monitor delivery protocols
- storage models
- replay and export flows
- UI-facing DTOs
- settings and admin APIs

If all of that were published as one library, the result would be harder to reuse and harder for the Go community to trust.

## Should I move my REST API into `proxykit`?

Usually no.

Expose your own routes in your application and keep `proxykit` as the transport and observation layer underneath.

## Does `proxykit` include a session store?

No.

It emits neutral observations. You decide how to store them, sample them, enrich them, or deliver them to clients.

## Does `proxykit` expose `/httpproxy` or `/wsproxy`?

No.

Those are route choices from an application. `proxykit` gives you handlers and resolvers, not opinionated route names.

## Why keep WebSocket and Socket.IO separate?

Because Socket.IO is a higher-level protocol on top of transport. The WebSocket transport should stay reusable even for projects that do not care about Socket.IO at all.

## Should capture and replay move into `proxykit`?

Only if they can be expressed as neutral reusable building blocks.

If the implementation assumes your product's session model, storage rules, or route protocol, it belongs in your application layer instead.

