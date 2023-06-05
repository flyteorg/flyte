package get

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// ExecutionConfig is duplicated struct from create with the same structure. This is to avoid the circular dependency. Only works with go-yaml.
// TODO : replace this with a cleaner design
type ExecutionConfig struct {
	IamRoleARN      string               `yaml:"iamRoleARN"`
	Inputs          map[string]yaml.Node `yaml:"inputs"`
	Envs            map[string]string    `yaml:"envs"`
	KubeServiceAcct string               `yaml:"kubeServiceAcct"`
	TargetDomain    string               `yaml:"targetDomain"`
	TargetProject   string               `yaml:"targetProject"`
	Task            string               `yaml:"task,omitempty"`
	Version         string               `yaml:"version"`
	Workflow        string               `yaml:"workflow,omitempty"`
}

func WriteExecConfigToFile(executionConfig ExecutionConfig, fileName string) error {
	d, err := yaml.Marshal(executionConfig)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	if _, err = os.Stat(fileName); err == nil {
		if !cmdUtil.AskForConfirmation(fmt.Sprintf("warning file %v will be overwritten", fileName), os.Stdin) {
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

func ParamMapForTask(task *admin.Task) (map[string]yaml.Node, error) {
	taskInputs := TaskInputs(task)
	paramMap := make(map[string]yaml.Node, len(taskInputs))
	for k, v := range taskInputs {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Type)
			return nil, err
		}
		var nativeLiteral interface{}
		if nativeLiteral, err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}

		if k == v.Description {
			// a: # a isn't very helpful
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, "")
		} else {
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, v.Description)
		}
		if err != nil {
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

func ParamMapForWorkflow(lp *admin.LaunchPlan) (map[string]yaml.Node, error) {
	workflowParams := WorkflowParams(lp)
	paramMap := make(map[string]yaml.Node, len(workflowParams))
	for k, v := range workflowParams {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.Var.Type)
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.Var.Type)
			return nil, err
		}
		var nativeLiteral interface{}
		if nativeLiteral, err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}
		// Override if there is a default value
		if paramsDefault, ok := v.Behavior.(*core.Parameter_Default); ok {
			if nativeLiteral, err = coreutils.ExtractFromLiteral(paramsDefault.Default); err != nil {
				return nil, err
			}
		}
		if k == v.Var.Description {
			// a: # a isn't very helpful
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, "")
		} else {
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, v.Var.Description)
		}

		if err != nil {
			return nil, err
		}
	}
	return paramMap, nil
}

func getCommentedYamlNode(input interface{}, comment string) (yaml.Node, error) {
	var node yaml.Node
	err := node.Encode(input)
	node.LineComment = comment
	return node, err
}
