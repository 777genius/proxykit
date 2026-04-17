package forward

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Example() {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s %s %s", r.Method, r.URL.Path, r.Header.Get("X-Proxy-Mode"))
	}))
	defer upstream.Close()

	handler := New(Options{
		MutateRequest: func(_ context.Context, req *http.Request) error {
			req.Header.Set("X-Proxy-Mode", "forward")
			return nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, upstream.URL+"/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 200
	// GET /status forward
}
