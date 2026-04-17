package mitm

import "strings"

// Policy decides whether a host should be intercepted.
type Policy struct {
	Authority   *Authority
	AllowSuffix []string
	DenySuffix  []string
}

// ShouldIntercept returns true when host is allowed by the policy and an authority exists.
func (p Policy) ShouldIntercept(host string) bool {
	if p.Authority == nil {
		return false
	}
	h := normalizeHost(host)
	if h == "" {
		return false
	}
	lh := strings.ToLower(h)

	for _, deny := range p.DenySuffix {
		deny = strings.ToLower(strings.TrimSpace(deny))
		if deny != "" && (lh == deny || strings.HasSuffix(lh, deny)) {
			return false
		}
	}
	if len(p.AllowSuffix) == 0 {
		return true
	}
	for _, allow := range p.AllowSuffix {
		allow = strings.ToLower(strings.TrimSpace(allow))
		if allow != "" && (lh == allow || strings.HasSuffix(lh, allow)) {
			return true
		}
	}
	return false
}
