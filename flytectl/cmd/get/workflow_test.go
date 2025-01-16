package get

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flytectl/cmd/config"
	"github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/workflow"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/ext/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/filters"
	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	argsWf    []string
	workflow1 *admin.Workflow
	workflows []*admin.Workflow
)

func getWorkflowSetup() {

	variableMap := map[string]*core.Variable{
		"var1": {
			Type: &core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
			Description: "var1",
		},
		"var2": {
			Type: &core.LiteralType{
				Type: &core.LiteralType_CollectionType{
					CollectionType: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
			Description: "var2 long descriptions probably needs truncate",
		},
	}
	workflow1 = &admin.Workflow{
		Id: &core.Identifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    "workflow1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Interface: &core.TypedInterface{
							Inputs: &core.VariableMap{
								Variables: variableMap,
							},
						},
					},
				},
			},
		},
	}
	workflow2 := &admin.Workflow{
		Id: &core.Identifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    "workflow2",
			Version: "v2",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
		},
	}
	workflows = []*admin.Workflow{workflow1, workflow2}
	argsWf = []string{"workflow1"}
	workflow.DefaultConfig.Latest = false
	workflow.DefaultConfig.Version = ""
	workflow.DefaultConfig.Filter = filters.DefaultFilter
}

func TestGetWorkflowFuncWithError(t *testing.T) {
	t.Run("failure fetch latest", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		workflow.DefaultConfig.Latest = true
		mockFetcher.OnFetchWorkflowLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		_, _, err := FetchWorkflowForName(s.Ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching version ", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		workflow.DefaultConfig.Version = "v1"
		mockFetcher.OnFetchWorkflowVersionMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching version"))
		_, _, err := FetchWorkflowForName(s.Ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching all version ", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		mockFetcher := new(mocks.AdminFetcherExtInterface)
		mockFetcher.OnFetchAllVerOfWorkflowMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching all version"))
		_, _, err := FetchWorkflowForName(s.Ctx, mockFetcher, "workflowName", projectValue, domainValue)
		assert.NotNil(t, err)
	})

	t.Run("failure fetching ", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		workflow.DefaultConfig.Latest = true
		args := []string{"workflowName"}
		s.FetcherExt.OnFetchWorkflowLatestVersionMatch(mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching latest version"))
		err := getWorkflowFunc(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
	})

	t.Run("fetching all workflow success", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		var args []string
		s.FetcherExt.OnFetchAllWorkflowsMatch(mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return([]*admin.NamedEntity{}, nil)
		err := getWorkflowFunc(s.Ctx, args, s.CmdCtx)
		assert.Nil(t, err)
	})

	t.Run("fetching all workflow error", func(t *testing.T) {
		s := testutils.Setup(t)

		getWorkflowSetup()
		var args []string
		s.FetcherExt.OnFetchAllWorkflowsMatch(mock.Anything, mock.Anything,
			mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error fetching all workflows"))
		err := getWorkflowFunc(s.Ctx, args, s.CmdCtx)
		assert.NotNil(t, err)
	})

}

func TestGetWorkflowFuncLatestWithTable(t *testing.T) {
	s := testutils.Setup(t)

	getWorkflowSetup()
	workflow.DefaultConfig.Latest = true
	config.GetConfig().Output = printer.OutputFormatTABLE.String()
	s.FetcherExt.OnFetchWorkflowLatestVersionMatch(s.Ctx, "workflow1", projectValue, domainValue).Return(workflow1, nil)
	err := getWorkflowFunc(s.Ctx, argsWf, s.CmdCtx)
	assert.Nil(t, err)
	s.TearDownAndVerify(t, `
 --------- ----------- --------------------------- --------- ---------------------- 
| VERSION | NAME      | INPUTS                    | OUTPUTS | CREATED AT           |
 --------- ----------- --------------------------- --------- ---------------------- 
| v1      | workflow1 | var1                      |         | 1970-01-01T00:00:00Z |
|         |           | var2: var2 long descri... |         |                      |
 --------- ----------- --------------------------- --------- ---------------------- 
1 rows`)
}

func TestListWorkflowFuncWithTable(t *testing.T) {
	s := testutils.Setup(t)

	getWorkflowSetup()
	workflow.DefaultConfig.Filter = filters.Filters{}
	config.GetConfig().Output = printer.OutputFormatTABLE.String()
	s.FetcherExt.OnFetchAllVerOfWorkflowMatch(s.Ctx, "workflow1", projectValue, domainValue, filters.Filters{}).Return(workflows, nil)
	err := getWorkflowFunc(s.Ctx, argsWf, s.CmdCtx)
	assert.Nil(t, err)
	s.TearDownAndVerify(t, `
 --------- ----------- ---------------------- 
| VERSION | NAME      | CREATED AT           |
 --------- ----------- ---------------------- 
| v1      | workflow1 | 1970-01-01T00:00:00Z |
 --------- ----------- ---------------------- 
| v2      | workflow2 | 1970-01-01T00:00:00Z |
 --------- ----------- ---------------------- 
2 rows`)
}
