package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func TestClientListProjects(t *testing.T) {
	var gotAction string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("parse form: %v", err)
		}
		gotAction = values.Get("Action")
		if values.Get("ProjectId") != "" || values.Get("Region") != "" {
			t.Fatalf("account API should not send project/region: ProjectId=%q Region=%q", values.Get("ProjectId"), values.Get("Region"))
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode":      0,
			"Action":       "GetProjectListResponse",
			"ProjectCount": 3,
			"ProjectSet": []map[string]any{
				{
					"ProjectId": "org-a", "ProjectName": "default", "IsDefault": true,
					"CreateTime": 1342434682, "MemberCount": 1,
				},
				{
					"ProjectId": "org-b", "ProjectName": "UDB临时测试(小于7天)", "IsDefault": false,
					"CreateTime": 1468225814, "MemberCount": 2,
				},
				{
					"ProjectId": "org-c", "ProjectName": "other", "IsDefault": false,
					"CreateTime": 1468225815, "MemberCount": 1,
				},
			},
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second

	reqCtx := client.CallContext{PublicKey: "test-public", PrivateKey: "test-private"}

	cloudClient := client.New(factory)
	out, err := cloudClient.ListProjects(context.Background(), reqCtx, types.ListProjectsInput{
		Name: "UDB临时测试(小于7天)",
	})
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if gotAction != "GetProjectList" {
		t.Fatalf("action: got %q", gotAction)
	}
	if out.TotalCount != 3 {
		t.Fatalf("total: got %d want 3", out.TotalCount)
	}
	if out.ReturnedCount != 1 || len(out.Projects) != 1 {
		t.Fatalf("returned: got %+v", out)
	}
	if out.Projects[0].ProjectID != "org-b" {
		t.Fatalf("project id: got %q want org-b", out.Projects[0].ProjectID)
	}
}

func TestClientListProjectsNameContains(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode":      0,
			"Action":       "GetProjectListResponse",
			"ProjectCount": 2,
			"ProjectSet": []map[string]any{
				{"ProjectId": "org-a", "ProjectName": "prod", "IsDefault": true, "CreateTime": 1, "MemberCount": 1},
				{"ProjectId": "org-b", "ProjectName": "UDB临时测试(小于7天)", "IsDefault": false, "CreateTime": 2, "MemberCount": 1},
			},
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk"}

	out, err := client.New(factory).ListProjects(context.Background(), reqCtx, types.ListProjectsInput{
		NameContains: "UDB临时",
	})
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if out.ReturnedCount != 1 || out.Projects[0].ProjectID != "org-b" {
		t.Fatalf("got %+v", out)
	}
}
