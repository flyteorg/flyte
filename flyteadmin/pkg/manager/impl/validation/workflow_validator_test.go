package validation

import (
	"context"
	"errors"
	"testing"

	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

var workflowConfig = testutils.GetApplicationConfigWithDefaultDomains()

func TestValidateWorkflowEmptyProject(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	request.Id.Project = ""
	err := ValidateWorkflow(context.Background(), request, testutils.GetRepoWithDefaultProject(), workflowConfig)
	assert.EqualError(t, err, "missing project")
}

func TestValidatWorkflowInvalidProjectAndDomain(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	err := ValidateWorkflow(context.Background(), request, testutils.GetRepoWithDefaultProjectAndErr(errors.New("foo")),
		workflowConfig)
	assert.EqualError(t, err, "failed to validate that project [project] and domain [domain] are registered, err: [foo]")
}

func TestValidateWorkflowEmptyDomain(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	request.Id.Domain = ""
	err := ValidateWorkflow(context.Background(), request, repositoryMocks.NewMockRepository(), workflowConfig)
	assert.EqualError(t, err, "missing domain")
}

func TestValidateWorkflowEmptyName(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	request.Id.Name = ""
	err := ValidateWorkflow(context.Background(), request, repositoryMocks.NewMockRepository(), workflowConfig)
	assert.EqualError(t, err, "missing name")
}

func TestValidateCompiledWorkflow(t *testing.T) {
	mockConfig := runtimeMocks.MockRegistrationValidationProvider{
		WorkflowSizeLimit: "1",
	}
	workflowClosure := admin.WorkflowClosure{
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Connections: &core.ConnectionSet{
					Downstream: map[string]*core.ConnectionSet_IdList{
						"such": {
							Ids: []string{
								"great",
								"amounts",
								"of",
								"data",
								"are",
								"being",
								"encapsulated",
								"here",
								"too",
								"much",
								"i",
								"dare",
								"say",
							},
						},
					},
				},
			},
		},
	}
	err := ValidateCompiledWorkflow(core.Identifier{}, workflowClosure, &mockConfig)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "Workflow closure size exceeds max limit [1]")
}
