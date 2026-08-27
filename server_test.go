package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

func TestRunSSEStopsOnCancel(t *testing.T) {
	addr := freeListenAddr(t)
	mcpServer := server.NewMCPServer(serverName, serverVersion, server.WithToolCapabilities(true))
	cfg := config{
		Transport:   transportSSE,
		Listen:      addr,
		SSEEndpoint: "/sse",
		MessagePath: "/message",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- runSSE(ctx, mcpServer, cfg)
	}()

	waitForListen(t, addr)

	resp, err := http.Get("http://" + addr + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /healthz status %d, want 404 (sse handler must be mounted)", resp.StatusCode)
	}

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("runSSE returned %v, want nil", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("runSSE did not return after cancel")
	}
}

func freeListenAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	return addr
}

func waitForListen(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server did not listen on %s", addr)
}
