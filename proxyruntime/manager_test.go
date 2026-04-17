package proxyruntime

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestManager_Apply_StartStop(t *testing.T) {
	log := zerolog.Nop()
	m := New(&log)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "ok") })

	if err := m.Apply(context.Background(), ApplyConfig{ForwardEnabled: true, ForwardAddr: "127.0.0.1:0"}, h); err != nil {
		t.Fatalf("apply start: %v", err)
	}
	if m.ForwardAddr() == "" {
		t.Fatalf("no forward listener")
	}

	if err := m.Apply(context.Background(), ApplyConfig{ForwardEnabled: true, ForwardAddr: "127.0.0.1:0", SocksEnabled: true, SocksAddr: "127.0.0.1:0"}, h); err != nil {
		t.Fatalf("apply update: %v", err)
	}
	if m.SocksAddr() == "" {
		t.Fatalf("no socks listener")
	}

	if err := m.Apply(context.Background(), ApplyConfig{}, h); err != nil {
		t.Fatalf("apply stop: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
}
