package ext

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminFetcherExtClient) FetchWorkflowAttributes(ctx context.Context, project, domain, name string,
	rsType admin.MatchableResource) (*admin.WorkflowAttributesGetResponse, error) {
	workflowAttr, err := a.AdminServiceClient().GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     name,
		ResourceType: rsType,
	})
	if err != nil {
		return nil, err
	}
	if workflowAttr.GetAttributes() == nil || workflowAttr.GetAttributes().GetMatchingAttributes() == nil {
		return nil, fmt.Errorf("attribute doesn't exist")
	}
	return workflowAttr, nil
}

func (a *AdminFetcherExtClient) FetchProjectDomainAttributes(ctx context.Context, project, domain string,
	rsType admin.MatchableResource) (*admin.ProjectDomainAttributesGetResponse, error) {
	projectDomainAttr, err := a.AdminServiceClient().GetProjectDomainAttributes(ctx,
		&admin.ProjectDomainAttributesGetRequest{
			Project:      project,
			Domain:       domain,
			ResourceType: rsType,
		})
	if err != nil {
		return nil, err
	}
	if projectDomainAttr.GetAttributes() == nil || projectDomainAttr.GetAttributes().GetMatchingAttributes() == nil {
		return nil, fmt.Errorf("attribute doesn't exist")
	}
	return projectDomainAttr, nil
}
