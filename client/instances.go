package client

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"

	"udb-mysql-mcp-server/types"
)

const (
	// DescribeUDBInstance list ClassType values are SQL / NOSQL / postgresql
	// (docs phrase this as "mysql: SQL"; live API rejects "mysql").
	udbInstanceListClassType        = "SQL"
	describeUDBInstanceAction       = "DescribeUDBInstance"
	describeUDBInstanceStateAction  = "DescribeUDBInstanceState"
	startUDBInstanceAction          = "StartUDBInstance"
	stopUDBInstanceAction           = "StopUDBInstance"
	restartUDBInstanceAction        = "RestartUDBInstance"
	modifyUDBInstanceNameAction     = "ModifyUDBInstanceName"
	createUDBMySQLInstanceAction    = "CreateUDBMySQLInstance"
	resizeUDBInstanceAction         = "ResizeUDBInstance"
	modifyUDBInstancePasswordAction = "ModifyUDBInstancePassword"

	defaultListPage     = 1
	defaultListPageSize = 20
	maxListPageSize     = 100

	followUpInstanceStateTool = "udb_mysql_get_instance_state"
	genericStatePollHint      = "Poll udb_mysql_get_instance_state until the instance reaches a stable state; stop on Fail or Recover fail."
)

type pageParams struct {
	Page     int
	PageSize int
	Offset   int
}

func normalizePagination(page, pageSize int) (pageParams, error) {
	if page <= 0 {
		page = defaultListPage
	}
	if pageSize <= 0 {
		pageSize = defaultListPageSize
	}
	if pageSize > maxListPageSize {
		pageSize = maxListPageSize
	}
	offset := 0
	if page > 1 {
		prev := page - 1
		if prev > math.MaxInt/pageSize {
			return pageParams{}, &InvalidInputError{Field: "page", Message: "page offset overflow"}
		}
		offset = prev * pageSize
	}
	return pageParams{Page: page, PageSize: pageSize, Offset: offset}, nil
}

func requireDBID(dbID string) (string, error) {
	id := strings.TrimSpace(dbID)
	if id == "" {
		return "", &InvalidInputError{Field: "db_id", Message: "is required"}
	}
	return id, nil
}

func (c *Client) sdkClient(reqCtx CallContext) (*udb.UDBClient, error) {
	if c.factory == nil {
		return nil, &InvalidInputError{Field: "factory", Message: "must not be nil"}
	}
	return c.factory.newUDBClient(reqCtx), nil
}

// GetInstance loads a single instance via DescribeUDBInstance.
func (c *Client) GetInstance(ctx context.Context, reqCtx CallContext, in types.GetInstanceInput) (types.InstanceOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.InstanceOutput{}, err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.InstanceOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.InstanceOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstanceRequest()
	sdkReq.DBId = ucloud.String(dbID)
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstance(sdkReq)
	if err != nil {
		return types.InstanceOutput{}, enhanceProjectScopeError(mapSDKError(describeUDBInstanceAction, err), reqCtx)
	}
	if resp == nil || len(resp.DataSet) == 0 {
		return types.InstanceOutput{}, &NotFoundError{InstanceID: dbID}
	}

	for _, item := range resp.DataSet {
		if item.DBId == dbID {
			return mapInstanceOutput(item), nil
		}
	}
	return types.InstanceOutput{}, &NotFoundError{InstanceID: dbID}
}

