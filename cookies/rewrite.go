package cookies

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/url"
	"strings"
)

const (
	ModeIsolate = "isolate"
	ModeAuto    = "auto"
	ModeOff     = "off"
)

// RewriteOptions controls how Set-Cookie and Cookie headers are rewritten at a
// reverse-proxy boundary.
type RewriteOptions struct {
	Mode            string
	DomainStrategy  string
	PathStrategy    string
	ProxyHost       string
	ProxyPathPrefix string
	HTTPS           bool
	Namespace       string
}

// Header is the minimal interface needed to rewrite cookie headers.
type Header interface {
	Values(key string) []string
	Set(key, value string)
	Add(key, value string)
	Del(key string)
}

// ParsedSetCookie is a lightweight Set-Cookie representation that preserves
// unrecognized tokens on rebuild.
type ParsedSetCookie struct {
	Name        string
	Value       string
	Attrs       map[string]string
	Flags       map[string]bool
	ExtraTokens []string
}

// NamespaceForURL derives a stable namespace from scheme and host.
func NamespaceForURL(u *url.URL) string {
	base := u.Scheme + "://" + u.Host
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])[:12]
}

// SanitizeHost strips a port and any leading dot from a host value.
func SanitizeHost(hostport string) string {
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		return h
	}
	return strings.TrimPrefix(hostport, ".")
}

// DomainAttrSafe reports whether a cookie Domain attribute is safe for the
// supplied host.
func DomainAttrSafe(host string) bool {
	h := strings.ToLower(host)
	if h == "localhost" {
		return false
	}
	if net.ParseIP(h) != nil {
		return false
	}
	return true
}

// NamespacePrefix returns the browser-storage prefix used for isolate mode.
func NamespacePrefix(ns string) string {
	if ns == "" {
		return "gpx__"
	}
	return "gpx_" + ns + "__"
}

// RewriteSetCookies mutates Set-Cookie headers in h according to opts.
func RewriteSetCookies(h Header, opts RewriteOptions) {
	if opts.Mode == ModeOff {
		return
	}
	orig := h.Values("Set-Cookie")
	if len(orig) == 0 {
		return
	}
	h.Del("Set-Cookie")

	pfx := NamespacePrefix(opts.Namespace)
	proxyPath := "/"
	if opts.PathStrategy == "prefix" && opts.ProxyPathPrefix != "" {
		proxyPath = opts.ProxyPathPrefix
	}

	for _, raw := range orig {
		p, ok := ParseSetCookieLine(raw)
		if !ok {
			h.Add("Set-Cookie", raw)
			continue
		}

		origName := p.Name
		isHostPrefix := strings.HasPrefix(origName, "__Host-")
		isSecurePrefix := strings.HasPrefix(origName, "__Secure-")

		if opts.Mode == ModeIsolate {
			p.Name = pfx + origName
		}

		if isHostPrefix {
			delete(p.Attrs, "domain")
		} else if opts.DomainStrategy == "hostOnly" {
			delete(p.Attrs, "domain")
		} else if opts.DomainStrategy == "proxyHost" {
			dh := SanitizeHost(opts.ProxyHost)
			if DomainAttrSafe(dh) {
				p.Attrs["domain"] = dh
			} else {
				delete(p.Attrs, "domain")
			}
		}

		if isHostPrefix {
			p.Attrs["path"] = "/"
		} else if opts.PathStrategy == "prefix" {
			p.Attrs["path"] = proxyPath
		} else {
			p.Attrs["path"] = "/"
		}

		if opts.HTTPS {
			if strings.EqualFold(p.Attrs["samesite"], "none") || isSecurePrefix || isHostPrefix {
				p.Flags["secure"] = true
			}
		}

		h.Add("Set-Cookie", BuildSetCookieLine(p))
	}
}

