package tensorflow

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

// TestTensorFlowChildPodsMatchTheTemplatesTheyCameFrom is the conformance check between the two
// halves of child pod discovery: the labels this plugin puts on the replica pod templates
// the training operator expands, and the selector it hands the framework to find the
// resulting pods. Label drift on either side would otherwise leave a GPU fault on a worker
// silently unclassified.
func TestTensorFlowChildPodsMatchTheTemplatesTheyCameFrom(t *testing.T) {
	taskTemplate := dummyTensorFlowTaskTemplate("job3", dummyTensorFlowCustomObj(2, 1, 1, 1))
	resource, err := tensorflowOperatorResourceHandler{}.BuildResource(context.TODO(), dummyTensorFlowTaskContext(taskTemplate, resourceRequirements, nil))
	require.NoError(t, err)

	job, ok := resource.(*kubeflowv1.TFJob)
	require.True(t, ok)
	// The plugin manager names the resource after it is built, and the operator labels the
	// replica pods with that name.
	job.Name = "job-name"

	selector, err := tensorflowOperatorResourceHandler{}.ChildPods(context.TODO(), attemptMetadata(), job)
	require.NoError(t, err)
	require.NotNil(t, selector)

	require.NotEmpty(t, job.Spec.TFReplicaSpecs)
	for replicaType, replicaSpec := range job.Spec.TFReplicaSpecs {
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
