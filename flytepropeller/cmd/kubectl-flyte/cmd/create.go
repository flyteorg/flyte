package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	compilerErrors "github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	protofileKey   = "proto-path"
	formatKey      = "format"
	executionIDKey = "execution-id"
	inputsKey      = "input-path"
	annotationsKey = "annotations"
)

type format = string

const (
	formatProto format = "proto"
	formatJSON  format = "json"
	formatYaml  format = "yaml"
)

const createCmdName = "create"

type CreateOpts struct {
	*RootOptions
	format      format
	execID      string
	inputsPath  string
	protoFile   string
	annotations *stringMapValue
	dryRun      bool
}

func NewCreateCommand(opts *RootOptions) *cobra.Command {

	createOpts := &CreateOpts{
		RootOptions: opts,
	}

	createCmd := &cobra.Command{
		Use:     createCmdName,
		Aliases: []string{"new", "compile"},
		Short:   "Creates a new workflow from proto-buffer files.",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requiredFlags(cmd, protofileKey, formatKey); err != nil {
				return err
			}

			fmt.Println("Line numbers in errors enabled")
			compilerErrors.SetIncludeSource()

			return createOpts.createWorkflowFromProto()
		},
	}

	createCmd.Flags().StringVarP(&createOpts.protoFile, protofileKey, "p", "", "Path of the workflow package proto-buffer file to be uploaded")
	createCmd.Flags().StringVarP(&createOpts.format, formatKey, "f", formatProto, "Format of the provided file. Supported formats: proto (default), json, yaml")
	createCmd.Flags().StringVarP(&createOpts.execID, executionIDKey, "", "", "Execution Id of the Workflow to create.")
	createCmd.Flags().StringVarP(&createOpts.inputsPath, inputsKey, "i", "", "Path to inputs file.")
	createOpts.annotations = newStringMapValue()
	createCmd.Flags().VarP(createOpts.annotations, annotationsKey, "a", "Defines extra annotations to declare on the created object.")
	createCmd.Flags().BoolVarP(&createOpts.dryRun, "dry-run", "d", false, "Compiles and transforms, but does not create a workflow. OutputsRef ts to STDOUT.")

	return createCmd
}

func unmarshal(in []byte, format format, message proto.Message) (err error) {
	switch format {
	case formatProto:
		err = proto.Unmarshal(in, message)
	case formatJSON:
		err = jsonpb.Unmarshal(bytes.NewReader(in), message)
		if err != nil {
			err = errors.Wrapf(err, "Failed to unmarshal converted Json. [%v]", string(in))
		}
	case formatYaml:
		jsonRaw, err := yaml.YAMLToJSON(in)
		if err != nil {
			return errors.Wrapf(err, "Failed to convert yaml to JSON. [%v]", string(in))
		}

		return unmarshal(jsonRaw, formatJSON, message)
	}

	return
}

var jsonPbMarshaler = jsonpb.Marshaler{}

func marshal(message proto.Message, format format) (raw []byte, err error) {
	switch format {
	case formatProto:
		return proto.Marshal(message)
	case formatJSON:
		b := &bytes.Buffer{}
		err := jsonPbMarshaler.Marshal(b, message)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal Json.")
		}
		return b.Bytes(), nil
	case formatYaml:
		b, err := marshal(message, formatJSON)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to marshal JSON")
		}
		return yaml.JSONToYAML(b)
	}
	return nil, errors.Errorf("Unknown format type")
}

func loadInputs(path string, format format) (c *core.LiteralMap, err error) {
	// Support reading from s3, etc.?
	var raw []byte
	raw, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	c = &core.LiteralMap{}
	err = unmarshal(raw, format, c)
	return
}

func compileTasks(tasks []*core.TaskTemplate) ([]*core.CompiledTask, error) {
	res := make([]*core.CompiledTask, 0, len(tasks))
	for _, task := range tasks {
		compiledTask, err := compiler.CompileTask(task)
		if err != nil {
			return nil, err
		}

		res = append(res, compiledTask)
	}

	return res, nil
}

func (c *CreateOpts) createWorkflowFromProto() error {
	fmt.Printf("Received protofiles : [%v] [%v].\n", c.protoFile, c.inputsPath)
	rawWf, err := ioutil.ReadFile(c.protoFile)
	if err != nil {
		return err
	}

	wfClosure := core.WorkflowClosure{}
	err = unmarshal(rawWf, c.format, &wfClosure)
	if err != nil {
		return err
	}

	compiledTasks, err := compileTasks(wfClosure.Tasks)
	if err != nil {
		return err
	}

	wf, err := compiler.CompileWorkflow(wfClosure.Workflow, []*core.WorkflowTemplate{}, compiledTasks, []common.InterfaceProvider{})
	if err != nil {
		return err
	}

	var inputs *core.LiteralMap
	if c.inputsPath != "" {
		inputs, err = loadInputs(c.inputsPath, c.format)
		if err != nil {
			return errors.Wrapf(err, "Failed to load inputs.")
		}
	}

	var executionID *core.WorkflowExecutionIdentifier
	if len(c.execID) > 0 {
		executionID = &core.WorkflowExecutionIdentifier{
			Name:    c.execID,
			Domain:  wfClosure.Workflow.Id.Domain,
			Project: wfClosure.Workflow.Id.Project,
		}
	}

	flyteWf, err := k8s.BuildFlyteWorkflow(wf, inputs, executionID, c.ConfigOverrides.Context.Namespace)
	if err != nil {
		return err
	}
	flyteWf.ExecutionID = v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: executionID,
	}
	if flyteWf.Annotations == nil {
		flyteWf.Annotations = *c.annotations.value
	} else {
		for key, val := range *c.annotations.value {
			flyteWf.Annotations[key] = val
		}
	}

	if c.dryRun {
		fmt.Printf("Dry Run mode enabled. Printing the compiled workflow.\n")
		j, err := json.Marshal(flyteWf)
		if err != nil {
			return errors.Wrapf(err, "Failed to marshal final workflow to Propeller format.")
		}
		y, err := yaml.JSONToYAML(j)
		if err != nil {
			return errors.Wrapf(err, "Failed to marshal final workflow from json to yaml.")
		}
		fmt.Println(string(y))
	} else {
		wf, err := c.flyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(c.ConfigOverrides.Context.Namespace).Create(context.TODO(), flyteWf, v1.CreateOptions{})
		if err != nil {
			return err
		}

		fmt.Printf("Successfully created Flyte Workflow %v.\n", wf.Name)
	}
	return nil
}
