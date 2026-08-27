package types

// ListDBTypesInput 是 udb_mysql_list_db_types 的类型化 MCP 输入。
type ListDBTypesInput struct {
	ScopeInput
	Zone                 string `json:"zone,required" jsonschema_description:"可用区，参见可用区列表。创建实例前先调用本接口查询支持的引擎版本"`
	BackupZone           string `json:"backup_zone,omitempty" jsonschema_description:"跨可用区高可用DB的备库所在区域，仅当该可用区支持跨可用区高可用时填入（DescribeUDBType BackupZone）"`
	DBClusterType        string `json:"db_cluster_type,omitempty" jsonschema_description:"DB实例类型过滤，如 mysql、sqlserver、mongo、postgresql"`
	InstanceMode         string `json:"instance_mode,omitempty" jsonschema:"enum=Normal,enum=HA,enum=sharded_cluster" jsonschema_description:"返回支持某种实例类型的DB类型，不传表示任何实例类型均可。Normal单点、HA高可用、sharded_cluster分片集群。区分大小写"`
	DiskType             string `json:"disk_type,omitempty" jsonschema_description:"返回支持某种磁盘类型的DB类型，如 Normal、SSD、NVMe_SSD、CLOUD_SSD_ESSENTIAL；不传表示任何磁盘类型均可"`
	CompatibleWithDBType string `json:"compatible_with_db_type,omitempty" jsonschema_description:"从备份创建实例时，该版本号所支持的备份创建版本；不传表示不是从备份创建"`
	DBSubVersion         string `json:"db_sub_version,omitempty" jsonschema_description:"从备份创建实例时，该小版本号所支持的备份创建小版本；不传表示不是从备份创建"`
}

// DBTypeSummary 是 DescribeUDBType 返回的一条支持的 DB 类型。
type DBTypeSummary struct {
	DBTypeID     string `json:"db_type_id" jsonschema_description:"DB类型id，按版本细分，如 mysql-8.0"`
	DBSubVersion string `json:"db_sub_version,omitempty" jsonschema_description:"mysql子版本，如 mysql-8.0.25"`
	Recommended  bool   `json:"recommended,omitempty" jsonschema_description:"该条目与API DedaultType（推荐DB版本）匹配时为true"`
}

// ListDBTypesOutput 是结构化 DB 类型响应。
type ListDBTypesOutput struct {
	Items                 []DBTypeSummary `json:"items" jsonschema_description:"DB类型列表"`
	RecommendedDBTypeID   string          `json:"recommended_db_type_id,omitempty" jsonschema_description:"API DedaultType 的 DBTypeId（推荐DB版本）"`
	RecommendedSubVersion string          `json:"recommended_db_sub_version,omitempty" jsonschema_description:"推荐DB版本的小版本号"`
}

// ListMachineTypesInput 是 udb_mysql_list_machine_types 的类型化 MCP 输入。
type ListMachineTypesInput struct {
	ScopeInput
	Zone         string `json:"zone,required" jsonschema_description:"可用区，参见可用区列表。创建实例前先调用本接口选择计算规格"`
	InstanceMode string `json:"instance_mode,omitempty" jsonschema:"enum=Normal,enum=HA" jsonschema_description:"UDB实例模式类型：Normal普通版（默认）、HA高可用版"`
}

// MachineTypeSummary 是一条计算规格。内存单位为GB（API 原样）。
type MachineTypeSummary struct {
	ID                 string `json:"id" jsonschema_description:"计算规格id，格式为 机型.配比.CPU核数规格，如 o.mysql4m.medium 表示快杰NVMe机型2C8G。机型 o/n 分别代表快杰NVMe和SSD云盘机型；配比 2m/4m/8m 代表CPU内存比1:2/1:4/1:8；规格 small=1C、medium=2C、xlarge=4C、2xlarge=8C、4xlarge=16C、8xlarge=32C、16xlarge=64C"`
	Description        string `json:"description" jsonschema_description:"计算规格描述，格式为 nCmG，表示n核mG内存"`
	CPU                int    `json:"cpu" jsonschema_description:"规格cpu核数"`
	MemoryGB           int    `json:"memory_gb" jsonschema_description:"规格内存大小，单位GB"`
	StorageClass       string `json:"storage_class,omitempty" jsonschema:"enum=CLOUD_SSD,enum=CLOUD_RSSD,enum=CLOUD_SSD_ESSENTIAL" jsonschema_description:"存储类型：CLOUD_SSD为SSD云盘、CLOUD_RSSD为RSSD云盘、CLOUD_SSD_ESSENTIAL为SSD Essential云盘"`
	SpecificationClass string `json:"specification_class,omitempty" jsonschema_description:"规格类型：O为NVMe型、OM为共享型、N为通用型"`
	Group              string `json:"group,omitempty" jsonschema_description:"内存/cpu配比"`
}

