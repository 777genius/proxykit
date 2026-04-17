package proxyhttp

import (
	"fmt"
	"net/http"
)

func Example() {
	header := http.Header{}
	header.Set("Connection", "keep-alive")
	header.Set("Upgrade", "websocket")

	ApplyForwardedHeaders(header, ForwardedHeaderConfig{
		ClientIP: "127.0.0.1",
		Proto:    "https",
		Via:      "proxykit",
	})
	RemoveHopHeaders(header)

	fmt.Println(header.Get("X-Forwarded-For"))
	fmt.Println(header.Get("X-Forwarded-Proto"))
	fmt.Println(header.Get("Via"))
	fmt.Println(header.Get("Connection") == "")
	fmt.Println(IsAbsoluteURL("https://example.com/api"))
	// Output:
	// 127.0.0.1
	// https
	// proxykit
	// true
	// true
}
