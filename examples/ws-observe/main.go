package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/777genius/proxykit/observe"
	"github.com/777genius/proxykit/wsproxy"
)

func main() {
	listen := flag.String("listen", ":8080", "listen address")
	target := flag.String("target", "", "default websocket target such as wss://echo.websocket.events")
	mount := flag.String("mount", "/ws", "mounted route prefix")
	flag.Parse()

	handler, err := wsproxy.New(wsproxy.Options{
		Resolver: wsproxy.QueryTargetResolver{
			DefaultTarget:   *target,
			NormalizeScheme: true,
		},
		Hooks: wsproxy.Hooks{
			OnSessionOpen: func(_ context.Context, session observe.Session) error {
				log.Printf("ws open target=%s session=%s", session.Target, session.ID)
				return nil
			},
			OnFrame: func(_ context.Context, frame observe.WSFrame) error {
				log.Printf("ws frame dir=%s type=%s bytes=%d", frame.Direction, frame.Type, len(frame.Payload))
				return nil
			},
			OnSessionClose: func(_ context.Context, info observe.CloseInfo) {
				log.Printf("ws close session=%s err=%v", info.SessionID, info.Err)
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle(*mount, handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("websocket proxy listening on %s with mount %s", *listen, *mount)
	log.Fatal(http.ListenAndServe(*listen, mux))
}
