package types

// ScopeInput 为云工具携带可选的项目与地域覆盖参数。
type ScopeInput struct {
	ProjectID string `json:"project_id,omitempty" jsonschema_description:"项目ID。不填写为默认项目（UCLOUD_PROJECT_ID），子帐号必须填写。资源若属于其他项目而未传该项目的 project_id，上游常报 7009/7048 类错误；可用 udb_mysql_list_projects 列出候选项目"`
	Region    string `json:"region,omitempty" jsonschema_description:"地域代码，如 cn-bj2。不传时使用 UCLOUD_REGION 或进程默认地域（通常 cn-bj2），仅查询该单一地域，不是全部地域。跨地域请先 udb_mysql_list_regions，再按地域多次调用。与 zone 同传时请明确传入 region"`
}

// GetInstanceInput 是 udb_mysql_get_instance 的类型化 MCP 输入。
type GetInstanceInput struct {
	ScopeInput
	DBID string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），如 udbha-xxxxx，可通过 DescribeUDBInstance（list_instances）获取"`
}

// InstanceOutput 是实例读取的紧凑结构化 MCP 响应。
type InstanceOutput struct {
	DBID         string `json:"db_id" jsonschema_description:"DB实例Id"`
	Name         string `json:"name" jsonschema_description:"实例名称，至少6位,最大63位"`
	State        string `json:"state" jsonschema_description:"实例状态（已去除空白）"`
	DBTypeID     string `json:"db_type_id,omitempty" jsonschema_description:"DB类型，按版本细分，如 mysql-8.0"`
	InstanceType string `json:"instance_type,omitempty" jsonschema_description:"UDB数据库机型"`
	Address      string `json:"address,omitempty" jsonschema_description:"DB实例虚IP（VirtualIP）"`
	Port         int    `json:"port,omitempty" jsonschema_description:"端口号，mysql默认3306"`
	CPU          int    `json:"cpu,omitempty" jsonschema_description:"CPU核数"`
	MemoryMB     int    `json:"memory_mb,omitempty" jsonschema_description:"内存限制(MB)"`
	DiskGB       int    `json:"disk_gb,omitempty" jsonschema_description:"磁盘空间(GB)"`
	Zone         string `json:"zone,omitempty" jsonschema_description:"DB实例所在可用区"`
	CreatedAt    string `json:"created_at,omitempty" jsonschema_description:"创建时间（UTC RFC3339）"`
}

// InstanceSummary 是列表返回的紧凑条目。
type InstanceSummary struct {
	DBID     string `json:"db_id" jsonschema_description:"DB实例Id"`
	Name     string `json:"name" jsonschema_description:"实例名称"`
	State    string `json:"state" jsonschema_description:"实例状态"`
	DBTypeID string `json:"db_type_id,omitempty" jsonschema_description:"DB类型，如 mysql-8.0"`
	Zone     string `json:"zone,omitempty" jsonschema_description:"DB实例所在可用区"`
}

// ListInstancesInput 是 udb_mysql_list_instances 的类型化 MCP 输入。
type ListInstancesInput struct {
	ScopeInput
	Page          int    `json:"page,omitempty" jsonschema_description:"页码（默认1）"`
	PageSize      int    `json:"page_size,omitempty" jsonschema_description:"每页数量（默认20，最大100）"`
	NameContains  string `json:"name_contains,omitempty" jsonschema_description:"客户端实例名称模糊过滤（在当前API页内后置过滤）"`
	State         string `json:"state,omitempty" jsonschema:"enum=Init,enum=Fail,enum=Starting,enum=Running,enum=Shutdown,enum=Shutoff,enum=Delete,enum=Upgrading,enum=Promoting,enum=Recovering,enum=Recover fail,enum=Remakeing,enum=RemakeFail,enum=VersionUpgrading,enum=VersionUpgradeWaitForSwitch,enum=VersionUpgradeFail,enum=UpdatingSSL,enum=UpdateSSLFail" jsonschema_description:"客户端实例状态过滤（在当前API页内后置过滤）。取值：Init初始化中、Fail安装失败、Starting启动中、Running运行、Shutdown关闭中、Shutoff已关闭、Delete已删除、Upgrading升级中、Promoting提升为独库进行中、Recovering恢复中、Recover fail恢复失败、Remakeing重做中、RemakeFail重做失败、VersionUpgrading小版本升级中、VersionUpgradeWaitForSwitch高可用等待切换、VersionUpgradeFail小版本升级失败、UpdatingSSL修改SSL中、UpdateSSLFail修改SSL失败"`
	IncludeSlaves bool   `json:"include_slaves,omitempty" jsonschema_description:"默认false排除MySQL从库（Role=slave或SrcDBId非空）；true时包含从库"`
	Zone          string `json:"zone,omitempty" jsonschema_description:"可选可用区过滤，不填时默认全部可用区"`
	Verbose       bool   `json:"verbose,omitempty" jsonschema_description:"true时每条返回完整实例字段"`
}

