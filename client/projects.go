package client

import (
	"context"
	"strings"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"udb-mysql-mcp-server/types"
)

const getProjectListAction = "GetProjectList"

// ListProjects returns account projects, optionally filtered by name.
func (c *Client) ListProjects(ctx context.Context, reqCtx CallContext, in types.ListProjectsInput) (types.ListProjectsOutput, error) {
	sdk, err := c.uaccountClient(reqCtx)
	if err != nil {
		return types.ListProjectsOutput{}, err
	}

	req := sdk.NewGetProjectListRequest()
	prepareRequest(&req.CommonBase, ctx, c.factory.Timeout)
	resp, err := sdk.GetProjectList(req)
	if err != nil {
		return types.ListProjectsOutput{}, mapSDKError(getProjectListAction, err)
	}

	projects := filterProjects(resp.ProjectSet, in)
	return types.ListProjectsOutput{
		TotalCount:    resp.ProjectCount,
		ReturnedCount: len(projects),
		Projects:      projects,
	}, nil
}

func filterProjects(items []uaccount.ProjectListInfo, in types.ListProjectsInput) []types.ProjectOutput {
	exact := strings.TrimSpace(in.Name)
	contains := strings.TrimSpace(in.NameContains)
	out := make([]types.ProjectOutput, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.ProjectName)
		if exact != "" && name != exact {
			continue
		}
		if exact == "" && contains != "" && !strings.Contains(name, contains) {
			continue
		}
		out = append(out, mapProject(item))
	}
	return out
}

func mapProject(item uaccount.ProjectListInfo) types.ProjectOutput {
	out := types.ProjectOutput{
		ProjectID:   strings.TrimSpace(item.ProjectId),
		ProjectName: strings.TrimSpace(item.ProjectName),
		IsDefault:   item.IsDefault,
		MemberCount: item.MemberCount,
	}
	if item.CreateTime > 0 {
		out.CreatedAt = time.Unix(int64(item.CreateTime), 0).UTC().Format(time.RFC3339)
	}
	return out
}
