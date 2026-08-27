package main

import (
	"testing"

	"udb-mysql-mcp-server/tools"
)

func TestParseConfigDefaults(t *testing.T) {
	t.Setenv("UCLOUD_MODE", "")
	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Transport != transportStdio || cfg.Mode != tools.ModeReadonly {
		t.Fatalf("got %+v", cfg)
	}
}

func TestParseConfigModeFromEnv(t *testing.T) {
	t.Setenv("UCLOUD_MODE", "admin")
	cfg, err := parseConfig([]string{"--transport", "sse", "--listen", "0.0.0.0:9000"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != tools.ModeAdmin {
		t.Fatalf("got %q", cfg.Mode)
	}
}

func TestParseConfigModeFlagOverridesEnv(t *testing.T) {
	t.Setenv("UCLOUD_MODE", "admin")
	cfg, err := parseConfig([]string{"--mode", "readonly"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != tools.ModeReadonly {
		t.Fatalf("got %q", cfg.Mode)
	}
}

func TestParseConfigSSE(t *testing.T) {
	cfg, err := parseConfig([]string{"--transport", "sse", "--listen", "127.0.0.1:9000"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Transport != transportSSE {
		t.Fatalf("got %q", cfg.Transport)
	}
}

func TestParseConfigRejectsUnknownTransport(t *testing.T) {
	_, err := parseConfig([]string{"--transport", "http"})
	if err == nil {
		t.Fatal("expected error")
	}
}
