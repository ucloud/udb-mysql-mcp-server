package client_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func TestClientListInstancesEchoesEffectiveScope(t *testing.T) {
	var gotRegion, gotProject string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("Action") != "DescribeUDBInstance" {
			t.Fatalf("action: %q", values.Get("Action"))
		}
		gotRegion = values.Get("Region")
		gotProject = values.Get("ProjectId")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBInstanceResponse", "TotalCount": 1,
			"DataSet": []map[string]any{{
				"DBId": "udb-1", "Name": "demo-db", "State": "Running",
				"DBTypeId": "mysql-8.0", "Zone": "cn-wlcb-01", "Role": "master",
			}},
		})
	}))
	defer srv.Close()

	api, _, ctx := catalogTestClient(t, srv.URL)
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-mcimio", Region: "cn-wlcb"}
	out, err := api.ListInstances(ctx, reqCtx, types.ListInstancesInput{})
	if err != nil {
		t.Fatalf("ListInstances: %v", err)
	}
	if gotRegion != "cn-wlcb" || gotProject != "org-mcimio" {
		t.Fatalf("wire scope: region=%q project=%q", gotRegion, gotProject)
	}
	if out.Region != "cn-wlcb" || out.ProjectID != "org-mcimio" {
		t.Fatalf("output scope: region=%q project=%q", out.Region, out.ProjectID)
	}
	if out.ReturnedCount != 1 || out.APITotalCount != 1 {
		t.Fatalf("counts: %+v", out)
	}
}

func TestClientListInstancesEchoesDefaultRegionFromCallContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBInstanceResponse", "TotalCount": 0, "DataSet": []any{},
		})
	}))
	defer srv.Close()

	api, reqCtx, ctx := catalogTestClient(t, srv.URL)
	out, err := api.ListInstances(ctx, reqCtx, types.ListInstancesInput{})
	if err != nil {
		t.Fatalf("ListInstances: %v", err)
	}
	if out.Region != "cn-bj2" || out.ProjectID != "org-1" {
		t.Fatalf("output scope: region=%q project=%q want cn-bj2/org-1", out.Region, out.ProjectID)
	}
}
