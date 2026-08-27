package tools

import (
	"strings"
	"testing"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func TestFormatListInstancesTextIncludesScope(t *testing.T) {
	got := formatListInstancesText(types.ListInstancesOutput{
		Region: "cn-bj2", ProjectID: "org-1", Page: 1, PageSize: 20, ReturnedCount: 3, APITotalCount: 3,
	})
	for _, want := range []string{"region=cn-bj2", "project=org-1", "returned 3", "api_total_count=3"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}

	noProject := formatListInstancesText(types.ListInstancesOutput{
		Region: "cn-wlcb", Page: 1, PageSize: 20, ReturnedCount: 0, APITotalCount: 0,
	})
	if !strings.Contains(noProject, "region=cn-wlcb") {
		t.Fatalf("missing region in %q", noProject)
	}
	if strings.Contains(noProject, "project=") {
		t.Fatalf("unexpected project in %q", noProject)
	}
}

func TestListInstancesBindingDocumentsSingleRegionSemantics(t *testing.T) {
	cfg := Config{
		Resolve: func() (client.CallContext, error) {
			return client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}, nil
		},
		Mode:   ModeAdmin,
		Client: client.New(&client.Factory{}),
	}
	binding, err := ListInstancesBinding(cfg)
	if err != nil {
		t.Fatalf("ListInstancesBinding: %v", err)
	}
	desc := binding.Catalog.Description
	for _, want := range []string{"单一地域", "不等于全部地域", "udb_mysql_list_regions", "region/project_id"} {
		if !strings.Contains(desc, want) {
			t.Errorf("description missing %q; got: %s", want, desc)
		}
	}
}
