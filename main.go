package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if isVersionRequest(os.Args[1:]) {
		fmt.Printf("udb-mysql-mcp-server %s (commit: %s, built: %s)\n", version, commit, date)
		return nil
	}

	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mcpServer, err := newMCPServer(cfg)
	if err != nil {
		return err
	}

	switch cfg.Transport {
	case transportStdio:
		return runStdio(ctx, mcpServer)
	case transportSSE:
		return runSSE(ctx, mcpServer, cfg)
	case transportStreamableHTTP:
		return runStreamableHTTP(ctx, mcpServer, cfg)
	default:
		return fmt.Errorf("unsupported transport %q", cfg.Transport)
	}
}
