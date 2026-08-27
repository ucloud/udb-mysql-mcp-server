package client

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

type secretBearingError struct {
	msg string
}

func (e *secretBearingError) Error() string { return e.msg }

func TestMapSDKErrorNonServerErrorDoesNotLeakCause(t *testing.T) {
	cause := &secretBearingError{
		msg: "PublicKey=pk-leak PrivateKey=sk-leak password=super-secret signedURL=https://backup.example/dl?token=abc",
	}
	mapped := mapSDKError("DescribeUDBInstance", cause)

	var upstream *UpstreamError
	if !errors.As(mapped, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T: %v", mapped, mapped)
	}
	if upstream.Message != "request failed" {
		t.Fatalf("message: got %q want %q", upstream.Message, "request failed")
	}
	if upstream.Action != "DescribeUDBInstance" {
		t.Fatalf("action: got %q", upstream.Action)
	}
	if !errors.Is(mapped, cause) {
		t.Fatal("expected errors.Is to find original cause via Unwrap")
	}
	var asCause *secretBearingError
	if !errors.As(mapped, &asCause) {
		t.Fatal("expected errors.As to reach original cause via Unwrap")
	}

	errText := mapped.Error()
	for _, secret := range []string{
		"PublicKey", "PrivateKey", "password", "signedURL",
		"pk-leak", "sk-leak", "super-secret", "token=abc",
	} {
		if strings.Contains(errText, secret) {
			t.Fatalf("error text leaked %q: %q", secret, errText)
		}
	}
}

func TestMapSDKErrorServerErrorPreservesRetCodeMessage(t *testing.T) {
	serverErr := uerr.NewServerCodeError(12345, "resource unavailable")
	mapped := mapSDKError("DescribeUDBInstance", serverErr)

	var upstream *UpstreamError
	if !errors.As(mapped, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T", mapped)
	}
	if upstream.RetCode != 12345 {
		t.Fatalf("ret_code: got %d want 12345", upstream.RetCode)
	}
	if upstream.Message != "resource unavailable" {
		t.Fatalf("message: got %q", upstream.Message)
	}
	if !errors.Is(mapped, serverErr) {
		t.Fatal("expected errors.Is to find server error cause")
	}
}

func TestMapSDKErrorWrappedSecretCause(t *testing.T) {
	inner := &secretBearingError{msg: "PrivateKey=inner-leak"}
	cause := fmt.Errorf("transport failed: %w", inner)
	mapped := mapSDKError("BackupUDBInstance", cause)

	var upstream *UpstreamError
	if !errors.As(mapped, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T", mapped)
	}
	if upstream.Message != "request failed" {
		t.Fatalf("message: got %q", upstream.Message)
	}
	if !errors.Is(mapped, inner) {
		t.Fatal("expected errors.Is to reach wrapped inner cause")
	}
	if strings.Contains(mapped.Error(), "inner-leak") || strings.Contains(mapped.Error(), "PrivateKey") {
		t.Fatalf("leaked wrapped secret: %q", mapped.Error())
	}
}

func TestEnhanceProjectScopeErrorAddsHintOnProjectMismatchRetCodes(t *testing.T) {
	reqCtx := CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-default", Region: "cn-bj2"}

	for _, retCode := range []int{7009, 7048} {
		mapped := mapSDKError("DescribeUDBInstanceState", uerr.NewServerCodeError(retCode, "describe error"))
		enhanced := enhanceProjectScopeError(mapped, reqCtx)

		var upstream *UpstreamError
		if !errors.As(enhanced, &upstream) {
			t.Fatalf("ret_code=%d: expected *UpstreamError, got %T: %v", retCode, enhanced, enhanced)
		}
		if upstream.RetCode != retCode {
			t.Fatalf("ret_code: got %d want %d", upstream.RetCode, retCode)
		}
		if !strings.Contains(upstream.Message, "org-default") {
			t.Fatalf("ret_code=%d: hint should mention current project, got %q", retCode, upstream.Message)
		}
		if !strings.Contains(upstream.Message, "project_id") || !strings.Contains(upstream.Message, "udb_mysql_list_projects") {
			t.Fatalf("ret_code=%d: hint should point to project_id and udb_mysql_list_projects, got %q", retCode, upstream.Message)
		}
		if !strings.Contains(upstream.Message, "describe error") {
			t.Fatalf("ret_code=%d: original message should be preserved, got %q", retCode, upstream.Message)
		}
		if !errors.Is(enhanced, mapped) {
			t.Fatalf("ret_code=%d: expected errors.Is to find original cause", retCode)
		}
	}
}

func TestEnhanceProjectScopeErrorLeavesOtherRetCodesUntouched(t *testing.T) {
	reqCtx := CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "org-default", Region: "cn-bj2"}

	mapped := mapSDKError("DescribeUDBInstance", uerr.NewServerCodeError(230, "ZoneID required"))
	enhanced := enhanceProjectScopeError(mapped, reqCtx)
	if enhanced.Error() != mapped.Error() {
		t.Fatalf("expected unchanged error, got %q want %q", enhanced.Error(), mapped.Error())
	}

	localErr := &InvalidInputError{Field: "db_id", Message: "is required"}
	if enhanced := enhanceProjectScopeError(localErr, reqCtx); enhanced != error(localErr) {
		t.Fatalf("non-upstream error must pass through unchanged, got %v", enhanced)
	}
}

func TestEnhanceProjectScopeErrorDefaultProjectWording(t *testing.T) {
	reqCtx := CallContext{PublicKey: "pk", PrivateKey: "sk", Region: "cn-bj2"}

	mapped := mapSDKError("DescribeUDBInstanceState", uerr.NewServerCodeError(7009, "describe error"))
	enhanced := enhanceProjectScopeError(mapped, reqCtx)

	var upstream *UpstreamError
	if !errors.As(enhanced, &upstream) {
		t.Fatalf("expected *UpstreamError, got %T: %v", enhanced, enhanced)
	}
	if !strings.Contains(upstream.Message, "the default project") {
		t.Fatalf("empty project should render as the default project, got %q", upstream.Message)
	}
	if strings.Contains(upstream.Message, `""`) {
		t.Fatalf("empty project must not render as empty quotes, got %q", upstream.Message)
	}
}
