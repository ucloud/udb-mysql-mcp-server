package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

const ToolListRegions = "udb_mysql_list_regions"

// ListRegionsBinding returns catalog metadata and an MCP handler for udb_mysql_list_regions.
func ListRegionsBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListRegions,
		Description: "获取地域和可用区列表。UCloud 目前拥有多个地域（Region），遍布全球，每个地域下开设有多个可用区（Zone），用户在使用公有云相关的 API 发送操作指令时，需要指定指令所指向的地域。",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListRegions,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListRegionsInput](),
		mcp.WithOutputSchema[types.ListRegionsOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedAccountCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListRegionsInput) (any, error) {
		return cloudClient.ListRegions(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListRegionsInput) (*mcp.CallToolResult, error) {
		out, err := invokeTypedAccountCloudTool[types.ListRegionsInput, types.ListRegionsOutput](binding, ctx, req, args)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatListRegionsText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

func formatListRegionsText(out types.ListRegionsOutput) string {
	if out.ReturnedCount == 0 {
		return fmt.Sprintf("UCloud regions: 0 matches (account total=%d)", out.TotalCount)
	}
	parts := make([]string, 0, out.ReturnedCount)
	for _, region := range out.Regions {
		zones := make([]string, 0, len(region.Zones))
		for _, zone := range region.Zones {
			zones = append(zones, zone.Zone)
		}
		parts = append(parts, fmt.Sprintf("%s(%s)[%s]", region.RegionName, region.Region, strings.Join(zones, ",")))
	}
	return fmt.Sprintf(
		"UCloud regions: %d match(es) of %d total: %s",
		out.ReturnedCount,
		out.TotalCount,
		strings.Join(parts, ", "),
	)
}
