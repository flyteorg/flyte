package ext

import (
	"context"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

//go:generate mockery -all -case=underscore

// AdminFetcherExtInterface Interface for exposing the fetch capabilities from the admin and also allow this to be injectable into other
// modules. eg : create execution which requires to fetch launchplan details to construct the execution spec.
type AdminFetcherExtInterface interface {
	AdminServiceClient() service.AdminServiceClient

	// FetchExecution fetches the execution based on name, project, domain
	FetchExecution(ctx context.Context, name, project, domain string) (*admin.Execution, error)

	// ListExecution fetches the all versions of  based on name, project, domain
	ListExecution(ctx context.Context, project, domain string, filter filters.Filters) (*admin.ExecutionList, error)

	// FetchAllVerOfLP fetches all versions of launch plan in a  project, domain
	FetchAllVerOfLP(ctx context.Context, lpName, project, domain string, filter filters.Filters) ([]*admin.LaunchPlan, error)

	// FetchLPLatestVersion fetches latest version of launch plan in a  project, domain
	FetchLPLatestVersion(ctx context.Context, name, project, domain string, filter filters.Filters) (*admin.LaunchPlan, error)

	// FetchLPVersion fetches particular version of launch plan in a  project, domain
	FetchLPVersion(ctx context.Context, name, version, project, domain string) (*admin.LaunchPlan, error)

	// FetchAllVerOfTask fetches all versions of task in a  project, domain
	FetchAllVerOfTask(ctx context.Context, name, project, domain string, filter filters.Filters) ([]*admin.Task, error)

	// FetchTaskLatestVersion fetches latest version of task in a  project, domain
	FetchTaskLatestVersion(ctx context.Context, name, project, domain string, filter filters.Filters) (*admin.Task, error)

	// FetchTaskVersion fetches particular version of task in a  project, domain
	FetchTaskVersion(ctx context.Context, name, version, project, domain string) (*admin.Task, error)

	// FetchAllVerOfWorkflow fetches all versions of task in a  project, domain
	FetchAllVerOfWorkflow(ctx context.Context, name, project, domain string, filter filters.Filters) ([]*admin.Workflow, error)

	// FetchWorkflowLatestVersion fetches latest version of workflow in a  project, domain
	FetchWorkflowLatestVersion(ctx context.Context, name, project, domain string, filter filters.Filters) (*admin.Workflow, error)

	// FetchWorkflowVersion fetches particular version of workflow in a  project, domain
	FetchWorkflowVersion(ctx context.Context, name, version, project, domain string) (*admin.Workflow, error)

	// FetchWorkflowAttributes fetches workflow attributes particular resource type in a  project, domain and workflow
	FetchWorkflowAttributes(ctx context.Context, project, domain, name string, rsType admin.MatchableResource) (*admin.WorkflowAttributesGetResponse, error)

	// FetchProjectDomainAttributes fetches project domain attributes particular resource type in a  project, domain
	FetchProjectDomainAttributes(ctx context.Context, project, domain string, rsType admin.MatchableResource) (*admin.ProjectDomainAttributesGetResponse, error)

	// ListProjects fetches all projects
	ListProjects(ctx context.Context, filter filters.Filters) (*admin.Projects, error)
}

// AdminFetcherExtClient is used for interacting with extended features used for fetching data from admin service
type AdminFetcherExtClient struct {
	AdminClient service.AdminServiceClient
}

func (a *AdminFetcherExtClient) AdminServiceClient() service.AdminServiceClient {
	if a == nil {
		return nil
	}
	return a.AdminClient
}
