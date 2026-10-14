package mpi

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

// TestMPIChildPodsMatchTheTemplatesTheyCameFrom is the conformance check between the two
// halves of child pod discovery: the labels this plugin puts on the replica pod templates
// the training operator expands, and the selector it hands the framework to find the
// resulting pods. Label drift on either side would otherwise leave a GPU fault on a worker
// silently unclassified.
func TestMPIChildPodsMatchTheTemplatesTheyCameFrom(t *testing.T) {
	taskTemplate := dummyMPITaskTemplate("job3", dummyMPICustomObj(2, 1, 1))
	resource, err := mpiOperatorResourceHandler{}.BuildResource(context.TODO(), dummyMPITaskContext(taskTemplate, resourceRequirements, nil))
	require.NoError(t, err)

	job, ok := resource.(*kubeflowv1.MPIJob)
	require.True(t, ok)
	// The plugin manager names the resource after it is built, and the operator labels the
	// replica pods with that name.
	job.Name = "job-name"

	selector, err := mpiOperatorResourceHandler{}.ChildPods(context.TODO(), attemptMetadata(), job)
	require.NoError(t, err)
	require.NotNil(t, selector)

	require.NotEmpty(t, job.Spec.MPIReplicaSpecs)
	for replicaType, replicaSpec := range job.Spec.MPIReplicaSpecs {
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
