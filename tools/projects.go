package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

const ToolListProjects = "udb_mysql_list_projects"

// ListProjectsBinding returns catalog metadata and an MCP handler for udb_mysql_list_projects.
func ListProjectsBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListProjects,
		Description: "获取项目 ID",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListProjects,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListProjectsInput](),
		mcp.WithOutputSchema[types.ListProjectsOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedAccountCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListProjectsInput) (any, error) {
		return cloudClient.ListProjects(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListProjectsInput) (*mcp.CallToolResult, error) {
		out, err := invokeTypedAccountCloudTool[types.ListProjectsInput, types.ListProjectsOutput](binding, ctx, req, args)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatListProjectsText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

func formatListProjectsText(out types.ListProjectsOutput) string {
	if out.ReturnedCount == 0 {
		return fmt.Sprintf("UCloud projects: 0 matches (account total=%d)", out.TotalCount)
	}
	names := make([]string, 0, out.ReturnedCount)
	for _, project := range out.Projects {
		names = append(names, fmt.Sprintf("%s(%s)", project.ProjectName, project.ProjectID))
	}
	return fmt.Sprintf(
		"UCloud projects: %d match(es) of %d total: %s",
		out.ReturnedCount,
		out.TotalCount,
		strings.Join(names, ", "),
	)
}
