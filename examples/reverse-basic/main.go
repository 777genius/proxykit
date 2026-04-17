package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/777genius/proxykit/reverse"
)

func main() {
	listen := flag.String("listen", ":8080", "listen address")
	mount := flag.String("mount", "/proxy", "mounted route prefix")
	target := flag.String("target", "", "default upstream target such as https://httpbin.org")
	flag.Parse()

	handler, err := reverse.New(reverse.Options{
		Resolver: reverse.QueryTargetResolver{
			MountPath:     *mount,
			DefaultTarget: *target,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle(*mount+"/", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("reverse proxy listening on %s with mount %s", *listen, *mount)
	log.Fatal(http.ListenAndServe(*listen, mux))
}
