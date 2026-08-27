package guard_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/guard"
	"udb-mysql-mcp-server/types"
)

func testClient(t *testing.T, baseURL string) (*client.Client, client.CallContext) {
	t.Helper()
	factory := &client.Factory{BaseURL: baseURL, Timeout: 5 * time.Second}
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	return client.New(factory), reqCtx
}

func ptrBool(v bool) *bool { return &v }

func TestDeleteBackupNameMismatchNeverCallsDelete(t *testing.T) {
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		actions = append(actions, values.Get("Action"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 1,
			"DataSet": []map[string]any{{"BackupId": 42, "BackupName": "backup-a", "BackupType": 1, "State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04"}},
		})
	}))
	defer srv.Close()

	api, reqCtx := testClient(t, srv.URL)
	_, err := guard.ConfirmBackupDelete(context.Background(), reqCtx, api, types.DeleteBackupInput{
		BackupID: 42, Name: "wrong-name", Zone: "cn-bj2-04", Confirm: ptrBool(true),
	})
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	var denial *client.ConflictError
	if !errors.As(err, &denial) {
		t.Fatalf("expected conflict denial, got %T: %v", err, err)
	}
	for _, action := range actions {
		if action == "DeleteUDBBackup" {
			t.Fatal("delete must not be called on name mismatch")
		}
	}
}

func TestDeleteBackupConfirmFalseNeverCallsLookup(t *testing.T) {
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		actions = append(actions, values.Get("Action"))
	}))
	defer srv.Close()

	api, reqCtx := testClient(t, srv.URL)
	_, err := guard.ConfirmBackupDelete(context.Background(), reqCtx, api, types.DeleteBackupInput{
		BackupID: 42, Name: "backup-a", Zone: "cn-bj2-04", Confirm: ptrBool(false),
	})
	if err == nil {
		t.Fatal("expected confirmation error")
	}
	var invalid *client.InvalidInputError
	if !errors.As(err, &invalid) || invalid.Field != "confirm" {
		t.Fatalf("expected confirm invalid_input, got %T: %v", err, err)
	}
	if len(actions) != 0 {
		t.Fatalf("lookup must not be called when confirm is false: %v", actions)
	}
}

func TestDeleteBackupExactMatchCallsLookupThenDelete(t *testing.T) {
	var actions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		actions = append(actions, values.Get("Action"))
		switch values.Get("Action") {
		case "DescribeUDBBackup":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 1,
				"DataSet": []map[string]any{{"BackupId": 42, "BackupName": "backup-a", "BackupType": 1, "State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04"}},
			})
		case "DeleteUDBBackup":
			_ = json.NewEncoder(w).Encode(map[string]any{"RetCode": 0, "Action": "DeleteUDBBackupResponse"})
		default:
			t.Fatalf("unexpected action %q", values.Get("Action"))
		}
	}))
	defer srv.Close()

	api, reqCtx := testClient(t, srv.URL)
	confirm := true
	current, err := guard.ConfirmBackupDelete(context.Background(), reqCtx, api, types.DeleteBackupInput{
		BackupID: 42, Name: "backup-a", Zone: "cn-bj2-04", Confirm: &confirm,
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if err := api.DeleteBackup(context.Background(), reqCtx, 42, "cn-bj2-04", ""); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if current.BackupName != "backup-a" {
		t.Fatalf("backup: %+v", current)
	}
	if len(actions) != 2 || actions[0] != "DescribeUDBBackup" || actions[1] != "DeleteUDBBackup" {
		t.Fatalf("actions: %v", actions)
	}
}

func TestDeleteBackupIncompleteLookupNeverCallsDelete(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		items := make([]map[string]any, 0, 100)
		for i := 0; i < 100; i++ {
			items = append(items, map[string]any{
				"BackupId": page*100 + i + 1, "BackupName": "other", "BackupType": 1,
				"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 6000, "DataSet": items,
		})
	}))
	defer srv.Close()

	api, reqCtx := testClient(t, srv.URL)
	_, err := guard.ConfirmBackupDelete(context.Background(), reqCtx, api, types.DeleteBackupInput{
		BackupID: 99999, Name: "backup-a", Zone: "cn-bj2-04", Confirm: ptrBool(true),
	})
	assertScanReason(t, err, client.ScanIncompleteReasonLimit)
}

func assertScanReason(t *testing.T, err error, want client.ScanIncompleteReason) {
	t.Helper()
	var scanErr *client.ScanIncompleteError
	if !errors.As(err, &scanErr) {
		t.Fatalf("expected ScanIncompleteError, got %T: %v", err, err)
	}
	if scanErr.Reason != want {
		t.Fatalf("reason: got %q want %q", scanErr.Reason, want)
	}
}
