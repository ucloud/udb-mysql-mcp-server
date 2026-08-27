package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

const (
	ToolGetInstance      = "udb_mysql_get_instance"
	ToolListInstances    = "udb_mysql_list_instances"
	ToolGetInstanceState = "udb_mysql_get_instance_state"
	ToolStartInstance    = "udb_mysql_start_instance"
	ToolStopInstance     = "udb_mysql_stop_instance"
	ToolRestartInstance  = "udb_mysql_restart_instance"
	ToolModifyName       = "udb_mysql_modify_name"
	ToolCreateInstance   = "udb_mysql_create_instance"
	ToolResizeInstance   = "udb_mysql_resize_instance"
	ToolResetPassword    = "udb_mysql_reset_password"

	followUpInstanceStateTool = "udb_mysql_get_instance_state"
	genericLifecycleHint      = "Poll udb_mysql_get_instance_state until the instance reaches a stable state; stop on Fail or Recover fail."
)

// GetInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_get_instance.
func GetInstanceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolGetInstance,
		Description: "获取UDB实例信息，支持两类操作：（1）指定DBId用于获取该db的信息；（2）指定ClassType、Offset、Limit用于列表操作，查询某一个类型db。",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolGetInstance,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.GetInstanceInput](),
		mcp.WithOutputSchema[types.InstanceOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.GetInstanceInput) (any, error) {
		return cloudClient.GetInstance(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.GetInstanceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, "")
		out, err := invokeTypedCloudTool[types.GetInstanceInput, types.InstanceOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatInstanceText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// ListInstancesBinding returns catalog metadata and an MCP handler for udb_mysql_list_instances.
func ListInstancesBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name: ToolListInstances,
		Description: "列出当前项目在单一地域下的 UDB MySQL 实例。" +
			"一次调用只覆盖一个 region：显式传入 region，或省略后回落到 UCLOUD_REGION/进程默认（通常 cn-bj2）；省略不等于全部地域。" +
			"跨地域请先 udb_mysql_list_regions，再按地域多次调用。返回的 region/project_id 为本次实际生效范围。",
		Risk: RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolListInstances,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ListInstancesInput](),
		mcp.WithOutputSchema[types.ListInstancesOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ListInstancesInput) (any, error) {
		return cloudClient.ListInstances(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ListInstancesInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ListInstancesInput, types.ListInstancesOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatListInstancesText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// GetInstanceStateBinding returns catalog metadata and an MCP handler for udb_mysql_get_instance_state.
func GetInstanceStateBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolGetInstanceState,
		Description: "获取UDB实例状态",
		Risk:        RiskRead,
	}
	mcpTool := mcp.NewTool(
		ToolGetInstanceState,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.GetInstanceStateInput](),
		mcp.WithOutputSchema[types.GetInstanceStateOutput](),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.GetInstanceStateInput) (any, error) {
		return cloudClient.GetInstanceState(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.GetInstanceStateInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, "")
		out, err := invokeTypedCloudTool[types.GetInstanceStateInput, types.GetInstanceStateOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatInstanceStateText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// StartInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_start_instance.
func StartInstanceBinding(cfg Config) (Binding, error) {
	return lifecycleBinding(cfg, lifecycleBindingSpec{
		name:        ToolStartInstance,
		description: "启动UDB实例",
		risk:        RiskWriteMid,
		idempotent:  true,
		run: func(ctx context.Context, reqCtx client.CallContext, api *client.Client, in types.LifecycleInput) error {
			return api.StartInstance(ctx, reqCtx, in)
		},
	})
}

// StopInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_stop_instance.
func StopInstanceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolStopInstance,
		Description: "关闭UDB实例",
		Risk:        RiskWriteMid,
	}
	mcpTool := mcp.NewTool(
		ToolStopInstance,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.StopInstanceInput](),
		mcp.WithOutputSchema[types.LifecycleMutationOutput](),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.StopInstanceInput) (any, error) {
		if err := cloudClient.StopInstance(ctx, reqCtx, in); err != nil {
			return types.LifecycleMutationOutput{}, err
		}
		return lifecycleOutput(in.DBID), nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.StopInstanceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.StopInstanceInput, types.LifecycleMutationOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatLifecycleText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// RestartInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_restart_instance.
func RestartInstanceBinding(cfg Config) (Binding, error) {
	return lifecycleBinding(cfg, lifecycleBindingSpec{
		name:        ToolRestartInstance,
		description: "重启UDB实例",
		risk:        RiskWriteMid,
		idempotent:  false,
		run: func(ctx context.Context, reqCtx client.CallContext, api *client.Client, in types.LifecycleInput) error {
			return api.RestartInstance(ctx, reqCtx, in)
		},
	})
}

// ModifyNameBinding returns catalog metadata and an MCP handler for udb_mysql_modify_name.
func ModifyNameBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolModifyName,
		Description: "重命名UDB实例",
		Risk:        RiskWriteLow,
	}
	mcpTool := mcp.NewTool(
		ToolModifyName,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ModifyNameInput](),
		mcp.WithOutputSchema[types.ModifyNameOutput](),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ModifyNameInput) (any, error) {
		if err := cloudClient.ModifyName(ctx, reqCtx, in); err != nil {
			return types.ModifyNameOutput{}, err
		}
		return types.ModifyNameOutput{DBID: in.DBID, NewName: strings.TrimSpace(in.Name)}, nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ModifyNameInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ModifyNameInput, types.ModifyNameOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, fmt.Sprintf("UDB instance %s renamed to %s", out.DBID, out.NewName)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// CreateInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_create_instance.
func CreateInstanceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolCreateInstance,
		Description: "创建UDB实例（包括创建mysql NVMe、共享型和O2实例以及从备份恢复实例）",
		Risk:        RiskWriteHigh,
	}
	mcpTool := mcp.NewTool(
		ToolCreateInstance,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.CreateInstanceInput](),
		mcp.WithOutputSchema[types.CreateInstanceOutput](),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.CreateInstanceInput) (any, error) {
		return cloudClient.CreateInstance(ctx, reqCtx, in)
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.CreateInstanceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.CreateInstanceInput, types.CreateInstanceOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, fmt.Sprintf("UDB instance %s created follow_up=%s", out.DBID, out.FollowUpTool)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// ResizeInstanceBinding returns catalog metadata and an MCP handler for udb_mysql_resize_instance.
func ResizeInstanceBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name: ToolResizeInstance,
		Description: "修改（升级和降级）UDB实例的配置，包括内存和磁盘的配置，对于内存升级无需关闭实例，其他场景需要事先关闭实例。两套参数可以配置升降机：1.配置UseSSD和SSDType  2.配置InstanceType，不需要配置InstanceMode。这两套第二套参数的优先级更高。" +
			"注意：同时传 machine_type 与 memory_limit_mb 时，实际生效规格以 machine_type 为准（其内存/CPU 覆盖 memory_limit_mb），请以 udb_mysql_list_machine_types 返回的规格为准，避免两者不一致",
		Risk: RiskWriteHigh,
	}
	mcpTool := mcp.NewTool(
		ToolResizeInstance,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ResizeInstanceInput](),
		mcp.WithOutputSchema[types.ResizeInstanceOutput](),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ResizeInstanceInput) (any, error) {
		if err := cloudClient.ResizeInstance(ctx, reqCtx, in); err != nil {
			return types.ResizeInstanceOutput{}, err
		}
		return types.ResizeInstanceOutput{DBID: in.DBID, FollowUpTool: followUpInstanceStateTool, Hint: genericLifecycleHint}, nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ResizeInstanceInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.ResizeInstanceInput, types.ResizeInstanceOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, fmt.Sprintf("UDB instance %s resize accepted follow_up=%s", out.DBID, out.FollowUpTool)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

// ResetPasswordBinding returns catalog metadata and an MCP handler for udb_mysql_reset_password.
func ResetPasswordBinding(cfg Config) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{
		Name:        ToolResetPassword,
		Description: "修改DB实例的管理员密码",
		Risk:        RiskWriteHigh,
	}
	mcpTool := mcp.NewTool(
		ToolResetPassword,
		mcp.WithDescription(catalogTool.Description),
		mcp.WithInputSchema[types.ResetPasswordInput](),
		mcp.WithOutputSchema[types.ResetPasswordOutput](),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
	)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.ResetPasswordInput) (any, error) {
		if err := cloudClient.ResetPassword(ctx, reqCtx, in); err != nil {
			return types.ResetPasswordOutput{}, err
		}
		return types.ResetPasswordOutput{DBID: in.DBID}, nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.ResetPasswordInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, "")
		out, err := invokeTypedCloudTool[types.ResetPasswordInput, types.ResetPasswordOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, fmt.Sprintf("UDB instance %s password reset accepted", out.DBID)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

type lifecycleBindingSpec struct {
	name        string
	description string
	risk        RiskLevel
	idempotent  bool
	run         func(context.Context, client.CallContext, *client.Client, types.LifecycleInput) error
}

func lifecycleBinding(cfg Config, spec lifecycleBindingSpec) (Binding, error) {
	cloudClient, err := requireCloudClient(cfg)
	if err != nil {
		return Binding{}, err
	}

	catalogTool := Tool{Name: spec.name, Description: spec.description, Risk: spec.risk}
	opts := []mcp.ToolOption{
		mcp.WithDescription(spec.description),
		mcp.WithInputSchema[types.LifecycleInput](),
		mcp.WithOutputSchema[types.LifecycleMutationOutput](),
		mcp.WithDestructiveHintAnnotation(false),
	}
	if spec.idempotent {
		opts = append(opts, mcp.WithIdempotentHintAnnotation(true))
	} else {
		opts = append(opts, mcp.WithIdempotentHintAnnotation(false))
	}
	mcpTool := mcp.NewTool(spec.name, opts...)

	binding, err := NewTypedCloudBinding(cfg, catalogTool, func(ctx context.Context, reqCtx client.CallContext, in types.LifecycleInput) (any, error) {
		if err := spec.run(ctx, reqCtx, cloudClient, in); err != nil {
			return types.LifecycleMutationOutput{}, err
		}
		return lifecycleOutput(in.DBID), nil
	})
	if err != nil {
		return Binding{}, err
	}

	handler := NewTypedToolHandler(func(ctx context.Context, req mcp.CallToolRequest, args types.LifecycleInput) (*mcp.CallToolResult, error) {
		projectID, region := scopeFromInput(args.ScopeInput, args.Zone)
		out, err := invokeTypedCloudTool[types.LifecycleInput, types.LifecycleMutationOutput](binding, ctx, req, args, projectID, region)
		if err != nil {
			return mapToolErrorResult(err), nil
		}
		return mcp.NewToolResultStructured(out, formatLifecycleText(out)), nil
	})

	return Binding{Catalog: catalogTool, MCP: mcpTool, Handler: handler}, nil
}

func lifecycleOutput(dbID string) types.LifecycleMutationOutput {
	return types.LifecycleMutationOutput{
		DBID:         dbID,
		FollowUpTool: followUpInstanceStateTool,
		Hint:         genericLifecycleHint,
	}
}

func formatInstanceText(out types.InstanceOutput) string {
	return fmt.Sprintf(
		"UDB instance %s (%s): state=%s type=%s zone=%s address=%s:%d cpu=%d memory_mb=%d disk_gb=%d created_at=%s",
		out.DBID, out.Name, out.State, out.DBTypeID, out.Zone, out.Address, out.Port, out.CPU, out.MemoryMB, out.DiskGB, out.CreatedAt,
	)
}

func formatListInstancesText(out types.ListInstancesOutput) string {
	scope := fmt.Sprintf("region=%s", out.Region)
	if out.ProjectID != "" {
		scope += fmt.Sprintf(" project=%s", out.ProjectID)
	}
	return fmt.Sprintf("UDB instances %s page %d size %d: returned %d (api_total_count=%d)",
		scope, out.Page, out.PageSize, out.ReturnedCount, out.APITotalCount)
}

func formatInstanceStateText(out types.GetInstanceStateOutput) string {
	return fmt.Sprintf("UDB instance %s state=%s follow_up=%s hint=%s", out.DBID, out.State, out.FollowUpTool, out.Hint)
}

func formatLifecycleText(out types.LifecycleMutationOutput) string {
	return fmt.Sprintf("UDB instance %s accepted follow_up=%s hint=%s", out.DBID, out.FollowUpTool, out.Hint)
}
