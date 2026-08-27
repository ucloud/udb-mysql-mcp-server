package types

// ListRegionsInput 是 udb_mysql_list_regions 的类型化 MCP 输入。
type ListRegionsInput struct {
	Region       string `json:"region,omitempty" jsonschema_description:"地域代码精确过滤，如 cn-bj2"`
	Zone         string `json:"zone,omitempty" jsonschema_description:"可用区精确过滤，如 cn-bj2-04"`
	NameContains string `json:"name_contains,omitempty" jsonschema_description:"地域名称或地域代码子串过滤（region 为空时生效）"`
}

// ZoneOutput 是 MCP 响应中的可用区条目。
type ZoneOutput struct {
	Zone      string `json:"zone" jsonschema_description:"可用区代码，如 cn-bj2-04"`
	IsDefault bool   `json:"is_default" jsonschema_description:"是否为账号默认机房"`
}

// RegionOutput 是带嵌套可用区的地域条目。
type RegionOutput struct {
	Region     string       `json:"region" jsonschema_description:"地域代码，如 cn-bj2"`
	RegionName string       `json:"region_name" jsonschema_description:"地域名称，如 北京二"`
	IsDefault  bool         `json:"is_default" jsonschema_description:"该地域下是否有账号默认机房"`
	Zones      []ZoneOutput `json:"zones" jsonschema_description:"该地域下的可用区列表"`
}

// ListRegionsOutput 是 udb_mysql_list_regions 的结构化响应。
type ListRegionsOutput struct {
	TotalCount    int            `json:"total_count" jsonschema_description:"GetRegion 返回的去重地域数（过滤前）"`
	ReturnedCount int            `json:"returned_count" jsonschema_description:"过滤后返回的地域数"`
	Regions       []RegionOutput `json:"regions" jsonschema_description:"匹配的地域及嵌套可用区"`
}
