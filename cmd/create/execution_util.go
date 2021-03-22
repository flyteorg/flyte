package create

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdGet "github.com/flyteorg/flytectl/cmd/get"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/google/uuid"
	"sigs.k8s.io/yaml"
)

func createExecutionRequestForWorkflow(ctx context.Context, workflowName string, project string, domain string, cmdCtx cmdCore.CommandContext) (*admin.ExecutionCreateRequest, error) {
	var lp *admin.LaunchPlan
	var err error
	// Fetch the launch plan
	if lp, err = cmdGet.FetchLPVersion(ctx, workflowName, executionConfig.Version, project, domain, cmdCtx); err != nil {
		return nil, err
	}
	// Create workflow params literal map
	var paramLiterals map[string]*core.Literal
	workflowParams := cmdGet.WorkflowParams(lp)
	if paramLiterals, err = MakeLiteralForParams(executionConfig.Inputs, workflowParams); err != nil {
		return nil, err
	}
	var inputs = &core.LiteralMap{
		Literals: paramLiterals,
	}
	ID := lp.Id
	return createExecutionRequest(ID, inputs, nil), nil
}

func createExecutionRequestForTask(ctx context.Context, taskName string, project string, domain string, cmdCtx cmdCore.CommandContext) (*admin.ExecutionCreateRequest, error) {
	var task *admin.Task
	var err error
	// Fetch the task
	if task, err = cmdGet.FetchTaskVersion(ctx, taskName, executionConfig.Version, project, domain, cmdCtx); err != nil {
		return nil, err
	}
	// Create task variables literal map
	var variableLiterals map[string]*core.Literal
	taskInputs := cmdGet.TaskInputs(task)
	if variableLiterals, err = MakeLiteralForVariables(executionConfig.Inputs, taskInputs); err != nil {
		return nil, err
	}
	var inputs = &core.LiteralMap{
		Literals: variableLiterals,
	}
	var authRole *admin.AuthRole
	if executionConfig.KubeServiceAcct != "" {
		authRole = &admin.AuthRole{Method: &admin.AuthRole_KubernetesServiceAccount{
			KubernetesServiceAccount: executionConfig.KubeServiceAcct}}
	} else {
		authRole = &admin.AuthRole{Method: &admin.AuthRole_AssumableIamRole{
			AssumableIamRole: executionConfig.IamRoleARN}}
	}
	ID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      project,
		Domain:       domain,
		Name:         task.Id.Name,
		Version:      task.Id.Version,
	}
	return createExecutionRequest(ID, inputs, authRole), nil
}

func createExecutionRequest(ID *core.Identifier, inputs *core.LiteralMap, authRole *admin.AuthRole) *admin.ExecutionCreateRequest {
	return &admin.ExecutionCreateRequest{
		Project: executionConfig.TargetProject,
		Domain:  executionConfig.TargetDomain,
		Name:    "f" + strings.ReplaceAll(uuid.New().String(), "-", "")[:19],
		Spec: &admin.ExecutionSpec{
			LaunchPlan: ID,
			Metadata: &admin.ExecutionMetadata{
				Mode:      admin.ExecutionMetadata_MANUAL,
				Principal: "sdk",
				Nesting:   0,
			},
			AuthRole: authRole,
		},
		Inputs: inputs,
	}
}

func readExecConfigFromFile(fileName string) (*ExecutionConfig, error) {
	data, _err := ioutil.ReadFile(fileName)
	if _err != nil {
		return nil, fmt.Errorf("unable to read from %v yaml file", fileName)
	}
	executionConfigRead := ExecutionConfig{}
	if _err = yaml.Unmarshal(data, &executionConfigRead); _err != nil {
		return nil, _err
	}
	return &executionConfigRead, nil
}

func resolveOverrides(readExecutionConfig *ExecutionConfig, project string, domain string) {
	if executionConfig.KubeServiceAcct != "" {
		readExecutionConfig.KubeServiceAcct = executionConfig.KubeServiceAcct
	}
	if executionConfig.IamRoleARN != "" {
		readExecutionConfig.IamRoleARN = executionConfig.IamRoleARN
	}
	if executionConfig.TargetProject != "" {
		readExecutionConfig.TargetProject = executionConfig.TargetProject
	}
	if executionConfig.TargetDomain != "" {
		readExecutionConfig.TargetDomain = executionConfig.TargetDomain
	}
	// Use the root project and domain to launch the task/workflow if target is unspecified
	if executionConfig.TargetProject == "" {
		readExecutionConfig.TargetProject = project
	}
	if executionConfig.TargetDomain == "" {
		readExecutionConfig.TargetDomain = domain
	}
}

func readConfigAndValidate(project string, domain string) (ExecutionParams, error) {
	executionParams := ExecutionParams{}
	if executionConfig.ExecFile == "" {
		return executionParams, errors.New("executionConfig can't be empty. Run the flytectl get task/launchplan to generate the config")
	}
	var readExecutionConfig *ExecutionConfig
	var err error
	if readExecutionConfig, err = readExecConfigFromFile(executionConfig.ExecFile); err != nil {
		return executionParams, err
	}
	resolveOverrides(readExecutionConfig, project, domain)
	// Update executionConfig pointer to readExecutionConfig as it contains all the updates.
	executionConfig = readExecutionConfig
	isTask := readExecutionConfig.Task != ""
	isWorkflow := readExecutionConfig.Workflow != ""
	if isTask == isWorkflow {
		return executionParams, errors.New("either one of task or workflow name should be specified to launch an execution")
	}
	name := readExecutionConfig.Task
	if !isTask {
		name = readExecutionConfig.Workflow
	}
	return ExecutionParams{name: name, isTask: isTask}, nil
}
