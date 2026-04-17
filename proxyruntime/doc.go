// Package proxyruntime manages long-lived proxy listeners such as HTTP
// forward-proxy and SOCKS5 endpoints.
//
// It is transport-agnostic at the application level: it owns lifecycle,
// restart semantics, and port conflict diagnostics, but it does not know
// about UI endpoints, persistence, or capture storage.
package proxyruntime
