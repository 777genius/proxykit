package httpcapture

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"github.com/cherrypick-agency/proxykit/observe"
	"github.com/cherrypick-agency/proxykit/proxyhttp"
)

func CloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	out := *u
	return &out
}

func CopyResponseHeaders(dst, src http.Header) {
	clean := observe.CloneHeader(src)
	proxyhttp.RemoveHopHeaders(clean)
	for key, values := range clean {
		dst[key] = append([]string(nil), values...)
	}
}

func SampleReadCloser(rc io.ReadCloser, max int, contentType string, contentEncoding string) (observe.InlineBody, io.ReadCloser, error) {
	if rc == nil {
		return observe.InlineBody{
			ContentType:     contentType,
			ContentEncoding: contentEncoding,
		}, nil, nil
	}
	if max <= 0 {
		return observe.InlineBody{
			ContentType:     contentType,
			ContentEncoding: contentEncoding,
		}, rc, nil
	}
	buf := make([]byte, max+1)
	n, err := io.ReadFull(rc, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		_ = rc.Close()
		return observe.InlineBody{}, nil, err
	}
	prefix := append([]byte(nil), buf[:n]...)
	inline := prefix
	truncated := false
	totalSize := int64(n)
	if n > max {
		inline = append([]byte(nil), prefix[:max]...)
		truncated = true
		totalSize = -1
	}
	restored := io.NopCloser(io.MultiReader(bytes.NewReader(prefix), rc))
	return observe.InlineBody{
		Bytes:           inline,
		TotalSize:       totalSize,
		Truncated:       truncated,
		ContentType:     contentType,
		ContentEncoding: contentEncoding,
	}, restored, nil
}
