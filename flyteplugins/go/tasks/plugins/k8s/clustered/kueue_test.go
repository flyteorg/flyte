package clustered

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"

	flyteerr "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/gang"
	plugink8s "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	stdutils "github.com/flyteorg/flyte/v2/flytestdlib/utils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const (
	platformQueue    = "platform-queue"
	teamQueue        = "team-a"
	execLabelsSource = "the execution labels"
)

// withConfig applies edit to a copy of the default plugin config for the duration of the test.
func withConfig(t *testing.T, edit func(*Config)) {
	t.Helper()
	prev := *GetConfig()
	cfg := defaultConfig
	edit(&cfg)
	require.NoError(t, SetConfig(&cfg))
	t.Cleanup(func() { require.NoError(t, SetConfig(&prev)) })
}

func kueueTestSpec() *clusteredpb.ClusteredTaskSpec {
	return &clusteredpb.ClusteredTaskSpec{
		Replicas:     2,
		NprocPerNode: 1,
		Runtime: &clusteredpb.Runtime{
			Kind: &clusteredpb.Runtime_Torchrun{Torchrun: &clusteredpb.TorchRuntime{}},
		},
	}
}

// podTemplateWithQueueLabel is a pod template whose user labels try to pick a Kueue queue.
func podTemplateWithQueueLabel(t *testing.T, queue string) *core.K8SPod {
	podSpec, err := stdutils.MarshalObjToStruct(corev1.PodSpec{
		Containers: []corev1.Container{{Name: primaryContainerName, Image: testImage}},
	})
	require.NoError(t, err)
	return &core.K8SPod{
		Metadata:             &core.K8SObjectMetadata{Labels: map[string]string{kueueQueueNameLabel: queue}},
		PodSpec:              podSpec,
		PrimaryContainerName: primaryContainerName,
	}
}

func buildJobSet(t *testing.T, podTemplate *core.K8SPod) *jobsetv1alpha2.JobSet {
	t.Helper()
	obj, err := clusteredResourceHandler{}.BuildResource(context.Background(),
		dummyTaskCtx(buildTaskTemplate(kueueTestSpec()), podTemplate))
	require.NoError(t, err)
	jobSet, ok := obj.(*jobsetv1alpha2.JobSet)
	require.True(t, ok)
	return jobSet
}

func primaryEnv(jobSet *jobsetv1alpha2.JobSet) map[string]string {
	env := map[string]string{}
	for _, c := range jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Spec.Containers {
		if c.Name != jobSet.Annotations[primaryContainerAnnotation] {
			continue
		}
		for _, e := range c.Env {
			env[e.Name] = e.Value
		}
	}
	return env
}

func TestBuildResource_KueueDisabled_Unchanged(t *testing.T) {
	baseline := buildJobSet(t, podTemplateWithQueueLabel(t, "test-queue"))

	// Every other Kueue setting is ignored while Kueue is disabled.
	withConfig(t, func(c *Config) {
		c.Kueue = KueueConfig{
			Enabled:            false,
			QueueName:          "other-queue",
			EvictAsSystemRetry: false,
			AdmissionTimeout:   stdconfig.Duration{Duration: time.Minute},
		}
	})
	got := buildJobSet(t, podTemplateWithQueueLabel(t, "test-queue"))

	assert.Equal(t, baseline, got)
	assert.Nil(t, got.Spec.Suspend)
	// Without Kueue enabled the plugin does not own the label: it passes through as before.
	assert.Equal(t, "test-queue", got.Labels[kueueQueueNameLabel])
}

func TestBuildResource_KueueEnabled_LabelAndSuspend(t *testing.T) {
	withConfig(t, func(c *Config) { c.Kueue.Enabled = true })

	jobSet := buildJobSet(t, nil)

	assert.Equal(t, "user-queue", jobSet.Labels[kueueQueueNameLabel])
	require.NotNil(t, jobSet.Spec.Suspend)
	assert.True(t, *jobSet.Spec.Suspend)
}

