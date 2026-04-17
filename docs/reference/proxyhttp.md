# `proxyhttp`

`proxyhttp` contains the narrow HTTP transport helpers shared across other packages.

Use it when you want:

- one place to build proxy-oriented `http.Transport`
- consistent hop-header stripping
- shared forwarded-header policy
- light protocol-detection helpers

## Main helpers

| Helper | Purpose |
| --- | --- |
| `NewTransport` | Builds a proxy-oriented `http.Transport` |
| `RemoveHopHeaders` | Strips hop-by-hop headers before forwarding |
| `ApplyForwardedHeaders` | Applies `X-Forwarded-*` and `Via` values |
| `IsWebSocketRequest` | Detects likely WS upgrades |
| `IsAbsoluteURL` | Detects absolute-URI request strings |

## Example

```go
header := http.Header{}
header.Set("Connection", "keep-alive")

proxyhttp.ApplyForwardedHeaders(header, proxyhttp.ForwardedHeaderConfig{
  ClientIP: "127.0.0.1",
  Proto:    "https",
  Via:      "proxykit",
})
proxyhttp.RemoveHopHeaders(header)
```

## Keep it narrow

`proxyhttp` should stay a supporting package, not grow into a second transport layer. If logic starts owning routes, storage, or capture policy, it probably belongs somewhere else.
