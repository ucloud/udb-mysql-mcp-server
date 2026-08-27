package guard

import (
	"context"
	"fmt"

	"udb-mysql-mcp-server/client"
	"udb-mysql-mcp-server/types"
)

// ConfirmBackupDelete validates confirm, locates the live backup, and checks the name match.
func ConfirmBackupDelete(
	ctx context.Context,
	req client.CallContext,
	api *client.Client,
	in types.DeleteBackupInput,
) (types.BackupSummary, error) {
	if err := client.ValidateRequiredConfirm(in.Confirm); err != nil {
		return types.BackupSummary{}, err
	}
	current, err := api.FindBackupByID(ctx, req, in.BackupID, in.Zone, in.BackupZone)
	if err != nil {
		return types.BackupSummary{}, err
	}
	err = client.ValidateNameMatch(fmt.Sprintf("%d", in.BackupID), in.Name, current.BackupName)
	return current, err
}
