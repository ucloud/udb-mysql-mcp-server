package tools

import (
	"testing"

	"udb-mysql-mcp-server/types"
)

func TestRegionFromZone(t *testing.T) {
	t.Parallel()
	cases := []struct {
		zone string
		want string
	}{
		{"cn-bj2-04", "cn-bj2"},
		{"cn-sh2-01", "cn-sh2"},
		{"hk-01", "hk"},
		{"us-ca-01", "us-ca"},
		{"idn-jakarta-01", "idn-jakarta"},
		{"  cn-gd-02  ", "cn-gd"},
		{"cn-bj2", ""},
		{"cn-bj2-", ""},
		{"", ""},
		{"no-dash", ""},
		{"cn-bj2-0a", ""},
	}
	for _, tc := range cases {
		if got := regionFromZone(tc.zone); got != tc.want {
			t.Fatalf("regionFromZone(%q)=%q want %q", tc.zone, got, tc.want)
		}
	}
}

func TestScopeFromInputDerivesRegionFromZone(t *testing.T) {
	t.Parallel()

	projectID, region := scopeFromInput(types.ScopeInput{ProjectID: "org-1"}, "cn-bj2-04")
	if projectID != "org-1" || region != "cn-bj2" {
		t.Fatalf("got project=%q region=%q", projectID, region)
	}

	_, region = scopeFromInput(types.ScopeInput{Region: "cn-sh2"}, "cn-bj2-04")
	if region != "cn-sh2" {
		t.Fatalf("explicit region should win, got %q", region)
	}

	_, region = scopeFromInput(types.ScopeInput{}, "")
	if region != "" {
		t.Fatalf("empty scope should stay empty, got %q", region)
	}
}
