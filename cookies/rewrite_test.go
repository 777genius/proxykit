package cookies

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestNamespaceForURL(t *testing.T) {
	u1, _ := url.Parse("https://example.com")
	u2, _ := url.Parse("http://example.com")
	u3, _ := url.Parse("https://example.com:443")

	ns1 := NamespaceForURL(u1)
	ns2 := NamespaceForURL(u2)
	ns3 := NamespaceForURL(u3)

	if len(ns1) != 12 || len(ns2) != 12 || len(ns3) != 12 {
		t.Fatalf("unexpected namespace lengths: %q %q %q", ns1, ns2, ns3)
	}
	if ns1 == ns2 || ns1 == ns3 {
		t.Fatalf("expected namespace to depend on scheme and host: %q %q %q", ns1, ns2, ns3)
	}
}

func TestRewriteSetCookies_IsolatePrefix(t *testing.T) {
	h := http.Header{}
	h.Add("Set-Cookie", "id=1; Path=/; Domain=example.com; SameSite=None")
	opts := RewriteOptions{
		Mode:            ModeIsolate,
		DomainStrategy:  "proxyHost",
		PathStrategy:    "prefix",
		ProxyHost:       "proxy.local:8443",
		ProxyPathPrefix: "/proxy",
		HTTPS:           true,
		Namespace:       "abcd",
	}

	RewriteSetCookies(h, opts)

	got := h.Get("Set-Cookie")
	if !strings.HasPrefix(got, "gpx_abcd__id=") {
		t.Fatalf("cookie name not isolated: %q", got)
	}
	if !strings.Contains(got, "; Domain=proxy.local") {
		t.Fatalf("proxy host domain not applied: %q", got)
	}
	if !strings.Contains(got, "; Path=/proxy") {
		t.Fatalf("proxy path not applied: %q", got)
	}
	if !strings.Contains(got, "; Secure") {
		t.Fatalf("secure not preserved for SameSite=None over HTTPS: %q", got)
	}
}

func TestRewriteOutboundCookies_Isolate(t *testing.T) {
	h := http.Header{}
	h.Add("Cookie", "gpx_abcd__a=1; other=2")
	h.Add("Cookie", "gpx_efgh__b=3; gpx_abcd__c=4")

	RewriteOutboundCookies(h, RewriteOptions{Mode: ModeIsolate, Namespace: "abcd"})

	if got := h.Get("Cookie"); got != "a=1; c=4" {
		t.Fatalf("unexpected outbound cookie rewrite: %q", got)
	}
}

func TestParseAndBuildSetCookieLine(t *testing.T) {
	raw := "sess=abc; Path=/; Domain=api.example.com; HttpOnly; Secure; SameSite=Lax; Priority=High; Partitioned; Foo=Bar"
	parsed, ok := ParseSetCookieLine(raw)
	if !ok {
		t.Fatal("ParseSetCookieLine returned ok=false")
	}
	if parsed.Name != "sess" || parsed.Value != "abc" {
		t.Fatalf("unexpected name/value: %+v", parsed)
	}
	out := BuildSetCookieLine(parsed)
	if !strings.Contains(out, "; Priority=High") || !strings.Contains(out, "; Foo=Bar") {
		t.Fatalf("BuildSetCookieLine lost attributes: %q", out)
	}
}
