package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/777genius/proxykit/connect"
	"github.com/777genius/proxykit/forward"
)

func main() {
	listen := flag.String("listen", ":8080", "listen address")
	flag.Parse()

	forwardHandler := forward.New(forward.Options{})
	connectHandler := connect.New(connect.Options{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			connectHandler.ServeHTTP(w, r)
			return
		}
		forwardHandler.ServeHTTP(w, r)
	})

	log.Printf("forward proxy with CONNECT support listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, handler))
}
