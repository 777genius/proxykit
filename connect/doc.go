// Package connect provides an HTTP CONNECT tunneling handler.
//
// The package owns CONNECT method validation, connection hijacking, upstream
// dialing, CONNECT acknowledgement, raw bidirectional tunneling, and neutral
// observation hooks. It does not own MITM, persistence, or application
// delivery contracts.
package connect
