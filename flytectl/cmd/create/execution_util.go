package create

import (
	"context"
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

func createExecutionRequestForWorkflow(ctx context.Context, workflowName, project, domain string,
	cmdCtx cmdCore.CommandContext) (*admin.ExecutionCreateRequest, error) {
	var lp *admin.LaunchPlan
	var err error

	// Fetch the launch plan
	if lp, err = cmdCtx.AdminFetcherExt().FetchLPVersion(ctx, workflowName, executionConfig.Version, project, domain); err != nil {
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

func createExecutionRequestForTask(ctx context.Context, taskName string, project string, domain string,
	cmdCtx cmdCore.CommandContext) (*admin.ExecutionCreateRequest, error) {
	var task *admin.Task
	var err error
	// Fetch the task
	if task, err = cmdCtx.AdminFetcherExt().FetchTaskVersion(ctx, taskName, executionConfig.Version, project, domain); err != nil {
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
	if executionConfig.KubeServiceAcct != "" || executionConfig.IamRoleARN != "" {
		authRole = &admin.AuthRole{
			AssumableIamRole:         executionConfig.IamRoleARN,
			KubernetesServiceAccount: executionConfig.KubeServiceAcct,
		}
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

func relaunchExecution(ctx context.Context, executionName string, project string, domain string,
	cmdCtx cmdCore.CommandContext) error {
	relaunchedExec, err := cmdCtx.AdminClient().RelaunchExecution(ctx, &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    executionName,
			Project: project,
			Domain:  domain,
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("execution identifier %v\n", relaunchedExec.Id)
	return nil
}

func createExecutionRequest(ID *core.Identifier, inputs *core.LiteralMap,
	authRole *admin.AuthRole) *admin.ExecutionCreateRequest {
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

func resolveOverrides(toBeOverridden *ExecutionConfig, project string, domain string) {
	if executionConfig.KubeServiceAcct != "" {
		toBeOverridden.KubeServiceAcct = executionConfig.KubeServiceAcct
	}
	if executionConfig.IamRoleARN != "" {
		toBeOverridden.IamRoleARN = executionConfig.IamRoleARN
	}
	if executionConfig.TargetProject != "" {
		toBeOverridden.TargetProject = executionConfig.TargetProject
	}
	if executionConfig.TargetDomain != "" {
		toBeOverridden.TargetDomain = executionConfig.TargetDomain
	}
	// Use the root project and domain to launch the task/workflow if target is unspecified
	if executionConfig.TargetProject == "" {
		toBeOverridden.TargetProject = project
	}
	if executionConfig.TargetDomain == "" {
		toBeOverridden.TargetDomain = domain
	}
}

func readConfigAndValidate(project string, domain string) (ExecutionParams, error) {
	executionParams := ExecutionParams{}
	if executionConfig.ExecFile == "" && executionConfig.Relaunch == "" {
		return executionParams, fmt.Errorf("executionConfig or relaunch can't be empty." +
			" Run the flytectl get task/launchplan to generate the config")
	}
	if executionConfig.Relaunch != "" {
		resolveOverrides(executionConfig, project, domain)
		return ExecutionParams{name: executionConfig.Relaunch, execType: Relaunch}, nil
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
		return executionParams, fmt.Errorf("either one of task or workflow name should be specified" +
			" to launch an execution")
	}
	name := readExecutionConfig.Task
	execType := Task
	if !isTask {
		name = readExecutionConfig.Workflow
		execType = Workflow
	}
	return ExecutionParams{name: name, execType: execType}, nil
}
