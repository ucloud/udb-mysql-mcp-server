package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"udb-mysql-mcp-server/client"
)

func TestDescribePriceBindingDocumentsInstanceMode(t *testing.T) {
	cfg := Config{
		Resolve: func() (client.CallContext, error) {
			return client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}, nil
		},
		Mode:   ModeAdmin,
		Client: client.New(&client.Factory{}),
	}
	binding, err := DescribePriceBinding(cfg)
	if err != nil {
		t.Fatalf("DescribePriceBinding: %v", err)
	}
	desc := binding.Catalog.Description
	for _, want := range []string{"Normal", "HA", "普通版", "高可用", "list_machine_types", "storage_class"} {
		if !strings.Contains(desc, want) {
			t.Errorf("description missing %q; got: %s", want, desc)
		}
	}
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(binding.MCP.RawInputSchema, &schema); err != nil {
		t.Fatalf("unmarshal schema: %v", err)
	}
	need := map[string]bool{"storage_class": false, "specification_class": false, "machine_type": false}
	for _, r := range schema.Required {
		if _, ok := need[r]; ok {
			need[r] = true
		}
	}
	for field, ok := range need {
		if !ok {
			t.Errorf("required missing %q; got %v", field, schema.Required)
		}
	}
}

// TestAllBindingsSchemaDescriptions asserts every tool exposes a non-empty
// description for every input property. mcp-go silently drops the whole
// description when a struct tag contains a backslash, so guard against it.
func TestAllBindingsSchemaDescriptions(t *testing.T) {
	cfg := Config{
		Resolve: func() (client.CallContext, error) {
			return client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}, nil
		},
		Mode:   ModeAdmin,
		Client: client.New(&client.Factory{}),
	}
	builders := map[string]func(Config) (Binding, error){
		ToolGetInstance:          GetInstanceBinding,
		ToolListInstances:        ListInstancesBinding,
		ToolGetInstanceState:     GetInstanceStateBinding,
		ToolStartInstance:        StartInstanceBinding,
		ToolStopInstance:         StopInstanceBinding,
		ToolRestartInstance:      RestartInstanceBinding,
		ToolModifyName:           ModifyNameBinding,
		ToolCreateInstance:       CreateInstanceBinding,
		ToolResizeInstance:       ResizeInstanceBinding,
		ToolResetPassword:        ResetPasswordBinding,
		ToolListDBTypes:          ListDBTypesBinding,
		ToolListMachineTypes:     ListMachineTypesBinding,
		ToolListParamGroups:      ListParamGroupsBinding,
		ToolDescribePrice:        DescribePriceBinding,
		ToolDescribeUpgradePrice: DescribeUpgradePriceBinding,
		ToolListBackups:          ListBackupsBinding,
		ToolGetBackupState:       GetBackupStateBinding,
		ToolGetBackupURL:         GetBackupURLBinding,
		ToolGetBackupStrategy:    GetBackupStrategyBinding,
		ToolCreateBackup:         CreateBackupBinding,
		ToolDeleteBackup:         DeleteBackupBinding,
	}
	for name, build := range builders {
		binding, err := build(cfg)
		if err != nil {
			t.Errorf("%s: build: %v", name, err)
			continue
		}
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(binding.MCP.RawInputSchema, &schema); err != nil {
			t.Errorf("%s: unmarshal schema: %v", name, err)
			continue
		}
		if len(schema.Properties) == 0 {
			continue
		}
		for field, raw := range schema.Properties {
			var prop struct {
				Description *string `json:"description"`
			}
			if err := json.Unmarshal(raw, &prop); err != nil {
				t.Errorf("%s.%s: unmarshal property: %v", name, field, err)
				continue
			}
			if prop.Description == nil || strings.TrimSpace(*prop.Description) == "" {
				t.Errorf("%s.%s: missing description (check struct tag for backslash)", name, field)
			}
		}
	}
}