func TestBuildResource_KueueEnabled_UserQueueRejected(t *testing.T) {
	withConfig(t, func(c *Config) {
		c.Kueue.Enabled = true
		c.Kueue.QueueName = platformQueue
	})

	_, err := clusteredResourceHandler{}.BuildResource(context.Background(),
		dummyTaskCtx(buildTaskTemplate(kueueTestSpec()), podTemplateWithQueueLabel(t, "chosen-by-user")))

	require.Error(t, err)
	assert.ErrorContains(t, err, "["+flyteerr.BadTaskSpecification+"]")
	assert.ErrorContains(t, err, "the pod template override")
	assert.ErrorContains(t, err, `"chosen-by-user"`)
	assert.ErrorContains(t, err, fmt.Sprintf("%q", platformQueue))
}

func TestBuildResource_KueueEnabled_UserNamesConfiguredQueue(t *testing.T) {
	withConfig(t, func(c *Config) {
		c.Kueue.Enabled = true
		c.Kueue.QueueName = platformQueue
	})

	jobSet := buildJobSet(t, podTemplateWithQueueLabel(t, platformQueue))

	assert.Equal(t, platformQueue, jobSet.Labels[kueueQueueNameLabel])
	podLabels := jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Labels
	assert.NotContains(t, podLabels, kueueQueueNameLabel, "pods must not carry a queue of their own")
	assert.Equal(t, "my-exec", podLabels["execution-id"], "other labels are untouched")
}

