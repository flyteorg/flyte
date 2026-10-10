package pytorch

import (
	"context"
	"testing"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func attemptMetadata() pluginsCore.TaskExecutionMetadata {
	meta := &mocks.TaskExecutionMetadata{}
	meta.EXPECT().GetLabels().Return(dummyLabels)
	return meta
}

// TestPytorchChildPodsMatchTheTemplatesTheyCameFrom is the conformance check between the two
// halves of child pod discovery: the labels this plugin puts on the replica pod templates
// the training operator expands, and the selector it hands the framework to find the
// resulting pods. Label drift on either side would otherwise leave a GPU fault on a worker
// silently unclassified.
func TestPytorchChildPodsMatchTheTemplatesTheyCameFrom(t *testing.T) {
	taskTemplate := dummyPytorchTaskTemplate("job3", dummyPytorchCustomObj(100))
	resource, err := pytorchOperatorResourceHandler{}.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, ""))
	require.NoError(t, err)

	job, ok := resource.(*kubeflowv1.PyTorchJob)
	require.True(t, ok)
	// The plugin manager names the resource after it is built, and the operator labels the
	// replica pods with that name.
	job.Name = "job-name"

	selector, err := pytorchOperatorResourceHandler{}.ChildPods(context.TODO(), attemptMetadata(), job)
	require.NoError(t, err)
	require.NotNil(t, selector)

	require.NotEmpty(t, job.Spec.PyTorchReplicaSpecs)
	for replicaType, replicaSpec := range job.Spec.PyTorchReplicaSpecs {
		t.Run(string(replicaType), func(t *testing.T) {
			templateLabels := replicaSpec.Template.GetLabels()
			require.NotEmpty(t, templateLabels)

			podLabels := make(map[string]string, len(templateLabels)+1)
			for k, v := range templateLabels {
				podLabels[k] = v
			}
			podLabels[kubeflowv1.JobNameLabel] = job.Name

			assert.True(t, selector.Matches(labels.Set(podLabels)),
				"the %s replica template's labels %v do not satisfy %s", replicaType, podLabels, selector)
		})
	}
}
