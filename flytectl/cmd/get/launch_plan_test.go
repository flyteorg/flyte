package get

import (
	"fmt"
	"os"
	"testing"

	structpb "github.com/golang/protobuf/ptypes/struct"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flytectl/pkg/printer"

	"github.com/flyteorg/flytectl/pkg/filters"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/launchplan"
	"github.com/flyteorg/flytectl/pkg/ext/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	resourceListRequest            *admin.ResourceListRequest
	resourceGetRequest             *admin.ResourceListRequest
	objectGetRequest               *admin.ObjectGetRequest
	namedIDRequest                 *admin.NamedEntityIdentifierListRequest
	launchPlanListResponse         *admin.LaunchPlanList
	filteredLaunchPlanListResponse *admin.LaunchPlanList
	argsLp                         []string
	namedIdentifierList            *admin.NamedEntityIdentifierList
	launchPlan2                    *admin.LaunchPlan
)

func getLaunchPlanSetup() {
	// TODO: migrate to new command context from testutils
	argsLp = []string{"launchplan1"}
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
				Description: "short desc",
			},
		},
		"numbers_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
				Description: "long description will be truncated in table",
			},
		},
		"run_local_at_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
				Description: "run_local_at_count",
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
		"generic": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_STRUCT,
					},
				},
				Description: "generic",
			},
			Behavior: &core.Parameter_Default{
				Default: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Generic{
								Generic: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"foo": {Kind: &structpb.Value_StringValue{StringValue: "foo"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	launchPlan1 := &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v1",
		},
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Name: "workflow1",
			},
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
	launchPlan2 = &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v2",
		},
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Name: "workflow2",
			},
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

	resourceListRequest = &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}

	resourceGetRequest = &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    argsLp[0],
		},
	}

	launchPlanListResponse = &admin.LaunchPlanList{
		LaunchPlans: launchPlans,
	}

	filteredLaunchPlanListResponse = &admin.LaunchPlanList{
		LaunchPlans: []*admin.LaunchPlan{launchPlan2},
	}

	objectGetRequest = &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Project:      projectValue,
			Domain:       domainValue,
			Name:         argsLp[0],
			Version:      "v2",
		},
	}

	namedIDRequest = &admin.NamedEntityIdentifierListRequest{
		Project: projectValue,
		Domain:  domainValue,
	}

	var entities []*admin.NamedEntityIdentifier
	id1 := &admin.NamedEntityIdentifier{
		Project: projectValue,
		Domain:  domainValue,
		Name:    "launchplan1",
	}
	id2 := &admin.NamedEntityIdentifier{
		Project: projectValue,
		Domain:  domainValue,
		Name:    "launchplan2",
	}
	entities = append(entities, id1, id2)
	namedIdentifierList = &admin.NamedEntityIdentifierList{
		Entities: entities,
	}

	launchplan.DefaultConfig.Latest = false
	launchplan.DefaultConfig.Version = ""
	launchplan.DefaultConfig.ExecFile = ""
	launchplan.DefaultConfig.Filter = filters.Filters{}
}

func TestGetLaunchPlanFuncWithError(t *testing.T) {
	t.Run("failure fetch latest", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		launchplan.DefaultConfig.Latest = true
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher.OnFetchLPLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		_, err := FetchLPForName(s.Ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching version ", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		launchplan.DefaultConfig.Version = "v1"
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher.OnFetchLPVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything).Return(nil, fmt.Errorf("error fetching version"))
		_, err := FetchLPForName(s.Ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching all version ", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		launchplan.DefaultConfig.Filter = filters.Filters{}
		launchplan.DefaultConfig.Filter = filters.Filters{}
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		mockFetcher.OnFetchAllVerOfLPMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching all version"))
		_, err := FetchLPForName(s.Ctx, mockFetcher, "lpName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching ", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		s.FetcherExt.OnFetchAllVerOfLP(s.Ctx, "launchplan1", "dummyProject", "dummyDomain", filters.Filters{}).Return(nil, fmt.Errorf("error fetching all version"))
		s.MockAdminClient.OnListLaunchPlansMatch(s.Ctx, resourceGetRequest).Return(nil, fmt.Errorf("error fetching all version"))
		s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, objectGetRequest).Return(nil, fmt.Errorf("error fetching lanuch plan"))
		s.MockAdminClient.OnListLaunchPlanIdsMatch(s.Ctx, namedIDRequest).Return(nil, fmt.Errorf("error listing lanuch plan ids"))
		err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching list", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		argsLp = []string{}
		s.FetcherExt.OnFetchAllVerOfLP(s.Ctx, "", "dummyProject", "dummyDomain", filters.Filters{}).Return(nil, fmt.Errorf("error fetching all version"))
		s.MockAdminClient.OnListLaunchPlansMatch(s.Ctx, resourceListRequest).Return(nil, fmt.Errorf("error fetching all version"))
		err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
		assert.NotNil(t, err)
	})
}

