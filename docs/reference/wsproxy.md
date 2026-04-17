# `wsproxy`

`wsproxy` provides a reusable WebSocket proxy handler with:

- target resolution
- optional `http`/`https` to `ws`/`wss` normalization
- bidirectional frame forwarding
- frame and protocol observation hooks
- optional plaintext fallback for TLS-mismatch targets

## Main entry point

```go
handler, err := wsproxy.New(wsproxy.Options{
  Resolver: wsproxy.QueryTargetResolver{
    NormalizeScheme: true,
  },
})
```

## Important options

| Option | Meaning |
| --- | --- |
| `Resolver` | Required target resolver |
| `ObserveTarget` | Predicate to suppress observation |
| `Hooks` | WebSocket-specific hook set |
| `HandshakeTimeout` | Upstream dial timeout |
| `InsecureTLS` | Skip upstream TLS verification |
| `AllowPlaintextFallback` | Retry `wss` targets as plaintext when appropriate |
| `BeforeForward` | Per-frame forward decision hook |
| `ForwardHeaders` | Selected request headers to forward upstream |
| `SynthesizeOrigin` | Optionally synthesize an upstream `Origin` |

## Why it matters

A lot of proxy libraries treat WebSocket as an afterthought. `wsproxy` is a first-class package with its own clear scope instead of an awkward branch inside an HTTP proxy object.

