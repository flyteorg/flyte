package ext

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminDeleterExtClient) DeleteWorkflowAttributes(ctx context.Context, project, domain, name string, rsType admin.MatchableResource) error {
	_, err := a.AdminServiceClient().DeleteWorkflowAttributes(ctx, &admin.WorkflowAttributesDeleteRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     name,
		ResourceType: rsType,
	})
	return err
}

func (a *AdminDeleterExtClient) DeleteProjectDomainAttributes(ctx context.Context, project, domain string, rsType admin.MatchableResource) error {
	_, err := a.AdminServiceClient().DeleteProjectDomainAttributes(ctx, &admin.ProjectDomainAttributesDeleteRequest{
		Project:      project,
		Domain:       domain,
		ResourceType: rsType,
	})
	return err
}

func (a *AdminDeleterExtClient) DeleteProjectAttributes(ctx context.Context, project string, rsType admin.MatchableResource) error {
	_, err := a.AdminServiceClient().DeleteProjectAttributes(ctx, &admin.ProjectAttributesDeleteRequest{
		Project:      project,
		ResourceType: rsType,
	})
	return err
}
