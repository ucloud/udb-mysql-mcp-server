package types

// ListBackupsInput 是 udb_mysql_list_backups 的类型化 MCP 输入。
type ListBackupsInput struct {
	ScopeInput
	Page       int    `json:"page,omitempty" jsonschema_description:"页码（默认1）"`
	PageSize   int    `json:"page_size,omitempty" jsonschema_description:"每页数量（默认20，最大100）"`
	Zone       string `json:"zone,omitempty" jsonschema_description:"可选可用区过滤，传给API"`
	DBID       string `json:"db_id,omitempty" jsonschema_description:"可选DB实例Id过滤；指定则只获取该db的备份信息"`
	BackupType string `json:"backup_type,omitempty" jsonschema:"enum=auto,enum=manual" jsonschema_description:"可选备份类型过滤：auto自动、manual手动（API侧0表示自动，1表示手动）"`
	BeginTime  int64  `json:"begin_time,omitempty" jsonschema_description:"可选起始时间过滤（Unix时间戳）"`
	EndTime    int64  `json:"end_time,omitempty" jsonschema_description:"可选结束时间过滤（Unix时间戳）"`
}

// BackupSummary 是备份列表条目。
type BackupSummary struct {
	BackupID        int    `json:"backup_id" jsonschema_description:"备份id"`
	BackupName      string `json:"backup_name" jsonschema_description:"备份名称"`
	BackupTimeUnix  int64  `json:"backup_time_unix,omitempty" jsonschema_description:"备份时间（Unix时间戳）"`
	BackupEndUnix   int    `json:"backup_end_time_unix,omitempty" jsonschema_description:"备份完成时间（Unix时间戳）"`
	BackupSizeBytes int    `json:"backup_size_bytes,omitempty" jsonschema_description:"备份文件大小（字节）"`
	BackupType      string `json:"backup_type" jsonschema_description:"备份类型：auto自动、manual手动"`
	State           string `json:"state" jsonschema_description:"备份状态：Backuping备份中、Success备份成功、Failed备份失败、Expired备份过期"`
	ErrorInfo       string `json:"error_info,omitempty" jsonschema_description:"备份错误信息"`
	DBID            string `json:"db_id" jsonschema_description:"备份所属DB实例Id"`
	DBName          string `json:"db_name,omitempty" jsonschema_description:"对应的db名称"`
	Zone            string `json:"zone" jsonschema_description:"备份所在可用区"`
	BackupZone      string `json:"backup_zone,omitempty" jsonschema_description:"跨机房高可用备库所在可用区"`
	Checksum        string `json:"checksum,omitempty" jsonschema_description:"备份文件的MD5值，备份完成后显示（仅Mysql NVMe机型与Mongo支持）"`
}

// ListBackupsOutput 是结构化备份列表响应。
type ListBackupsOutput struct {
	Items         []BackupSummary `json:"items" jsonschema_description:"备份列表"`
	APITotalCount int             `json:"api_total_count" jsonschema_description:"DescribeUDBBackup TotalCount（与本次请求上游过滤条件匹配的总数）"`
	ReturnedCount int             `json:"returned_count" jsonschema_description:"当前页返回条数"`
	Page          int             `json:"page" jsonschema_description:"规范化后的页码（>=1）"`
	PageSize      int             `json:"page_size" jsonschema_description:"规范化后的每页数量（默认20，最大100）"`
}

// GetBackupStateInput 是 udb_mysql_get_backup_state 的类型化 MCP 输入。
type GetBackupStateInput struct {
	ScopeInput
	BackupID   int    `json:"backup_id,required" jsonschema_description:"备份记录ID，取值见 BackupUDBInstance 响应或 DescribeUDBBackup（list_backups）"`
	Zone       string `json:"zone,required" jsonschema_description:"可用区（API必填），参见可用区列表"`
	BackupZone string `json:"backup_zone,omitempty" jsonschema_description:"跨可用区高可用备库所在可用区"`
}

