package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"

	"udb-mysql-mcp-server/types"
)

const (
	udbBackupListClassType               = "sql"
	describeUDBBackupAction              = "DescribeUDBBackup"
	describeUDBInstanceBackupStateAction = "DescribeUDBInstanceBackupState"
	describeUDBInstanceBackupURLAction   = "DescribeUDBInstanceBackupURL"
	describeUDBBackupStrategyAction      = "DescribeUDBBackupStrategy"
	backupUDBInstanceAction              = "BackupUDBInstance"
	deleteUDBBackupAction                = "DeleteUDBBackup"

	backupLookupPageSize = 100
	backupLookupMaxPages = 50

	followUpBackupStateTool = "udb_mysql_get_backup_state"
	followUpListBackupsTool = "udb_mysql_list_backups"
	genericBackupPollHint   = "Poll udb_mysql_get_backup_state until the backup reaches a stable state; stop on Failed or Expired."
	missingBackupIDListHint = "Call udb_mysql_list_backups with the provided list_backups_follow_up parameters, then match backup_name on the current page items (API does not filter by BackupName server-side)."
)

// ListBackups lists backups with pagination and optional API filters.
func (c *Client) ListBackups(ctx context.Context, reqCtx CallContext, in types.ListBackupsInput) (types.ListBackupsOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.ListBackupsOutput{}, err
	}
	page, err := normalizePagination(in.Page, in.PageSize)
	if err != nil {
		return types.ListBackupsOutput{}, err
	}
	if in.BeginTime > 0 && in.EndTime > 0 && in.BeginTime > in.EndTime {
		return types.ListBackupsOutput{}, &InvalidInputError{Field: "begin_time", Message: "must be less than or equal to end_time"}
	}
	if in.BeginTime < 0 {
		return types.ListBackupsOutput{}, &InvalidInputError{Field: "begin_time", Message: "must not be negative"}
	}
	if in.EndTime < 0 {
		return types.ListBackupsOutput{}, &InvalidInputError{Field: "end_time", Message: "must not be negative"}
	}
	backupType, backupTypeSet, err := parseBackupType(in.BackupType)
	if err != nil {
		return types.ListBackupsOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.ListBackupsOutput{}, err
	}

	sdkReq := client.NewDescribeUDBBackupRequest()
	sdkReq.ClassType = ucloud.String(udbBackupListClassType)
	sdkReq.Offset = ucloud.Int(page.Offset)
	sdkReq.Limit = ucloud.Int(page.PageSize)
	if zone := strings.TrimSpace(in.Zone); zone != "" {
		sdkReq.Zone = ucloud.String(zone)
	}
	if dbID := strings.TrimSpace(in.DBID); dbID != "" {
		sdkReq.DBId = ucloud.String(dbID)
	}
	if backupTypeSet {
		sdkReq.BackupType = ucloud.Int(int(backupType))
	}
	if in.BeginTime > 0 {
		sdkReq.BeginTime = ucloud.Int(int(in.BeginTime))
	}
	if in.EndTime > 0 {
		sdkReq.EndTime = ucloud.Int(int(in.EndTime))
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBBackup(sdkReq)
	if err != nil {
		return types.ListBackupsOutput{}, mapSDKError(describeUDBBackupAction, err)
	}

	apiTotal := 0
	if resp != nil {
		apiTotal = resp.TotalCount
	}
	items := make([]types.BackupSummary, 0)
	if resp != nil {
		for _, item := range resp.DataSet {
			items = append(items, mapBackupSummary(item))
		}
	}
	return types.ListBackupsOutput{
		Items:         items,
		APITotalCount: apiTotal,
		ReturnedCount: len(items),
		Page:          page.Page,
		PageSize:      page.PageSize,
	}, nil
}