// ListInstancesOutput 是结构化列表响应。
type ListInstancesOutput struct {
	Items         []any  `json:"items" jsonschema_description:"实例摘要或完整实例对象"`
	APITotalCount int    `json:"api_total_count" jsonschema_description:"DescribeUDBInstance TotalCount（账号级上游总数；非过滤后的匹配数）"`
	ReturnedCount int    `json:"returned_count" jsonschema_description:"当前页经客户端过滤后返回的条数"`
	Page          int    `json:"page" jsonschema_description:"规范化后的页码（>=1）"`
	PageSize      int    `json:"page_size" jsonschema_description:"规范化后的每页数量（默认20，最大100）"`
	Region        string `json:"region" jsonschema_description:"本次实际查询的地域（入参或 UCLOUD_REGION/进程默认），单次仅覆盖该地域，不是全部地域"`
	ProjectID     string `json:"project_id,omitempty" jsonschema_description:"本次实际使用的项目ID"`
	PostFiltered  *bool  `json:"post_filtered,omitempty" jsonschema_description:"当名称/状态过滤在当前API页生效时为true"`
	FilterScope   string `json:"filter_scope,omitempty" jsonschema_description:"post_filtered为true时的过滤范围：current_page"`
}

// GetInstanceStateInput 是 udb_mysql_get_instance_state 的类型化 MCP 输入。
type GetInstanceStateInput struct {
	ScopeInput
	DBID string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
}

// GetInstanceStateOutput 是结构化状态响应。
type GetInstanceStateOutput struct {
	DBID         string `json:"db_id" jsonschema_description:"DB实例Id"`
	State        string `json:"state" jsonschema_description:"DescribeUDBInstanceState 返回的状态（已去除空白）"`
	FollowUpTool string `json:"follow_up_tool,omitempty" jsonschema_description:"建议轮询的后续 MCP 工具名"`
	Hint         string `json:"hint" jsonschema_description:"通用轮询指引；服务端不做内部轮询"`
}

// LifecycleInput 是 start/restart 生命周期变更的共享输入。
type LifecycleInput struct {
	ScopeInput
	DBID string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	Zone string `json:"zone,omitempty" jsonschema_description:"可选可用区，参见可用区列表"`
}

// StopInstanceInput 是 udb_mysql_stop_instance 的类型化 MCP 输入。
type StopInstanceInput struct {
	ScopeInput
	DBID        string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	Zone        string `json:"zone,omitempty" jsonschema_description:"可选可用区，参见可用区列表"`
	ForceToKill bool   `json:"force_to_kill,omitempty" jsonschema_description:"是否使用强制手段关闭DB，默认false"`
}

// LifecycleMutationOutput 是结构化生命周期变更响应。
type LifecycleMutationOutput struct {
	DBID         string `json:"db_id" jsonschema_description:"DB实例Id"`
	FollowUpTool string `json:"follow_up_tool,omitempty" jsonschema_description:"建议轮询的后续 MCP 工具名"`
	Hint         string `json:"hint" jsonschema_description:"通用轮询指引；服务端不做内部轮询"`
}