func TestGetLaunchPlanFunc(t *testing.T) {
	s := setup()
	getLaunchPlanSetup()
	s.FetcherExt.OnFetchAllVerOfLPMatch(mock.Anything, mock.Anything, "dummyProject", "dummyDomain", filters.Filters{}).Return(launchPlanListResponse.LaunchPlans, nil)
	err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "FetchAllVerOfLP", s.Ctx, "launchplan1", "dummyProject", "dummyDomain", launchplan.DefaultConfig.Filter)
	s.TearDownAndVerify(t, `[{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}},{"id": {"name": "launchplan1","version": "v1"},"spec": {"workflowId": {"name": "workflow1"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:00Z"}}]`)
}

func TestGetLaunchPlanFuncLatest(t *testing.T) {
	s := setup()
	getLaunchPlanSetup()
	launchplan.DefaultConfig.Latest = true
	launchplan.DefaultConfig.Filter = filters.Filters{}
	s.FetcherExt.OnFetchLPLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(launchPlan2, nil)
	err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "FetchLPLatestVersion", s.Ctx, "launchplan1", projectValue, domainValue, launchplan.DefaultConfig.Filter)
	s.TearDownAndVerify(t, `{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}}`)
}

func TestGetLaunchPlanWithVersion(t *testing.T) {
	s := testutils.SetupWithExt()
	getLaunchPlanSetup()
	launchplan.DefaultConfig.Version = "v2"
	s.FetcherExt.OnFetchLPVersion(s.Ctx, "launchplan1", "v2", "dummyProject", "dummyDomain").Return(launchPlan2, nil)
	err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "FetchLPVersion", s.Ctx, "launchplan1", "v2", "dummyProject", "dummyDomain")
	s.TearDownAndVerify(t, `{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}}`)
}

func TestGetLaunchPlans(t *testing.T) {
	t.Run("no workflow filter", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		s.FetcherExt.OnFetchAllVerOfLP(s.Ctx, "", "dummyProject", "dummyDomain", filters.Filters{}).Return(launchPlanListResponse.LaunchPlans, nil)
		argsLp = []string{}
		err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `[{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}},{"id": {"name": "launchplan1","version": "v1"},"spec": {"workflowId": {"name": "workflow1"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:00Z"}}]`)
	})
	t.Run("workflow filter", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		s.FetcherExt.OnFetchAllVerOfLP(s.Ctx, "", "dummyProject", "dummyDomain", filters.Filters{
			FieldSelector: "workflow.name=workflow2",
		}).Return(launchPlanListResponse.LaunchPlans, nil)
		argsLp = []string{}
		launchplan.DefaultConfig.Workflow = "workflow2"
		err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
		assert.Nil(t, err)
		s.TearDownAndVerify(t, `[{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}},{"id": {"name": "launchplan1","version": "v1"},"spec": {"workflowId": {"name": "workflow1"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:00Z"}}]`)
	})
	t.Run("workflow filter error", func(t *testing.T) {
		s := setup()
		getLaunchPlanSetup()
		argsLp = []string{}
		launchplan.DefaultConfig.Workflow = "workflow2"
		launchplan.DefaultConfig.Filter.FieldSelector = "workflow.name"
		err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("fieldSelector cannot be specified with workflow flag"), err)
	})
}

