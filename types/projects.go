package types

// ListProjectsInput 是 udb_mysql_list_projects 的类型化 MCP 输入。
type ListProjectsInput struct {
	Name         string `json:"name,omitempty" jsonschema_description:"项目名称精确匹配，如 UDB临时测试(小于7天)"`
	NameContains string `json:"name_contains,omitempty" jsonschema_description:"项目名称子串过滤（name 为空时生效）"`
}

// ProjectOutput 是 MCP 响应中的项目条目。
type ProjectOutput struct {
	ProjectID   string `json:"project_id" jsonschema_description:"UCloud项目ID（org-xxx）"`
	ProjectName string `json:"project_name" jsonschema_description:"项目名称"`
	IsDefault   bool   `json:"is_default" jsonschema_description:"是否为账号默认项目"`
	MemberCount int    `json:"member_count,omitempty" jsonschema_description:"项目成员数（可用时）"`
	CreatedAt   string `json:"created_at,omitempty" jsonschema_description:"创建时间（UTC RFC3339，可用时）"`
}

// ListProjectsOutput 是 udb_mysql_list_projects 的结构化响应。
type ListProjectsOutput struct {
	TotalCount    int             `json:"total_count" jsonschema_description:"账号API返回的项目总数（名称过滤前）"`
	ReturnedCount int             `json:"returned_count" jsonschema_description:"名称过滤后返回的项目数"`
	Projects      []ProjectOutput `json:"projects" jsonschema_description:"匹配的项目列表"`
}
