# `forward`

`forward` provides classic HTTP forward proxying for absolute-URI requests.

It is intentionally scoped to HTTP forwarding only.

`CONNECT` is handled by the separate `connect` package.

## Use it for

- clients configured to use your app as an HTTP proxy
- traffic capture tools
- forward-proxy flows where you want custom request/response mutation and neutral observation

## Main entry point

```go
handler := forward.New(forward.Options{
  ForwardedHeaders: proxyhttp.ForwardedHeaderConfig{
    ClientIP: "203.0.113.10",
    Proto:    "http",
    Via:      "proxykit",
  },
})
```

## Important options

| Option | Meaning |
| --- | --- |
| `RoundTripper` | Custom transport |
| `MutateRequest` | Request mutation before forwarding |
| `MutateResponse` | Response mutation before returning to client |
| `Hooks` | Shared observation hooks |
| `ObserveRequest` | Predicate to suppress observation |
| `ForwardedHeaders` | Standard forwarded header policy |
| `SampleRequestBodyBytes` | Observed request body sample limit |
| `SampleResponseBodyBytes` | Observed response body sample limit |

## Important behavior

- requires an absolute request URL
- rejects `CONNECT`
- strips hop-by-hop headers
- applies forwarding headers explicitly instead of implicitly hiding policy in the transport

