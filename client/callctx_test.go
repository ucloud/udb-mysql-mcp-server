package client

import (
	"errors"
	"testing"
)

func TestResolveFromEnv(t *testing.T) {
	t.Setenv("UCLOUD_PUBLIC_KEY", "pk")
	t.Setenv("UCLOUD_PRIVATE_KEY", "sk")
	t.Setenv("UCLOUD_PROJECT_ID", "org-1")
	t.Setenv("UCLOUD_REGION", "")
	got, err := ResolveFromEnv("cn-bj2")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got.PublicKey != "pk" || got.ProjectID != "org-1" || got.Region != "cn-bj2" {
		t.Fatalf("got %+v", got)
	}
}

func TestResolveFromEnvMissingKeys(t *testing.T) {
	t.Setenv("UCLOUD_PUBLIC_KEY", "")
	t.Setenv("UCLOUD_PRIVATE_KEY", "")
	_, err := ResolveFromEnv("cn-bj2")
	var cred *CredentialsError
	if !errors.As(err, &cred) {
		t.Fatalf("got %T: %v", err, err)
	}
}

func TestWithScopeOverlay(t *testing.T) {
	base := CallContext{PublicKey: "pk", PrivateKey: "sk", ProjectID: "env-org", Region: "cn-bj2"}
	got, err := base.WithScope("org-2", "cn-sh2", true)
	if err != nil {
		t.Fatalf("scope: %v", err)
	}
	if got.ProjectID != "org-2" || got.Region != "cn-sh2" {
		t.Fatalf("got %+v", got)
	}
}