func TestGetLaunchPlansWithExecFile(t *testing.T) {
	s := testutils.SetupWithExt()
	getLaunchPlanSetup()
	s.MockAdminClient.OnListLaunchPlansMatch(s.Ctx, resourceListRequest).Return(launchPlanListResponse, nil)
	s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, objectGetRequest).Return(launchPlan2, nil)
	s.MockAdminClient.OnListLaunchPlanIdsMatch(s.Ctx, namedIDRequest).Return(namedIdentifierList, nil)
	s.FetcherExt.OnFetchLPVersion(s.Ctx, "launchplan1", "v2", "dummyProject", "dummyDomain").Return(launchPlan2, nil)
	launchplan.DefaultConfig.Version = "v2"
	launchplan.DefaultConfig.ExecFile = testDataFolder + "exec_file"
	err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
	assert.Nil(t, err)

	data, err := os.ReadFile(launchplan.DefaultConfig.ExecFile)
	assert.Nil(t, err)
	assert.Equal(t, `iamRoleARN: ""
inputs:
    generic:
        foo: foo
    numbers:
        - 0
    numbers_count: 0 # long description will be truncated in table
    run_local_at_count: 10 # short desc
envs: {}
kubeServiceAcct: ""
targetDomain: ""
targetProject: ""
version: v2
workflow: launchplan1
`, string(data))
	os.Remove(launchplan.DefaultConfig.ExecFile)

	s.FetcherExt.AssertCalled(t, "FetchLPVersion", s.Ctx, "launchplan1", "v2", "dummyProject", "dummyDomain")
	s.TearDownAndVerify(t, `{"id": {"name": "launchplan1","version": "v2"},"spec": {"workflowId": {"name": "workflow2"},"defaultInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}}},"closure": {"expectedInputs": {"parameters": {"generic": {"var": {"type": {"simple": "STRUCT"},"description": "generic"},"default": {"scalar": {"generic": {"foo": "foo"}}}},"numbers": {"var": {"type": {"collectionType": {"simple": "INTEGER"}},"description": "short desc"}},"numbers_count": {"var": {"type": {"simple": "INTEGER"},"description": "long description will be truncated in table"}},"run_local_at_count": {"var": {"type": {"simple": "INTEGER"},"description": "run_local_at_count"},"default": {"scalar": {"primitive": {"integer": "10"}}}}}},"createdAt": "1970-01-01T00:00:01Z"}}`)
}

func TestGetLaunchPlanTableFunc(t *testing.T) {
	s := testutils.SetupWithExt()
	getLaunchPlanSetup()
	s.MockAdminClient.OnListLaunchPlansMatch(s.Ctx, resourceGetRequest).Return(launchPlanListResponse, nil)
	s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, objectGetRequest).Return(launchPlan2, nil)
	s.MockAdminClient.OnListLaunchPlanIdsMatch(s.Ctx, namedIDRequest).Return(namedIdentifierList, nil)
	s.FetcherExt.OnFetchAllVerOfLP(s.Ctx, "launchplan1", "dummyProject", "dummyDomain", filters.Filters{}).Return(launchPlanListResponse.LaunchPlans, nil)
	config.GetConfig().Output = printer.OutputFormatTABLE.String()
	err := getLaunchPlanFunc(s.Ctx, argsLp, s.CmdCtx)
	assert.Nil(t, err)
	s.FetcherExt.AssertCalled(t, "FetchAllVerOfLP", s.Ctx, "launchplan1", "dummyProject", "dummyDomain", filters.Filters{})
	s.TearDownAndVerify(t, `
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| VERSION | NAME        | TYPE | STATE | SCHEDULE | INPUTS                    | OUTPUTS | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| v2      | launchplan1 |      |       |          | generic                   |         |
|         |             |      |       |          | numbers: short desc       |         |
|         |             |      |       |          | numbers_count: long de... |         |
|         |             |      |       |          | run_local_at_count        |         | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
| v1      | launchplan1 |      |       |          | generic                   |         |
|         |             |      |       |          | numbers: short desc       |         |
|         |             |      |       |          | numbers_count: long de... |         |
|         |             |      |       |          | run_local_at_count        |         | 
--------- ------------- ------ ------- ---------- --------------------------- --------- 
2 rows`)
}
