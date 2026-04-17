package proxyruntime

import (
	"context"
	"fmt"
	"net/http"
)

func Example() {
	manager := New(nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if err := manager.Apply(context.Background(), ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    "127.0.0.1:0",
		SocksEnabled:   true,
		SocksAddr:      "127.0.0.1:0",
	}, handler); err != nil {
		panic(err)
	}
	defer func() {
		_ = manager.StopForward(context.Background())
		_ = manager.StopSocks(context.Background())
	}()

	resp, err := http.Get("http://" + manager.ForwardAddr() + "/healthz")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	fmt.Println(manager.SocksAddr() != "")
	// Output:
	// 200
	// true
}
