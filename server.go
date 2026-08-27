package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/tools"
)

const (
	serverName    = "udb-mysql-mcp-server"
	serverVersion = "0.2.0"
)

func newMCPServer(cfg config) (*server.MCPServer, error) {
	mcpServer := server.NewMCPServer(
		serverName,
		serverVersion,
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)
	toolCfg := tools.Config{
		Resolve: func() (client.CallContext, error) {
			return client.ResolveFromEnv(cfg.DefaultRegion)
		},
		Mode:   cfg.Mode,
		Client: client.New(client.NewFactory()),
	}
	if err := tools.Register(mcpServer, toolCfg); err != nil {
		return nil, err
	}
	return mcpServer, nil
}

func runStdio(ctx context.Context, mcpServer *server.MCPServer) error {
	stdio := server.NewStdioServer(mcpServer)
	return stdio.Listen(ctx, os.Stdin, os.Stdout)
}

func runSSE(ctx context.Context, mcpServer *server.MCPServer, cfg config) error {
	netServer := &http.Server{
		Addr:              cfg.Listen,
		ReadHeaderTimeout: 5 * time.Second,
	}
	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithSSEEndpoint(cfg.SSEEndpoint),
		server.WithMessageEndpoint(cfg.MessagePath),
		server.WithAppendQueryToMessageEndpoint(),
		server.WithHTTPServer(netServer),
	)
	netServer.Handler = sseServer
	outputLocalConfig(cfg, fmt.Sprintf("http://%s%s", displayHost(cfg.Listen), cfg.SSEEndpoint), "sse")
	return serveUntilCancel(ctx, netServer.ListenAndServe, sseServer.Shutdown)
}

func runStreamableHTTP(ctx context.Context, mcpServer *server.MCPServer, cfg config) error {
	netServer := &http.Server{
		Addr:              cfg.Listen,
		ReadHeaderTimeout: 5 * time.Second,
	}
	streamable := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithStreamableHTTPServer(netServer),
		server.WithStateLess(true),
		server.WithEndpointPath(cfg.HTTPEndpoint),
	)
	mux := http.NewServeMux()
	mux.Handle(cfg.HTTPEndpoint, streamable)
	netServer.Handler = mux

	outputLocalConfig(cfg, fmt.Sprintf("http://%s%s", displayHost(cfg.Listen), cfg.HTTPEndpoint), "streamable-http")
	return serveUntilCancel(ctx, netServer.ListenAndServe, streamable.Shutdown)
}

func serveUntilCancel(ctx context.Context, start func() error, shutdown func(context.Context) error) error {
	errCh := make(chan error, 1)
	go func() { errCh <- start() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = shutdown(shutdownCtx)
		return nil
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func displayHost(listen string) string {
	host, port, ok := splitHostPort(listen)
	if !ok {
		return listen
	}
	if host == "0.0.0.0" || host == "" {
		host = "127.0.0.1"
	}
	return host + ":" + port
}

func splitHostPort(listen string) (string, string, bool) {
	for i := len(listen) - 1; i >= 0; i-- {
		if listen[i] == ':' {
			return listen[:i], listen[i+1:], true
		}
	}
	return "", "", false
}

func outputLocalConfig(cfg config, url, kind string) {
	configJSON := map[string]any{
		"mcpServers": map[string]any{
			serverName: map[string]any{
				"url":  url,
				"type": kind,
			},
		},
	}
	raw, _ := json.MarshalIndent(configJSON, "", "  ")
	log.Printf("listening on %s (%s mode=%s)", cfg.Listen, cfg.Transport, cfg.Mode)
	fmt.Println("=== MCP Server Configuration ===")
	fmt.Println(string(raw))
}
