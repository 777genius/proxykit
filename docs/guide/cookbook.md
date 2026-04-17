# Cookbook

These recipes focus on practical embedding patterns, not on building a full gateway product.

## Mounted reverse proxy with a stable mount path

Use this when your app serves `/proxy/...` and forwards to a target chosen by `_target`.

```go
package main

import (
  "log"
  "net/http"

  reverseproxy "github.com/777genius/proxykit/reverse"
)

func main() {
  proxy, err := reverseproxy.New(reverseproxy.Options{
    Resolver: reverseproxy.QueryTargetResolver{
      MountPath: "/proxy",
    },
  })
  if err != nil {
    log.Fatal(err)
  }

  mux := http.NewServeMux()
  mux.Handle("/proxy/", proxy)

  log.Fatal(http.ListenAndServe(":8080", mux))
}
```

## Forward proxy plus CONNECT on one listener

Use this when HTTP requests and HTTPS tunnels share the same port.

```go
package main

import (
  "net/http"

  "github.com/777genius/proxykit/connect"
  "github.com/777genius/proxykit/forward"
)

func main() {
  forwardHandler := forward.New(forward.Options{})
  connectHandler := connect.New(connect.Options{})

  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodConnect {
      connectHandler.ServeHTTP(w, r)
      return
    }
    forwardHandler.ServeHTTP(w, r)
  })

  _ = http.ListenAndServe(":8080", handler)
}
```

## Observe HTTP traffic without choosing a storage model yet

Use this when you want capture or metrics, but you do not want the transport package to own persistence.

```go
package main

import (
  "context"
  "log"

  "github.com/777genius/proxykit/observe"
  reverseproxy "github.com/777genius/proxykit/reverse"
)

func main() {
  _, _ = reverseproxy.New(reverseproxy.Options{
    Resolver: reverseproxy.QueryTargetResolver{
      MountPath: "/proxy",
    },
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
}
```

## Decode Socket.IO events without making WebSocket transport Socket.IO-specific

Use this when only some of your WebSocket traffic carries Socket.IO event frames.

```go
package main

import (
  "context"
  "log"

  "github.com/777genius/proxykit/observe"
  "github.com/777genius/proxykit/socketio"
  "github.com/777genius/proxykit/wsproxy"
)

func main() {
  _, _ = wsproxy.New(wsproxy.Options{
    Resolver: wsproxy.QueryTargetResolver{
      NormalizeScheme: true,
    },
    Hooks: wsproxy.Hooks{
      OnFrame: func(_ context.Context, frame observe.WSFrame) error {
        if frame.Type != observe.WSMessageText {
          return nil
        }
        namespace, event, argsJSON, ok := socketio.ParseEvent(string(frame.Payload))
        if !ok {
          return nil
        }
        log.Printf("socket.io %s %s %s", namespace, event, argsJSON)
        return nil
      },
    },
  })
}
```

## Rewrite cookies at a mounted reverse-proxy boundary

Use this when browser cookies from multiple upstreams would otherwise collide at the proxy host.

```go
package main

import (
  "net/http"

  "github.com/777genius/proxykit/cookies"
)

func rewrite(headers http.Header, namespace string) {
  opts := cookies.RewriteOptions{
    Mode:            cookies.ModeIsolate,
    Namespace:       namespace,
    DomainStrategy:  "hostOnly",
    PathStrategy:    "prefix",
    ProxyPathPrefix: "/proxy",
    HTTPS:           true,
  }

  cookies.RewriteSetCookies(headers, opts)
  cookies.RewriteOutboundCookies(headers, opts)
}
```

## Start and stop listeners dynamically

Use this when your app turns forward and SOCKS listeners on or off from settings or a desktop UI.

```go
package main

import (
  "context"
  "net/http"

  "github.com/777genius/proxykit/proxyruntime"
)

func main() {
  manager := proxyruntime.New(nil)
  handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
  })

  _ = manager.Apply(context.Background(), proxyruntime.ApplyConfig{
    ForwardEnabled: true,
    ForwardAddr:    "127.0.0.1:8080",
    SocksEnabled:   true,
    SocksAddr:      "127.0.0.1:1080",
  }, handler)
}
```

## More examples

The module also ships compile-checked examples in:

- [`reverse/example_test.go`](https://github.com/777genius/proxykit/blob/main/reverse/example_test.go)
- [`forward/example_test.go`](https://github.com/777genius/proxykit/blob/main/forward/example_test.go)
- [`connect/example_test.go`](https://github.com/777genius/proxykit/blob/main/connect/example_test.go)
- [`wsproxy/example_test.go`](https://github.com/777genius/proxykit/blob/main/wsproxy/example_test.go)
- [`proxyruntime/example_test.go`](https://github.com/777genius/proxykit/blob/main/proxyruntime/example_test.go)
