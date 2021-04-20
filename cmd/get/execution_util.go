package get

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"sigs.k8s.io/yaml"
)

// ExecutionConfig is duplicated struct from create with the same structure. This is to avoid the circular dependency.
// TODO : replace this with a cleaner design
type ExecutionConfig struct {
	TargetDomain    string                 `json:"targetDomain"`
	TargetProject   string                 `json:"targetProject"`
	KubeServiceAcct string                 `json:"kubeServiceAcct"`
	IamRoleARN      string                 `json:"iamRoleARN"`
	Workflow        string                 `json:"workflow,omitempty"`
	Task            string                 `json:"task,omitempty"`
	Version         string                 `json:"version"`
	Inputs          map[string]interface{} `json:"inputs"`
}

func (f FetcherImpl) FetchExecution(ctx context.Context, name string, project string, domain string, cmdCtx cmdCore.CommandContext) (*admin.Execution, error) {
	e, err := cmdCtx.AdminClient().GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	})
	if err != nil {
		return nil, err
	}
	return e, nil
}

func WriteExecConfigToFile(executionConfig ExecutionConfig, fileName string) error {
	d, err := yaml.Marshal(executionConfig)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	if _, err = os.Stat(fileName); err == nil {
		if !cmdUtil.AskForConfirmation(fmt.Sprintf("warning file %v will be overwritten", fileName)) {
			return errors.New("backup the file before continuing")
		}
	}
	return ioutil.WriteFile(fileName, d, 0600)
}

func CreateAndWriteExecConfigForTask(task *admin.Task, fileName string) error {
	var err error
	executionConfig := ExecutionConfig{Task: task.Id.Name, Version: task.Id.Version}
	if executionConfig.Inputs, err = ParamMapForTask(task); err != nil {
		return err
	}
	return WriteExecConfigToFile(executionConfig, fileName)
}

func CreateAndWriteExecConfigForWorkflow(wlp *admin.LaunchPlan, fileName string) error {
	var err error
	executionConfig := ExecutionConfig{Workflow: wlp.Id.Name, Version: wlp.Id.Version}
	if executionConfig.Inputs, err = ParamMapForWorkflow(wlp); err != nil {
		return err
	}
	return WriteExecConfigToFile(executionConfig, fileName)
}

func TaskInputs(task *admin.Task) map[string]*core.Variable {
	taskInputs := map[string]*core.Variable{}
	if task == nil || task.Closure == nil {
		return taskInputs
	}
	if task.Closure.CompiledTask == nil {
		return taskInputs
	}
	if task.Closure.CompiledTask.Template == nil {
		return taskInputs
	}
	if task.Closure.CompiledTask.Template.Interface == nil {
		return taskInputs
	}
	if task.Closure.CompiledTask.Template.Interface.Inputs == nil {
		return taskInputs
	}
	return task.Closure.CompiledTask.Template.Interface.Inputs.Variables
}

func ParamMapForTask(task *admin.Task) (map[string]interface{}, error) {
	taskInputs := TaskInputs(task)
	paramMap := make(map[string]interface{}, len(taskInputs))
	for k, v := range taskInputs {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Type)
			return nil, err
		}
		if paramMap[k], err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}
	}
	return paramMap, nil
}

func WorkflowParams(lp *admin.LaunchPlan) map[string]*core.Parameter {
	workflowParams := map[string]*core.Parameter{}
	if lp == nil || lp.Spec == nil {
		return workflowParams
	}
	if lp.Spec.DefaultInputs == nil {
		return workflowParams
	}
	return lp.Spec.DefaultInputs.Parameters
}

func ParamMapForWorkflow(lp *admin.LaunchPlan) (map[string]interface{}, error) {
	workflowParams := WorkflowParams(lp)
	paramMap := make(map[string]interface{}, len(workflowParams))
	for k, v := range workflowParams {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Var.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Var.Type)
			return nil, err
		}
		if paramMap[k], err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}
		// Override if there is a default value
		if paramsDefault, ok := v.Behavior.(*core.Parameter_Default); ok {
			if paramMap[k], err = coreutils.ExtractFromLiteral(paramsDefault.Default); err != nil {
				return nil, err
			}
		}
	}
	return paramMap, nil
}
