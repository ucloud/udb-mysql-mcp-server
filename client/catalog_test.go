package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

func intPtr(v int) *int {
	return &v
}

func catalogTestClient(t *testing.T, srvURL string) (*client.Client, client.CallContext, context.Context) {
	t.Helper()
	api := client.New(&client.Factory{BaseURL: srvURL, Timeout: 5 * time.Second})
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	return api, reqCtx, context.Background()
}

func describePriceBaseInput() types.DescribePriceInput {
	return types.DescribePriceInput{
		Zone:               "cn-bj2-04",
		DBTypeID:           "mysql-8.0",
		MemoryLimitMB:      8000,
		DiskSpaceGB:        200,
		MachineType:        "o.mysql4m.medium",
		StorageClass:       "CLOUD_RSSD",
		SpecificationClass: "O",
	}
}

func TestClientListParamGroupsPagination(t *testing.T) {
	var gotOffset, gotLimit int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		gotOffset, _ = strconv.Atoi(values.Get("Offset"))
		gotLimit, _ = strconv.Atoi(values.Get("Limit"))
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0, "TotalCount": 100,
			"DataSet": []map[string]any{{
				"GroupId": 1, "GroupName": "g1", "DBTypeId": "mysql-8.0", "GroupType": 1,
			}},
		})
	}))
	defer srv.Close()

	factory := &client.Factory{BaseURL: srv.URL, Timeout: 5 * time.Second}
	api := client.New(factory)
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	ctx := context.Background()

	out, err := api.ListParamGroups(ctx, reqCtx, types.ListParamGroupsInput{Page: 3, PageSize: 20})
	if err != nil {
		t.Fatalf("ListParamGroups: %v", err)
	}
	if gotOffset != 40 || gotLimit != 20 {
		t.Fatalf("wire offset=%d limit=%d", gotOffset, gotLimit)
	}
	if out.APITotalCount != 100 || out.ReturnedCount != 1 || out.Page != 3 || out.PageSize != 20 {
		t.Fatalf("output: %+v", out)
	}
}

func TestClientListParamGroupsPageOverflow(t *testing.T) {
	api := client.New(&client.Factory{BaseURL: "http://unused", Timeout: time.Second})
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	_, err := api.ListParamGroups(context.Background(), reqCtx, types.ListParamGroupsInput{
		Page: mathMaxInt, PageSize: 100,
	})
	if err == nil {
		t.Fatal("expected page overflow error")
	}
}

func TestClientListDBTypesWire(t *testing.T) {
	var gotAction string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(body))
		gotAction = values.Get("Action")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode":     0,
			"DedaultType": map[string]any{"DBTypeId": "mysql-8.0", "DBSubVersion": "8.0"},
			"DataSet":     []map[string]any{{"DBTypeId": "mysql-8.0", "DBSubVersion": "8.0"}},
		})
	}))
	defer srv.Close()

	factory := &client.Factory{BaseURL: srv.URL, Timeout: 5 * time.Second}
	api := client.New(factory)
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	ctx := context.Background()

	out, err := api.ListDBTypes(ctx, reqCtx, types.ListDBTypesInput{Zone: "cn-bj2-04"})
	if err != nil {
		t.Fatalf("ListDBTypes: %v", err)
	}
	if gotAction != "DescribeUDBType" {
		t.Fatalf("action: %q", gotAction)
	}
	if out.RecommendedDBTypeID != "mysql-8.0" || len(out.Items) != 1 || !out.Items[0].Recommended {
		t.Fatalf("output: %+v", out)
	}
}

const mathMaxInt = int(^uint(0) >> 1)

func TestClientDescribeUpgradePrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"RetCode": 0, "Price": 999})
	}))
	defer srv.Close()

	api := client.New(&client.Factory{BaseURL: srv.URL, Timeout: 5 * time.Second})
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}
	ctx := context.Background()

	out, err := api.DescribeUpgradePrice(ctx, reqCtx, types.DescribeUpgradePriceInput{
		DBID: "udb-1", MemoryLimitMB: 4000, DiskSpaceGB: 100,
	})
	if err != nil {
		t.Fatalf("DescribeUpgradePrice: %v", err)
	}
	if out.PriceCents != 999 {
		t.Fatalf("price_cents: %d", out.PriceCents)
	}
}

