package client

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	sdklog "github.com/ucloud/ucloud-sdk-go/ucloud/log"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

const defaultRequestTimeout = 30 * time.Second

// Factory builds per-request SDK clients from identity and scope.
type Factory struct {
	BaseURL string
	Timeout time.Duration
}

// NewFactory constructs a factory with default timeout.
// Optional UCLOUD_API_BASE_URL overrides the SDK default API host (for stubs / alternate envs).
func NewFactory() *Factory {
	f := &Factory{Timeout: defaultRequestTimeout}
	if base := strings.TrimSpace(os.Getenv("UCLOUD_API_BASE_URL")); base != "" {
		f.BaseURL = base
	}
	return f
}

// Client performs UCloud UDB API calls through per-request SDK clients.
type Client struct {
	factory *Factory
}

// New constructs a client backed by the given factory.
func New(factory *Factory) *Client {
	return &Client{factory: factory}
}

func (f *Factory) newUAccountClient(req CallContext) *uaccount.UAccountClient {
	cfg := ucloud.NewConfig()
	if f.BaseURL != "" {
		cfg.BaseUrl = f.BaseURL
	}
	timeout := f.Timeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}
	cfg.Timeout = timeout
	cfg.MaxRetries = 0

	cred := auth.NewCredential()
	cred.PublicKey = req.PublicKey
	cred.PrivateKey = req.PrivateKey
	return newUAccountSDKClient(&cfg, &cred)
}

func newUAccountSDKClient(cfg *ucloud.Config, cred *auth.Credential) *uaccount.UAccountClient {
	cfg.MaxRetries = 0
	client := uaccount.NewClient(cfg, cred)
	sdklog.SetOutput(os.Stderr)
	if logger := client.GetLogger(); logger != nil {
		logger.SetOutput(os.Stderr)
	}
	return client
}

func (c *Client) uaccountClient(reqCtx CallContext) (*uaccount.UAccountClient, error) {
	if c.factory == nil {
		return nil, &InvalidInputError{Field: "factory", Message: "must not be nil"}
	}
	return c.factory.newUAccountClient(reqCtx), nil
}

func (f *Factory) newUDBClient(req CallContext) *udb.UDBClient {
	cfg := ucloud.NewConfig()
	cfg.Region = req.Region
	cfg.ProjectId = req.ProjectID
	if f.BaseURL != "" {
		cfg.BaseUrl = f.BaseURL
	}
	timeout := f.Timeout
	if timeout == 0 {
		timeout = defaultRequestTimeout
	}
	cfg.Timeout = timeout
	cfg.MaxRetries = 0

	cred := auth.NewCredential()
	cred.PublicKey = req.PublicKey
	cred.PrivateKey = req.PrivateKey
	return newSDKClient(&cfg, &cred)
}

func newSDKClient(cfg *ucloud.Config, cred *auth.Credential) *udb.UDBClient {
	cfg.MaxRetries = 0
	client := udb.NewClient(cfg, cred)
	// ucloud-sdk-go defaults loggers to stdout; that breaks MCP stdio framing.
	sdklog.SetOutput(os.Stderr)
	if logger := client.GetLogger(); logger != nil {
		logger.SetOutput(os.Stderr)
	}
	return client
}

func prepareRequest(req request.Common, ctx context.Context, fallback time.Duration) {
	req.SetRetryable(false)
	req.WithTimeout(requestTimeout(ctx, fallback))
}

func requestTimeout(ctx context.Context, fallback time.Duration) time.Duration {
	if fallback <= 0 {
		fallback = defaultRequestTimeout
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return fallback
	}
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return time.Millisecond
	}
	if remaining < fallback {
		return remaining
	}
	return fallback
}
