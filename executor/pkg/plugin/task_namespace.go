package plugin

import (
	"fmt"
	"slices"
	"strings"

	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/validation"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorconfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// TaskNamespace resolves the namespace declared by a task, falling back to the
// TaskAction namespace. It never creates a namespace or derives one from identity.
func TaskNamespace(ta *executorv1.TaskAction) (string, error) {
	if len(ta.Spec.TaskTemplate) == 0 {
		return ta.Namespace, nil
	}
	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(ta.Spec.TaskTemplate, tmpl); err != nil {
		return "", fmt.Errorf("failed to read task namespace: %w", err)
	}
	namespace := tmpl.GetConfig()[flytek8s.TaskNamespaceConfigKey]
	if namespace == "" {
		return ta.Namespace, nil
	}
	if errs := validation.IsDNS1123Label(namespace); len(errs) > 0 {
		return "", fmt.Errorf("invalid task namespace %q: %s", namespace, strings.Join(errs, "; "))
	}

	return namespace, nil
}

// ValidateTaskNamespace checks placement before the first launch. Already accepted
// actions must remain readable and cleanable after the deployment allowlist changes.
func ValidateTaskNamespace(ta *executorv1.TaskAction) error {
	namespace, err := TaskNamespace(ta)
	if err != nil {
		return err
	}
	allowed := executorconfig.GetConfig().AllowedTaskNamespaces
	if namespace != ta.Namespace && !slices.Contains(allowed, namespace) && !slices.Contains(allowed, "*") {
		return fmt.Errorf("task namespace %q is not in executor.allowedTaskNamespaces", namespace)
	}
	return nil
}
