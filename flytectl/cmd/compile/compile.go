package compile

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	config "github.com/flyteorg/flyte/flytectl/cmd/config/subcommand/compile"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/cmd/register"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
)

// Utility function for compiling a list of Tasks
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

/*
Utility to compile a packaged workflow locally.
compilation is done locally so no flyte cluster is required.
*/
func compileFromPackage(packagePath string) error {
	args := []string{packagePath}
	fileList, tmpDir, err := register.GetSerializeOutputFiles(context.Background(), args, true)
	defer os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Println("Error found while extracting package..")
		return err
	}
	fmt.Println("Successfully extracted package...")
	fmt.Println("Processing Protobuf files...")
	workflows := make(map[string]*admin.WorkflowSpec)
	plans := make(map[string]*admin.LaunchPlan)
	tasks := []*admin.TaskSpec{}

	for _, pbFilePath := range fileList {
		rawTsk, err := ioutil.ReadFile(pbFilePath)
		if err != nil {
			fmt.Printf("error unmarshalling task..")
			return err
		}
		spec, err := register.UnMarshalContents(context.Background(), rawTsk, pbFilePath)
		if err != nil {
			return err
		}

		switch v := spec.(type) {
		case *admin.TaskSpec:
			tasks = append(tasks, v)
		case *admin.WorkflowSpec:
			workflows[v.Template.Id.Name] = v
		case *admin.LaunchPlan:
			plans[v.Id.Name] = v
		}
	}

	// compile tasks
	taskTemplates := []*core.TaskTemplate{}
	for _, task := range tasks {
		taskTemplates = append(taskTemplates, task.Template)
	}

	fmt.Println("\nCompiling tasks...")
	compiledTasks, err := compileTasks(taskTemplates)
	if err != nil {
		fmt.Println("Error while compiling tasks...")
		return err
	}

	var providers []common.InterfaceProvider
	var compiledWorkflows = map[string]*core.CompiledWorkflowClosure{}

	// compile workflows
	for _, workflow := range workflows {
		providers, err = handleWorkflow(workflow, compiledTasks, compiledWorkflows, providers, plans, workflows)

		if err != nil {
			return err
		}
	}

	fmt.Println("All Workflows compiled successfully!")
	fmt.Println("\nSummary:")
	fmt.Println(len(workflows), " workflows found in package")
	fmt.Println(len(tasks), " Tasks found in package")
	fmt.Println(len(plans), " Launch plans found in package")
	return nil
}

func handleWorkflow(
	workflow *admin.WorkflowSpec,
	compiledTasks []*core.CompiledTask,
	compiledWorkflows map[string]*core.CompiledWorkflowClosure,
	compiledLaunchPlanProviders []common.InterfaceProvider,
	plans map[string]*admin.LaunchPlan,
	workflows map[string]*admin.WorkflowSpec) ([]common.InterfaceProvider, error) {
	reqs, _ := compiler.GetRequirements(workflow.Template, workflow.SubWorkflows)
	wfName := workflow.Template.Id.Name

	// Check if all the subworkflows referenced by launchplan are compiled
	for i := range reqs.GetRequiredLaunchPlanIds() {
		lpID := &reqs.GetRequiredLaunchPlanIds()[i]
		lpWfName := plans[lpID.Name].Spec.WorkflowId.Name
		missingWorkflow := workflows[lpWfName]
		if compiledWorkflows[lpWfName] == nil {
			// Recursively compile the missing workflow first
			err := error(nil)
			compiledLaunchPlanProviders, err = handleWorkflow(missingWorkflow, compiledTasks, compiledWorkflows, compiledLaunchPlanProviders, plans, workflows)
			if err != nil {
				return nil, err
			}
		}
	}

	fmt.Println("\nCompiling workflow:", wfName)

	wf, err := compiler.CompileWorkflow(workflow.Template,
		workflow.SubWorkflows,
		compiledTasks,
		compiledLaunchPlanProviders)

	if err != nil {
		fmt.Println(":( Error Compiling workflow:", wfName)
		return nil, err
	}
	compiledWorkflows[wfName] = wf

	// Update the expected inputs and outputs for the launchplans which reference this workflow
	for _, plan := range plans {
		if plan.Spec.WorkflowId.Name == wfName {
			plan.Closure.ExpectedOutputs = wf.Primary.Template.Interface.Outputs
			newMap := make(map[string]*core.Parameter)

			for key, value := range wf.Primary.Template.Interface.Inputs.Variables {
				newMap[key] = &core.Parameter{
					Var: value,
				}
			}
			plan.Closure.ExpectedInputs = &core.ParameterMap{
				Parameters: newMap,
			}
			compiledLaunchPlanProviders = append(compiledLaunchPlanProviders, compiler.NewLaunchPlanInterfaceProvider(*plan))
		}
	}

	return compiledLaunchPlanProviders, nil
}

const (
	compileShort = `Validate flyte packages without registration needed.`
	compileLong  = `
Validate workflows by compiling flyte's serialized protobuf files  (task, workflows and launch plans). This is useful for testing workflows and tasks without neededing to talk with a flyte cluster.

::

 flytectl compile --file my-flyte-package.tgz

::

 flytectl compile --file /home/user/dags/my-flyte-package.tgz

.. note::
   Input file is a path to a tgz. This file is generated by either pyflyte or jflyte. tgz file contains protobuf files describing workflows, tasks and launch plans.

`
)

func compile(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {
	packageFilePath := config.DefaultCompileConfig.File
	if packageFilePath == "" {
		return fmt.Errorf("path to package tgz's file is a required flag")
	}
	return compileFromPackage(packageFilePath)
}

func CreateCompileCommand() map[string]cmdCore.CommandEntry {
	compileResourcesFuncs := map[string]cmdCore.CommandEntry{
		"compile": {
			Short:                    compileShort,
			Long:                     compileLong,
			CmdFunc:                  compile,
			PFlagProvider:            config.DefaultCompileConfig,
			ProjectDomainNotRequired: true,
			DisableFlyteClient:       true,
		},
	}
	return compileResourcesFuncs
}
