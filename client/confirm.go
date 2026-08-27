package client

// ValidateRequiredConfirm rejects missing or false destructive confirmations.
func ValidateRequiredConfirm(confirm *bool) error {
	if confirm == nil {
		return &InvalidInputError{Field: "confirm", Message: "is required"}
	}
	if !*confirm {
		return &InvalidInputError{Field: "confirm", Message: "must be true"}
	}
	return nil
}
