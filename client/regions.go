package client

import (
	"context"
	"sort"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"udb-mysql-mcp-server/types"
)

const getRegionAction = "GetRegion"

// ListRegions returns account regions and availability zones, optionally filtered.
func (c *Client) ListRegions(ctx context.Context, reqCtx CallContext, in types.ListRegionsInput) (types.ListRegionsOutput, error) {
	sdk, err := c.uaccountClient(reqCtx)
	if err != nil {
		return types.ListRegionsOutput{}, err
	}

	req := sdk.NewGetRegionRequest()
	prepareRequest(&req.CommonBase, ctx, c.factory.Timeout)
	resp, err := sdk.GetRegion(req)
	if err != nil {
		return types.ListRegionsOutput{}, mapSDKError(getRegionAction, err)
	}

	grouped := groupRegions(nil)
	if resp != nil {
		grouped = groupRegions(resp.Regions)
	}
	regions := filterRegions(grouped, in)
	return types.ListRegionsOutput{
		TotalCount:    len(grouped),
		ReturnedCount: len(regions),
		Regions:       regions,
	}, nil
}

func groupRegions(items []uaccount.RegionInfo) []types.RegionOutput {
	type acc struct {
		region     string
		regionName string
		isDefault  bool
		zones      []types.ZoneOutput
		seenZones  map[string]struct{}
	}

	order := make([]string, 0)
	byRegion := map[string]*acc{}
	for _, item := range items {
		region := strings.TrimSpace(item.Region)
		if region == "" {
			continue
		}
		entry, ok := byRegion[region]
		if !ok {
			entry = &acc{
				region:     region,
				regionName: strings.TrimSpace(item.RegionName),
				seenZones:  map[string]struct{}{},
			}
			byRegion[region] = entry
			order = append(order, region)
		}
		if entry.regionName == "" {
			entry.regionName = strings.TrimSpace(item.RegionName)
		}
		if item.IsDefault {
			entry.isDefault = true
		}
		zone := strings.TrimSpace(item.Zone)
		if zone == "" {
			continue
		}
		if _, dup := entry.seenZones[zone]; dup {
			continue
		}
		entry.seenZones[zone] = struct{}{}
		entry.zones = append(entry.zones, types.ZoneOutput{
			Zone:      zone,
			IsDefault: item.IsDefault,
		})
	}

	sort.Strings(order)
	out := make([]types.RegionOutput, 0, len(order))
	for _, region := range order {
		entry := byRegion[region]
		sort.Slice(entry.zones, func(i, j int) bool {
			return entry.zones[i].Zone < entry.zones[j].Zone
		})
		out = append(out, types.RegionOutput{
			Region:     entry.region,
			RegionName: entry.regionName,
			IsDefault:  entry.isDefault,
			Zones:      entry.zones,
		})
	}
	return out
}

func filterRegions(items []types.RegionOutput, in types.ListRegionsInput) []types.RegionOutput {
	exactRegion := strings.TrimSpace(in.Region)
	exactZone := strings.TrimSpace(in.Zone)
	contains := strings.TrimSpace(in.NameContains)

	out := make([]types.RegionOutput, 0, len(items))
	for _, item := range items {
		if exactRegion != "" && item.Region != exactRegion {
			continue
		}
		if exactRegion == "" && contains != "" &&
			!strings.Contains(item.RegionName, contains) &&
			!strings.Contains(item.Region, contains) {
			continue
		}

		zones := item.Zones
		if exactZone != "" {
			filtered := make([]types.ZoneOutput, 0, 1)
			for _, zone := range item.Zones {
				if zone.Zone == exactZone {
					filtered = append(filtered, zone)
				}
			}
			if len(filtered) == 0 {
				continue
			}
			zones = filtered
		}

		out = append(out, types.RegionOutput{
			Region:     item.Region,
			RegionName: item.RegionName,
			IsDefault:  item.IsDefault,
			Zones:      zones,
		})
	}
	return out
}
