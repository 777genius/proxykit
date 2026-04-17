package cookies

import (
	"fmt"
	"net/http"
)

func Example() {
	headers := http.Header{}
	headers.Add("Set-Cookie", "session=abc; Path=/; HttpOnly")
	headers.Set("Cookie", "gpx_demo__session=abc; theme=dark")

	RewriteSetCookies(headers, RewriteOptions{
		Mode:            ModeIsolate,
		Namespace:       "demo",
		DomainStrategy:  "hostOnly",
		PathStrategy:    "prefix",
		ProxyPathPrefix: "/proxy",
	})
	RewriteOutboundCookies(headers, RewriteOptions{
		Mode:      ModeIsolate,
		Namespace: "demo",
	})

	fmt.Println(headers.Values("Set-Cookie")[0])
	fmt.Println(headers.Get("Cookie"))
	// Output:
	// gpx_demo__session=abc; Path=/proxy; HttpOnly
	// session=abc
}