// ListInstances lists instances with pagination and current-page post-filters.
func (c *Client) ListInstances(ctx context.Context, reqCtx CallContext, in types.ListInstancesInput) (types.ListInstancesOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.ListInstancesOutput{}, err
	}

	page, err := normalizePagination(in.Page, in.PageSize)
	if err != nil {
		return types.ListInstancesOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.ListInstancesOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstanceRequest()
	sdkReq.ClassType = ucloud.String(udbInstanceListClassType)
	sdkReq.Offset = ucloud.Int(page.Offset)
	sdkReq.Limit = ucloud.Int(page.PageSize)
	if zone := strings.TrimSpace(in.Zone); zone != "" {
		sdkReq.Zone = ucloud.String(zone)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstance(sdkReq)
	if err != nil {
		return types.ListInstancesOutput{}, mapSDKError(describeUDBInstanceAction, err)
	}

	apiTotal := 0
	if resp != nil {
		apiTotal = resp.TotalCount
	}

	items := make([]any, 0)
	if resp != nil {
		for _, item := range resp.DataSet {
			if !in.IncludeSlaves && isMySQLSlave(item.Role, item.SrcDBId) {
				continue
			}
			out := mapInstanceOutput(item)
			if !matchesListFilters(out, in.NameContains, in.State) {
				continue
			}
			if in.Verbose {
				items = append(items, out)
			} else {
				items = append(items, mapInstanceSummary(out))
			}
		}
	}

	result := types.ListInstancesOutput{
		Items:         items,
		APITotalCount: apiTotal,
		ReturnedCount: len(items),
		Page:          page.Page,
		PageSize:      page.PageSize,
		Region:        reqCtx.Region,
		ProjectID:     reqCtx.ProjectID,
	}
	if strings.TrimSpace(in.NameContains) != "" || strings.TrimSpace(in.State) != "" {
		v := true
		result.PostFiltered = &v
		result.FilterScope = "current_page"
	}
	return result, nil
}

// GetInstanceState returns the raw upstream state with generic polling guidance.
func (c *Client) GetInstanceState(ctx context.Context, reqCtx CallContext, in types.GetInstanceStateInput) (types.GetInstanceStateOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.GetInstanceStateOutput{}, err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.GetInstanceStateOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.GetInstanceStateOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstanceStateRequest()
	sdkReq.DBId = ucloud.String(dbID)
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstanceState(sdkReq)
	if err != nil {
		return types.GetInstanceStateOutput{}, enhanceProjectScopeError(mapSDKError(describeUDBInstanceStateAction, err), reqCtx)
	}
	if resp == nil {
		return types.GetInstanceStateOutput{}, &UpstreamError{Action: describeUDBInstanceStateAction, Message: "empty response"}
	}

	return types.GetInstanceStateOutput{
		DBID:         dbID,
		State:        strings.TrimSpace(resp.State),
		FollowUpTool: followUpInstanceStateTool,
		Hint:         genericStatePollHint,
	}, nil
}

// StartInstance starts a UDB instance.
func (c *Client) StartInstance(ctx context.Context, reqCtx CallContext, in types.LifecycleInput) error {
	return c.lifecycleMutation(ctx, reqCtx, startUDBInstanceAction, in.DBID, func(client *udb.UDBClient) error {
		req := client.NewStartUDBInstanceRequest()
		req.DBId = ucloud.String(strings.TrimSpace(in.DBID))
		if z := strings.TrimSpace(in.Zone); z != "" {
			req.Zone = ucloud.String(z)
		}
		prepareRequest(req, ctx, c.factory.Timeout)
		_, err := client.StartUDBInstance(req)
		return err
	})
}

// StopInstance stops a UDB instance.
func (c *Client) StopInstance(ctx context.Context, reqCtx CallContext, in types.StopInstanceInput) error {
	return c.lifecycleMutation(ctx, reqCtx, stopUDBInstanceAction, in.DBID, func(client *udb.UDBClient) error {
		req := client.NewStopUDBInstanceRequest()
		req.DBId = ucloud.String(strings.TrimSpace(in.DBID))
		if z := strings.TrimSpace(in.Zone); z != "" {
			req.Zone = ucloud.String(z)
		}
		if in.ForceToKill {
			req.ForceToKill = ucloud.Bool(true)
		}
		prepareRequest(req, ctx, c.factory.Timeout)
		_, err := client.StopUDBInstance(req)
		return err
	})
}

// RestartInstance restarts a UDB instance.
func (c *Client) RestartInstance(ctx context.Context, reqCtx CallContext, in types.LifecycleInput) error {
	return c.lifecycleMutation(ctx, reqCtx, restartUDBInstanceAction, in.DBID, func(client *udb.UDBClient) error {
		req := client.NewRestartUDBInstanceRequest()
		req.DBId = ucloud.String(strings.TrimSpace(in.DBID))
		if z := strings.TrimSpace(in.Zone); z != "" {
			req.Zone = ucloud.String(z)
		}
		prepareRequest(req, ctx, c.factory.Timeout)
		_, err := client.RestartUDBInstance(req)
		return err
	})
}

// ModifyName changes an instance display name.
func (c *Client) ModifyName(ctx context.Context, reqCtx CallContext, in types.ModifyNameInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return &InvalidInputError{Field: "name", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return err
	}

	req := client.NewModifyUDBInstanceNameRequest()
	req.DBId = ucloud.String(dbID)
	req.Name = ucloud.String(name)
	if z := strings.TrimSpace(in.Zone); z != "" {
		req.Zone = ucloud.String(z)
	}
	prepareRequest(req, ctx, c.factory.Timeout)
	_, err = client.ModifyUDBInstanceName(req)
	if err != nil {
		return mapSDKError(modifyUDBInstanceNameAction, err)
	}
	return nil
}

// CreateInstance provisions a new UDB MySQL instance.
func (c *Client) CreateInstance(ctx context.Context, reqCtx CallContext, in types.CreateInstanceInput) (types.CreateInstanceOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.CreateInstanceOutput{}, err
	}
	if err := validateCreateInstanceInput(in); err != nil {
		return types.CreateInstanceOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.CreateInstanceOutput{}, err
	}

	req := client.NewCreateUDBMySQLInstanceRequest()
	req.Zone = ucloud.String(strings.TrimSpace(in.Zone))
	req.Name = ucloud.String(strings.TrimSpace(in.Name))
	req.AdminPassword = ucloud.String(in.AdminPassword)
	req.DBTypeId = ucloud.String(strings.TrimSpace(in.DBTypeID))
	req.DiskSpace = ucloud.Int(in.DiskSpaceGB)
	req.ParamGroupId = ucloud.Int(in.ParamGroupID)
	req.MachineType = ucloud.String(strings.TrimSpace(in.MachineType))
	req.StorageClass = ucloud.String(strings.TrimSpace(in.StorageClass))
	req.SpecificationClass = ucloud.String(strings.TrimSpace(in.SpecificationClass))
	// Live CreateUDBMySQLInstance requires Port (ret_code=210 when omitted).
	port := in.Port
	if port <= 0 {
		port = 3306
	}
	req.Port = ucloud.Int(port)
	if v := strings.TrimSpace(in.ChargeType); v != "" {
		req.ChargeType = ucloud.String(v)
	}
	if in.Quantity != nil {
		req.Quantity = ucloud.Int(*in.Quantity)
	}
	if v := strings.TrimSpace(in.InstanceMode); v != "" {
		req.InstanceMode = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.VPCID); v != "" {
		req.VPCId = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.SubnetID); v != "" {
		req.SubnetId = ucloud.String(v)
	}
	if in.BackupID > 0 {
		req.BackupId = ucloud.Int(in.BackupID)
	}
	if v := strings.TrimSpace(in.BackupZone); v != "" {
		req.BackupZone = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.DBSubVersion); v != "" {
		req.DBSubVersion = ucloud.String(v)
	}
	prepareRequest(req, ctx, c.factory.Timeout)

	resp, err := client.CreateUDBMySQLInstance(req)
	if err != nil {
		return types.CreateInstanceOutput{}, mapSDKError(createUDBMySQLInstanceAction, err)
	}
	if resp == nil || strings.TrimSpace(resp.DBId) == "" {
		return types.CreateInstanceOutput{}, &UpstreamError{Action: createUDBMySQLInstanceAction, Message: "create instance returned empty DBId"}
	}
	return types.CreateInstanceOutput{
		DBID:         resp.DBId,
		FollowUpTool: followUpInstanceStateTool,
		Hint:         genericStatePollHint,
	}, nil
}