// ListMachineTypesOutput 是结构化计算规格响应。
type ListMachineTypesOutput struct {
	Items   []MachineTypeSummary `json:"items" jsonschema_description:"计算规格列表"`
	Default MachineTypeSummary   `json:"default,omitempty" jsonschema_description:"默认计算规格"`
}

// ListParamGroupsInput 是 udb_mysql_list_param_groups 的类型化 MCP 输入。
type ListParamGroupsInput struct {
	ScopeInput
	Page       int    `json:"page,omitempty" jsonschema_description:"页码（默认1）"`
	PageSize   int    `json:"page_size,omitempty" jsonschema_description:"每页数量（默认20，最大100）"`
	GroupID    int    `json:"group_id,omitempty" jsonschema_description:"参数组id；指定则获取单个参数组描述（此时 zone 必填），否则为列表操作"`
	RegionFlag bool   `json:"region_flag,omitempty" jsonschema_description:"请求不带zone时，true表示只拉取跨可用区相关配置文件（跨可用区HA实例的参数组须从此查询），否则拉取所有机房配置文件（DescribeUDBParamGroup RegionFlag）"`
	IsInUDBC   *bool  `json:"is_in_udbc,omitempty" jsonschema_description:"是否选取专区中配置（DescribeUDBParamGroup IsInUDBC）；不传用API默认，true限定专区DB，false排除专区"`
	Zone       string `json:"zone,omitempty" jsonschema_description:"可选可用区；group_id 单查时必填（上游 DescribeUDBParamGroup 要求），列表操作时可省略"`
}

// ParamGroupSummary 是创建流程使用的参数组摘要。
type ParamGroupSummary struct {
	GroupID     int    `json:"group_id" jsonschema_description:"参数组id"`
	Name        string `json:"name" jsonschema_description:"参数组名称"`
	DBTypeID    string `json:"db_type_id" jsonschema_description:"DB类型id，须与创建实例的 db_type_id 匹配"`
	Description string `json:"description,omitempty" jsonschema_description:"参数组描述"`
	GroupType   int    `json:"group_type" jsonschema_description:"参数组类型：1稳定版（默认）、2高性能版"`
	Modifiable  bool   `json:"modifiable" jsonschema_description:"参数组是否可修改"`
	RegionFlag  bool   `json:"region_flag" jsonschema_description:"是否为跨可用区配置文件"`
	Zone        string `json:"zone,omitempty" jsonschema_description:"可用区"`
}

// ListParamGroupsOutput 是结构化参数组响应。
type ListParamGroupsOutput struct {
	Items         []ParamGroupSummary `json:"items" jsonschema_description:"参数组列表"`
	APITotalCount int                 `json:"api_total_count" jsonschema_description:"DescribeUDBParamGroup TotalCount（上游账号级总数）"`
	ReturnedCount int                 `json:"returned_count" jsonschema_description:"当前页返回条数"`
	Page          int                 `json:"page" jsonschema_description:"规范化后的页码（>=1）"`
	PageSize      int                 `json:"page_size" jsonschema_description:"规范化后的每页数量（默认20，最大100）"`
}

