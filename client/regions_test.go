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

func TestClientListRegions(t *testing.T) {
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
			"RetCode": 0,
			"Action":  "GetRegionResponse",
			"Regions": []map[string]any{
				{"Region": "cn-sh2", "RegionName": "上海二", "Zone": "cn-sh2-01", "IsDefault": false, "RegionId": 2, "BitMaps": "1"},
				{"Region": "cn-bj2", "RegionName": "北京二", "Zone": "cn-bj2-04", "IsDefault": true, "RegionId": 1, "BitMaps": "1"},
				{"Region": "cn-bj2", "RegionName": "北京二", "Zone": "cn-bj2-02", "IsDefault": false, "RegionId": 1, "BitMaps": "1"},
			},
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second
	reqCtx := client.CallContext{PublicKey: "test-public", PrivateKey: "test-private"}

	out, err := client.New(factory).ListRegions(context.Background(), reqCtx, types.ListRegionsInput{})
	if err != nil {
		t.Fatalf("ListRegions: %v", err)
	}
	if gotAction != "GetRegion" {
		t.Fatalf("action: got %q", gotAction)
	}
	if out.TotalCount != 2 || out.ReturnedCount != 2 || len(out.Regions) != 2 {
		t.Fatalf("counts: got %+v", out)
	}
	if out.Regions[0].Region != "cn-bj2" {
		t.Fatalf("sort: first region got %q want cn-bj2", out.Regions[0].Region)
	}
	if !out.Regions[0].IsDefault || out.Regions[0].RegionName != "北京二" {
		t.Fatalf("beijing: got %+v", out.Regions[0])
	}
	if len(out.Regions[0].Zones) != 2 || out.Regions[0].Zones[0].Zone != "cn-bj2-02" || out.Regions[0].Zones[1].Zone != "cn-bj2-04" {
		t.Fatalf("zones: got %+v", out.Regions[0].Zones)
	}
	if !out.Regions[0].Zones[1].IsDefault {
		t.Fatalf("default zone: got %+v", out.Regions[0].Zones)
	}
}

func TestClientListRegionsFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"RetCode": 0,
			"Action":  "GetRegionResponse",
			"Regions": []map[string]any{
				{"Region": "cn-bj2", "RegionName": "北京二", "Zone": "cn-bj2-02", "IsDefault": false},
				{"Region": "cn-bj2", "RegionName": "北京二", "Zone": "cn-bj2-04", "IsDefault": true},
				{"Region": "cn-sh2", "RegionName": "上海二", "Zone": "cn-sh2-01", "IsDefault": false},
			},
		})
	}))
	defer srv.Close()

	factory := client.NewFactory()
	factory.BaseURL = srv.URL
	factory.Timeout = 5 * time.Second
	reqCtx := client.CallContext{PublicKey: "pk", PrivateKey: "sk"}
	cloud := client.New(factory)

	byName, err := cloud.ListRegions(context.Background(), reqCtx, types.ListRegionsInput{NameContains: "北京"})
	if err != nil {
		t.Fatalf("name_contains: %v", err)
	}
	if byName.TotalCount != 2 || byName.ReturnedCount != 1 || byName.Regions[0].Region != "cn-bj2" {
		t.Fatalf("name_contains: got %+v", byName)
	}

	byZone, err := cloud.ListRegions(context.Background(), reqCtx, types.ListRegionsInput{Zone: "cn-bj2-04"})
	if err != nil {
		t.Fatalf("zone: %v", err)
	}
	if byZone.ReturnedCount != 1 || len(byZone.Regions[0].Zones) != 1 || byZone.Regions[0].Zones[0].Zone != "cn-bj2-04" {
		t.Fatalf("zone: got %+v", byZone)
	}

	byRegion, err := cloud.ListRegions(context.Background(), reqCtx, types.ListRegionsInput{Region: "cn-sh2"})
	if err != nil {
		t.Fatalf("region: %v", err)
	}
	if byRegion.ReturnedCount != 1 || byRegion.Regions[0].Region != "cn-sh2" {
		t.Fatalf("region: got %+v", byRegion)
	}
}
