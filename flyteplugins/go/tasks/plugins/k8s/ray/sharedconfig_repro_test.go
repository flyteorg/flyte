package ray

import (
	"context"
	"maps"
	"testing"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// A RayJob that leaves RayStartParams unset — the common case, and the one that falls through to
// the plugin's configured defaults.
func rayJobWithoutStartParams() *plugins.RayJob {
	return &plugins.RayJob{
		RayCluster: &plugins.RayCluster{
			HeadGroupSpec:   &plugins.HeadGroupSpec{},
			WorkerGroupSpec: []*plugins.WorkerGroupSpec{{GroupName: workerGroupName, Replicas: 1, MinReplicas: 1, MaxReplicas: 1}},
		},
		ShutdownAfterJobFinishes: true,
		TtlSecondsAfterFinished:  120,
	}
}

// Other tests in this package replace the global config without restoring it, which can leave the
// default start parameters empty and skip the branch under test. Pin a config with non-empty
// defaults, and put the original back afterwards.
func withDefaultStartParams(t *testing.T) {
	origConfig := *GetConfig()
	t.Cleanup(func() { assert.NoError(t, SetConfig(&origConfig)) })

	assert.NoError(t, SetConfig(&Config{
		Defaults: DefaultConfig{
			HeadNode: NodeConfig{
				StartParameters: map[string]string{DisableUsageStatsStartParameter: "true"},
				IPAddress:       "$MY_POD_IP",
			},
			WorkerNode: NodeConfig{
				StartParameters: map[string]string{DisableUsageStatsStartParameter: "true"},
				IPAddress:       "$MY_POD_IP",
			},
		},
	}))
}

// BuildResource must not write into the plugin's shared configuration.
//
// When a task leaves RayStartParams unset, the head/worker start params fall back to
// cfg.Defaults.{Head,Worker}Node.StartParameters. Those maps belong to the process-wide plugin
// config, so filling in include-dashboard / node-ip-address / dashboard-host mutates config that
// every later task reads.
func TestBuildResource_DoesNotMutateSharedConfig(t *testing.T) {
	withDefaultStartParams(t)

	cfg := GetConfig()
	headDefaults := cfg.Defaults.HeadNode.StartParameters
	workerDefaults := cfg.Defaults.WorkerNode.StartParameters

	headBefore := maps.Clone(headDefaults)
	workerBefore := maps.Clone(workerDefaults)

	handler := rayJobResourceHandler{}
	taskTemplate := dummyRayTaskTemplate("ray-id", rayJobWithoutStartParams())
	taskCtx := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)

	_, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.NoError(t, err)

	assert.Equal(t, headBefore, cfg.Defaults.HeadNode.StartParameters,
		"BuildResource mutated the shared head-node start parameters")
	assert.Equal(t, workerBefore, cfg.Defaults.WorkerNode.StartParameters,
		"BuildResource mutated the shared worker-node start parameters")
}

// Two RayJobs built independently must not share one RayStartParams map.
//
// constructRayJob assigns the resolved start params straight into the CR
// (RayStartParams: headNodeRayStartParams). When they came from the shared defaults, every CR
// built that way — and the plugin config itself — is the same map instance.
func TestBuildResource_RayJobsDoNotShareStartParamsMap(t *testing.T) {
	withDefaultStartParams(t)

	handler := rayJobResourceHandler{}

	build := func() *rayv1.RayJob {
		taskTemplate := dummyRayTaskTemplate("ray-id", rayJobWithoutStartParams())
		taskCtx := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)
		obj, err := handler.BuildResource(context.TODO(), taskCtx)
		assert.NoError(t, err)
		return obj.(*rayv1.RayJob)
	}

	first := build()
	second := build()

	first.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams["num-cpus"] = "tainted-by-first-rayjob"

	assert.NotContains(t, second.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams, "num-cpus",
		"editing one RayJob's start params changed another RayJob's")
	assert.NotContains(t, GetConfig().Defaults.HeadNode.StartParameters, "num-cpus",
		"editing a RayJob's start params changed the shared plugin config")
}

// Start params a task supplies for itself reach the CR alongside the defaults the plugin injects.
// Covers the head and worker branches that copy the task's own map rather than the config's.
func TestBuildResource_KeepsTaskSuppliedStartParams(t *testing.T) {
	withDefaultStartParams(t)

	rayJob := rayJobWithoutStartParams()
	rayJob.RayCluster.HeadGroupSpec.RayStartParams = map[string]string{"num-cpus": "1"}
	rayJob.RayCluster.WorkerGroupSpec[0].RayStartParams = map[string]string{"num-cpus": "2"}

	handler := rayJobResourceHandler{}
	taskTemplate := dummyRayTaskTemplate("ray-id", rayJob)
	taskCtx := dummyRayTaskContext(taskTemplate, resourceRequirements, nil, "", serviceAccount)

	obj, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.NoError(t, err)
	built := obj.(*rayv1.RayJob)

	head := built.Spec.RayClusterSpec.HeadGroupSpec.RayStartParams
	assert.Equal(t, "1", head["num-cpus"])
	assert.Contains(t, head, NodeIPAddress)

	worker := built.Spec.RayClusterSpec.WorkerGroupSpecs[0].RayStartParams
	assert.Equal(t, "2", worker["num-cpus"])
	assert.Contains(t, worker, NodeIPAddress)
}
