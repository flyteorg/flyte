package ext

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminUpdaterExtClient) UpdateWorkflowAttributes(ctx context.Context, project, domain, name string, matchingAttr *admin.MatchingAttributes) error {
	_, err := a.AdminServiceClient().UpdateWorkflowAttributes(ctx, &admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            project,
			Domain:             domain,
			Workflow:           name,
			MatchingAttributes: matchingAttr,
		},
	})
	return err
}

func (a *AdminUpdaterExtClient) UpdateProjectDomainAttributes(ctx context.Context, project, domain string, matchingAttr *admin.MatchingAttributes) error {
	_, err := a.AdminServiceClient().UpdateProjectDomainAttributes(ctx,
		&admin.ProjectDomainAttributesUpdateRequest{
			Attributes: &admin.ProjectDomainAttributes{
				Project:            project,
				Domain:             domain,
				MatchingAttributes: matchingAttr,
			},
		})
	return err
}

func (a *AdminUpdaterExtClient) UpdateProjectAttributes(ctx context.Context, project string, matchingAttr *admin.MatchingAttributes) error {
	_, err := a.AdminServiceClient().UpdateProjectAttributes(ctx,
		&admin.ProjectAttributesUpdateRequest{
			Attributes: &admin.ProjectAttributes{
				Project:            project,
				MatchingAttributes: matchingAttr,
			},
		})
	return err
}
