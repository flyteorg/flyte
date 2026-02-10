package get

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	cmdUtil "github.com/flyteorg/flyte/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.yaml.in/yaml/v3"
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
	executionConfig := ExecutionConfig{Task: task.GetId().GetName(), Version: task.GetId().GetVersion()}
	if executionConfig.Inputs, err = ParamMapForTask(task); err != nil {
		return err
	}
	return WriteExecConfigToFile(executionConfig, fileName)
}

func CreateAndWriteExecConfigForWorkflow(wlp *admin.LaunchPlan, fileName string) error {
	var err error
	executionConfig := ExecutionConfig{Workflow: wlp.GetId().GetName(), Version: wlp.GetId().GetVersion()}
	if executionConfig.Inputs, err = ParamMapForWorkflow(wlp); err != nil {
		return err
	}
	return WriteExecConfigToFile(executionConfig, fileName)
}

func TaskInputs(task *admin.Task) map[string]*core.Variable {
	taskInputs := map[string]*core.Variable{}
	if task == nil || task.GetClosure() == nil {
		return taskInputs
	}
	if task.GetClosure().GetCompiledTask() == nil {
		return taskInputs
	}
	if task.GetClosure().GetCompiledTask().GetTemplate() == nil {
		return taskInputs
	}
	if task.GetClosure().GetCompiledTask().GetTemplate().GetInterface() == nil {
		return taskInputs
	}
	if task.GetClosure().GetCompiledTask().GetTemplate().GetInterface().GetInputs() == nil {
		return taskInputs
	}
	return task.GetClosure().GetCompiledTask().GetTemplate().GetInterface().GetInputs().GetVariables()
}

func ParamMapForTask(task *admin.Task) (map[string]yaml.Node, error) {
	taskInputs := TaskInputs(task)
	paramMap := make(map[string]yaml.Node, len(taskInputs))
	for k, v := range taskInputs {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.GetType())
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.GetType())
			return nil, err
		}
		var nativeLiteral interface{}
		if nativeLiteral, err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}

		if k == v.GetDescription() {
			// a: # a isn't very helpful
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, "")
		} else {
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, v.GetDescription())
		}
		if err != nil {
			return nil, err
		}
	}
	return paramMap, nil
}

func WorkflowParams(lp *admin.LaunchPlan) map[string]*core.Parameter {
	workflowParams := map[string]*core.Parameter{}
	if lp == nil || lp.GetSpec() == nil {
		return workflowParams
	}
	if lp.GetSpec().GetDefaultInputs() == nil {
		return workflowParams
	}
	return lp.GetSpec().GetDefaultInputs().GetParameters()
}

func ParamMapForWorkflow(lp *admin.LaunchPlan) (map[string]yaml.Node, error) {
	workflowParams := WorkflowParams(lp)
	paramMap := make(map[string]yaml.Node, len(workflowParams))
	for k, v := range workflowParams {
		varTypeValue, err := coreutils.MakeDefaultLiteralForType(v.GetVar().GetType())
		if err != nil {
			fmt.Println("error creating default value for literal type ", v.GetVar().GetType())
			return nil, err
		}
		var nativeLiteral interface{}
		if nativeLiteral, err = coreutils.ExtractFromLiteral(varTypeValue); err != nil {
			return nil, err
		}
		// Override if there is a default value
		if paramsDefault, ok := v.GetBehavior().(*core.Parameter_Default); ok {
			if nativeLiteral, err = coreutils.ExtractFromLiteral(paramsDefault.Default); err != nil {
				return nil, err
			}
		}
		if k == v.GetVar().GetDescription() {
			// a: # a isn't very helpful
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, "")
		} else {
			paramMap[k], err = getCommentedYamlNode(nativeLiteral, v.GetVar().GetDescription())
		}

		if err != nil {
			return nil, err
		}
	}
	return paramMap, nil
}

func getCommentedYamlNode(input interface{}, comment string) (yaml.Node, error) {
	var node yaml.Node
	var err error

	if s, ok := input.(*structpb.Struct); ok {
		err = node.Encode(s.AsMap())
	} else {
		err = node.Encode(input)
	}

	node.LineComment = comment
	return node, err
}