// RewriteOutboundCookies rewrites Cookie headers for upstream requests in
// isolate mode by filtering to the current namespace and unwrapping names.
func RewriteOutboundCookies(h Header, opts RewriteOptions) {
	if opts.Mode != ModeIsolate {
		return
	}
	cookies := h.Values("Cookie")
	if len(cookies) == 0 {
		return
	}
	combined := strings.Join(cookies, "; ")
	pairs := strings.Split(combined, ";")
	pfx := NamespacePrefix(opts.Namespace)
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(p[:eq])
		val := strings.TrimSpace(p[eq+1:])
		if strings.HasPrefix(name, pfx) {
			out = append(out, strings.TrimPrefix(name, pfx)+"="+val)
		}
	}
	h.Del("Cookie")
	if len(out) > 0 {
		h.Set("Cookie", strings.Join(out, "; "))
	}
}

// ParseSetCookieLine parses a Set-Cookie line while preserving unrecognized
// tokens for round-trip rebuilding.
func ParseSetCookieLine(raw string) (ParsedSetCookie, bool) {
	parts := strings.Split(raw, ";")
	if len(parts) == 0 {
		return ParsedSetCookie{}, false
	}
	first := strings.TrimSpace(parts[0])
	eq := strings.Index(first, "=")
	if eq <= 0 {
		return ParsedSetCookie{}, false
	}
	p := ParsedSetCookie{
		Name:  strings.TrimSpace(first[:eq]),
		Value: strings.TrimSpace(first[eq+1:]),
		Attrs: map[string]string{},
		Flags: map[string]bool{},
	}
	for i := 1; i < len(parts); i++ {
		tok := strings.TrimSpace(parts[i])
		if tok == "" {
			continue
		}
		lower := strings.ToLower(tok)
		switch lower {
		case "secure":
			p.Flags["secure"] = true
			continue
		case "httponly":
			p.Flags["httponly"] = true
			continue
		case "partitioned":
			p.Flags["partitioned"] = true
			continue
		}
		if kv := strings.Index(tok, "="); kv > 0 {
			key := strings.ToLower(strings.TrimSpace(tok[:kv]))
			val := strings.TrimSpace(tok[kv+1:])
			switch key {
			case "domain", "path", "expires", "max-age", "samesite", "priority":
				p.Attrs[key] = val
			default:
				p.ExtraTokens = append(p.ExtraTokens, tok)
			}
		} else {
			p.ExtraTokens = append(p.ExtraTokens, tok)
		}
	}
	return p, true
}

// BuildSetCookieLine serializes a parsed Set-Cookie value back to a header line.
func BuildSetCookieLine(p ParsedSetCookie) string {
	var b strings.Builder
	b.WriteString(p.Name)
	b.WriteByte('=')
	b.WriteString(p.Value)
	if v := p.Attrs["domain"]; v != "" {
		b.WriteString("; Domain=")
		b.WriteString(v)
	}
	if v := p.Attrs["path"]; v != "" {
		b.WriteString("; Path=")
		b.WriteString(v)
	}
	if v := p.Attrs["expires"]; v != "" {
		b.WriteString("; Expires=")
		b.WriteString(v)
	}
	if v := p.Attrs["max-age"]; v != "" {
		b.WriteString("; Max-Age=")
		b.WriteString(v)
	}
	if p.Flags["httponly"] {
		b.WriteString("; HttpOnly")
	}
	if p.Flags["secure"] {
		b.WriteString("; Secure")
	}
	if v := p.Attrs["samesite"]; v != "" {
		switch strings.ToLower(v) {
		case "lax":
			b.WriteString("; SameSite=Lax")
		case "strict":
			b.WriteString("; SameSite=Strict")
		case "none":
			b.WriteString("; SameSite=None")
		default:
			b.WriteString("; SameSite=")
			b.WriteString(v)
		}
	}
	if v := p.Attrs["priority"]; v != "" {
		b.WriteString("; Priority=")
		b.WriteString(v)
	}
	if p.Flags["partitioned"] {
		b.WriteString("; Partitioned")
	}
	for _, token := range p.ExtraTokens {
		b.WriteString("; ")
		b.WriteString(token)
	}
	return b.String()
}