// ResizeInstance changes instance capacity.
func (c *Client) ResizeInstance(ctx context.Context, reqCtx CallContext, in types.ResizeInstanceInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateResizeInstanceInput(in); err != nil {
		return err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return err
	}

	req := client.NewResizeUDBInstanceRequest()
	req.DBId = ucloud.String(strings.TrimSpace(in.DBID))
	req.MemoryLimit = ucloud.Int(in.MemoryLimitMB)
	req.DiskSpace = ucloud.Int(in.DiskSpaceGB)
	if z := strings.TrimSpace(in.Zone); z != "" {
		req.Zone = ucloud.String(z)
	}
	if in.CPU > 0 {
		req.CPU = ucloud.Int(in.CPU)
	}
	if v := strings.TrimSpace(in.InstanceType); v != "" {
		req.InstanceType = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.MachineType); v != "" {
		req.MachineType = ucloud.String(v)
	}
	if in.SpecificationType > 0 {
		req.SpecificationType = ucloud.String(strconv.Itoa(in.SpecificationType))
	}
	prepareRequest(req, ctx, c.factory.Timeout)
	_, err = client.ResizeUDBInstance(req)
	if err != nil {
		return mapSDKError(resizeUDBInstanceAction, err)
	}
	return nil
}

// ResetPassword changes the administrator password.
func (c *Client) ResetPassword(ctx context.Context, reqCtx CallContext, in types.ResetPasswordInput) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Password) == "" {
		return &InvalidInputError{Field: "password", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return err
	}

	req := client.NewModifyUDBInstancePasswordRequest()
	req.DBId = ucloud.String(dbID)
	req.Password = ucloud.String(in.Password)
	prepareRequest(req, ctx, c.factory.Timeout)
	_, err = client.ModifyUDBInstancePassword(req)
	if err != nil {
		return mapSDKError(modifyUDBInstancePasswordAction, err)
	}
	return nil
}

