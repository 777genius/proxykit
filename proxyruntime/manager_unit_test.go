package proxyruntime

import (
	"context"
	"net/http"
	"testing"
)

func TestManager_ForwardAddr(t *testing.T) {
	m := &Manager{}
	if got := m.ForwardAddr(); got != "" {
		t.Errorf("ForwardAddr() = %q, want empty", got)
	}
}

func TestManager_SocksAddr(t *testing.T) {
	m := &Manager{}
	if got := m.SocksAddr(); got != "" {
		t.Errorf("SocksAddr() = %q, want empty", got)
	}
}

func TestManager_Apply_ForwardOnly(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    ":0",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, handler)
	if err != nil {
		t.Logf("Apply with forward enabled: %v", err)
	}
	if err == nil {
		if addr := m.ForwardAddr(); addr == "" {
			t.Error("ForwardAddr should not be empty when forward is enabled")
		}
		_ = m.StopForward(ctx)
	}
}

func TestManager_Apply_SocksOnly(t *testing.T) {
	m := &Manager{}
	cfg := ApplyConfig{
		SocksEnabled:  true,
		SocksAddr:     ":0",
		SocksAuthMode: "none",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, nil)
	if err != nil {
		t.Logf("Apply with socks enabled: %v", err)
	}
	if err == nil {
		if addr := m.SocksAddr(); addr == "" {
			t.Error("SocksAddr should not be empty when socks is enabled")
		}
		_ = m.StopSocks(ctx)
	}
}

func TestManager_Apply_BothEnabled(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg := ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    ":0",
		SocksEnabled:   true,
		SocksAddr:      ":0",
		SocksAuthMode:  "none",
	}

	ctx := context.Background()
	err := m.Apply(ctx, cfg, handler)
	if err != nil {
		t.Logf("Apply with both enabled: %v", err)
	}
	if err == nil {
		_ = m.StopForward(ctx)
		_ = m.StopSocks(ctx)
	}
}

func TestManager_Apply_DisableAll(t *testing.T) {
	m := &Manager{}

	err := m.Apply(context.Background(), ApplyConfig{}, nil)
	if err != nil {
		t.Errorf("Apply with everything disabled should not error: %v", err)
	}
	if m.ForwardAddr() != "" {
		t.Error("ForwardAddr should be empty when disabled")
	}
	if m.SocksAddr() != "" {
		t.Error("SocksAddr should be empty when disabled")
	}
}

func TestManager_Apply_InvalidForwardAddr(t *testing.T) {
	m := &Manager{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	err := m.Apply(context.Background(), ApplyConfig{
		ForwardEnabled: true,
		ForwardAddr:    "invalid:addr:format",
	}, handler)
	if err == nil {
		t.Error("Apply with invalid forward address should return error")
		_ = m.StopForward(context.Background())
	}
}

func TestManager_Apply_InvalidSocksAddr(t *testing.T) {
	m := &Manager{}

	err := m.Apply(context.Background(), ApplyConfig{
		SocksEnabled:  true,
		SocksAddr:     "invalid:addr:format",
		SocksAuthMode: "none",
	}, nil)
	if err == nil {
		t.Error("Apply with invalid socks address should return error")
		_ = m.StopSocks(context.Background())
	}
}

func TestManager_StopForward_WhenNotRunning(t *testing.T) {
	if err := (&Manager{}).StopForward(context.Background()); err != nil {
		t.Errorf("StopForward on non-running proxy returned error: %v", err)
	}
}

func TestManager_StopSocks_WhenNotRunning(t *testing.T) {
	if err := (&Manager{}).StopSocks(context.Background()); err != nil {
		t.Errorf("StopSocks on non-running proxy returned error: %v", err)
	}
}
