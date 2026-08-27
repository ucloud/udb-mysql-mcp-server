package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

const (
	ToolListDBTypes          = "udb_mysql_list_db_types"
	ToolListMachineTypes     = "udb_mysql_list_machine_types"
	ToolListParamGroups      = "udb_mysql_list_param_groups"
	ToolDescribePrice        = "udb_mysql_describe_price"
	ToolDescribeUpgradePrice = "udb_mysql_describe_upgrade_price"

	mcpCurrencyCNY = "CNY"
)

// ListDBTypesBinding returns catalog metadata and an MCP handler for udb_mysql_list_db_types.
func ListDBTypesBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListDBTypes,
		Description: "获取UDB支持的类型信息",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListDBTypes,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListDBTypesInput](),
		mcp.WithOutputSchema[types.ListDBTypesOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListDBTypesInput) (any, error) {
		return cloudClient.ListDBTypes(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListDBTypesInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ListDBTypesInput, types.ListDBTypesOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatDBTypesText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// ListMachineTypesBinding returns catalog metadata and an MCP handler for udb_mysql_list_machine_types.
func ListMachineTypesBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListMachineTypes,
		Description: "获取UDB云数据库支持的计算规格列表，暂不支持获取跨可用区实例的计算规格，目前支持的数据库品类包括：NVMe版和SSD云盘版MySQL",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListMachineTypes,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListMachineTypesInput](),
		mcp.WithOutputSchema[types.ListMachineTypesOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListMachineTypesInput) (any, error) {
		return cloudClient.ListMachineTypes(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListMachineTypesInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ListMachineTypesInput, types.ListMachineTypesOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatMachineTypesText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// ListParamGroupsBinding returns catalog metadata and an MCP handler for udb_mysql_list_param_groups.
func ListParamGroupsBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListParamGroups,
		Description: "获取参数组详细参数信息",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListParamGroups,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListParamGroupsInput](),
		mcp.WithOutputSchema[types.ListParamGroupsOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListParamGroupsInput) (any, error) {
		return cloudClient.ListParamGroups(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListParamGroupsInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ListParamGroupsInput, types.ListParamGroupsOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatParamGroupsText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// DescribePriceBinding returns catalog metadata and an MCP handler for udb_mysql_describe_price.
func DescribePriceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name: ToolDescribePrice,
		Description: "获取UDB MySQL 创建询价。" +
			"推荐先调 udb_mysql_list_machine_types，将同条规格的 id→machine_type、storage_class、specification_class 原样传入，并传 memory_limit_mb（按 1000 进制，如 8GB=8000）与 disk_space_gb。" +
			"instance_mode：普通版传 Normal（默认，可不传）、高可用版传 HA、从库传 Slave；不传时按 Normal。" +
			"storage_class 与 specification_class 必填，缺省会导致上游报错或返回不可信价格。",
		Risk: RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolDescribePrice,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.DescribePriceInput](),
		mcp.WithOutputSchema[types.DescribePriceOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.DescribePriceInput) (any, error) {
		return cloudClient.DescribePrice(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.DescribePriceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.DescribePriceInput, types.DescribePriceOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatDescribePriceText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// DescribeUpgradePriceBinding returns catalog metadata and an MCP handler for udb_mysql_describe_upgrade_price.
func DescribeUpgradePriceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolDescribeUpgradePrice,
		Description: "获取UDB实例升降级价格信息",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolDescribeUpgradePrice,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.DescribeUpgradePriceInput](),
		mcp.WithOutputSchema[types.DescribeUpgradePriceOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.DescribeUpgradePriceInput) (any, error) {
		out, err := cloudClient.DescribeUpgradePrice(ctx, reqCtx, in)
		if err != nil {
			return types.DescribeUpgradePriceOutput{}, err
		}
		out.Currency = mcpCurrencyCNY
		return out, nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.DescribeUpgradePriceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.DescribeUpgradePriceInput, types.DescribeUpgradePriceOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatDescribeUpgradePriceText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

func formatDBTypesText(out types.ListDBTypesOutput) string {
	return fmt.Sprintf("UDB DB types: %d entries", len(out.Items))
}

func formatMachineTypesText(out types.ListMachineTypesOutput) string {
	return fmt.Sprintf("UDB machine types: %d entries", len(out.Items))
}

func formatParamGroupsText(out types.ListParamGroupsOutput) string {
	return fmt.Sprintf("UDB param groups page %d size %d: returned %d (api_total_count=%d)", out.Page, out.PageSize, out.ReturnedCount, out.APITotalCount)
}

func formatDescribePriceText(out types.DescribePriceOutput) string {
	return fmt.Sprintf("UDB create price: %d charge options", len(out.Items))
}

func formatDescribeUpgradePriceText(out types.DescribeUpgradePriceOutput) string {
	return fmt.Sprintf("UDB upgrade price: %d fen (分)", out.PriceCents)
}
