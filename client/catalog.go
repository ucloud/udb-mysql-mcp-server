package client

import (
	"context"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/ucloud"

	"udb-mysql-mcp-server/types"
)

const (
	describeUDBTypeAction                 = "DescribeUDBType"
	listUDBMachineTypeAction              = "ListUDBMachineType"
	describeUDBParamGroupAction           = "DescribeUDBParamGroup"
	describeUDBInstancePriceAction        = "DescribeUDBInstancePrice"
	describeUDBInstanceUpgradePriceAction = "DescribeUDBInstanceUpgradePrice"
	udbParamGroupClassType                = "sql"
)

// ListDBTypes lists supported DB engine versions via DescribeUDBType.
func (c *Client) ListDBTypes(ctx context.Context, reqCtx CallContext, in types.ListDBTypesInput) (types.ListDBTypesOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.ListDBTypesOutput{}, err
	}
	if strings.TrimSpace(in.Zone) == "" {
		return types.ListDBTypesOutput{}, &InvalidInputError{Field: "zone", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.ListDBTypesOutput{}, err
	}

	sdkReq := client.NewDescribeUDBTypeRequest()
	sdkReq.Zone = ucloud.String(strings.TrimSpace(in.Zone))
	if v := strings.TrimSpace(in.BackupZone); v != "" {
		sdkReq.BackupZone = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.DBClusterType); v != "" {
		sdkReq.DBClusterType = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.InstanceMode); v != "" {
		sdkReq.InstanceMode = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.DiskType); v != "" {
		sdkReq.DiskType = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.CompatibleWithDBType); v != "" {
		sdkReq.CompatibleWithDBType = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.DBSubVersion); v != "" {
		sdkReq.DBSubVersion = ucloud.String(v)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBType(sdkReq)
	if err != nil {
		return types.ListDBTypesOutput{}, mapSDKError(describeUDBTypeAction, err)
	}
	if resp == nil {
		return types.ListDBTypesOutput{Items: []types.DBTypeSummary{}}, nil
	}

	out := types.ListDBTypesOutput{Items: make([]types.DBTypeSummary, 0, len(resp.DataSet))}
	recID := strings.TrimSpace(resp.DedaultType.DBTypeId)
	hasRecommended := recID != ""
	if hasRecommended {
		out.RecommendedDBTypeID = recID
		out.RecommendedSubVersion = resp.DedaultType.DBSubVersion
	}
	for _, item := range resp.DataSet {
		entry := types.DBTypeSummary{DBTypeID: item.DBTypeId, DBSubVersion: item.DBSubVersion}
		if hasRecommended && item.DBTypeId == recID {
			entry.Recommended = true
		}
		out.Items = append(out.Items, entry)
	}
	return out, nil
}

// ListMachineTypes lists compute specifications via ListUDBMachineType.
func (c *Client) ListMachineTypes(ctx context.Context, reqCtx CallContext, in types.ListMachineTypesInput) (types.ListMachineTypesOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.ListMachineTypesOutput{}, err
	}
	if strings.TrimSpace(in.Zone) == "" {
		return types.ListMachineTypesOutput{}, &InvalidInputError{Field: "zone", Message: "is required"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.ListMachineTypesOutput{}, err
	}

	sdkReq := client.NewListUDBMachineTypeRequest()
	sdkReq.Zone = ucloud.String(strings.TrimSpace(in.Zone))
	if v := strings.TrimSpace(in.InstanceMode); v != "" {
		sdkReq.InstanceMode = ucloud.String(v)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.ListUDBMachineType(sdkReq)
	if err != nil {
		return types.ListMachineTypesOutput{}, mapSDKError(listUDBMachineTypeAction, err)
	}
	if resp == nil {
		return types.ListMachineTypesOutput{Items: []types.MachineTypeSummary{}}, nil
	}

	items := make([]types.MachineTypeSummary, 0, len(resp.DataSet))
	for _, item := range resp.DataSet {
		items = append(items, mapMachineTypeSummary(item))
	}
	return types.ListMachineTypesOutput{
		Items:   items,
		Default: mapMachineTypeSummary(resp.DefaultMachineType),
	}, nil
}

// ListParamGroups lists parameter groups with pagination via DescribeUDBParamGroup.
func (c *Client) ListParamGroups(ctx context.Context, reqCtx CallContext, in types.ListParamGroupsInput) (types.ListParamGroupsOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.ListParamGroupsOutput{}, err
	}
	page, err := normalizePagination(in.Page, in.PageSize)
	if err != nil {
		return types.ListParamGroupsOutput{}, err
	}
	// DescribeUDBParamGroup requires Zone for single-group lookup (ret_code=210
	// otherwise); list mode works without it. Fail fast with a clear message.
	if in.GroupID > 0 && strings.TrimSpace(in.Zone) == "" {
		return types.ListParamGroupsOutput{}, &InvalidInputError{
			Field:   "zone",
			Message: "is required when group_id is set (single-group lookup); zones via udb_mysql_list_regions",
		}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.ListParamGroupsOutput{}, err
	}

	sdkReq := client.NewDescribeUDBParamGroupRequest()
	sdkReq.Offset = ucloud.Int(page.Offset)
	sdkReq.Limit = ucloud.Int(page.PageSize)
	sdkReq.ClassType = ucloud.String(udbParamGroupClassType)
	if in.GroupID > 0 {
		sdkReq.GroupId = ucloud.Int(in.GroupID)
	}
	if in.RegionFlag {
		sdkReq.RegionFlag = ucloud.Bool(true)
	}
	if in.IsInUDBC != nil {
		sdkReq.IsInUDBC = ucloud.Bool(*in.IsInUDBC)
	}
	if zone := strings.TrimSpace(in.Zone); zone != "" {
		sdkReq.Zone = ucloud.String(zone)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBParamGroup(sdkReq)
	if err != nil {
		return types.ListParamGroupsOutput{}, mapSDKError(describeUDBParamGroupAction, err)
	}

	apiTotal := 0
	items := make([]types.ParamGroupSummary, 0)
	if resp != nil {
		apiTotal = resp.TotalCount
		for _, item := range resp.DataSet {
			items = append(items, mapParamGroupSummary(item))
		}
	}
	return types.ListParamGroupsOutput{
		Items:         items,
		APITotalCount: apiTotal,
		ReturnedCount: len(items),
		Page:          page.Page,
		PageSize:      page.PageSize,
	}, nil
}

// DescribePrice quotes create pricing via DescribeUDBInstancePrice.
func (c *Client) DescribePrice(ctx context.Context, reqCtx CallContext, in types.DescribePriceInput) (types.DescribePriceOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.DescribePriceOutput{}, err
	}
	if strings.TrimSpace(in.Zone) == "" {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "zone", Message: "is required"}
	}
	if strings.TrimSpace(in.DBTypeID) == "" {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "db_type_id", Message: "is required"}
	}
	if in.MemoryLimitMB <= 0 {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "memory_limit_mb", Message: "must be positive"}
	}
	if in.DiskSpaceGB <= 0 {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "disk_space_gb", Message: "must be positive"}
	}
	if strings.TrimSpace(in.MachineType) == "" {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "machine_type", Message: "is required"}
	}
	if strings.TrimSpace(in.StorageClass) == "" {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "storage_class", Message: "is required"}
	}
	if strings.TrimSpace(in.SpecificationClass) == "" {
		return types.DescribePriceOutput{}, &InvalidInputError{Field: "specification_class", Message: "is required"}
	}
	count, err := wireCountOrDefault(in.Count)
	if err != nil {
		return types.DescribePriceOutput{}, err
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.DescribePriceOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstancePriceRequest()
	sdkReq.Zone = ucloud.String(strings.TrimSpace(in.Zone))
	sdkReq.DBTypeId = ucloud.String(strings.TrimSpace(in.DBTypeID))
	sdkReq.MemoryLimit = ucloud.Int(in.MemoryLimitMB)
	sdkReq.DiskSpace = ucloud.Int(in.DiskSpaceGB)
	sdkReq.Count = ucloud.Int(count)
	// Always use machine-type purchase path (SpecificationType=1).
	sdkReq.SpecificationType = ucloud.Int(1)
	sdkReq.MachineType = ucloud.String(strings.TrimSpace(in.MachineType))
	sdkReq.StorageClass = ucloud.String(strings.TrimSpace(in.StorageClass))
	sdkReq.SpecificationClass = ucloud.String(strings.TrimSpace(in.SpecificationClass))
	// Schema documents Normal as default; upstream fails with ret_code=130 when omitted.
	instanceMode := strings.TrimSpace(in.InstanceMode)
	if instanceMode == "" {
		instanceMode = "Normal"
	}
	sdkReq.InstanceMode = ucloud.String(instanceMode)
	if v := strings.TrimSpace(in.ChargeType); v != "" {
		sdkReq.ChargeType = ucloud.String(v)
	}
	if in.Quantity != nil {
		sdkReq.Quantity = ucloud.Int(*in.Quantity)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstancePrice(sdkReq)
	if err != nil {
		return types.DescribePriceOutput{}, mapSDKError(describeUDBInstancePriceAction, err)
	}

	items := make([]types.PriceItemOutput, 0)
	if resp != nil {
		for _, item := range resp.DataSet {
			items = append(items, types.PriceItemOutput{ChargeType: item.ChargeType, PriceCents: item.Price})
		}
	}
	return types.DescribePriceOutput{Items: items}, nil
}

// DescribeUpgradePrice quotes upgrade pricing via DescribeUDBInstanceUpgradePrice.
func (c *Client) DescribeUpgradePrice(ctx context.Context, reqCtx CallContext, in types.DescribeUpgradePriceInput) (types.DescribeUpgradePriceOutput, error) {
	if err := ctx.Err(); err != nil {
		return types.DescribeUpgradePriceOutput{}, err
	}
	dbID, err := requireDBID(in.DBID)
	if err != nil {
		return types.DescribeUpgradePriceOutput{}, err
	}
	if in.MemoryLimitMB <= 0 {
		return types.DescribeUpgradePriceOutput{}, &InvalidInputError{Field: "memory_limit_mb", Message: "must be positive"}
	}
	if in.DiskSpaceGB <= 0 {
		return types.DescribeUpgradePriceOutput{}, &InvalidInputError{Field: "disk_space_gb", Message: "must be positive"}
	}

	client, err := c.sdkClient(reqCtx)
	if err != nil {
		return types.DescribeUpgradePriceOutput{}, err
	}

	sdkReq := client.NewDescribeUDBInstanceUpgradePriceRequest()
	sdkReq.DBId = ucloud.String(dbID)
	sdkReq.MemoryLimit = ucloud.Int(in.MemoryLimitMB)
	sdkReq.DiskSpace = ucloud.Int(in.DiskSpaceGB)
	if zone := strings.TrimSpace(in.Zone); zone != "" {
		sdkReq.Zone = ucloud.String(zone)
	}
	if in.CPU > 0 {
		sdkReq.CPU = ucloud.Int(in.CPU)
	}
	if v := strings.TrimSpace(in.InstanceType); v != "" {
		sdkReq.InstanceType = ucloud.String(v)
	}
	if v := strings.TrimSpace(in.MachineType); v != "" {
		sdkReq.MachineType = ucloud.String(v)
	}
	if in.SpecificationType > 0 {
		sdkReq.SpecificationType = ucloud.Int(in.SpecificationType)
	}
	if in.OrderStartTime > 0 {
		sdkReq.OrderStartTime = ucloud.Int(in.OrderStartTime)
	}
	prepareRequest(sdkReq, ctx, c.factory.Timeout)

	resp, err := client.DescribeUDBInstanceUpgradePrice(sdkReq)
	if err != nil {
		return types.DescribeUpgradePriceOutput{}, mapSDKError(describeUDBInstanceUpgradePriceAction, err)
	}

	priceCents := 0
	if resp != nil {
		priceCents = resp.Price
	}
	return types.DescribeUpgradePriceOutput{PriceCents: priceCents}, nil
}

func mapMachineTypeSummary(item udb.MachineType) types.MachineTypeSummary {
	return types.MachineTypeSummary{
		ID:                 item.ID,
		Description:        item.Description,
		CPU:                item.Cpu,
		MemoryGB:           item.Memory,
		StorageClass:       item.StorageClass,
		SpecificationClass: item.SpecificationClass,
		Group:              item.Group,
	}
}

func mapParamGroupSummary(item udb.UDBParamGroupSet) types.ParamGroupSummary {
	return types.ParamGroupSummary{
		GroupID:     item.GroupId,
		Name:        item.GroupName,
		DBTypeID:    item.DBTypeId,
		Description: item.Description,
		GroupType:   item.GroupType,
		Modifiable:  item.Modifiable,
		RegionFlag:  item.RegionFlag,
		Zone:        item.Zone,
	}
}
