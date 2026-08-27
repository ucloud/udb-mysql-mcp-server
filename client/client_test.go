package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func TestClientAppliesRequestScopeAndDisablesRetries(t *testing.T) {
	var (
		gotAction, gotDBID, gotRegion, gotProject string
		requestCount                              atomic.Int32
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatalf("parse form: %v", err)
		}
		gotAction = values.Get("Action")
		gotDBID = values.Get("DBId")
		gotRegion = values.Get("Region")
		gotProject = values.Get("ProjectId")

		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0,
			"Action":  "DescribeUDBInstanceResponse",
			"DataSet": []map[string]any{
				{
					"DBId":         "udb-test",
					"Name":         "demo-db",
					"State":        " Running ",
					"DBTypeId":     "mysql-8.0",
					"InstanceType": "SATA_SSD",
					"VirtualIP":    "10.1.2.3",
					"Port":         3306,
					"CPU":          2,
					"MemoryLimit":  4000,
					"DiskSpace":    100,
					"Zone":         "cn-bj2-04",
					"CreateTime":   1344810776,
				},
			},
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second

	reqCtx := client.CallContext{PublicKey: "test-public", PrivateKey: "test-private", ProjectID: "org-1", Region: "cn-bj2"}

	cloudClient := client.New(factory)
	out, err := cloudClient.GetInstance(context.Background(), reqCtx, types.GetInstanceInput{DBID: "udb-test"})
	if err != nil {
		t.Fatalf("GetInstance: %v", err)
	}

	if gotAction != "DescribeUDBInstance" {
		t.Fatalf("action: got %q", gotAction)
	}
	if gotDBID != "udb-test" {
		t.Fatalf("DBId: got %q", gotDBID)
	}
	if gotRegion != "cn-bj2" {
		t.Fatalf("Region: got %q", gotRegion)
	}
	if gotProject != "org-1" {
		t.Fatalf("ProjectId: got %q", gotProject)
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count: got %d want 1", requestCount.Load())
	}
	if out.State != "Running" || out.DBID != "udb-test" {
		t.Fatalf("output: %+v", out)
	}
}

func TestClientDoesNotRetryOnUpstreamFailure(t *testing.T) {
	var requestCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 500,
			"Action":  "DescribeUDBInstanceResponse",
			"Message": "temporary upstream failure",
		})
	}))
	defer srv.Close()

	reqCtx := client.CallContext{PublicKey: "test-public", PrivateKey: "test-private", ProjectID: "org-1", Region: "cn-bj2"}
	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second

	_, err := client.New(factory).GetInstance(context.Background(), reqCtx, types.GetInstanceInput{DBID: "udb-test"})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if requestCount.Load() != 1 {
		t.Fatalf("request count: got %d want 1 (retries disabled)", requestCount.Load())
	}
}

func TestClientMapsRetCodeErrorOnHTTP200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 99999,
			"Action":  "DescribeUDBInstanceResponse",
			"Message": "resource unavailable",
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second
	reqCtx := client.CallContext{PublicKey: "test-public-key", PrivateKey: "test-private-key", ProjectID: "org-1", Region: "cn-bj2"}

	_, err := client.New(factory).GetInstance(context.Background(), reqCtx, types.GetInstanceInput{DBID: "udb-test"})
	if err == nil {
		t.Fatal("expected upstream error")
	}

	var upstream *client.UpstreamError
	if !errors.As(err, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T: %v", err, err)
	}
	if upstream.Action != "DescribeUDBInstance" {
		t.Fatalf("action: got %q", upstream.Action)
	}
	if upstream.RetCode != 99999 {
		t.Fatalf("ret_code: got %d", upstream.RetCode)
	}
	if !strings.Contains(upstream.Message, "resource unavailable") {
		t.Fatalf("message: got %q", upstream.Message)
	}

	errText := err.Error()
	for _, secret := range []string{"test-public-key", "test-private-key", "PrivateKey", "PublicKey"} {
		if strings.Contains(errText, secret) {
			t.Fatalf("error leaked credential material %q: %q", secret, errText)
		}
	}
}

func TestUpstreamErrorDoesNotLeakSecrets(t *testing.T) {
	err := &client.UpstreamError{
		Action:  "DescribeUDBInstance",
		RetCode: 10001,
		Message: "permission denied",
		Cause:   nil,
	}
	msg := err.Error()
	if strings.Contains(msg, "test-private") || strings.Contains(msg, "PrivateKey") {
		t.Fatalf("error leaked secret: %q", msg)
	}
	if err.UpstreamRetCode() != 10001 {
		t.Fatalf("ret code: got %d", err.UpstreamRetCode())
	}
}

func TestNewFactoryReadsAPIBaseURLEnv(t *testing.T) {
	t.Setenv("UCLOUD_API_BASE_URL", "https://api.example.test/")
	factory := client.NewFactory()
	if factory.BaseURL != "https://api.example.test/" {
		t.Fatalf("BaseURL: got %q", factory.BaseURL)
	}
	t.Setenv("UCLOUD_API_BASE_URL", "   ")
	factory = client.NewFactory()
	if factory.BaseURL != "" {
		t.Fatalf("blank BaseURL should be ignored, got %q", factory.BaseURL)
	}
}
