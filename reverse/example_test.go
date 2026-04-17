package reverse

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func Example() {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "upstream %s", r.URL.Path)
	}))
	defer upstream.Close()

	handler, err := New(Options{
		Resolver: QueryTargetResolver{MountPath: "/proxy"},
	})
	if err != nil {
		panic(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/proxy/api/hello?_target="+url.QueryEscape(upstream.URL), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(strings.TrimSpace(rec.Body.String()))
	// Output:
	// 200
	// upstream /api/hello
}
