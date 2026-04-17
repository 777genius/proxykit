package proxyruntime

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestSocks5_HandshakeAndConnect(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "pong") }))
	defer upstream.Close()
	_, upPortStr, _ := net.SplitHostPort(upstream.Listener.Addr().String())

	m := New(nil)
	if err := m.StartSocks("127.0.0.1:0", "none", "", ""); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = m.StopSocks(nil) }()
	addr := m.SocksAddr()
	if addr == "" {
		t.Fatalf("no socks addr")
	}

	c, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close()
	br := bufio.NewReader(c)

	_, _ = c.Write([]byte{0x05, 0x01, 0x00})
	v, _ := br.ReadByte()
	mth, _ := br.ReadByte()
	if v != 0x05 || mth != 0x00 {
		t.Fatalf("bad greet resp: %x %x", v, mth)
	}

	port, _ := strconv.Atoi(upPortStr)
	req := []byte{0x05, 0x01, 0x00, 0x01, 127, 0, 0, 1, byte(port >> 8), byte(port & 0xff)}
	_, _ = c.Write(req)
	b := make([]byte, 10)
	if _, err := io.ReadFull(br, b); err != nil {
		t.Fatalf("read resp: %v", err)
	}
	if b[0] != 0x05 || b[1] != 0x00 {
		t.Fatalf("bad resp: %x %x", b[0], b[1])
	}

	_, _ = c.Write([]byte("GET / HTTP/1.1\r\nHost: 127.0.0.1\r\nConnection: close\r\n\r\n"))
	resp, _ := io.ReadAll(br)
	if len(resp) == 0 {
		t.Fatalf("empty response")
	}
}
