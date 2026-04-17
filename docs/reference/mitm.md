# `mitm`

`mitm` provides optional TLS interception building blocks for tools that need development-grade certificate issuance and host filtering.

It exists as a separate package because MITM is useful, but it should not leak into plain reverse, forward, CONNECT, or WebSocket transport by default.

## Main pieces

| Type or function | Purpose |
| --- | --- |
| `GenerateDevCA` | Generates a development root CA |
| `LoadAuthority` / `LoadAuthorityFromPEM` | Loads CA material |
| `Authority.IssueFor` | Issues and caches per-host leaf certificates |
| `Policy` | Controls whether a host should be intercepted |

## Example

```go
certPEM, keyPEM, _ := mitm.GenerateDevCA("proxykit dev ca", 1)
authority, _ := mitm.LoadAuthorityFromPEM(certPEM, keyPEM)

policy := mitm.Policy{
  Authority:   authority,
  AllowSuffix: []string{".example.com"},
}

leaf, _ := authority.IssueFor("api.example.com:443")
_ = leaf
```

## Important boundary

This package does **not** provide:

- a full MITM proxy handler
- CA installation into an operating system trust store
- product-specific certificate management workflows

Those concerns belong in your application layer or in an optional higher-level package.