func TestApplyKueue_UserSources(t *testing.T) {
	cfg := &KueueConfig{Enabled: true, QueueName: platformQueue}
	newJobSet := func() *jobsetv1alpha2.JobSet {
		return &jobsetv1alpha2.JobSet{Spec: jobsetv1alpha2.JobSetSpec{
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{{Name: workersReplicatedJobName}},
		}}
	}

	t.Run("execution label naming another queue is rejected", func(t *testing.T) {
		err := applyKueue(newJobSet(), cfg, []labelSource{
			{name: execLabelsSource, labels: map[string]string{kueueQueueNameLabel: teamQueue}},
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, execLabelsSource)
	})

	t.Run("no user label is labelled and suspended", func(t *testing.T) {
		js := newJobSet()
		require.NoError(t, applyKueue(js, cfg, []labelSource{{name: execLabelsSource}}))
		assert.Equal(t, platformQueue, js.Labels[kueueQueueNameLabel])
		require.NotNil(t, js.Spec.Suspend)
		assert.True(t, *js.Spec.Suspend)
	})

	t.Run("disabled ignores user labels", func(t *testing.T) {
		js := newJobSet()
		require.NoError(t, applyKueue(js, &KueueConfig{QueueName: platformQueue}, []labelSource{
			{name: execLabelsSource, labels: map[string]string{kueueQueueNameLabel: teamQueue}},
		}))
		assert.Nil(t, js.Spec.Suspend)
		assert.NotContains(t, js.Labels, kueueQueueNameLabel)
	})
}

func TestBuildResource_KueueEnabled_InvalidQueueName(t *testing.T) {
	for name, queue := range map[string]string{
		"empty":         "  ",
		"invalid label": "not a/valid label",
	} {
		t.Run(name, func(t *testing.T) {
			withConfig(t, func(c *Config) {
				c.Kueue.Enabled = true
				c.Kueue.QueueName = queue
			})
			_, err := clusteredResourceHandler{}.BuildResource(context.Background(),
				dummyTaskCtx(buildTaskTemplate(kueueTestSpec()), nil))
			require.Error(t, err)
			assert.ErrorContains(t, err, "["+flyteerr.BadTaskSpecification+"]")
			assert.ErrorContains(t, err, "plugins.clustered.kueue.queue-name")
		})
	}
}

func TestBuildResource_StartupTimeout(t *testing.T) {
	t.Run("injected in whole seconds", func(t *testing.T) {
		withConfig(t, func(c *Config) {
			c.StartupTimeout = stdconfig.Duration{Duration: 20*time.Minute + 500*time.Millisecond}
		})
		assert.Equal(t, "1200", primaryEnv(buildJobSet(t, nil))[startupTimeoutEnv])
	})
	t.Run("independent of Kueue", func(t *testing.T) {
		withConfig(t, func(c *Config) {
			c.StartupTimeout = stdconfig.Duration{Duration: time.Minute}
			c.Kueue.Enabled = true
		})
		assert.Equal(t, "60", primaryEnv(buildJobSet(t, nil))[startupTimeoutEnv])
	})
	t.Run("unset leaves the launcher default", func(t *testing.T) {
		assert.NotContains(t, primaryEnv(buildJobSet(t, nil)), startupTimeoutEnv)
	})
}

// suspendedSince is a never-started JobSet held suspended since created.
func suspendedSince(created time.Time) *jobsetv1alpha2.JobSet {
	js := makeJobSet(jobsetv1alpha2.JobSetSuspended, metav1.ConditionTrue, true)
	js.CreationTimestamp = metav1.NewTime(created)
	return js
}

func TestGetTaskPhase_AdmissionTimeout(t *testing.T) {
	ctx := context.Background()

	t.Run("exceeded is a system retry with cleanup", func(t *testing.T) {
		withConfig(t, func(c *Config) { c.Kueue.AdmissionTimeout = stdconfig.Duration{Duration: 2 * time.Hour} })

		phase, err := clusteredResourceHandler{}.GetTaskPhase(ctx,
			dummyPluginCtx(twoNodeSpec(), emptyK8sReader()), suspendedSince(time.Now().Add(-3*time.Hour)))
		require.NoError(t, err)

		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
		require.NotNil(t, phase.Err())
		assert.Equal(t, gang.CodeGangAdmissionTimeout, phase.Err().GetCode())
		assert.Equal(t, core.ExecutionError_SYSTEM, phase.Err().GetKind())
		assert.Contains(t, phase.Err().GetMessage(), "not admitted within 2h0m0s")
		assert.True(t, phase.CleanupOnFailure())
		assert.False(t, gang.IsEviction(phase.Err()), "nothing ran, so it is not charged to the eviction budget")
	})

	t.Run("not yet exceeded keeps holding", func(t *testing.T) {
		withConfig(t, func(c *Config) { c.Kueue.AdmissionTimeout = stdconfig.Duration{Duration: 2 * time.Hour} })

		phase, err := clusteredResourceHandler{}.GetTaskPhase(ctx,
			dummyPluginCtx(twoNodeSpec(), emptyK8sReader()), suspendedSince(time.Now().Add(-time.Hour)))
		require.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	})

	t.Run("disabled keeps holding", func(t *testing.T) {
		phase, err := clusteredResourceHandler{}.GetTaskPhase(ctx,
			dummyPluginCtx(twoNodeSpec(), emptyK8sReader()), suspendedSince(time.Now().Add(-48*time.Hour)))
		require.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	})

	t.Run("a started gang is evicted, not timed out", func(t *testing.T) {
		withConfig(t, func(c *Config) { c.Kueue.AdmissionTimeout = stdconfig.Duration{Duration: time.Minute} })

		pCtx := dummyPluginCtxWithState(twoNodeSpec(), emptyK8sReader(),
			plugink8s.PluginState{Phase: pluginsCore.PhaseRunning, PhaseVersion: 1}, nil)
		phase, err := clusteredResourceHandler{}.GetTaskPhase(ctx, pCtx, suspendedSince(time.Now().Add(-time.Hour)))
		require.NoError(t, err)
		require.NotNil(t, phase.Err())
		assert.Equal(t, gang.CodeGangEvicted, phase.Err().GetCode())
	})
}
