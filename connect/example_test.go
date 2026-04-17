package connect

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"strings"
)

func Example() {
	upstreamLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer upstreamLn.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := upstreamLn.Accept()
		if err == nil {
			_, _ = io.Copy(io.Discard, conn)
			_ = conn.Close()
		}
	}()

	server := httptest.NewServer(New(Options{}))
	defer server.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, _ = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", upstreamLn.Addr().String(), upstreamLn.Addr().String())
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		panic(err)
	}

	fmt.Println(strings.TrimSpace(line))
	_ = conn.Close()
	<-done
	// Output:
	// HTTP/1.1 200 Connection Established
}
