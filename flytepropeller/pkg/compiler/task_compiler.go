package compiler

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
)

func validateResource(resourceName core.Resources_ResourceName, resourceVal string, errs errors.CompileErrors) (ok bool) {
	if _, err := resource.ParseQuantity(resourceVal); err != nil {
		errs.Collect(errors.NewUnrecognizedValueErr(fmt.Sprintf("resources.%v", resourceName), resourceVal))
		return true
	}
	return false
}

func validateKnownResources(resources []*core.Resources_ResourceEntry, errs errors.CompileErrors) {
	for _, r := range resources {
		validateResource(r.GetName(), r.GetValue(), errs.NewScope())
	}
}

func validateResources(resources *core.Resources, errs errors.CompileErrors) (ok bool) {
	// Validate known resource keys.
	validateKnownResources(resources.GetRequests(), errs.NewScope())
	validateKnownResources(resources.GetLimits(), errs.NewScope())

	return !errs.HasErrors()
}

func validateContainerCommand(task *core.TaskTemplate, errs errors.CompileErrors) (ok bool) {
	if task.GetInterface() == nil {
		// Nothing to validate.
		return
	}
	hasInputs := task.GetInterface().GetInputs() != nil && len(task.GetInterface().GetInputs().GetVariables()) > 0
	hasOutputs := task.GetInterface().GetOutputs() != nil && len(task.GetInterface().GetOutputs().GetVariables()) > 0
	if !(hasInputs || hasOutputs) {
		// Nothing to validate.
		return
	}
	if task.GetContainer().Command == nil && task.GetContainer().Args == nil {
		// When an interface with inputs or outputs is defined, the container command + args together must not be empty.
		errs.Collect(errors.NewValueRequiredErr("container", "command"))
	}

	return !errs.HasErrors()
}

func validateContainer(task *core.TaskTemplate, errs errors.CompileErrors) (ok bool) {
	if task.GetContainer() == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "container"))
		return
	}

	validateContainerCommand(task, errs)

	container := task.GetContainer()
	if container.GetImage() == "" {
		errs.Collect(errors.NewValueRequiredErr("container", "image"))
	}

	if container.GetResources() != nil {
		validateResources(container.GetResources(), errs.NewScope())
	}

	return !errs.HasErrors()
}

func validateK8sPod(task *core.TaskTemplate, errs errors.CompileErrors) (ok bool) {
	if task.GetK8SPod() == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "k8s pod"))
		return
	}
	var podSpec v1.PodSpec
	if err := utils.UnmarshalStructToObj(task.GetK8SPod().GetPodSpec(), &podSpec); err != nil {
		errs.Collect(errors.NewInvalidValueErr("root", "k8s pod spec"))
		return
	}
	for _, container := range podSpec.Containers {
		if containerErrs := validation.IsDNS1123Label(container.Name); len(containerErrs) > 0 {
			errs.Collect(errors.NewInvalidValueErr("root", "k8s pod spec container name"))
		}
	}
	return !errs.HasErrors()
}

func compileTaskInternal(task *core.TaskTemplate, errs errors.CompileErrors) common.Task {
	if task.GetId() == nil {
		errs.Collect(errors.NewValueRequiredErr("root", "Id"))
	}

	switch task.GetTarget().(type) {
	case *core.TaskTemplate_Container:
		validateContainer(task, errs.NewScope())
	case *core.TaskTemplate_K8SPod:
		validateK8sPod(task, errs.NewScope())
	}

	return taskBuilder{flyteTask: task}
}

// CompileTask compiles a given Task into an executable Task. It validates all required parameters and ensures a Task
// is well-formed.
func CompileTask(task *core.TaskTemplate) (*core.CompiledTask, error) {
	errs := errors.NewCompileErrors()
	t := compileTaskInternal(task, errs.NewScope())
	if errs.HasErrors() {
		return nil, errs
	}

	return &core.CompiledTask{Template: t.GetCoreTask()}, nil
}
