package ext

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func (a *AdminFetcherExtClient) FetchWorkflowAttributes(ctx context.Context, project, domain, name string,
	rsType admin.MatchableResource) (*admin.WorkflowAttributesGetResponse, error) {
	response, err := a.AdminServiceClient().GetWorkflowAttributes(ctx, &admin.WorkflowAttributesGetRequest{
		Project:      project,
		Domain:       domain,
		Workflow:     name,
		ResourceType: rsType,
	})
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}
	if status.Code(err) == codes.NotFound ||
		response.GetAttributes() == nil ||
		response.GetAttributes().GetMatchingAttributes() == nil {
		return nil, NewNotFoundError("attribute")
	}
	return response, nil
}

func (a *AdminFetcherExtClient) FetchProjectDomainAttributes(ctx context.Context, project, domain string,
	rsType admin.MatchableResource) (*admin.ProjectDomainAttributesGetResponse, error) {
	response, err := a.AdminServiceClient().GetProjectDomainAttributes(ctx,
		&admin.ProjectDomainAttributesGetRequest{
			Project:      project,
			Domain:       domain,
			ResourceType: rsType,
		})
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}
	if status.Code(err) == codes.NotFound ||
		response.GetAttributes() == nil ||
		response.GetAttributes().GetMatchingAttributes() == nil {
		return nil, NewNotFoundError("attribute")
	}
	return response, nil
}

func (a *AdminFetcherExtClient) FetchProjectAttributes(ctx context.Context, project string,
	rsType admin.MatchableResource) (*admin.ProjectAttributesGetResponse, error) {
	response, err := a.AdminServiceClient().GetProjectAttributes(ctx,
		&admin.ProjectAttributesGetRequest{
			Project:      project,
			ResourceType: rsType,
		})
	if err != nil && status.Code(err) != codes.NotFound {
		return nil, err
	}
	if status.Code(err) == codes.NotFound ||
		response.GetAttributes() == nil ||
		response.GetAttributes().GetMatchingAttributes() == nil {
		return nil, NewNotFoundError("attribute")
	}
	return response, nil
}
