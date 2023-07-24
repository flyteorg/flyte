package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	compilerErrors "github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type CompileOpts struct {
	*RootOptions
	inputFormat     format
	outputFormat    format
	protoFile       string
	outputPath      string
	dumpClosureYaml bool
}

func NewCompileCommand(opts *RootOptions) *cobra.Command {

	compileOpts := &CompileOpts{
		RootOptions: opts,
	}
	compileCmd := &cobra.Command{
		Use:     "compile",
		Aliases: []string{"new", "compile"},
		Short:   "Compile a workflow from core proto-buffer files and output a closure.",
		Long:    ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requiredFlags(cmd, protofileKey, formatKey); err != nil {
				return err
			}

			fmt.Println("Line numbers in errors enabled")
			compilerErrors.SetIncludeSource()

			return compileOpts.compileWorkflowCmd()
		},
	}

	compileCmd.Flags().StringVarP(&compileOpts.protoFile, "input-file", "i", "", "Path of the workflow package proto-buffer file to be uploaded")
	compileCmd.Flags().StringVarP(&compileOpts.inputFormat, "input-format", "f", formatProto, "Format of the provided file. Supported formats: proto (default), json, yaml")
	compileCmd.Flags().StringVarP(&compileOpts.outputPath, "output-file", "o", "", "Path of the generated output file.")
	compileCmd.Flags().StringVarP(&compileOpts.outputFormat, "output-format", "m", formatProto, "Format of the generated file. Supported formats: proto (default), json, yaml")
	compileCmd.Flags().BoolVarP(&compileOpts.dumpClosureYaml, "dump-closure-yaml", "d", false, "Compiles and transforms, but does not create a workflow. OutputsRef ts to STDOUT.")

	return compileCmd
}

func (c *CompileOpts) compileWorkflowCmd() error {
	if c.protoFile == "" {
		return errors.Errorf("Input file not specified")
	}
	fmt.Printf("Received protofiles : [%v].\n", c.protoFile)

	rawWf, err := ioutil.ReadFile(c.protoFile)
	if err != nil {
		return err
	}

	wfClosure := core.WorkflowClosure{}
	err = unmarshal(rawWf, c.inputFormat, &wfClosure)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal input Workflow")
	}

	if c.dumpClosureYaml {
		b, err := marshal(&wfClosure, formatYaml)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(c.protoFile+".yaml", b, os.ModePerm)
		if err != nil {
			return err
		}
	}

	compiledTasks, err := compileTasks(wfClosure.Tasks)
	if err != nil {
		return err
	}

	compileWfClosure, err := compiler.CompileWorkflow(wfClosure.Workflow, []*core.WorkflowTemplate{}, compiledTasks, []common.InterfaceProvider{})
	if err != nil {
		return err
	}

	fmt.Printf("Workflow compiled successfully, creating output location: [%v] format [%v]\n", c.outputPath, c.outputFormat)

	o, err := marshal(compileWfClosure, c.outputFormat)
	if err != nil {
		return errors.Wrapf(err, "Failed to marshal final workflow.")
	}

	if c.outputPath != "" {
		return ioutil.WriteFile(c.outputPath, o, os.ModePerm)
	}
	fmt.Printf("%v", string(o))
	return nil
}
