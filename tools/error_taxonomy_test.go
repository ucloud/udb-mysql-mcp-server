package tools

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"udb-mysql-mcp-server/client"
)

func TestMapToolErrorResultTaxonomy(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code string
	}{
		{name: "credentials", err: &client.CredentialsError{}, code: "credentials_required"},
		{name: "name conflict", err: &client.ConflictError{}, code: "conflict"},
		{name: "confirm required", err: client.ValidateRequiredConfirm(nil), code: "invalid_input"},
		{name: "upstream", err: &client.UpstreamError{Action: "ResizeUDBInstance", RetCode: 500, Message: "fail"}, code: "upstream_error"},
		{name: "internal", err: &InternalError{Message: "typed payload missing"}, code: "internal_error"},
		{name: "unknown", err: errors.New("confirm must be true"), code: "internal_error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := mapToolErrorResult(tc.err)
			if !result.IsError {
				t.Fatal("expected error result")
			}
			raw, _ := json.Marshal(result.StructuredContent)
			var toolErr ToolError
			if err := json.Unmarshal(raw, &toolErr); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if toolErr.Code != tc.code {
				t.Fatalf("code: got %q want %q", toolErr.Code, tc.code)
			}
		})
	}
}

func TestMapToolErrorResultInternalErrorDoesNotLeakDetail(t *testing.T) {
	result := mapToolErrorResult(&InternalError{Message: "typed payload missing"})
	raw, _ := json.Marshal(result.StructuredContent)
	if strings.Contains(string(raw), "typed payload") {
		t.Fatalf("leaked internal detail: %s", raw)
	}
}

func TestModeAllowsRisk(t *testing.T) {
	if ModeAllowsRisk(ModeReadonly, RiskWriteLow) {
		t.Fatal("readonly must reject writes")
	}
	if !ModeAllowsRisk(ModeReadWrite, RiskWriteMid) {
		t.Fatal("readwrite must allow write-mid")
	}
	if ModeAllowsRisk(ModeReadWrite, RiskCritical) {
		t.Fatal("readwrite must reject critical")
	}
	if !ModeAllowsRisk(ModeAdmin, RiskCritical) {
		t.Fatal("admin must allow critical")
	}
}
