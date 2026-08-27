package tools

import (
	"context"
	"strings"
	"unicode"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

const invalidInputBindMessage = "invalid tool arguments"

type typedCloudBinding[T any] struct {
	cfg            Config
	tool           Tool
	requireProject bool
	execute        func(context.Context, client.CallContext, T) (any, error)
}

// NewTypedCloudBinding constructs a cloud tool that requires project + region after env overlay.
func NewTypedCloudBinding[T any](cfg Config, tool Tool, execute func(context.Context, client.CallContext, T) (any, error)) (*typedCloudBinding[T], error) {
	if _, err := requireCloudClient(cfg); err != nil {
		return nil, err
	}
	return &typedCloudBinding[T]{cfg: cfg, tool: tool, requireProject: false, execute: execute}, nil
}

func invokeTypedCloudTool[T any, R any](binding *typedCloudBinding[T], ctx context.Context, _ mcp.CallToolRequest, in T, projectID, region string) (R, error) {
	return invokeCloud[T, R](binding.cfg, ctx, in, projectID, region, binding.requireProject, binding.execute)
}

// scopeFromInput returns project/region overlays for a cloud call.
// When region is omitted but zone is present, region is derived from zone
// (UCloud form {region}-{az}, e.g. cn-bj2-04 → cn-bj2). Tool schemas keep
// advertising region as the explicit field; this is internal compatibility only.
func scopeFromInput(in types.ScopeInput, zone string) (projectID, region string) {
	region = strings.TrimSpace(in.Region)
	if region == "" {
		region = regionFromZone(zone)
	}
	return in.ProjectID, region
}

// regionFromZone extracts region from a UCloud zone whose last segment is a
// numeric AZ suffix. Returns empty when the zone cannot be parsed.
func regionFromZone(zone string) string {
	zone = strings.TrimSpace(zone)
	i := strings.LastIndex(zone, "-")
	if i <= 0 {
		return ""
	}
	suffix := zone[i+1:]
	if suffix == "" {
		return ""
	}
	for _, r := range suffix {
		if !unicode.IsDigit(r) {
			return ""
		}
	}
	return zone[:i]
}

type typedAccountCloudBinding[T any] struct {
	cfg     Config
	tool    Tool
	execute func(context.Context, client.CallContext, T) (any, error)
}

// NewTypedAccountCloudBinding constructs an account-scoped cloud tool that requires signing keys only.
func NewTypedAccountCloudBinding[T any](cfg Config, tool Tool, execute func(context.Context, client.CallContext, T) (any, error)) (*typedAccountCloudBinding[T], error) {
	if _, err := requireCloudClient(cfg); err != nil {
		return nil, err
	}
	return &typedAccountCloudBinding[T]{cfg: cfg, tool: tool, execute: execute}, nil
}

func invokeTypedAccountCloudTool[T any, R any](binding *typedAccountCloudBinding[T], ctx context.Context, _ mcp.CallToolRequest, in T) (R, error) {
	return invokeCloud[T, R](binding.cfg, ctx, in, "", "", false, binding.execute)
}

// NewTypedToolHandler binds MCP arguments before invoking the handler.
func NewTypedToolHandler[T any](handler func(context.Context, mcp.CallToolRequest, T) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args T
		if err := request.BindArguments(&args); err != nil {
			return toolResultError(ToolError{Code: "invalid_input", Message: invalidInputBindMessage}), nil
		}
		return handler(ctx, request, args)
	}
}
