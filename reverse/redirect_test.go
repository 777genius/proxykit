package reverse

import (
	"net/url"
	"testing"
)

func TestRewriteRedirectLocation_Relative(t *testing.T) {
	upstream, err := url.Parse("https://example.com/base/api")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	rewritten, err := RewriteRedirectLocation("/final?x=1&x=2", upstream, RedirectRewriteOptions{
		MountPath: "/proxy",
	})
	if err != nil {
		t.Fatalf("RewriteRedirectLocation() error = %v", err)
	}

	got, err := url.Parse(rewritten)
	if err != nil {
		t.Fatalf("Parse(rewritten) error = %v", err)
	}
	if got.Path != "/proxy/final" {
		t.Fatalf("path = %q, want %q", got.Path, "/proxy/final")
	}
	if got.Query().Get("_target") != "https://example.com" {
		t.Fatalf("_target = %q, want %q", got.Query().Get("_target"), "https://example.com")
	}
	values := got.Query()["x"]
	if len(values) != 2 || values[0] != "1" || values[1] != "2" {
		t.Fatalf("x query = %#v, want [1 2]", values)
	}
}

func TestRewriteRedirectLocation_AbsoluteAndFragment(t *testing.T) {
	upstream, err := url.Parse("https://example.com/base")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	rewritten, err := RewriteRedirectLocation("https://other.example/login?from=proxy#done", upstream, RedirectRewriteOptions{
		MountPath:   "/httpproxy",
		TargetParam: "_upstream",
	})
	if err != nil {
		t.Fatalf("RewriteRedirectLocation() error = %v", err)
	}

	got, err := url.Parse(rewritten)
	if err != nil {
		t.Fatalf("Parse(rewritten) error = %v", err)
	}
	if got.Path != "/httpproxy/login" {
		t.Fatalf("path = %q, want %q", got.Path, "/httpproxy/login")
	}
	if got.Query().Get("_upstream") != "https://other.example" {
		t.Fatalf("_upstream = %q, want %q", got.Query().Get("_upstream"), "https://other.example")
	}
	if got.Query().Get("from") != "proxy" {
		t.Fatalf("from = %q, want %q", got.Query().Get("from"), "proxy")
	}
	if got.Fragment != "done" {
		t.Fatalf("fragment = %q, want %q", got.Fragment, "done")
	}
}

func TestRewriteRedirectLocation_RootMountAndRootPath(t *testing.T) {
	upstream, err := url.Parse("http://example.com/base")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	rewritten, err := RewriteRedirectLocation("/", upstream, RedirectRewriteOptions{
		MountPath: "/proxy/",
	})
	if err != nil {
		t.Fatalf("RewriteRedirectLocation() error = %v", err)
	}
	if rewritten != "/proxy/?_target=http%3A%2F%2Fexample.com" {
		t.Fatalf("rewritten = %q, want %q", rewritten, "/proxy/?_target=http%3A%2F%2Fexample.com")
	}
}