// GetBackupStateOutput 是结构化备份状态响应。
type GetBackupStateOutput struct {
	BackupID          int    `json:"backup_id" jsonschema_description:"备份记录ID"`
	State             string `json:"state" jsonschema_description:"DescribeUDBInstanceBackupState 返回的状态（已去除空白）"`
	FollowUpTool      string `json:"follow_up_tool,omitempty" jsonschema_description:"建议轮询的后续 MCP 工具名"`
	Hint              string `json:"hint" jsonschema_description:"通用轮询指引；服务端不做内部轮询"`
	BackupEndTimeUnix int    `json:"backup_end_time_unix,omitempty" jsonschema_description:"备份完成时间（Unix时间戳）"`
	BackupSizeBytes   int    `json:"backup_size_bytes,omitempty" jsonschema_description:"备份文件大小（字节）"`
}

// GetBackupURLInput 是 udb_mysql_get_backup_url 的类型化 MCP 输入。
type GetBackupURLInput struct {
	ScopeInput
	BackupID  int    `json:"backup_id,required" jsonschema_description:"备份记录ID，可通过 DescribeUDBBackup（list_backups）获取"`
	DBID      string `json:"db_id,required" jsonschema_description:"DB实例Id，可通过 DescribeUDBInstance 获取"`
	Zone      string `json:"zone,omitempty" jsonschema_description:"可选可用区；不传时自动通过 db_id 解析实例所在可用区（额外调用一次 DescribeUDBInstance）"`
	ValidTime int    `json:"valid_time,omitempty" jsonschema_description:"返回URL的过期时间（秒），最小默认4小时、最大7天；不填默认4小时"`
}

// GetBackupURLOutput 是结构化备份下载地址响应。
type GetBackupURLOutput struct {
	BackupID         int    `json:"backup_id" jsonschema_description:"备份记录ID"`
	DBID             string `json:"db_id" jsonschema_description:"DB实例Id"`
	PublicBackupPath string `json:"public_backup_path,omitempty" jsonschema_description:"备份文件公网下载地址"`
	InnerBackupPath  string `json:"inner_backup_path,omitempty" jsonschema_description:"备份文件内网下载地址"`
	Checksum         string `json:"checksum,omitempty" jsonschema_description:"备份文件的md5值"`
}

// GetBackupStrategyInput 是 udb_mysql_get_backup_strategy 的类型化 MCP 输入。
type GetBackupStrategyInput struct {
	ScopeInput
	DBID string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId）"`
	Zone string `json:"zone,required" jsonschema_description:"可用区（API必填），参见可用区列表"`
}

// UFileBackupTarget 是可选的用户 UFile 转存配置（仅bucket；token永不透出）。
type UFileBackupTarget struct {
	Bucket string `json:"bucket,omitempty" jsonschema_description:"转存备份的UFile bucket名（已配置时返回）"`
}

// GetBackupStrategyOutput 是结构化备份策略响应。
type GetBackupStrategyOutput struct {
	DBID            string            `json:"db_id" jsonschema_description:"DB实例Id"`
	BackupBeginHour int               `json:"backup_begin_hour" jsonschema_description:"备份策略开始时间（单位小时；API未设置时默认3点）"`
	BackupDate      string            `json:"backup_date" jsonschema_description:"备份日期标记位，7位，每位为一周中一天的备份开关（0关/1开），最右一位为周日，从右到左依次为周一至周六；如1100000表示打开周六和周五的自动备份"`
	BackupMethod    string            `json:"backup_method" jsonschema_description:"默认备份方式：nobackup不备份、snapshot快照备份、logic逻辑备份、xtrabackup物理备份、ark_snapshot方舟快照备份"`
	SaveDays        int               `json:"save_days" jsonschema_description:"备份保留天数"`
	UserUFile       UFileBackupTarget `json:"user_ufile,omitempty" jsonschema_description:"用户转存备份到自己的UFILE配置"`
}