// ModifyNameInput 是 udb_mysql_modify_name 的类型化 MCP 输入。
type ModifyNameInput struct {
	ScopeInput
	DBID string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	Name string `json:"name,required" jsonschema_description:"新实例名称，至少6位,最大63位"`
	Zone string `json:"zone,omitempty" jsonschema_description:"可选可用区，参见可用区列表"`
}

// ModifyNameOutput 是结构化重命名响应。
type ModifyNameOutput struct {
	DBID    string `json:"db_id" jsonschema_description:"DB实例Id"`
	NewName string `json:"new_name" jsonschema_description:"新实例名称"`
}

// CreateInstanceInput 是 udb_mysql_create_instance 的类型化 MCP 输入。
type CreateInstanceInput struct {
	ScopeInput
	Zone               string `json:"zone,required" jsonschema_description:"可用区，参见可用区列表"`
	Name               string `json:"name,required" jsonschema_description:"实例名称，至少6位,最大63位"`
	AdminPassword      string `json:"admin_password,required" jsonschema_description:"管理员密码。8-36位，支持大小写字母、数字及特殊字符@#$%^*-+=_,?!&()~.|，须包含两类及以上字符"`
	DBTypeID           string `json:"db_type_id,required" jsonschema_description:"DB类型，按版本细分 mysql-8.4/mysql-8.0/mysql-5.7/percona-5.7/mysql-5.6/percona-5.6/mysql-5.5，取值见 DescribeUDBType 返回的 DBTypeId"`
	Port               int    `json:"port,omitempty" jsonschema_description:"端口号，mysql默认3306（不填服务端按3306发送）"`
	DiskSpaceGB        int    `json:"disk_space_gb,required" jsonschema_description:"磁盘空间(GB)，支持约20G–32T，步长通常为10"`
	ParamGroupID       int    `json:"param_group_id,required" jsonschema_description:"配置参数组id，取值见 DescribeUDBParamGroup 返回的 GroupId，且须与 db_type_id 匹配"`
	MachineType        string `json:"machine_type,required" jsonschema_description:"规格类型ID，取值见 ListUDBMachineType 返回的 ID 字段，格式如 o.mysql4m.medium（机型.配比.CPU规格）"`
	StorageClass       string `json:"storage_class,required" jsonschema:"enum=CLOUD_SSD,enum=CLOUD_RSSD,enum=CLOUD_SSD_ESSENTIAL" jsonschema_description:"存储类型：CLOUD_RSSD为RSSD云盘（对应O型）、CLOUD_SSD为SSD云盘、CLOUD_SSD_ESSENTIAL为SSD Essential云盘（对应OM型）。须与 SpecificationClass 组合使用，取值从 ListUDBMachineType 同条规格读取"`
	SpecificationClass string `json:"specification_class,required" jsonschema:"enum=O,enum=O2,enum=OM" jsonschema_description:"规格类型：O为NVMe型、O2、OM为共享型。与 StorageClass 组合使用"`
	ChargeType         string `json:"charge_type,omitempty" jsonschema:"enum=Year,enum=Month,enum=Dynamic,enum=Trial" jsonschema_description:"付费类型：Year按年、Month按月、Dynamic按需付费（需开启权限）、Trial试用（需开启权限）；默认Month"`
	Quantity           *int   `json:"quantity,omitempty" jsonschema_description:"购买时长（计费时间单位个数），默认1。如买2个月则传2；计费单位为Month且传0表示购买到月底"`
	InstanceMode       string `json:"instance_mode,omitempty" jsonschema:"enum=Normal,enum=HA" jsonschema_description:"UDB实例模式类型：Normal普通版（默认）、HA高可用版。跨可用区高可用需同时传 backup_zone，且 param_group_id 须为跨可用区配置文件（DescribeUDBParamGroup RegionFlag=true 查询）"`
	VPCID              string `json:"vpc_id,omitempty" jsonschema_description:"VPC ID，与 subnet_id 成对使用；取值见 UVPC 相关接口"`
	SubnetID           string `json:"subnet_id,omitempty" jsonschema_description:"子网ID，与 vpc_id 须同属一个 VPC"`
	BackupID           int    `json:"backup_id,omitempty" jsonschema_description:"备份ID；指定则从备份恢复，取值见 DescribeUDBBackup"`
	BackupZone         string `json:"backup_zone,omitempty" jsonschema_description:"跨可用区高可用备库所在可用区，参见可用区列表"`
	DBSubVersion       string `json:"db_sub_version,omitempty" jsonschema_description:"MySQL小版本号，指定小版本创建，可用版本通过 DescribeUDBType 获取"`
}

