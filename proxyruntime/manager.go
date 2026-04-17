package proxyruntime

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"syscall"
	"time"

	socks5 "github.com/armon/go-socks5"
	"github.com/rs/zerolog"
)

// Manager manages separate listeners for forward-proxy and SOCKS5.
// Listener restarts are graceful via Shutdown/Close.
type Manager struct {
	log *zerolog.Logger

	mu sync.Mutex

	// HTTP forward-proxy
	fwdSrv *http.Server
	fwdLn  net.Listener

	// SOCKS5
	socksSrv *socks5.Server
	socksLn  net.Listener
}

func New(log *zerolog.Logger) *Manager { return &Manager{log: log} }

// StartForward starts an HTTP server on addr with the provided handler.
func (m *Manager) StartForward(addr string, handler http.Handler) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fwdLn != nil {
		_ = m.stopForwardLocked(context.Background())
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			health := probeHTTPHealth(addrsForPortProbe(addr))
			if health.ok {
				return fmt.Errorf("forward listen %s: %w (port is already in use; existing service answers /healthz: %s)", addr, err, health.summary)
			}
			if health.summary != "" {
				return fmt.Errorf("forward listen %s: %w (port is already in use; existing service does NOT answer /healthz: %s)", addr, err, health.summary)
			}
			return fmt.Errorf("forward listen %s: %w (port is already in use; existing service does NOT answer /healthz)", addr, err)
		}
		return fmt.Errorf("forward listen %s: %w", addr, err)
	}
	m.fwdLn = ln
	m.fwdSrv = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	srv := m.fwdSrv
	go func() {
		err := srv.Serve(ln)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			if m.log != nil {
				m.log.Error().Err(err).Msg("forward proxy server error")
			}
			_ = ln.Close()
			m.mu.Lock()
			if m.fwdLn == ln {
				m.fwdLn = nil
			}
			if m.fwdSrv == srv {
				m.fwdSrv = nil
			}
			m.mu.Unlock()
		}
	}()
	if m.log != nil {
		m.log.Info().Str("addr", addr).Msg("forward proxy started")
	}
	return nil
}

func (m *Manager) StopForward(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopForwardLocked(ctx)
}

func (m *Manager) stopForwardLocked(ctx context.Context) error {
	if m.fwdSrv == nil {
		return nil
	}
	srv := m.fwdSrv
	ln := m.fwdLn
	m.fwdSrv = nil
	m.fwdLn = nil
	if ln != nil {
		_ = ln.Close()
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	if m.log != nil {
		m.log.Info().Msg("forward proxy stopped")
	}
	return nil
}

// StartSocks starts a SOCKS5 server on addr with optional authentication.
// authMode: "none" | "userpass"; user/pass are only used for userpass.
func (m *Manager) StartSocks(addr, authMode, user, pass string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.socksLn != nil {
		_ = m.stopSocksLocked()
	}

	conf := &socks5.Config{}
	switch authMode {
	case "userpass":
		creds := socks5.StaticCredentials{user: pass}
		auth := socks5.UserPassAuthenticator{Credentials: creds}
		conf.AuthMethods = []socks5.Authenticator{auth}
	default:
		conf.AuthMethods = []socks5.Authenticator{}
	}
	srv, err := socks5.New(conf)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		if errors.Is(err, syscall.EADDRINUSE) {
			health := probeHTTPHealth(addrsForPortProbe(addr))
			if health.ok {
				return fmt.Errorf("socks listen %s: %w (port is already in use; existing service answers /healthz: %s)", addr, err, health.summary)
			}
			if health.summary != "" {
				return fmt.Errorf("socks listen %s: %w (port is already in use; existing service does NOT answer /healthz: %s)", addr, err, health.summary)
			}
			return fmt.Errorf("socks listen %s: %w (port is already in use; existing service does NOT answer /healthz)", addr, err)
		}
		return fmt.Errorf("socks listen %s: %w", addr, err)
	}
	m.socksSrv = srv
	m.socksLn = ln
	go func() {
		if err := srv.Serve(ln); err != nil && m.log != nil {
			m.log.Error().Err(err).Msg("socks server error")
		}
	}()
	if m.log != nil {
		m.log.Info().Str("addr", addr).Msg("socks5 started")
	}
	return nil
}

func (m *Manager) StopSocks(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.stopSocksLocked()
}

func (m *Manager) stopSocksLocked() error {
	if m.socksLn == nil {
		return nil
	}
	_ = m.socksLn.Close()
	m.socksLn = nil
	m.socksSrv = nil
	if m.log != nil {
		m.log.Info().Msg("socks5 stopped")
	}
	return nil
}

// ApplyConfig enables or restarts listeners according to the provided configuration.
type ApplyConfig struct {
	ForwardEnabled bool
	ForwardAddr    string

	SocksEnabled  bool
	SocksAddr     string
	SocksAuthMode string
	SocksUser     string
	SocksPass     string
}

// ForwardAddr returns the actual forward-proxy address if running.
func (m *Manager) ForwardAddr() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fwdLn == nil {
		return ""
	}
	return m.fwdLn.Addr().String()
}

// SocksAddr returns the actual SOCKS5 address if running.
func (m *Manager) SocksAddr() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.socksLn == nil {
		return ""
	}
	return m.socksLn.Addr().String()
}

// Apply accepts a handler for forward-proxy.
func (m *Manager) Apply(ctx context.Context, cfg ApplyConfig, forwardHandler http.Handler) error {
	if cfg.ForwardEnabled {
		if err := m.StartForward(cfg.ForwardAddr, forwardHandler); err != nil {
			return err
		}
	} else {
		_ = m.StopForward(ctx)
	}
	if cfg.SocksEnabled {
		if err := m.StartSocks(cfg.SocksAddr, cfg.SocksAuthMode, cfg.SocksUser, cfg.SocksPass); err != nil {
			return err
		}
	} else {
		_ = m.StopSocks(ctx)
	}
	return nil
}

type healthProbeResult struct {
	ok      bool
	summary string
}

func addrsForPortProbe(listenAddr string) []string {
	port := ""
	if strings.HasPrefix(listenAddr, ":") {
		port = strings.TrimPrefix(listenAddr, ":")
	} else {
		_, p, err := net.SplitHostPort(listenAddr)
		if err == nil {
			port = p
		}
	}
	if port == "" {
		return nil
	}
	return []string{
		net.JoinHostPort("127.0.0.1", port),
		net.JoinHostPort("localhost", port),
		net.JoinHostPort("::1", port),
	}
}

func probeHTTPHealth(addrs []string) healthProbeResult {
	client := &http.Client{Timeout: 750 * time.Millisecond}
	var parts []string
	for _, addr := range addrs {
		url := "http://" + addr + "/healthz"
		resp, err := client.Get(url)
		if err != nil {
			parts = append(parts, fmt.Sprintf("%s -> error: %v", addr, err))
			continue
		}
		body := ""
		if resp.Body != nil {
			reader := bufio.NewReader(resp.Body)
			line, _ := reader.ReadString('\n')
			body = strings.TrimSpace(line)
			_ = resp.Body.Close()
		}
		parts = append(parts, fmt.Sprintf("%s -> %s %s", addr, resp.Status, body))
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return healthProbeResult{ok: true, summary: strings.Join(parts, "; ")}
		}
	}
	return healthProbeResult{summary: strings.Join(parts, "; ")}
}
