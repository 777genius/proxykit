// Package socketio parses Socket.IO event-style packets from text frames.
//
// The parser is intentionally lightweight so higher-level proxy packages can
// optionally decode Socket.IO traffic without taking a hard dependency on a
// specific capture or storage model.
package socketio
