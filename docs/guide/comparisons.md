# Comparisons

This page explains where `proxykit` sits relative to well-known Go proxy projects.

## Short version

`proxykit` is strongest when you want:

- an embeddable library, not a full platform
- one foundation for reverse HTTP, forward HTTP, CONNECT, and WebSocket
- explicit boundaries between transport and product-specific API layers

It is **not** trying to be:

- a gateway platform like Traefik or Skipper
- a full programmable policy engine with its own domain-specific routing model

## Against `goproxy`

Official repo: [elazarl/goproxy](https://github.com/elazarl/goproxy)

| Topic | `goproxy` | `proxykit` |
| --- | --- | --- |
| Main shape | Mature programmable HTTP/HTTPS proxy | Composable transport foundation |
| Forward proxy | Strong | Strong |
| CONNECT | Strong | Strong |
| MITM | Strong | Optional helper package |
| Reverse proxy story | Not the main focus | First-class |
| WebSocket story | Available, but not the main identity | First-class package |
| Architecture style | More monolithic proxy object | Smaller packages with explicit seams |

Choose `proxykit` over `goproxy` when you care more about clean package boundaries and multi-transport composition than about inheriting a long-standing all-in-one proxy abstraction.

## Against `oxy`

Official repo: [vulcand/oxy](https://github.com/vulcand/oxy)

| Topic | `oxy` | `proxykit` |
| --- | --- | --- |
| Main shape | HTTP reverse-proxy middleware toolkit | Multi-transport proxy foundation |
| Reverse proxy | Excellent | Strong |
| Forward proxy | Not the focus | First-class |
| CONNECT | Not the focus | First-class |
| WebSocket | Limited compared to `proxykit` scope | First-class |
| Listener runtime | External concern | Included via `proxyruntime` |

Choose `proxykit` over `oxy` when your application is not only a reverse proxy and you want forward, CONNECT, and WebSocket handled under the same library family.

## Against `martian`

Official repo: [google/martian](https://github.com/google/martian)

| Topic | `martian` | `proxykit` |
| --- | --- | --- |
| Main shape | Powerful modifier-based HTTP/S proxy library | Practical transport foundation |
| HTTP mutation | Very strong | Strong enough for adapter-driven apps |
| Reverse proxy | Not the main shape | First-class |
| WebSocket | Not the main focus | First-class |
| Mental model | Rich modifier graph | Small packages and explicit hooks |

Choose `proxykit` when you want simpler integration and broader transport coverage, not a deep modifier framework.

## Against gateway products

Examples:

- [traefik/traefik](https://github.com/traefik/traefik)
- [zalando/skipper](https://github.com/zalando/skipper)
- [caddyserver/caddy](https://github.com/caddyserver/caddy)

These are better when you want:

- a production gateway platform
- service discovery
- ingress control
- platform-level routing and deployment features

`proxykit` is better when you want:

- a library you embed in your own Go process
- full control over storage and app protocol
- no gateway control plane
- a narrower surface area

## Where `proxykit` is currently weaker

Be honest about current trade-offs:

- fewer examples than older incumbents
- less ecosystem recognition
- fewer third-party integrations
- much newer public identity

That is why this docs site exists: the code is already cleaner than many incumbent alternatives for certain use cases, but the documentation now needs to make that clear quickly.
