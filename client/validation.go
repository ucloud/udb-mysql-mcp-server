package client

import "strings"

const genericUpstreamFailureMessage = "request failed"

// wireCountOrDefault normalizes instance count for upstream wire: zero means default 1.
func wireCountOrDefault(count int) (int, error) {
	if count < 0 {
		return 0, &InvalidInputError{Field: "count", Message: "must not be negative"}
	}
	if count == 0 {
		return 1, nil
	}
	return count, nil
}

func validateVPCSubnetPair(vpcID, subnetID string) error {
	vpc := strings.TrimSpace(vpcID)
	subnet := strings.TrimSpace(subnetID)
	if (vpc == "") == (subnet == "") {
		return nil
	}
	if vpc == "" {
		return &InvalidInputError{Field: "vpc_id", Message: "must be provided together with subnet_id"}
	}
	return &InvalidInputError{Field: "subnet_id", Message: "must be provided together with vpc_id"}
}
