package tools

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"udb-mysql-mcp-server/client"
)

// Resolver returns process identity for one cloud call.
type Resolver func() (client.CallContext, error)

// Config carries local-process dependencies for UDB tools.
type Config struct {
	Resolve Resolver
	Mode    Mode
	Client  *client.Client
	Now     func() time.Time
}

func (cfg Config) now() time.Time {
	if cfg.Now != nil {
		return cfg.Now()
	}
	return time.Now()
}

func (cfg Config) callContext(projectID, region string, requireProject bool) (client.CallContext, error) {
	if cfg.Resolve == nil {
		return client.CallContext{}, &client.CredentialsError{}
	}
	base, err := cfg.Resolve()
	if err != nil {
		return client.CallContext{}, err
	}
	return base.WithScope(projectID, region, requireProject)
}

func requireCloudClient(cfg Config) (*client.Client, error) {
	if cfg.Client == nil {
		return nil, fmt.Errorf("tools: client must not be nil")
	}
	if cfg.Resolve == nil {
		return nil, fmt.Errorf("tools: credential resolver must not be nil")
	}
	return cfg.Client, nil
}

// Binding pairs local metadata with an MCP tool handler.
type Binding struct {
	Catalog Tool
	MCP     mcp.Tool
	Handler server.ToolHandlerFunc
}

// ToolError is a structured MCP error payload without credentials.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	RetCode int    `json:"ret_code,omitempty"`
}

// InternalError indicates an unexpected local failure surfaced to MCP callers.
type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	if e.Message == "" {
		return "internal error"
	}
	return e.Message
}

func mapExecutionError(err error) ToolError {
	var cred *client.CredentialsError
	if errors.As(err, &cred) {
		return ToolError{
			Code:    "credentials_required",
			Message: cred.Error(),
		}
	}
	var conflict *client.ConflictError
	if errors.As(err, &conflict) {
		return ToolError{Code: "conflict", Message: conflict.Error()}
	}
	var internal *InternalError
	if errors.As(err, &internal) {
		return ToolError{Code: "internal_error", Message: "internal error"}
	}
	var backupNotFound *client.BackupNotFoundError
	if errors.As(err, &backupNotFound) {
		return ToolError{Code: "not_found", Message: backupNotFound.Error()}
	}
	var scanIncomplete *client.ScanIncompleteError
	if errors.As(err, &scanIncomplete) {
		return mapScanIncompleteError(scanIncomplete)
	}
	var notFound *client.NotFoundError
	if errors.As(err, &notFound) {
		return ToolError{Code: "not_found", Message: notFound.Error()}
	}
	var invalid *client.InvalidInputError
	if errors.As(err, &invalid) {
		return ToolError{Code: "invalid_input", Message: invalid.Error()}
	}
	var upstream *client.UpstreamError
	if errors.As(err, &upstream) {
		return ToolError{
			Code:    "upstream_error",
			Message: upstream.Error(),
			RetCode: upstream.RetCode,
		}
	}
	return ToolError{Code: "internal_error", Message: "internal error"}
}

func mapScanIncompleteError(err *client.ScanIncompleteError) ToolError {
	var message string
	switch err.Reason {
	case client.ScanIncompleteReasonLimit:
		message = "backup lookup could not complete within scan limits; narrow scope with zone filters or retry"
	case client.ScanIncompleteReasonCountChanged:
		message = "backup list changed during lookup; retry after listing backups again"
	case client.ScanIncompleteReasonEmptyPage:
		message = "backup lookup ended before the list was fully scanned; retry or narrow scope with zone filters"
	default:
		message = "backup lookup did not finish; retry or narrow scope with zone filters"
	}
	return ToolError{Code: "upstream_error", Message: message}
}

func mapToolErrorResult(err error) *mcp.CallToolResult {
	return toolResultError(mapExecutionError(err))
}

func toolResultError(toolErr ToolError) *mcp.CallToolResult {
	result := mcp.NewToolResultStructured(
		toolErr,
		fmt.Sprintf("%s: %s", toolErr.Code, toolErr.Message),
	)
	result.IsError = true
	return result
}

func invokeCloud[T any, R any](
	cfg Config,
	ctx context.Context,
	in T,
	projectID, region string,
	requireProject bool,
	execute func(context.Context, client.CallContext, T) (any, error),
) (R, error) {
	var zero R
	call, err := cfg.callContext(projectID, region, requireProject)
	if err != nil {
		return zero, err
	}
	result, err := execute(ctx, call, in)
	if err != nil {
		return zero, err
	}
	out, ok := result.(R)
	if !ok {
		return zero, &InternalError{Message: "unexpected handler result type"}
	}
	return out, nil
}