func (c *Client) lifecycleMutation(ctx context.Context, reqCtx CallContext, action, dbID string, call func(*udb.UDBClient) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := requireDBID(dbID); err != nil {
		return err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return err
	}
	if err := call(client); err != nil {
		return mapSDKError(action, err)
	}
	return nil
}

func validateCreateInstanceInput(in types.CreateInstanceInput) error {
	checks := []struct {
		field string
		value string
	}{
		{"zone", in.Zone},
		{"name", in.Name},
		{"admin_password", in.AdminPassword},
		{"db_type_id", in.DBTypeID},
		{"machine_type", in.MachineType},
		{"storage_class", in.StorageClass},
		{"specification_class", in.SpecificationClass},
	}
	for _, c := range checks {
		if strings.TrimSpace(c.value) == "" {
			return &InvalidInputError{Field: c.field, Message: "is required"}
		}
	}
	if in.DiskSpaceGB <= 0 {
		return &InvalidInputError{Field: "disk_space_gb", Message: "is required"}
	}
	if in.ParamGroupID <= 0 {
		return &InvalidInputError{Field: "param_group_id", Message: "is required"}
	}
	return validateVPCSubnetPair(in.VPCID, in.SubnetID)
}

func validateResizeInstanceInput(in types.ResizeInstanceInput) error {
	if _, err := requireDBID(in.DBID); err != nil {
		return err
	}
	if in.MemoryLimitMB <= 0 {
		return &InvalidInputError{Field: "memory_limit_mb", Message: "is required"}
	}
	if in.DiskSpaceGB <= 0 {
		return &InvalidInputError{Field: "disk_space_gb", Message: "is required"}
	}
	return ValidateRequiredConfirm(in.Confirm)
}

func isMySQLSlave(role, srcDBID string) bool {
	if strings.EqualFold(strings.TrimSpace(role), "slave") {
		return true
	}
	return strings.TrimSpace(srcDBID) != ""
}

func matchesListFilters(inst types.InstanceOutput, nameContains, state string) bool {
	if name := strings.TrimSpace(nameContains); name != "" {
		if !strings.Contains(strings.ToLower(inst.Name), strings.ToLower(name)) {
			return false
		}
	}
	if state := strings.TrimSpace(state); state != "" {
		if !strings.EqualFold(inst.State, state) {
			return false
		}
	}
	return true
}

func mapInstanceSummary(out types.InstanceOutput) types.InstanceSummary {
	return types.InstanceSummary{
		DBID:     out.DBID,
		Name:     out.Name,
		State:    out.State,
		DBTypeID: out.DBTypeID,
		Zone:     out.Zone,
	}
}

func mapInstanceOutput(item udb.UDBInstanceSet) types.InstanceOutput {
	out := types.InstanceOutput{
		DBID:         item.DBId,
		Name:         item.Name,
		State:        strings.TrimSpace(item.State),
		DBTypeID:     item.DBTypeId,
		InstanceType: item.InstanceType,
		Address:      item.VirtualIP,
		Port:         item.Port,
		CPU:          item.CPU,
		MemoryMB:     item.MemoryLimit,
		DiskGB:       item.DiskSpace,
		Zone:         item.Zone,
	}
	if item.CreateTime > 0 {
		out.CreatedAt = time.Unix(int64(item.CreateTime), 0).UTC().Format(time.RFC3339)
	}
	return out
}
