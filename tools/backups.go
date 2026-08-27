package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/guard"
	"udb-mysql-mcp-server/types"
)

const (
	ToolListBackups       = "udb_mysql_list_backups"
	ToolGetBackupState    = "udb_mysql_get_backup_state"
	ToolGetBackupURL      = "udb_mysql_get_backup_url"
	ToolGetBackupStrategy = "udb_mysql_get_backup_strategy"
	ToolCreateBackup      = "udb_mysql_create_backup"
	ToolDeleteBackup      = "udb_mysql_delete_backup"
)

// ListBackupsBinding returns catalog metadata and an MCP handler for udb_mysql_list_backups.
func ListBackupsBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolListBackups,
		Description: "列表UDB实例备份信息.Zone不填表示多可用区，填代表单可用区",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListBackups,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListBackupsInput](),
		mcp.WithOutputSchema[types.ListBackupsOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListBackupsInput) (any, error) {
		return cloudClient.ListBackups(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListBackupsInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ListBackupsInput, types.ListBackupsOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatListBackupsText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// GetBackupStateBinding returns catalog metadata and an MCP handler for udb_mysql_get_backup_state.
func GetBackupStateBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolGetBackupState,
		Description: "获取UDB实例备份状态",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolGetBackupState,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.GetBackupStateInput](),
		mcp.WithOutputSchema[types.GetBackupStateOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.GetBackupStateInput) (any, error) {
		return cloudClient.GetBackupState(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.GetBackupStateInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.GetBackupStateInput, types.GetBackupStateOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatBackupStateText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// GetBackupURLBinding returns catalog metadata and an MCP handler for udb_mysql_get_backup_url.
func GetBackupURLBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolGetBackupURL,
		Description: "获取UDB备份下载地址",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolGetBackupURL,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.GetBackupURLInput](),
		mcp.WithOutputSchema[types.GetBackupURLOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.GetBackupURLInput) (any, error) {
		return cloudClient.GetBackupURL(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.GetBackupURLInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.GetBackupURLInput, types.GetBackupURLOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatBackupURLText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// GetBackupStrategyBinding returns catalog metadata and an MCP handler for udb_mysql_get_backup_strategy.
func GetBackupStrategyBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolGetBackupStrategy,
		Description: "获取实例备份策略",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolGetBackupStrategy,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.GetBackupStrategyInput](),
		mcp.WithOutputSchema[types.GetBackupStrategyOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.GetBackupStrategyInput) (any, error) {
		return cloudClient.GetBackupStrategy(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.GetBackupStrategyInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.GetBackupStrategyInput, types.GetBackupStrategyOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatBackupStrategyText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// CreateBackupBinding returns catalog metadata and an MCP handler for udb_mysql_create_backup.
func CreateBackupBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}
	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	catalogTool := Tool{
		Name:        ToolCreateBackup,
		Description: "发起实例备份。",
		Risk:        RiskWriteLow,
	}
	mcpTool := mcp.NewTool(
		ToolCreateBackup,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.CreateBackupInput](),
		mcp.WithOutputSchema[types.CreateBackupOutput](),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.CreateBackupInput) (any, error) {
		return cloudClient.CreateBackup(ctx, reqCtx, in, nowFn())
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.CreateBackupInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.CreateBackupInput, types.CreateBackupOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatCreateBackupText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// DeleteBackupBinding returns catalog metadata and an MCP handler for udb_mysql_delete_backup.
func DeleteBackupBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolDeleteBackup,
		Description: "删除UDB实例备份",
		Risk:        RiskCritical,
	}
	mcpTool := mcp.NewTool(
		ToolDeleteBackup,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.DeleteBackupInput](),
		mcp.WithOutputSchema[types.DeleteBackupOutput](),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.DeleteBackupInput) (any, error) {
		current, err := guard.ConfirmBackupDelete(ctx, reqCtx, cloudClient, in)
		if err != nil {
			return types.DeleteBackupOutput{}, err
		}
		if err := cloudClient.DeleteBackup(ctx, reqCtx, in.BackupID, in.Zone, in.BackupZone); err != nil {
			return types.DeleteBackupOutput{}, err
		}
		return types.DeleteBackupOutput{BackupID: in.BackupID, BackupName: current.BackupName}, nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.DeleteBackupInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.DeleteBackupInput, types.DeleteBackupOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, fmt.Sprintf("UDB backup %d deleted", out.BackupID)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

func formatListBackupsText(out types.ListBackupsOutput) string {
	return fmt.Sprintf("UDB backups page %d size %d: returned %d (api_total_count=%d)", out.Page, out.PageSize, out.ReturnedCount, out.APITotalCount)
}

func formatBackupStateText(out types.GetBackupStateOutput) string {
	return fmt.Sprintf("UDB backup %d state=%s hint=%s", out.BackupID, out.State, out.Hint)
}

func formatBackupURLText(out types.GetBackupURLOutput) string {
	return fmt.Sprintf("UDB backup %d for %s: download URLs returned (not logged in audit)", out.BackupID, out.DBID)
}

func formatBackupStrategyText(out types.GetBackupStrategyOutput) string {
	return fmt.Sprintf("UDB backup strategy for %s: method=%s save_days=%d", out.DBID, out.BackupMethod, out.SaveDays)
}

func formatCreateBackupText(out types.CreateBackupOutput) string {
	if out.BackupIDKnown {
		return fmt.Sprintf("UDB backup started for %s backup_id=%d follow_up=%s", out.DBID, out.BackupID, out.FollowUpTool)
	}
	return fmt.Sprintf("UDB backup started for %s backup_name=%s requested_at=%d follow_up=%s", out.DBID, out.BackupName, out.RequestedAtUnix, out.FollowUpTool)
}