func TestClientDescribePriceMachineTypePath(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0,
			"DataSet": []map[string]any{{"ChargeType": "Month", "Price": 1}},
		})
	}

	t.Run("requires machine_type", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.MachineType = ""
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if invalid.Field != "machine_type" {
			t.Fatalf("field: got %q want machine_type", invalid.Field)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream calls: got %d want 0", requestCount.Load())
		}
	})

	t.Run("always wires SpecificationType=1", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		_, err := api.DescribePrice(ctx, reqCtx, describePriceBaseInput())
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("SpecificationType") != "1" {
			t.Fatalf("SpecificationType: got %q want 1", got.Get("SpecificationType"))
		}
		if got.Get("MachineType") != "o.mysql4m.medium" {
			t.Fatalf("MachineType: got %q want o.mysql4m.medium", got.Get("MachineType"))
		}
		if got.Get("InstanceMode") != "Normal" {
			t.Fatalf("InstanceMode: got %q want Normal (default)", got.Get("InstanceMode"))
		}
		if got.Get("StorageClass") != "CLOUD_RSSD" {
			t.Fatalf("StorageClass: got %q want CLOUD_RSSD", got.Get("StorageClass"))
		}
		if got.Get("SpecificationClass") != "O" {
			t.Fatalf("SpecificationClass: got %q want O", got.Get("SpecificationClass"))
		}
	})

	t.Run("requires storage_class", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.StorageClass = ""
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if invalid.Field != "storage_class" {
			t.Fatalf("field: got %q want storage_class", invalid.Field)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream calls: got %d want 0", requestCount.Load())
		}
	})

	t.Run("requires specification_class", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.SpecificationClass = ""
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if invalid.Field != "specification_class" {
			t.Fatalf("field: got %q want specification_class", invalid.Field)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream calls: got %d want 0", requestCount.Load())
		}
	})

	t.Run("wires explicit instance_mode HA", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.InstanceMode = "HA"
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("InstanceMode") != "HA" {
			t.Fatalf("InstanceMode: got %q want HA", got.Get("InstanceMode"))
		}
	})
}

func TestClientDescribePriceQuantityPresence(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0,
			"DataSet": []map[string]any{{"ChargeType": "Month", "Price": 1}},
		})
	}

	t.Run("omit when nil", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		_, err := api.DescribePrice(ctx, reqCtx, describePriceBaseInput())
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if _, ok := got["Quantity"]; ok {
			t.Fatalf("Quantity must be omitted when nil, got wire: %+v", got)
		}
	})

	t.Run("wire zero when pointer to zero", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Quantity = intPtr(0)
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("Quantity") != "0" {
			t.Fatalf("Quantity: got %q want 0", got.Get("Quantity"))
		}
	})

	t.Run("wire positive", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Quantity = intPtr(3)
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("Quantity") != "3" {
			t.Fatalf("Quantity: got %q want 3", got.Get("Quantity"))
		}
	})
}

func TestClientDescribePriceCountValidation(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0,
			"DataSet": []map[string]any{{"ChargeType": "Month", "Price": 1}},
		})
	}

	t.Run("negative rejected without upstream", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Count = -1
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err == nil {
			t.Fatal("expected invalid_input error")
		}
		var invalid *client.InvalidInputError
		if !errors.As(err, &invalid) {
			t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
		}
		if invalid.Field != "count" {
			t.Fatalf("field: got %q want count", invalid.Field)
		}
		if requestCount.Load() != 0 {
			t.Fatalf("upstream must not be called, got %d", requestCount.Load())
		}
	})

	t.Run("zero defaults to one on wire", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Count = 0
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("Count") != "1" {
			t.Fatalf("Count: got %q want 1", got.Get("Count"))
		}
	})

	t.Run("positive passed through", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Count = 3
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("Count") != "3" {
			t.Fatalf("Count: got %q want 3", got.Get("Count"))
		}
	})

	t.Run("large count passed through without local cap", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := describePriceBaseInput()
		in.Count = 15
		_, err := api.DescribePrice(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("DescribePrice: %v", err)
		}
		if got.Get("Count") != "15" {
			t.Fatalf("Count: got %q want 15", got.Get("Count"))
		}
	})
}

func TestClientCreateInstanceQuantityPresence(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{"RetCode": 0, "DBId": "udb-new"})
	}
	createBase := types.CreateInstanceInput{
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

	t.Run("omit when nil", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		_, err := api.CreateInstance(ctx, reqCtx, createBase)
		if err != nil {
			t.Fatalf("CreateInstance: %v", err)
		}
		if _, ok := got["Quantity"]; ok {
			t.Fatalf("Quantity must be omitted when nil, got wire: %+v", got)
		}
	})

	t.Run("wire zero when pointer to zero", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		in := createBase
		in.Quantity = intPtr(0)
		_, err := api.CreateInstance(ctx, reqCtx, in)
		if err != nil {
			t.Fatalf("CreateInstance: %v", err)
		}
		if got.Get("Quantity") != "0" {
			t.Fatalf("Quantity: got %q want 0", got.Get("Quantity"))
		}
	})
}

