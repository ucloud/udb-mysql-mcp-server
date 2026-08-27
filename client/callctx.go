package client

import (
	"fmt"
	"os"
	"strings"
)

// CallContext is the local-process identity and API scope for one cloud call.
// Credentials always come from the process environment (or a test resolver).
type CallContext struct {
	PublicKey  string
	PrivateKey string
	ProjectID  string
	Region     string
}

// CredentialsError indicates signing keys are missing.
type CredentialsError struct{}

func (e *CredentialsError) Error() string {
	return "UCloud public key and private key are required before calling cloud tools"
}

// ResolveFromEnv reads UCLOUD_* process environment variables.
// Region falls back to defaultRegion when UCLOUD_REGION is unset.
func ResolveFromEnv(defaultRegion string) (CallContext, error) {
	publicKey := strings.TrimSpace(os.Getenv("UCLOUD_PUBLIC_KEY"))
	privateKey := strings.TrimSpace(os.Getenv("UCLOUD_PRIVATE_KEY"))
	if publicKey == "" || privateKey == "" {
		return CallContext{}, &CredentialsError{}
	}
	region := strings.TrimSpace(os.Getenv("UCLOUD_REGION"))
	if region == "" {
		region = strings.TrimSpace(defaultRegion)
	}
	return CallContext{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		ProjectID:  strings.TrimSpace(os.Getenv("UCLOUD_PROJECT_ID")),
		Region:     region,
	}, nil
}

// WithScope overlays per-call project/region arguments onto env defaults.
func (c CallContext) WithScope(projectID, region string, requireProject bool) (CallContext, error) {
	if p := strings.TrimSpace(projectID); p != "" {
		c.ProjectID = p
	}
	if r := strings.TrimSpace(region); r != "" {
		c.Region = r
	}
	if requireProject && strings.TrimSpace(c.ProjectID) == "" {
		return CallContext{}, &InvalidInputError{Field: "project_id", Message: "is required"}
	}
	if strings.TrimSpace(c.Region) == "" {
		return CallContext{}, &InvalidInputError{Field: "region", Message: "is required"}
	}
	if strings.TrimSpace(c.PublicKey) == "" || strings.TrimSpace(c.PrivateKey) == "" {
		return CallContext{}, &CredentialsError{}
	}
	return c, nil
}

// ConflictError indicates a destructive confirm name does not match live state.
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	if e.Message == "" {
		return "provided name does not exactly match the current resource name"
	}
	return e.Message
}

// ValidateNameMatch requires an exact live name match for destructive operations.
func ValidateNameMatch(resourceID, expectedName, currentName string) error {
	if strings.TrimSpace(resourceID) == "" {
		return &ConflictError{Message: "resource id is required"}
	}
	expected := strings.TrimSpace(expectedName)
	current := strings.TrimSpace(currentName)
	if expected == "" || current == "" || expected != current {
		return &ConflictError{}
	}
	return nil
}

func (c CallContext) String() string {
	return fmt.Sprintf("CallContext{region=%s project=%s}", c.Region, c.ProjectID)
}
