// Package forward provides a reusable HTTP forward proxy handler for absolute
// URI requests.
//
// The package owns request normalization, hop-by-hop header stripping,
// forwarding header policy, outbound transport execution, and neutral
// observation hooks. CONNECT tunneling, WebSocket upgrades, persistence, and
// app-specific monitoring remain outside the package.
package forward
