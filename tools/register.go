package tools

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

// Register adds mode-allowed UDB tools to the MCP server.
func Register(s *server.MCPServer, cfg Config) error {
	builders := []struct {
		name  string
		build func() (Binding, error)
	}{
		{"get instance", func() (Binding, error) { return GetInstanceBinding(cfg) }},
		{"list projects", func() (Binding, error) { return ListProjectsBinding(cfg) }},
		{"list regions", func() (Binding, error) { return ListRegionsBinding(cfg) }},
		{"list instances", func() (Binding, error) { return ListInstancesBinding(cfg) }},
		{"get instance state", func() (Binding, error) { return GetInstanceStateBinding(cfg) }},
		{"list db types", func() (Binding, error) { return ListDBTypesBinding(cfg) }},
		{"list machine types", func() (Binding, error) { return ListMachineTypesBinding(cfg) }},
		{"list param groups", func() (Binding, error) { return ListParamGroupsBinding(cfg) }},
		{"describe price", func() (Binding, error) { return DescribePriceBinding(cfg) }},
		{"describe upgrade price", func() (Binding, error) { return DescribeUpgradePriceBinding(cfg) }},
		{"list backups", func() (Binding, error) { return ListBackupsBinding(cfg) }},
		{"get backup state", func() (Binding, error) { return GetBackupStateBinding(cfg) }},
		{"get backup url", func() (Binding, error) { return GetBackupURLBinding(cfg) }},
		{"get backup strategy", func() (Binding, error) { return GetBackupStrategyBinding(cfg) }},
		{"start instance", func() (Binding, error) { return StartInstanceBinding(cfg) }},
		{"stop instance", func() (Binding, error) { return StopInstanceBinding(cfg) }},
		{"restart instance", func() (Binding, error) { return RestartInstanceBinding(cfg) }},
		{"create backup", func() (Binding, error) { return CreateBackupBinding(cfg) }},
		{"modify name", func() (Binding, error) { return ModifyNameBinding(cfg) }},
		{"create instance", func() (Binding, error) { return CreateInstanceBinding(cfg) }},
		{"resize instance", func() (Binding, error) { return ResizeInstanceBinding(cfg) }},
		{"reset password", func() (Binding, error) { return ResetPasswordBinding(cfg) }},
		{"delete backup", func() (Binding, error) { return DeleteBackupBinding(cfg) }},
	}

	seen := map[string]struct{}{}
	for _, b := range builders {
		binding, err := b.build()
		if err != nil {
			return fmt.Errorf("udb: %s binding: %w", b.name, err)
		}
		if _, dup := seen[binding.Catalog.Name]; dup {
			return fmt.Errorf("udb: duplicate tool binding %q", binding.Catalog.Name)
		}
		seen[binding.Catalog.Name] = struct{}{}
		if !ModeAllowsRisk(cfg.Mode, binding.Catalog.Risk) {
			continue
		}
		s.AddTool(binding.MCP, binding.Handler)
	}
	return nil
}