// GetBackupState returns raw backup state from DescribeUDBInstanceBackupState.
func (c *Client) GetBackupState(ctx context.Context, reqCtx CallContext, in types.GetBackupStateInput) (types.GetBackupStateOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.GetBackupStateOutput{}, err
	}
	if in.BackupID <= 0 {
		return types.GetBackupStateOutput{}, &InvalidInputError{Field: "backup_id", Message: "must be positive"}
	}
	zone := strings.TrimSpace(in.Zone)
	if zone == "" {
		return types.GetBackupStateOutput{}, &InvalidInputError{Field: "zone", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.GetBackupStateOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstanceBackupStateRequest()
	sdkReq.BackupId = ucloud.Int(in.BackupID)
	sdkReq.Zone = ucloud.String(zone)
	if backupZone := strings.TrimSpace(in.BackupZone); backupZone != "" {
		sdkReq.BackupZone = ucloud.String(backupZone)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstanceBackupState(sdkReq)
	if err != nil {
		return types.GetBackupStateOutput{}, mapSDKError(describeUDBInstanceBackupStateAction, err)
	}
	if resp == nil {
		return types.GetBackupStateOutput{}, &UpstreamError{Action: describeUDBInstanceBackupStateAction, Message: "empty response"}
	}
	return types.GetBackupStateOutput{
		BackupID:          in.BackupID,
		State:             strings.TrimSpace(resp.State),
		FollowUpTool:      followUpBackupStateTool,
		Hint:              genericBackupPollHint,
		BackupEndTimeUnix: resp.BackupEndTime,
		BackupSizeBytes:   resp.BackupSize,
	}, nil
}

// GetBackupURL returns backup download URLs without exposing secrets in errors.
func (c *Client) GetBackupURL(ctx context.Context, reqCtx CallContext, in types.GetBackupURLInput) (types.GetBackupURLOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.GetBackupURLOutput{}, err
	}
	if in.BackupID <= 0 {
		return types.GetBackupURLOutput{}, &InvalidInputError{Field: "backup_id", Message: "must be positive"}
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.GetBackupURLOutput{}, err
	}
	if in.ValidTime < 0 {
		return types.GetBackupURLOutput{}, &InvalidInputError{Field: "valid_time", Message: "must not be negative"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.GetBackupURLOutput{}, err
	}

	// DescribeUDBInstanceBackupURL rejects requests without Zone upstream.
	// Resolve it from the instance so callers can omit zone entirely.
	zone := strings.TrimSpace(in.Zone)
	if zone == "" {
		inst, err := c.GetInstance(ctx, reqCtx, types.GetInstanceInput{DBID: dbID})
		if err != nil {
			return types.GetBackupURLOutput{}, err
		}
		zone = strings.TrimSpace(inst.Zone)
		if zone == "" {
			return types.GetBackupURLOutput{}, &UpstreamError{
				Action:  describeUDBInstanceBackupURLAction,
				Message: "instance zone is empty; pass zone explicitly",
			}
		}
	}

	sdkReq := client.NewDescribeUDBInstanceBackupURLRequest()
	sdkReq.BackupId = ucloud.Int(in.BackupID)
	sdkReq.DBId = ucloud.String(dbID)
	sdkReq.Zone = ucloud.String(zone)
	if in.ValidTime > 0 {
		sdkReq.ValidTime = ucloud.Int(in.ValidTime)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstanceBackupURL(sdkReq)
	if err != nil {
		return types.GetBackupURLOutput{}, mapSDKError(describeUDBInstanceBackupURLAction, err)
	}
	if resp == nil {
		return types.GetBackupURLOutput{}, &UpstreamError{Action: describeUDBInstanceBackupURLAction, Message: "empty response"}
	}
	return types.GetBackupURLOutput{
		BackupID:         in.BackupID,
		DBID:             dbID,
		PublicBackupPath: resp.BackupPath,
		InnerBackupPath:  resp.InnerBackupPath,
		Checksum:         resp.MD5,
	}, nil
}

// GetBackupStrategy returns automatic backup strategy; UFile output is bucket only.
func (c *Client) GetBackupStrategy(ctx context.Context, reqCtx CallContext, in types.GetBackupStrategyInput) (types.GetBackupStrategyOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.GetBackupStrategyOutput{}, err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.GetBackupStrategyOutput{}, err
	}
	zone := strings.TrimSpace(in.Zone)
	if zone == "" {
		return types.GetBackupStrategyOutput{}, &InvalidInputError{Field: "zone", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.GetBackupStrategyOutput{}, err
	}

	sdkReq := client.NewDescribeUDBBackupStrategyRequest()
	sdkReq.DBId = ucloud.String(dbID)
	sdkReq.Zone = ucloud.String(zone)
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBBackupStrategy(sdkReq)
	if err != nil {
		return types.GetBackupStrategyOutput{}, enhanceProjectScopeError(mapSDKError(describeUDBBackupStrategyAction, err), reqCtx)
	}
	if resp == nil {
		return types.GetBackupStrategyOutput{}, &UpstreamError{Action: describeUDBBackupStrategyAction, Message: "empty response"}
	}
	out := types.GetBackupStrategyOutput{
		DBID:            dbID,
		BackupBeginHour: resp.BackupBeginTime,
		BackupDate:      resp.BackupDate,
		BackupMethod:    resp.BackupMethod,
		SaveDays:        resp.SaveDays,
	}
	if bucket := strings.TrimSpace(resp.UserUFileData.Bucket); bucket != "" {
		out.UserUFile.Bucket = bucket
	}
	return out, nil
}

// CreateBackup starts a manual backup without pre-reading instance state.
func (c *Client) CreateBackup(ctx context.Context, reqCtx CallContext, in types.CreateBackupInput, requestedAt time.Time) (types.CreateBackupOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.CreateBackupOutput{}, err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.CreateBackupOutput{}, err
	}
	backupName := strings.TrimSpace(in.BackupName)
	if backupName == "" {
		return types.CreateBackupOutput{}, &InvalidInputError{Field: "backup_name", Message: "is required"}
	}
	if requestedAt.IsZero() {
		return types.CreateBackupOutput{}, &InvalidInputError{Field: "requested_at", Message: "must be captured before BackupUDBInstance"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.CreateBackupOutput{}, err
	}

	sdkReq := client.NewBackupUDBInstanceRequest()
	sdkReq.DBId = ucloud.String(dbID)
	sdkReq.BackupName = ucloud.String(backupName)
	if zone := strings.TrimSpace(in.Zone); zone != "" {
		sdkReq.Zone = ucloud.String(zone)
	}
	if method := strings.TrimSpace(in.BackupMethod); method != "" {
		sdkReq.BackupMethod = ucloud.String(method)
	}
	if blacklist := strings.TrimSpace(in.Blacklist); blacklist != "" {
		sdkReq.Blacklist = ucloud.String(blacklist)
	}
	if in.ForceBackup {
		sdkReq.ForceBackup = ucloud.Bool(true)
	}
	if in.UseBlacklist {
		sdkReq.UseBlacklist = ucloud.Bool(true)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.BackupUDBInstance(sdkReq)
	if err != nil {
		return types.CreateBackupOutput{}, mapSDKError(backupUDBInstanceAction, err)
	}

	zone := strings.TrimSpace(in.Zone)
	requestedUnix := requestedAt.UTC().Unix()
	if resp != nil && resp.BackupId > 0 {
		return types.CreateBackupOutput{
			DBID:            dbID,
			BackupName:      backupName,
			Zone:            zone,
			BackupID:        resp.BackupId,
			BackupIDKnown:   true,
			RequestedAtUnix: requestedUnix,
			FollowUpTool:    followUpBackupStateTool,
			Hint:            fmt.Sprintf("Poll %s until Success, Failed, or Expired.", followUpBackupStateTool),
		}, nil
	}

	followUp := &types.ListBackupsFollowUp{
		DBID:            dbID,
		BeginTime:       requestedUnix,
		BackupType:      "manual",
		Zone:            zone,
		MatchBackupName: backupName,
	}
	return types.CreateBackupOutput{
		DBID:                dbID,
		BackupName:          backupName,
		Zone:                zone,
		BackupIDKnown:       false,
		RequestedAtUnix:     requestedUnix,
		FollowUpTool:        followUpListBackupsTool,
		Hint:                fmt.Sprintf("%s Then call %s once backup_id is known.", missingBackupIDListHint, followUpBackupStateTool),
		ListBackupsFollowUp: followUp,
	}, nil
}

// FindBackupByID scans paginated DescribeUDBBackup results until a match or conclusive absence.
func (c *Client) FindBackupByID(ctx context.Context, reqCtx CallContext, backupID int, zone, backupZone string) (types.BackupSummary, error) {
	if backupID <= 0 {
		return types.BackupSummary{}, &InvalidInputError{Field: "backup_id", Message: "is required"}
	}

	var (
		baselineTotal int
		hasBaseline   bool
		itemsScanned  int
		scanComplete  bool
		lastPage      int
	)

	for page := 1; page <= backupLookupMaxPages; page++ {
		lastPage = page
		pageResult, err := c.ListBackups(ctx, reqCtx, types.ListBackupsInput{
			Page: page, PageSize: backupLookupPageSize, Zone: zone,
		})
		if err != nil {
			return types.BackupSummary{}, err
		}

		pageTotal := pageResult.APITotalCount
		if !hasBaseline {
			baselineTotal = pageTotal
			hasBaseline = true
		} else if pageTotal != baselineTotal {
			return types.BackupSummary{}, &ScanIncompleteError{
				BackupID: backupID, APITotalCount: pageTotal, ScannedCount: itemsScanned,
				Reason: ScanIncompleteReasonCountChanged,
			}
		}

		itemCount := len(pageResult.Items)
		if itemCount == 0 {
			if baselineTotal > 0 && itemsScanned < baselineTotal {
				return types.BackupSummary{}, &ScanIncompleteError{
					BackupID: backupID, APITotalCount: baselineTotal, ScannedCount: itemsScanned,
					Reason: ScanIncompleteReasonEmptyPage,
				}
			}
			scanComplete = itemsScanned >= baselineTotal
			break
		}

		itemsScanned += itemCount
		for _, item := range pageResult.Items {
			if item.BackupID != backupID {
				continue
			}
			if want := strings.TrimSpace(backupZone); want != "" && item.BackupZone != want {
				continue
			}
			return item, nil
		}

		if itemCount < backupLookupPageSize {
			if itemsScanned >= baselineTotal {
				scanComplete = true
				break
			}
			continue
		}
		if itemsScanned >= baselineTotal {
			scanComplete = true
			break
		}
	}

	if !scanComplete {
		reason := ScanIncompleteReasonIncomplete
		if lastPage >= backupLookupMaxPages {
			reason = ScanIncompleteReasonLimit
		}
		return types.BackupSummary{}, &ScanIncompleteError{
			BackupID: backupID, APITotalCount: baselineTotal, ScannedCount: itemsScanned,
			Reason: reason,
		}
	}
	return types.BackupSummary{}, &BackupNotFoundError{BackupID: backupID}
}

// DeleteBackup deletes a backup by ID.
func (c *Client) DeleteBackup(ctx context.Context, reqCtx CallContext, backupID int, zone, backupZone string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if backupID <= 0 {
		return &InvalidInputError{Field: "backup_id", Message: "is required"}
	}
	zone = strings.TrimSpace(zone)
	if zone == "" {
		return &InvalidInputError{Field: "zone", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return err
	}

	req := client.NewDeleteUDBBackupRequest()
	req.BackupId = ucloud.Int(backupID)
	req.Zone = ucloud.String(zone)
	if v := strings.TrimSpace(backupZone); v != "" {
		req.BackupZone = ucloud.String(v)
	}
	prepareRequest(req, ctx, c.factory.Timeout)

	_, err = client.DeleteUDBBackup(req)
	if err != nil {
		return mapSDKError(deleteUDBBackupAction, err)
	}
	return nil
}

func mapBackupSummary(item udb.UDBBackupSet) types.BackupSummary {
	return types.BackupSummary{
		BackupID:        item.BackupId,
		BackupName:      item.BackupName,
		BackupTimeUnix:  int64(item.BackupTime),
		BackupEndUnix:   item.BackupEndTime,
		BackupSizeBytes: item.BackupSize,
		BackupType:      backupTypeLabel(item.BackupType),
		State:           strings.TrimSpace(item.State),
		ErrorInfo:       item.ErrorInfo,
		DBID:            item.DBId,
		DBName:          item.DBName,
		Zone:            item.Zone,
		BackupZone:      item.BackupZone,
		Checksum:        item.MD5,
	}
}

func backupTypeLabel(raw int) string {
	switch raw {
	case 0:
		return "auto"
	case 1:
		return "manual"
	default:
		return ""
	}
}

func parseBackupType(raw string) (int, bool, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return 0, false, nil
	}
	switch value {
	case "auto":
		return 0, true, nil
	case "manual":
		return 1, true, nil
	default:
		return 0, false, &InvalidInputError{Field: "backup_type", Message: "must be auto or manual"}
	}
}