// CreateInstanceOutput 是结构化创建实例响应。
type CreateInstanceOutput struct {
	DBID         string `json:"db_id" jsonschema_description:"创建成功的DB实例Id"`
	FollowUpTool string `json:"follow_up_tool" jsonschema_description:"建议轮询的后续 MCP 工具名"`
	Hint         string `json:"hint" jsonschema_description:"通用轮询指引；服务端不做内部轮询"`
}

// ResizeInstanceInput 是 udb_mysql_resize_instance 的类型化 MCP 输入。
type ResizeInstanceInput struct {
	ScopeInput
	DBID              string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	Zone              string `json:"zone,omitempty" jsonschema_description:"可选可用区，参见可用区列表"`
	MemoryLimitMB     int    `json:"memory_limit_mb,required" jsonschema_description:"内存限制(MB)，支持档位：2000/4000/6000/8000/12000/16000/24000/32000/48000/64000/96000/128000/192000/256000/320000。内存升级无需关机，其他场景需先关机。同时传 machine_type 时以 machine_type 对应规格为准，此值会被覆盖"`
	DiskSpaceGB       int    `json:"disk_space_gb,required" jsonschema_description:"磁盘空间(GB)，支持20G-32T"`
	CPU               int    `json:"cpu,omitempty" jsonschema_description:"数据库的CPU核数（只对普通版SQLServer有用）"`
	InstanceType      string `json:"instance_type,omitempty" jsonschema:"enum=Normal,enum=SATA_SSD,enum=PCIE_SSD,enum=Normal_Volume,enum=SATA_SSD_Volume,enum=PCIE_SSD_Volume,enum=NVMe_SSD" jsonschema_description:"UDB数据库机型：Normal标准机型、SATA_SSD为SSD机型、PCIE_SSD为SSD高性能机型、Normal_Volume标准大容量机型、SATA_SSD_Volume为SSD大容量机型、PCIE_SSD_Volume为SSD高性能大容量机型、NVMe_SSD快杰机型"`
	InstanceMode      string `json:"instance_mode,omitempty" jsonschema:"enum=Normal,enum=HA" jsonschema_description:"UDB实例模式类型：Normal普通版（默认）、HA高可用版"`
	MachineType       string `json:"machine_type,omitempty" jsonschema_description:"规格类型ID，当 specification_type=1 时有效，可通过 ListUDBMachineType 查询。传入时实际生效规格（内存/CPU）以该规格为准，覆盖 memory_limit_mb"`
	SpecificationType int    `json:"specification_type,omitempty" jsonschema_description:"实例计算规格类型：0或不传代表使用内存方式购买，1代表使用内存-cpu可选配比方式购买（需填写 machine_type）"`
	Confirm           *bool  `json:"confirm,required" jsonschema_description:"必须为true才执行升降配操作"`
}

// ResizeInstanceOutput 是结构化升降配响应。
type ResizeInstanceOutput struct {
	DBID         string `json:"db_id" jsonschema_description:"DB实例Id"`
	FollowUpTool string `json:"follow_up_tool" jsonschema_description:"建议轮询的后续 MCP 工具名"`
	Hint         string `json:"hint" jsonschema_description:"通用轮询指引；服务端不做内部轮询"`
}

// ResetPasswordInput 是 udb_mysql_reset_password 的类型化 MCP 输入。
type ResetPasswordInput struct {
	ScopeInput
	DBID     string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	Password string `json:"password,required" jsonschema_description:"实例的新密码。8-36位，支持大小写字母、数字及特殊字符@#$%^*-+=_,?!&()~.|，须包含两类及以上字符"`
}

// ResetPasswordOutput 是结构化重置密码响应。
type ResetPasswordOutput struct {
	DBID string `json:"db_id" jsonschema_description:"DB实例Id"`
}
