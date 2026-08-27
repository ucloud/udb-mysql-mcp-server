package client_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func createInstanceBaseInput() types.CreateInstanceInput {
	return types.CreateInstanceInput{
		Zone:               "cn-bj2-04",
		Name:               "demo-db",
		AdminPassword:      "pw",
		DBTypeID:           "mysql-8.0",
		DiskSpaceGB:        20,
		ParamGroupID:       1,
		MachineType:        "o.mysql4m.medium",
		StorageClass:       "CLOUD_RSSD",
		SpecificationClass: "O",
	}
}

func TestClientCreateInstanceVPCSubnetPairValidation(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{"RetCode": 0, "DBId": "udb-new"})
	}

	t.Run("both present", func(t *testing.T) {
		var got url.Values
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := createInstanceBaseInput()
		in.VPCID = "vpc-1"
		in.SubnetID = "subnet-1"
		_, err := api.CreateInstance(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("CreateInstance: %v", err)
		}
		if requestCount.Load() != 1 {
			t.Fatalf("upstream calls: got %d want 1", requestCount.Load())
		}
		if got.Get("VPCId") != "vpc-1" || got.Get("SubnetId") != "subnet-1" {
			t.Fatalf("wire: %+v", got)
		}
	})

	t.Run("both absent", func(t *testing.T) {
		var got url.Values
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		_, err := api.CreateInstance(ctx, reqCtx, createInstanceBaseInput())
		if err != nil {
			t.Fatalf("CreateInstance: %v", err)
		}
		if requestCount.Load() != 1 {
			t.Fatalf("upstream calls: got %d want 1", requestCount.Load())
		}
		if _, ok := got["VPCId"]; ok {
			t.Fatalf("VPCId must be omitted, got wire: %+v", got)
		}
		if _, ok := got["SubnetId"]; ok {
			t.Fatalf("SubnetId must be omitted, got wire: %+v", got)
		}
	})

	t.Run("only vpc_id", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := createInstanceBaseInput()
		in.VPCID = "vpc-1"
		_, err := api.CreateInstance(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream must not be called, got %d", requestCount.Load())
		}
	})

	t.Run("only subnet_id", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := createInstanceBaseInput()
		in.SubnetID = "subnet-1"
		_, err := api.CreateInstance(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream must not be called, got %d", requestCount.Load())
		}
	})

	t.Run("whitespace only treated as absent both", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := createInstanceBaseInput()
		in.VPCID = "  "
		in.SubnetID = "\t"
		_, err := api.CreateInstance(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("CreateInstance: %v", err)
		}
		if requestCount.Load() != 1 {
			t.Fatalf("upstream calls: got %d want 1", requestCount.Load())
		}
	})
}
