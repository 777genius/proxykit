# `reverse`

`reverse` provides a mounted reverse HTTP proxy handler with:

- target resolution
- path and query merging
- hop-by-hop header stripping
- request and response mutation hooks
- neutral observation hooks

## Use it for

- `/proxy/...?_target=https://upstream.example`
- reverse-proxy adapters in debuggers or developer tools
- product-specific reverse proxy endpoints without hard-coding those routes into the library

## Main entry point

```go
handler, err := reverse.New(reverse.Options{
  Resolver: reverse.QueryTargetResolver{
    MountPath: "/proxy",
  },
})
```

## Important options

| Option | Meaning |
| --- | --- |
| `Resolver` | Required target resolver |
| `RoundTripper` | Custom `http.RoundTripper` |
| `MutateRequest` | Request mutation before transport round trip |
| `MutateResponse` | Response mutation before response is copied downstream |
| `Hooks` | Shared observation hooks |
| `ObserveTarget` | Predicate to suppress observation for selected targets |
| `PreserveHost` | Keep incoming `Host` instead of replacing with upstream host |
| `SampleRequestBodyBytes` | Sample limit for observed request body |
| `SampleResponseBodyBytes` | Sample limit for observed response body |

## Built-in resolver

`QueryTargetResolver` resolves `_target`, applies an optional default target, strips selected query parameters, and merges the request path and query into the upstream URL.

That gives you mounted reverse-proxy behavior without forcing route names into the package itself.

