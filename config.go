package main

import (
	"fmt"
	"os"
	"strings"

	"udb-mysql-mcp-server/tools"
)

type transport string

const (
	transportStdio          transport = "stdio"
	transportSSE            transport = "sse"
	transportStreamableHTTP transport = "streamable_http"
)

type config struct {
	Transport     transport
	Mode          tools.Mode
	DefaultRegion string
	Listen        string
	SSEEndpoint   string
	MessagePath   string
	HTTPEndpoint  string
}

func defaultConfig() config {
	return config{
		Transport:     transportStdio,
		Mode:          tools.ModeReadonly,
		DefaultRegion: "cn-bj2",
		Listen:        "127.0.0.1:9000",
		SSEEndpoint:   "/sse",
		MessagePath:   "/message",
		HTTPEndpoint:  "/mcp",
	}
}

func parseConfig(args []string) (config, error) {
	cfg := defaultConfig()
	if v := strings.TrimSpace(os.Getenv("UCLOUD_MCP_TRANSPORT")); v != "" {
		cfg.Transport = transport(v)
	}
	if v := strings.TrimSpace(os.Getenv("UCLOUD_MODE")); v != "" {
		cfg.Mode = tools.Mode(v)
	}
	if v := strings.TrimSpace(os.Getenv("UCLOUD_DEFAULT_REGION")); v != "" {
		cfg.DefaultRegion = v
	}
	if v, ok := os.LookupEnv("UCLOUD_MCP_LISTEN"); ok {
		cfg.Listen = strings.TrimSpace(v)
	}
	if v := strings.TrimSpace(os.Getenv("UCLOUD_MCP_ENDPOINT_PATH")); v != "" {
		cfg.HTTPEndpoint = v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_SERVER_SSE_ENDPOINT")); v != "" {
		cfg.SSEEndpoint = v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_SERVER_MESSAGE_ENDPOINT")); v != "" {
		cfg.MessagePath = v
	}
	if v := strings.TrimSpace(os.Getenv("MCP_SERVER_SSE_PORT")); v != "" && cfg.Listen == defaultConfig().Listen {
		cfg.Listen = "127.0.0.1:" + v
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		need := func(field string) (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing value for %s", field)
			}
			i++
			return args[i], nil
		}
		switch arg {
		case "--transport":
			v, err := need("transport")
			if err != nil {
				return config{}, err
			}
			cfg.Transport = transport(v)
		case "--mode":
			v, err := need("mode")
			if err != nil {
				return config{}, err
			}
			cfg.Mode = tools.Mode(v)
		case "--default-region":
			v, err := need("default-region")
			if err != nil {
				return config{}, err
			}
			cfg.DefaultRegion = v
		case "--listen":
			v, err := need("listen")
			if err != nil {
				return config{}, err
			}
			cfg.Listen = v
		case "--endpoint-path":
			v, err := need("endpoint-path")
			if err != nil {
				return config{}, err
			}
			cfg.HTTPEndpoint = v
		default:
			return config{}, fmt.Errorf("unknown argument %q", arg)
		}
	}

	switch cfg.Transport {
	case transportStdio:
	case transportSSE, transportStreamableHTTP:
		if strings.TrimSpace(cfg.Listen) == "" {
			return config{}, fmt.Errorf("listen address must not be empty when transport is %s", cfg.Transport)
		}
	default:
		return config{}, fmt.Errorf("unsupported transport %q", cfg.Transport)
	}
	switch cfg.Mode {
	case tools.ModeReadonly, tools.ModeReadWrite, tools.ModeAdmin:
	default:
		return config{}, fmt.Errorf("unsupported mode %q", cfg.Mode)
	}
	if strings.TrimSpace(cfg.DefaultRegion) == "" {
		return config{}, fmt.Errorf("default region must not be empty")
	}
	return cfg, nil
}

func isVersionRequest(args []string) bool {
	for _, a := range args {
		if a == "--version" || a == "-v" {
			return true
		}
	}
	return false
}