func TestCatalogReadRespectsContextDeadline(t *testing.T) {
	requestReached := make(chan struct{}, 1)
	serverCanceled := make(chan struct{}, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReached <- struct{}{}
		go func() {
			<-r.Context().Done()
			serverCanceled <- struct{}{}
		}()
		w.WriteHeader(http.StatusOK)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Error("expected flusher")
			return
		}
		for {
			select {
			case <-r.Context().Done():
				return
			default:
				flusher.Flush()
				time.Sleep(10 * time.Millisecond)
			}
		}
	}))
	defer srv.Close()

	api, reqCtx, ctx := catalogTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	clientDone := make(chan error, 1)
	go func() {
		_, err := api.ListDBTypes(ctx, reqCtx, types.ListDBTypesInput{Zone: "cn-bj2-04"})
		clientDone <- err
	}()

	select {
	case <-requestReached:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for request to reach server")
	}

	var clientErr error
	select {
	case clientErr = <-clientDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for client to return after context deadline")
	}
	if clientErr == nil {
		t.Fatal("expected error from context deadline")
	}
	if !errors.Is(clientErr, context.DeadlineExceeded) && !errors.Is(clientErr, context.Canceled) {
		var upstream *client.UpstreamError
		if !errors.As(clientErr, &upstream) {
			t.Fatalf("expected deadline/cancel or upstream timeout, got %T: %v", clientErr, clientErr)
		}
	}

	select {
	case <-serverCanceled:
	case <-time.After(5 * time.Second):
		t.Fatal("server did not observe request context cancellation")
	}
}

func TestListDBTypesReturnsCanceledContextError(t *testing.T) {
	api, reqCtx, ctx := catalogTestClient(t, "http://unused")
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	_, err := api.ListDBTypes(ctx, reqCtx, types.ListDBTypesInput{Zone: "cn-bj2-04"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %T: %v", err, err)
	}
}

func TestCatalogMapsRetCodeErrorOnHTTP200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 88888,
			"Action":  "DescribeUDBTypeResponse",
			"Message": "catalog rejected",
		})
	}))
	defer srv.Close()

	api, reqCtx, ctx := catalogTestClient(t, srv.URL)
	_, err := api.ListDBTypes(ctx, reqCtx, types.ListDBTypesInput{Zone: "cn-bj2-04"})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	var upstream *client.UpstreamError
	if !errors.As(err, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T: %v", err, err)
	}
	if upstream.Action != "DescribeUDBType" {
		t.Fatalf("action: got %q want DescribeUDBType", upstream.Action)
	}
	if upstream.RetCode != 88888 {
		t.Fatalf("ret_code: got %d want 88888", upstream.RetCode)
	}
}

func TestClientListParamGroupsIsInUDBCPresence(t *testing.T) {
	respond := func(w http.ResponseWriter) {
		_ = json.NewEncoder(w).Encode(map[string]any{"RetCode": 0, "TotalCount": 0, "DataSet": []any{}})
	}

	t.Run("omit when nil", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		_, err := api.ListParamGroups(ctx, reqCtx, types.ListParamGroupsInput{})
		if err != nil {
			t.Fatalf("ListParamGroups: %v", err)
		}
		if _, ok := got["IsInUDBC"]; ok {
			t.Fatalf("IsInUDBC must be omitted when nil, got wire: %+v", got)
		}
	})

	t.Run("wire false", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		falseVal := false
		_, err := api.ListParamGroups(ctx, reqCtx, types.ListParamGroupsInput{IsInUDBC: &falseVal})
		if err != nil {
			t.Fatalf("ListParamGroups: %v", err)
		}
		if got.Get("IsInUDBC") != "false" {
			t.Fatalf("IsInUDBC: got %q want false", got.Get("IsInUDBC"))
		}
	})

	t.Run("wire true", func(t *testing.T) {
		var got url.Values
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			got, _ = url.ParseQuery(string(body))
			respond(w)
		}))
		defer srv.Close()

		api, reqCtx, ctx := catalogTestClient(t, srv.URL)
		trueVal := true
		_, err := api.ListParamGroups(ctx, reqCtx, types.ListParamGroupsInput{IsInUDBC: &trueVal})
		if err != nil {
			t.Fatalf("ListParamGroups: %v", err)
		}
		if got.Get("IsInUDBC") != "true" {
			t.Fatalf("IsInUDBC: got %q want true", got.Get("IsInUDBC"))
		}
	})
}

func TestClientListParamGroupsSingleGroupRequiresZone(t *testing.T) {
	// No stub server needed: validation must fire before any upstream call.
	api := client.New(&client.Factory{Timeout: client.NewFactory().Timeout})
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-1", Region: "cn-bj2"}

	_, err := api.ListParamGroups(context.Background(), reqCtx, types.ListParamGroupsInput{GroupID: 15694})
	if err == nil {
		t.Fatal("expected InvalidInputError when group_id is set without zone")
	}
	var invalid *client.InvalidInputError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected *client.InvalidInputError, got %T: %v", err, err)
	}
	if invalid.Field != "zone" {
		t.Fatalf("field: got %q want zone", invalid.Field)
	}
	if !strings.Contains(invalid.Error(), "group_id") {
		t.Fatalf("message should tie the requirement to group_id lookup, got %q", invalid.Error())
	}

	// List mode without zone stays valid: the error above must come from the
	// group_id pairing, not from zone being required in general.
	if err := func() error {
		_, err := api.ListParamGroups(context.Background(), reqCtx, types.ListParamGroupsInput{})
		if _, ok := err.(*client.InvalidInputError); ok && strings.Contains(err.Error(), "zone") {
			return err
		}
		return nil
	}(); err != nil {
		t.Fatalf("list mode without zone must not fail zone validation early, got %v", err)
	}
}
