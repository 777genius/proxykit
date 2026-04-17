# Getting Started

## Install

```bash
go get github.com/777genius/proxykit@v0.1.6
```

`proxykit` is a library-first project. You embed handlers inside your application instead of running a prebuilt control plane.

## First reverse proxy

This is the smallest useful mounted reverse proxy:

```go
package main

import (
  "log"
  "net/http"

  reverseproxy "github.com/777genius/proxykit/reverse"
)

func main() {
  handler, err := reverseproxy.New(reverseproxy.Options{
    Resolver: reverseproxy.QueryTargetResolver{
      MountPath: "/proxy",
    },
  })
  if err != nil {
    log.Fatal(err)
  }

  mux := http.NewServeMux()
  mux.Handle("/proxy/", handler)

  log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Request flow:

```text
GET /proxy/api/users?_target=https://example.com
-> upstream target https://example.com/api/users
```

## Add observation hooks

You can attach session, HTTP, and WebSocket observation without coupling to a storage engine:

```go
handler, err := reverseproxy.New(reverseproxy.Options{
  Resolver: reverseproxy.QueryTargetResolver{MountPath: "/proxy"},
  Hooks: observe.Hooks{
    OnSessionOpen: func(_ context.Context, s observe.Session) error {
      log.Printf("open %s %s", s.Kind, s.Target)
      return nil
    },
    OnHTTPRoundTrip: func(_ context.Context, rt observe.HTTPRoundTrip) error {
      log.Printf("%s -> %d", rt.Request.URL, rt.Response.StatusCode)
      return nil
    },
  },
})
```

## Recommended adoption path

If you are introducing `proxykit` into an existing application, this order usually keeps risk down:

1. Start with one transport package such as `reverse` or `wsproxy`.
2. Keep your existing routes and DTOs in your own app layer.
3. Move capture, persistence, and delivery logic behind observation hooks.
4. Add `forward`, `connect`, `cookies`, or `proxyruntime` only when the product really needs them.

## Where to go next

- [Architecture](/guide/architecture) for what belongs inside `proxykit`
- [Packages](/guide/packages) for the package map
- [Observation Hooks](/guide/observation-hooks) for the main extension surface
- [Comparisons](/guide/comparisons) for how this differs from `goproxy`, `oxy`, and `martian`
