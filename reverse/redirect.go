package reverse

import (
	"errors"
	"net/url"
	"strings"
)

// RedirectRewriteOptions configures how upstream redirect locations are
// rewritten back into the local proxy mount.
type RedirectRewriteOptions struct {
	MountPath   string
	TargetParam string
}

// RewriteRedirectLocation rewrites an upstream redirect location so clients
// continue following redirects through the local proxy mount.
//
// Example:
//
//	/next?x=1 -> /proxy/next?_target=https://example.com&x=1
func RewriteRedirectLocation(location string, upstream *url.URL, opts RedirectRewriteOptions) (string, error) {
	if location == "" {
		return "", nil
	}
	if upstream == nil {
		return "", errors.New("reverse: upstream target is required")
	}

	redirectURL, err := url.Parse(location)
	if err != nil {
		return "", err
	}

	base := upstream
	if redirectURL.IsAbs() {
		base = redirectURL
	}
	resolved := base.ResolveReference(redirectURL)

	targetParam := opts.TargetParam
	if targetParam == "" {
		targetParam = "_target"
	}

	proxyURL := url.URL{
		Path:     joinMountPath(opts.MountPath, resolved.EscapedPath()),
		Fragment: resolved.Fragment,
	}
	query := url.Values{}
	query.Set(targetParam, resolved.Scheme+"://"+resolved.Host)
	for key, values := range resolved.Query() {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	proxyURL.RawQuery = query.Encode()
	return proxyURL.String(), nil
}

func joinMountPath(mountPath string, escapedPath string) string {
	mountPath = strings.TrimRight(mountPath, "/")
	if escapedPath == "" {
		escapedPath = "/"
	}
	if mountPath == "" {
		return escapedPath
	}
	return mountPath + escapedPath
}
