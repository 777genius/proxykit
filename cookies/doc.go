// Package cookies provides proxy-safe cookie rewriting helpers for mounted
// reverse proxies.
//
// The package focuses on a narrow but important boundary: rewriting Set-Cookie
// and Cookie headers so a mounted reverse proxy can safely expose multiple
// upstreams without browser-cookie collisions. It does not own routing,
// storage, or session semantics.
package cookies
