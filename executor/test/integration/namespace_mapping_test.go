/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

// TestEnqueueReconcile_MappedNamespace runs a project-domain in its own namespace:
// the root TaskAction is created in "<project>-<domain>", the executor creates its
// pod there, a child follows its parent, and the shared watcher still reports the
// action to the run service.
func TestEnqueueReconcile_MappedNamespace(t *testing.T) {
	// Enqueue through a client with the mapping; the suite's actionsClient keeps
	// watching every namespace through the shared cache.
	mappedClient, err := actionsk8s.NewActionsClient(k8sClient, sharedCache, "flyte", 32, 1, runClient, 1024,
		promutils.NewScope(fmt.Sprintf("integration_actions_mapped_%d", time.Now().UnixNano())),
		actionsk8s.WithNamespaceMapping("{{ project }}-{{ domain }}", true))
	require.NoError(t, err)

	runID := &common.RunIdentifier{
		Org:     "test-org",
		Project: "mapped-project",
		Domain:  "development",
		Name:    uniqueRunName("ns-run"),
	}
	const namespace = "mapped-project-development"
	require.NoError(t, mappedClient.Enqueue(ctx, newRootTaskAction(runID), nil))

	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Name: namespace}, &corev1.Namespace{}),
		"the run's namespace should be created")
	crKey := types.NamespacedName{Name: runID.Name + "-a0", Namespace: namespace}
	ta := &flyteorgv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, crKey, ta))

	// Two reconciles take the CR through the plugin, which creates the pod next to it.
	req := reconcile.Request{NamespacedName: crKey}
	for i := 0; i < 2; i++ {
		_, err := reconciler.Reconcile(ctx, req)
		require.NoError(t, err)
	}
	pod := &corev1.Pod{}
	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Name: runID.Name + "-a0-0", Namespace: namespace}, pod),
		"the task pod should be created in the run's namespace")
	require.NotEmpty(t, pod.OwnerReferences)
	assert.Equal(t, crKey.Name, pod.OwnerReferences[0].Name)

	// A child is created in its parent's namespace.
	child := newRootTaskAction(runID)
	child.ActionId = &common.ActionIdentifier{Run: runID, Name: "child1"}
	child.ParentActionName = &runID.Name
	require.NoError(t, mappedClient.Enqueue(ctx, child, nil))
	childCR := &flyteorgv1.TaskAction{}
	require.NoError(t, k8sClient.Get(ctx, types.NamespacedName{Name: runID.Name + "-child1", Namespace: namespace}, childCR))
	require.NotEmpty(t, childCR.OwnerReferences)
	assert.Equal(t, crKey.Name, childCR.OwnerReferences[0].Name)

	// The watcher looks the action up in its own namespace and records it.
	require.Eventually(t, func() bool {
		for _, r := range runClient.recordedActions() {
			if r.GetActionId().GetRun().GetName() == runID.Name && r.GetActionId().GetName() == runID.Name {
				return true
			}
		}
		return false
	}, 15*time.Second, 200*time.Millisecond, "RecordAction should reach the internal run service")
}
