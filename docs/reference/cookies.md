# `cookies`

`cookies` provides helpers for mounted reverse proxies that need browser-safe cookie isolation.

## Main capabilities

- rewrite `Set-Cookie` headers for reverse-proxy boundaries
- isolate browser storage by upstream namespace
- preserve unknown `Set-Cookie` tokens on round trip
- rewrite outbound `Cookie` headers back to upstream form

## Main concepts

| Concept | Meaning |
| --- | --- |
| `ModeOff` | Do not rewrite cookies |
| `ModeAuto` | Conservative automatic rewriting |
| `ModeIsolate` | Namespace browser cookies per upstream target |
| `NamespaceForURL` | Stable namespace derived from upstream URL |
| `RewriteSetCookies` | Rewrites upstream `Set-Cookie` values |
| `RewriteOutboundCookies` | Filters and unwraps cookies on the way back upstream |

## Why it is separate

Cookie rewriting is policy-heavy. It should not be hard-coded into a reverse proxy engine.

Keeping it as a focused package lets applications choose:

- whether they need isolation at all
- which namespace strategy to use
- how mounted paths should affect cookie paths

