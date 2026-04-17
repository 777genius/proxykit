# `connect`

`connect` handles plain HTTP `CONNECT` tunneling.

It owns:

- hijacking the downstream connection
- dialing the upstream target
- sending `200 Connection Established`
- copying bytes in both directions
- lifecycle hooks

It does **not** own MITM policy. MITM belongs in a separate optional layer.

## Main entry point

```go
handler := connect.New(connect.Options{})
```

## Important options

| Option | Meaning |
| --- | --- |
| `DialContext` | Custom network dialer |
| `ObserveRequest` | Predicate to suppress observation |
| `Hooks` | Shared observation hooks |
| `Network` | Dial network, default `tcp` |
| `WriteError` | Custom HTTP error writer before hijack |

## Protocol events

When a tunnel is established, `connect` can emit a neutral protocol event:

- namespace: `/_sys`
- name: `tunnel_established`

Your application can use that signal or ignore it.

