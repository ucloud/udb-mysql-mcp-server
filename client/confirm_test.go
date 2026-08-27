package client_test

import (
	"errors"
	"testing"

	"udb-mysql-mcp-server/client"
)

func TestValidateRequiredConfirmNil(t *testing.T) {
	err := client.ValidateRequiredConfirm(nil)
	var invalid *client.InvalidInputError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
	}
	if invalid.Field != "confirm" || invalid.Message != "is required" {
		t.Fatalf("invalid: %+v", invalid)
	}
}

func TestValidateRequiredConfirmFalse(t *testing.T) {
	confirm := false
	err := client.ValidateRequiredConfirm(&confirm)
	var invalid *client.InvalidInputError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidInputError, got %T: %v", err, err)
	}
	if invalid.Field != "confirm" || invalid.Message != "must be true" {
		t.Fatalf("invalid: %+v", invalid)
	}
}

func TestValidateRequiredConfirmTrue(t *testing.T) {
	confirm := true
	if err := client.ValidateRequiredConfirm(&confirm); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
