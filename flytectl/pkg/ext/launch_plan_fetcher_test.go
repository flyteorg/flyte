package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	launchPlanListResponse *admin.LaunchPlanList
	lpFilters              = filters.Filters{}
	launchPlan1            *admin.LaunchPlan
)

func getLaunchPlanFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}

	parameterMap := map[string]*core.Parameter{
		"numbers": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
			},
		},
		"numbers_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
		"run_local_at_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
			Behavior: &core.Parameter_Default{
				Default: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 10,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	launchPlan1 = &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v1",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
	}
	launchPlan2 := &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v2",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
	}

	launchPlans := []*admin.LaunchPlan{launchPlan2, launchPlan1}

	launchPlanListResponse = &admin.LaunchPlanList{
		LaunchPlans: launchPlans,
	}
}

func TestFetchAllVerOfLP(t *testing.T) {
	getLaunchPlanFetcherSetup()
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(launchPlanListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfLP(ctx, "lpName", "project", "domain", lpFilters)
	assert.Nil(t, err)
}

func TestFetchLPVersion(t *testing.T) {
	getLaunchPlanFetcherSetup()
	adminClient.OnGetLaunchPlanMatch(mock.Anything, mock.Anything).Return(launchPlan1, nil)
	_, err := adminFetcherExt.FetchLPVersion(ctx, "launchplan1", "v1", "project", "domain")
	assert.Nil(t, err)
}

func TestFetchAllVerOfLPError(t *testing.T) {
	getLaunchPlanFetcherSetup()
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchAllVerOfLP(ctx, "lpName", "project", "domain", lpFilters)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchAllVerOfLPFilterError(t *testing.T) {
	getLaunchPlanFetcherSetup()
	lpFilters.FieldSelector = "hello="
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Please add a valid field selector"))
	_, err := adminFetcherExt.FetchAllVerOfLP(ctx, "lpName", "project", "domain", lpFilters)
	assert.Equal(t, fmt.Errorf("Please add a valid field selector"), err)
}

func TestFetchAllVerOfLPEmptyResponse(t *testing.T) {
	launchPlanListResponse := &admin.LaunchPlanList{}
	getLaunchPlanFetcherSetup()
	lpFilters.FieldSelector = ""
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(launchPlanListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfLP(ctx, "lpName", "project", "domain", lpFilters)
	assert.Equal(t, fmt.Errorf("no launchplans retrieved for lpName"), err)
}

func TestFetchLPLatestVersion(t *testing.T) {
	getLaunchPlanFetcherSetup()
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(launchPlanListResponse, nil)
	_, err := adminFetcherExt.FetchLPLatestVersion(ctx, "lpName", "project", "domain", lpFilters)
	assert.Nil(t, err)
}

func TestFetchLPLatestVersionError(t *testing.T) {
	launchPlanListResponse := &admin.LaunchPlanList{}
	getLaunchPlanFetcherSetup()
	adminClient.OnListLaunchPlansMatch(mock.Anything, mock.Anything).Return(launchPlanListResponse, nil)
	_, err := adminFetcherExt.FetchLPLatestVersion(ctx, "lpName", "project", "domain", lpFilters)
	assert.Equal(t, fmt.Errorf("no launchplans retrieved for lpName"), err)
}
