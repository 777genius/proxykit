# `proxyruntime`

`proxyruntime` manages forward-proxy and SOCKS listeners for applications that need runtime enable, disable, or restart behavior.

It is intentionally headless.

It does **not** know about:

- your config DTOs
- your settings page
- your admin REST endpoints

## Main entry point

```go
manager := proxyruntime.New(logger)
err := manager.Apply(ctx, proxyruntime.ApplyConfig{
  ForwardEnabled: true,
  ForwardAddr:    "127.0.0.1:0",
  SocksEnabled:   true,
  SocksAddr:      "127.0.0.1:0",
}, handler)
```

## Main capabilities

- start and stop forward listeners
- start and stop SOCKS listeners
- graceful listener restart
- actual bound address lookup via `ForwardAddr()` and `SocksAddr()`
- human-readable port conflict diagnostics

## Good usage pattern

Store your configuration elsewhere, then translate it into `ApplyConfig` in your app layer.

That keeps `proxyruntime` reusable across products with different configuration APIs.

