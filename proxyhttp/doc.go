// Package proxyhttp provides shared HTTP proxy transport primitives.
//
// The package intentionally stays narrow: transport construction, protocol
// detection helpers, and hop-by-hop header handling that can be reused by
// forward, reverse, and MITM proxy implementations.
package proxyhttp
