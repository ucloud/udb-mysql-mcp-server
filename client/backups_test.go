package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

type describeWireCall struct {
	offset  int
	limit   int
	zone    string
	region  string
	project string
}

func TestFindBackupByIDFindsOnSecondPageWithOffset100(t *testing.T) {
	var calls []describeWireCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("Action") != "DescribeUDBBackup" {
			t.Fatalf("action: %q", values.Get("Action"))
		}
		offset, _ := parseInt(values.Get("Offset"))
		limit, _ := parseInt(values.Get("Limit"))
		calls = append(calls, describeWireCall{
			offset: offset, limit: limit, zone: values.Get("Zone"),
			region: values.Get("Region"), project: values.Get("ProjectId"),
		})

		page := offset/limit + 1
		if page == 1 {
			items := make([]map[string]any, 0, 100)
			for i := 0; i < 100; i++ {
				items = append(items, map[string]any{
					"BackupId": i + 1, "BackupName": fmt.Sprintf("b-%d", i), "BackupType": 1,
					"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
				})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 101, "DataSet": items,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 101,
			"DataSet": []map[string]any{{
				"BackupId": 142, "BackupName": "target-backup", "BackupType": 1,
				"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
			}},
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	backup, err := api.FindBackupByID(context.Background(), reqCtx, 142, "cn-bj2-04", "")
	if err != nil {
		t.Fatalf("FindBackupByID: %v", err)
	}
	if backup.BackupID != 142 || backup.BackupName != "target-backup" {
		t.Fatalf("backup: %+v", backup)
	}
	if len(calls) != 2 || calls[0].offset != 0 || calls[1].offset != 100 {
		t.Fatalf("pagination offsets: %+v", calls)
	}
	for _, call := range calls {
		if call.limit != 100 || call.zone != "cn-bj2-04" || call.region != "cn-bj2" || call.project != "org-1" {
			t.Fatalf("wire scope/zone: %+v", call)
		}
	}
}

func TestFindBackupByIDNotFoundAfterFullScan(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 1,
			"DataSet": []map[string]any{{
				"BackupId": 1, "BackupName": "other", "BackupType": 1,
				"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
			}},
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	_, err := api.FindBackupByID(context.Background(), reqCtx, 99, "cn-bj2-04", "")
	var notFound *client.BackupNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected not found, got %T: %v", err, err)
	}
}

func TestFindBackupByIDAbnormalEmptyPageReturnsScanIncomplete(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		if page == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 2,
				"DataSet": []map[string]any{{
					"BackupId": 1, "BackupName": "one", "BackupType": 1,
					"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
				}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 2, "DataSet": []map[string]any{},
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	_, err := api.FindBackupByID(context.Background(), reqCtx, 99, "cn-bj2-04", "")
	assertScanReason(t, err, client.ScanIncompleteReasonEmptyPage)
}

func TestFindBackupByIDTotalCountChangeReturnsScanIncomplete(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		total := 2
		if page > 1 {
			total = 3
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": total,
			"DataSet": []map[string]any{{
				"BackupId": page, "BackupName": "one", "BackupType": 1,
				"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
			}},
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	_, err := api.FindBackupByID(context.Background(), reqCtx, 99, "cn-bj2-04", "")
	assertScanReason(t, err, client.ScanIncompleteReasonCountChanged)
}

func TestFindBackupByIDScanCapReturnsScanIncompleteLimit(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		items := make([]map[string]any, 0, 100)
		for i := 0; i < 100; i++ {
			items = append(items, map[string]any{
				"BackupId": page*100 + i + 1, "BackupName": fmt.Sprintf("b-%d", i), "BackupType": 1,
				"State": "Success", "DBId": "udb-demo", "Zone": "cn-bj2-04",
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBBackupResponse", "TotalCount": 6000, "DataSet": items,
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	_, err := api.FindBackupByID(context.Background(), reqCtx, 99999, "cn-bj2-04", "")
	assertScanReason(t, err, client.ScanIncompleteReasonLimit)
}

func backupTestClient(t *testing.T, baseURL string) (*client.Client, client.CallContext) {
	t.Helper()
	factory := &client.Factory{BaseURL: baseURL, Timeout: client.NewFactory().Timeout}
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	return client.New(factory), reqCtx
}

func assertScanReason(t *testing.T, err error, want client.ScanIncompleteReason) {
	t.Helper()
	var scanErr *client.ScanIncompleteError
	if !errors.As(err, &scanErr) {
		t.Fatalf("expected ScanIncompleteError, got %T: %v", err, err)
	}
	if scanErr.Reason != want {
		t.Fatalf("reason: got %q want %q (message=%q)", scanErr.Reason, want, scanErr.Error())
	}
}

func parseInt(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	var n int
	_, err := fmt.Sscanf(raw, "%d", &n)
	return n, err
}

func TestGetBackupURLResolvesZoneFromInstanceWhenOmitted(t *testing.T) {
	var actions []string
	var urlZones []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		action := values.Get("Action")
		actions = append(actions, action)
		switch action {
		case "DescribeUDBInstance":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"RetCode": 0, "Action": "DescribeUDBInstanceResponse", "TotalCount": 1,
				"DataSet": []map[string]any{{
					"DBId": "udb-demo", "Name": "demo", "State": "Running",
					"DBTypeId": "mysql-8.0", "Zone": "cn-bj2-03",
				}},
			})
		case "DescribeUDBInstanceBackupURL":
			urlZones = append(urlZones, values.Get("Zone"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"RetCode": 0, "Action": "DescribeUDBInstanceBackupURLResponse",
				"BackupPath": "https://pub.example/b.gz", "InnerBackupPath": "http://inner.example/b.gz", "MD5": "checksum",
			})
		default:
			t.Fatalf("unexpected action %q", action)
		}
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	out, err := api.GetBackupURL(context.Background(), reqCtx, types.GetBackupURLInput{
		BackupID: 42, DBID: "udb-demo",
	})
	if err != nil {
		t.Fatalf("GetBackupURL: %v", err)
	}
	if out.PublicBackupPath != "https://pub.example/b.gz" || out.Checksum != "checksum" {
		t.Fatalf("out: %+v", out)
	}
	if len(actions) != 2 || actions[0] != "DescribeUDBInstance" || actions[1] != "DescribeUDBInstanceBackupURL" {
		t.Fatalf("expected zone lookup then URL call, got %v", actions)
	}
	if len(urlZones) != 1 || urlZones[0] != "cn-bj2-03" {
		t.Fatalf("expected resolved zone cn-bj2-03 on URL call, got %v", urlZones)
	}
}

func TestGetBackupURLSkipsLookupWhenZoneProvided(t *testing.T) {
	var sawInstanceLookup bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		if values.Get("Action") == "DescribeUDBInstance" {
			sawInstanceLookup = true
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "Action": "DescribeUDBInstanceBackupURLResponse",
			"BackupPath": "https://pub.example/b.gz",
		})
	}))
	defer srv.Close()

	api, reqCtx := backupTestClient(t, srv.URL)
	_, err := api.GetBackupURL(context.Background(), reqCtx, types.GetBackupURLInput{
		BackupID: 42, DBID: "udb-demo", Zone: "cn-bj2-03",
	})
	if err != nil {
		t.Fatalf("GetBackupURL: %v", err)
	}
	if sawInstanceLookup {
		t.Fatal("explicit zone must skip the DescribeUDBInstance lookup")
	}
}
