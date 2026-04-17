package wsproxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

func Example() {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}
		_ = conn.WriteMessage(mt, []byte("upstream:"+string(msg)))
	}))
	defer upstream.Close()

	handler, err := New(Options{
		Resolver: QueryTargetResolver{NormalizeScheme: true},
	})
	if err != nil {
		panic(err)
	}
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	target := url.QueryEscape(upstream.URL + "/echo")
	conn, _, err := websocket.DefaultDialer.Dial(httpToWS(proxy.URL)+"/proxy?_target="+target, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
		panic(err)
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	fmt.Println(string(msg))
	// Output:
	// upstream:hello
}

func httpToWS(raw string) string {
	return "ws" + strings.TrimPrefix(raw, "http")
}
