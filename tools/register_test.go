package tools

import (
	"testing"

	"github.com/mark3labs/mcp-go/server"

	"udb-mysql-mcp-server/client"
)

func TestRegisterFiltersByMode(t *testing.T) {
	resolve := func() (client.CallContext, error) {
		return client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}, nil
	}
	cfg := Config{Resolve: resolve, Mode: ModeReadonly, Client: client.New(&client.Factory{})}
	s := server.NewMCPServer("test", "0.0.1", server.WithToolCapabilities(true))
	if err := Register(s, cfg); err != nil {
		t.Fatalf("register: %v", err)
	}

	admin := Config{Resolve: resolve, Mode: ModeAdmin, Client: client.New(&client.Factory{})}
	sAdmin := server.NewMCPServer("test", "0.0.1", server.WithToolCapabilities(true))
	if err := Register(sAdmin, admin); err != nil {
		t.Fatalf("register admin: %v", err)
	}
}

func TestRegisterRequiresClient(t *testing.T) {
	s := server.NewMCPServer("test", "0.0.1", server.WithToolCapabilities(true))
	err := Register(s, Config{
		Resolve: func() (client.CallContext, error) { return client.CallContext{}, nil },
		Mode:    ModeReadonly,
	})
	if err == nil {
		t.Fatal("expected error when client is nil")
	}
}