// CreateBackupInput 是 udb_mysql_create_backup 的类型化 MCP 输入。
type CreateBackupInput struct {
	ScopeInput
	DBID         string `json:"db_id,required" jsonschema_description:"DB实例Id（DBId），可通过 DescribeUDBInstance 获取"`
	BackupName   string `json:"backup_name,required" jsonschema_description:"备份名称（建议全局唯一，便于后续查找）"`
	Zone         string `json:"zone,omitempty" jsonschema_description:"可用区（支持时）"`
	BackupMethod string `json:"backup_method,omitempty" jsonschema:"enum=snapshot,enum=xtrabackup" jsonschema_description:"备份方式，不填默认逻辑备份：SSD版本mysql/mongodb支持snapshot（快照/物理备份），NVMe版本mysql支持xtrabackup"`
	Blacklist    string `json:"blacklist,omitempty" jsonschema_description:"备份黑名单列表，以 ; 分隔；仅逻辑备份下生效，快照备份下无效"`
	ForceBackup  bool   `json:"force_backup,omitempty" jsonschema_description:"逻辑备份时是否使用 --force 参数（true使用）；物理备份此参数无效"`
	UseBlacklist bool   `json:"use_blacklist,omitempty" jsonschema_description:"是否使用黑名单备份，默认false"`
}

// ListBackupsFollowUp 将缺失 BackupId 的响应映射为可直接执行的 list_backups 参数。
type ListBackupsFollowUp struct {
	DBID            string `json:"db_id" jsonschema_description:"作为 udb_mysql_list_backups 的 db_id 传入"`
	BeginTime       int64  `json:"begin_time" jsonschema_description:"作为 udb_mysql_list_backups 的 begin_time 传入（备份调用前捕获的Unix秒）"`
	BackupType      string `json:"backup_type" jsonschema_description:"作为 udb_mysql_list_backups 的 backup_type=manual 传入"`
	Zone            string `json:"zone" jsonschema_description:"作为 udb_mysql_list_backups 的 zone 传入"`
	MatchBackupName string `json:"match_backup_name" jsonschema_description:"在当前页items中匹配此 backup_name；API不支持按 BackupName 服务端过滤"`
}

// CreateBackupOutput 是结构化创建备份响应。
type CreateBackupOutput struct {
	DBID                string               `json:"db_id" jsonschema_description:"DB实例Id"`
	BackupName          string               `json:"backup_name" jsonschema_description:"备份名称"`
	Zone                string               `json:"zone,omitempty" jsonschema_description:"可用区"`
	BackupID            int                  `json:"backup_id,omitempty" jsonschema_description:"备份记录ID（API返回时）"`
	BackupIDKnown       bool                 `json:"backup_id_known" jsonschema_description:"API返回了BackupId时为true"`
	RequestedAtUnix     int64                `json:"requested_at_unix" jsonschema_description:"备份调用前捕获的Unix秒，用于BackupId缺失时查找"`
	FollowUpTool        string               `json:"follow_up_tool" jsonschema_description:"下一步建议调用的 MCP 工具（轮询完成或发现BackupId）"`
	Hint                string               `json:"hint" jsonschema_description:"后续指引；服务端不做内部轮询"`
	ListBackupsFollowUp *ListBackupsFollowUp `json:"list_backups_follow_up,omitempty" jsonschema_description:"BackupId缺失时可直接执行的 list_backups 参数"`
}

// DeleteBackupInput 是 udb_mysql_delete_backup 的类型化 MCP 输入。
type DeleteBackupInput struct {
	ScopeInput
	BackupID   int    `json:"backup_id,required" jsonschema_description:"备份id，可通过 DescribeUDBBackup（list_backups）获取"`
	Name       string `json:"name,required" jsonschema_description:"当前备份名称，用于删除前实时比对确认"`
	Zone       string `json:"zone,required" jsonschema_description:"可用区（API必填），参见可用区列表"`
	BackupZone string `json:"backup_zone,omitempty" jsonschema_description:"跨可用区高可用备库所在可用区"`
	Confirm    *bool  `json:"confirm,required" jsonschema_description:"必须为true以确认不可逆删除"`
}

// DeleteBackupOutput 是结构化删除备份响应。
type DeleteBackupOutput struct {
	BackupID   int    `json:"backup_id" jsonschema_description:"已删除的备份id"`
	BackupName string `json:"backup_name" jsonschema_description:"删除前实时确认的备份名称"`
}