// DescribePriceInput 是 udb_mysql_describe_price 的类型化 MCP 输入。
type DescribePriceInput struct {
	ScopeInput
	Zone          string `json:"zone,required" jsonschema_description:"可用区，参见可用区列表"`
	DBTypeID      string `json:"db_type_id,required" jsonschema_description:"UDB实例的DB版本字符串，如 mysql-8.0"`
	MemoryLimitMB int    `json:"memory_limit_mb,required" jsonschema_description:"内存限制(MB)，目前支持2000‑96000，按1000进制(1GB=1000MB)计算"`
	DiskSpaceGB   int    `json:"disk_space_gb,required" jsonschema_description:"磁盘空间(GB)，支持20-3000，输入不带单位"`
	Count         int    `json:"count,omitempty" jsonschema_description:"购买DB实例数量，最大10台，默认1台"`
	ChargeType    string `json:"charge_type,omitempty" jsonschema:"enum=Year,enum=Month,enum=Dynamic,enum=Trial" jsonschema_description:"付费类型：Year按年、Month按月（默认）、Dynamic按需付费（需开启权限）、Trial试用（需开启权限）"`
	Quantity      *int   `json:"quantity,omitempty" jsonschema_description:"购买多少个计费时间单位，默认1。如买2个月传2；计费单位为Month且传0表示购买到月底"`
	InstanceMode  string `json:"instance_mode,omitempty" jsonschema:"enum=Normal,enum=Slave,enum=HA" jsonschema_description:"实例部署类型：Normal=普通版/单点（默认，不传时服务端按 Normal）、HA=高可用版、Slave=从库。询普通版传 Normal，询高可用传 HA"`
	//CPU               int    `json:"cpu,omitempty" jsonschema_description:"CPU个数，db类型为sqlserver时必填"`
	//InstanceType      string `json:"instance_type,omitempty" jsonschema:"enum=SATA_SSD,enum=NVMe_SSD" jsonschema_description:"机型字段（快杰机型请改用 specification_class+storage_class，优先级更高）：SATA_SSD为SATA SSD机型（仅部分地域支持）、NVMe_SSD为快杰机型"`
	//SpecificationType int    `json:"specification_type,omitempty" jsonschema_description:"实例计算规格类型：0或不传代表使用内存方式购买，1代表使用内存-cpu可选配比方式购买（需填写 machine_type）"`
	MachineType        string `json:"machine_type,required" jsonschema_description:"规格类型ID，取值见 udb_mysql_list_machine_types 返回的 id，须与同条规格的 storage_class/specification_class 一起传入"`
	StorageClass       string `json:"storage_class,required" jsonschema:"enum=CLOUD_SSD,enum=CLOUD_RSSD,enum=CLOUD_SSD_ESSENTIAL" jsonschema_description:"存储类型：CLOUD_SSD为SSD云盘、CLOUD_RSSD为RSSD云盘、CLOUD_SSD_ESSENTIAL为SSD Essential云盘。须与 specification_class 组合，取值从 list_machine_types 同条规格读取"`
	SpecificationClass string `json:"specification_class,required" jsonschema:"enum=O,enum=OM,enum=N" jsonschema_description:"规格类型：O为NVME、OM为共享型、N为通用型。须与 storage_class 组合，取值从 list_machine_types 同条规格读取"`
}

// PriceItemOutput 保留整数最小计价单位（分）。
type PriceItemOutput struct {
	ChargeType string `json:"charge_type" jsonschema_description:"付费类型：Year/Month/Dynamic/Trial"`
	PriceCents int    `json:"price_cents" jsonschema_description:"价格，单位为分"`
}

// DescribePriceOutput 是结构化创建询价响应。
type DescribePriceOutput struct {
	Items []PriceItemOutput `json:"items" jsonschema_description:"各计费方式的价格列表"`
}

// DescribeUpgradePriceInput 是 udb_mysql_describe_upgrade_price 的类型化 MCP 输入。
type DescribeUpgradePriceInput struct {
	ScopeInput
	DBID              string `json:"db_id,required" jsonschema_description:"实例的Id（DBId）"`
	Zone              string `json:"zone,omitempty" jsonschema_description:"可选可用区，参见可用区列表"`
	MemoryLimitMB     int    `json:"memory_limit_mb,required" jsonschema_description:"内存限制(MB)"`
	DiskSpaceGB       int    `json:"disk_space_gb,required" jsonschema_description:"磁盘空间(GB)，支持20G-500G"`
	CPU               int    `json:"cpu,omitempty" jsonschema_description:"CPU核数，快杰SQLServer升降级必传"`
	InstanceType      string `json:"instance_type,omitempty" jsonschema:"enum=SATA_SSD,enum=NVMe_SSD" jsonschema_description:"机型：SATA_SSD或NVMe_SSD"`
	MachineType       string `json:"machine_type,omitempty" jsonschema_description:"规格类型ID，当 specification_type=1 时有效"`
	SpecificationType int    `json:"specification_type,omitempty" jsonschema_description:"实例计算规格类型：0或不传代表使用内存方式购买，1代表使用内存-cpu可选配比方式购买（需填写 machine_type）"`
	OrderStartTime    int    `json:"order_start_time,omitempty" jsonschema_description:"获取指定时间开始之后的升级价格（Unix时间戳）；不填默认当前时间"`
}

// DescribeUpgradePriceOutput 是结构化升降级询价响应。
type DescribeUpgradePriceOutput struct {
	PriceCents int    `json:"price_cents" jsonschema_description:"升级价格，单位为分"`
	Currency   string `json:"currency" jsonschema_description:"MCP约定字段：分计价恒为 CNY（人民币），非SDK响应字段"`
}
