# `socketio`

`socketio` parses event-style Socket.IO text packets without forcing Socket.IO into the WebSocket transport itself.

That design is intentional:

- `wsproxy` stays reusable for all WebSocket traffic
- Socket.IO decoding remains optional
- higher-level protocol interpretation happens in your adapter layer

## Main entry point

```go
namespace, event, argsJSON, ok := socketio.ParseEvent(packet)
```

## What it supports

`ParseEvent` understands common Socket.IO v3 and v4 event-like text packet forms:

- `42` event packets
- `45` binary event packets
- `43` ack packets
- `46` binary ack packets

## Example

```go
namespace, event, argsJSON, ok := socketio.ParseEvent(`42/chat,17["message",{"body":"hello"}]`)
```

Typical result:

- `namespace`: `/chat`
- `event`: `message`
- `argsJSON`: raw JSON array payload

## Why raw JSON output is useful

The package returns the event payload as JSON text instead of binding it to a fixed struct.

That keeps the package:

- lightweight
- schema-agnostic
- useful for debuggers, inspectors, and custom adapters
