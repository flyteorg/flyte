package plugin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytesecret "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func TestResolveServiceAccount(t *testing.T) {
	// A service account set on the task's security context always wins over the default.
	sc := &core.SecurityContext{RunAs: &core.Identity{K8SServiceAccount: "custom-sa"}}
	require.Equal(t, "custom-sa", resolveServiceAccount(sc, "config-default-sa"))

	// With no run account, it falls back to the provided default. Getters are nil-safe.
	require.Equal(t, "config-default-sa", resolveServiceAccount(nil, "config-default-sa"))
	require.Equal(t, "config-default-sa", resolveServiceAccount(&core.SecurityContext{}, "config-default-sa"))

	// An empty default resolves to empty, so Kubernetes uses the pod namespace's `default`.
	require.Equal(t, "", resolveServiceAccount(nil, ""))
}

func TestNewTaskExecutionMetadata_UsesProjectedRunContext(t *testing.T) {
	interruptible := true
	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			EnvVars:       map[string]string{"TRACE_ID": "root-abc"},
			Interruptible: &interruptible,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)
	require.Equal(t, "root-abc", meta.GetEnvironmentVariables()["TRACE_ID"])
	require.True(t, meta.IsInterruptible())
}

func TestNewTaskExecutionMetadata_Interruptible(t *testing.T) {
	marshalTemplate := func(t *testing.T, tmpl *core.TaskTemplate) []byte {
		data, err := proto.Marshal(tmpl)
		require.NoError(t, err)
		return data
	}
	templateWith := func(t *testing.T, interruptible bool) []byte {
		return marshalTemplate(t, &core.TaskTemplate{
			Metadata: &core.TaskMetadata{
				InterruptibleValue: &core.TaskMetadata_Interruptible{Interruptible: interruptible},
			},
		})
	}
	newMeta := func(t *testing.T, runLevel *bool, taskTemplate []byte) pluginsCore.TaskExecutionMetadata {
		meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{
			Spec: flyteorgv1.TaskActionSpec{
				Project:       "project",
				Domain:        "development",
				RunName:       "run-name",
				ActionName:    "action-name",
				RunOutputBase: "s3://bucket/run",
				Interruptible: runLevel,
				TaskTemplate:  taskTemplate,
			},
		})
		require.NoError(t, err)
		return meta
	}

	t.Run("defaults to false", func(t *testing.T) {
		require.False(t, newMeta(t, nil, nil).IsInterruptible())
	})

	t.Run("template metadata applies when run level is unset", func(t *testing.T) {
		require.True(t, newMeta(t, nil, templateWith(t, true)).IsInterruptible())
		require.False(t, newMeta(t, nil, templateWith(t, false)).IsInterruptible())
	})

	t.Run("template without the flag leaves the default", func(t *testing.T) {
		require.False(t, newMeta(t, nil, marshalTemplate(t, &core.TaskTemplate{Metadata: &core.TaskMetadata{}})).IsInterruptible())
	})

	t.Run("run level wins over template metadata", func(t *testing.T) {
		runTrue, runFalse := true, false
		require.True(t, newMeta(t, &runTrue, templateWith(t, false)).IsInterruptible())
		require.False(t, newMeta(t, &runFalse, templateWith(t, true)).IsInterruptible())
	})
}

func TestNewTaskExecutionMetadata_UserEnvVarsCannotClobberInternal(t *testing.T) {
	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			EnvVars: map[string]string{
				"ACTION_NAME": "malicious-override",
				"RUN_NAME":    "malicious-override",
				"_U_RUN_BASE": "malicious-override",
				"USER_VAR":    "allowed",
			},
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	env := meta.GetEnvironmentVariables()
	require.Equal(t, "action-name", env["ACTION_NAME"])
	require.Equal(t, "run-name", env["RUN_NAME"])
	require.Equal(t, "s3://bucket/run", env["_U_RUN_BASE"])
	require.Equal(t, "allowed", env["USER_VAR"])
}

func TestNewTaskExecutionMetadata_UsesTaskTemplateID(t *testing.T) {
	taskID := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "development",
		Name:         "task-name",
		Version:      "version",
	}
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{Id: taskID})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			TaskTemplate:  taskTemplate,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	got := meta.GetTaskExecutionID().GetID().GetTaskId()
	require.NotNil(t, got)
	require.Equal(t, taskID.GetResourceType(), got.GetResourceType())
	require.Equal(t, taskID.GetProject(), got.GetProject())
	require.Equal(t, taskID.GetDomain(), got.GetDomain())
	require.Equal(t, taskID.GetName(), got.GetName())
	require.Equal(t, taskID.GetVersion(), got.GetVersion())
}

func TestTaskExecutionID_GetGeneratedNameWith(t *testing.T) {
	t.Run("within max length", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "short-name",
		}
		name, err := execID.GetGeneratedNameWith(0, 50)
		require.NoError(t, err)
		require.Equal(t, "short-name", name)
	})

	t.Run("exceeds max length uses FixedLengthUniqueID", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "a-very-long-generated-name-that-exceeds-the-max-length-limit-12345",
		}
		maxLength := 20
		name, err := execID.GetGeneratedNameWith(0, maxLength)
		require.NoError(t, err)
		require.LessOrEqual(t, len(name), maxLength)

		expected, err := encoding.FixedLengthUniqueID(execID.generatedName, maxLength)
		require.NoError(t, err)
		require.Equal(t, expected, name)
	})

	t.Run("max length too small returns error", func(t *testing.T) {
		execID := &taskExecutionID{
			generatedName: "long-name",
		}
		_, err := execID.GetGeneratedNameWith(0, 2)
		require.Error(t, err)
	})
}

func TestNewTaskExecutionMetadata_ManagedLabel(t *testing.T) {
	// The manager cache selects on this label, so a Pod without it is invisible
	// to the executor. Plugins merge these labels into the Pod templates they
	// build, which is how operator-created Pods inherit it.
	newMeta := func(labels map[string]string) pluginsCore.TaskExecutionMetadata {
		meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{
			ObjectMeta: metav1.ObjectMeta{Labels: labels},
			Spec: flyteorgv1.TaskActionSpec{
				Project:       "project",
				Domain:        "development",
				RunName:       "run-name",
				ActionName:    "action-name",
				RunOutputBase: "s3://bucket/run",
			},
		})
		require.NoError(t, err)
		return meta
	}

	t.Run("always set", func(t *testing.T) {
		require.Equal(t, flytek8s.ManagedLabelValue,
			newMeta(nil).GetLabels()[flytek8s.ManagedLabelKey])
	})

	t.Run("matches the cache selector", func(t *testing.T) {
		selector := labels.SelectorFromSet(labels.Set{
			flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
		})
		require.True(t, selector.Matches(labels.Set(newMeta(nil).GetLabels())))
	})

	t.Run("user labels cannot override it", func(t *testing.T) {
		meta := newMeta(map[string]string{flytek8s.ManagedLabelKey: "false"})
		require.Equal(t, flytek8s.ManagedLabelValue,
			meta.GetLabels()[flytek8s.ManagedLabelKey])
	})
}

func TestNewTaskExecutionMetadata_StampsIdentifyingLabels(t *testing.T) {
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{
		Id: &core.Identifier{Name: "my_module.my_task"},
	})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:       "project",
			Domain:        "development",
			RunName:       "run-name",
			ActionName:    "action-name",
			RunOutputBase: "s3://bucket/run",
			TaskTemplate:  taskTemplate,
		},
		Status: flyteorgv1.TaskActionStatus{Attempts: 3},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name", labels[RunLabel])
	require.Equal(t, "action-name", labels[ActionLabel])
	// Attempts are 1-based, and the third attempt is the second retry.
	require.Equal(t, "3", labels[AttemptLabel])
	require.Equal(t, "my_module.my_task", labels[TaskNameLabel])

	// The labels injected for secret scoping are still there.
	require.Equal(t, "project", labels[flytesecret.ProjectLabel])
	require.Equal(t, "development", labels[flytesecret.DomainLabel])
}

func TestNewTaskExecutionMetadata_LabelsAreSanitizedAndOptional(t *testing.T) {
	longName := strings.Repeat("a", 70)
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{
		Id: &core.Identifier{Name: longName},
	})
	require.NoError(t, err)

	taskAction := &flyteorgv1.TaskAction{
		Spec: flyteorgv1.TaskActionSpec{
			Project:      "project",
			Domain:       "development",
			RunName:      "run/name:with bad chars",
			TaskTemplate: taskTemplate,
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name-with-bad-chars", labels[RunLabel])
	require.Len(t, labels[TaskNameLabel], 63)
	// A first attempt that has not been recorded yet still reads as attempt 1.
	require.Equal(t, "1", labels[AttemptLabel])
	// An empty action name is left unstamped rather than stamped blank.
	require.NotContains(t, labels, ActionLabel)

	for key, value := range labels {
		require.Empty(t, validation.IsValidLabelValue(value), "label %s has invalid value %q", key, value)
	}
}

func TestNewTaskExecutionMetadata_CRDLabelsDoNotClobberIdentifyingLabels(t *testing.T) {
	taskAction := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{RunLabel: "spoofed", "team": "data"},
		},
		Spec: flyteorgv1.TaskActionSpec{
			Project:    "project",
			Domain:     "development",
			RunName:    "run-name",
			ActionName: "action-name",
		},
	}

	meta, err := NewTaskExecutionMetadata(taskAction)
	require.NoError(t, err)

	labels := meta.GetLabels()
	require.Equal(t, "run-name", labels[RunLabel])
	require.Equal(t, "data", labels["team"])
}

func TestNewTaskExecutionMetadata_TaskResourceDefaults(t *testing.T) {
	entries := func(kv ...string) []*core.Resources_ResourceEntry {
		names := map[string]core.Resources_ResourceName{
			"cpu": core.Resources_CPU, "memory": core.Resources_MEMORY,
			"gpu": core.Resources_GPU, "storage": core.Resources_EPHEMERAL_STORAGE,
		}
		var out []*core.Resources_ResourceEntry
		for i := 0; i < len(kv); i += 2 {
			out = append(out, &core.Resources_ResourceEntry{Name: names[kv[i]], Value: kv[i+1]})
		}
		return out
	}

	tests := []struct {
		name         string
		taskRequests []*core.Resources_ResourceEntry
		taskLimits   []*core.Resources_ResourceEntry
		defaults     *task.TaskResourceDefaults
		want         v1.ResourceRequirements
	}{
		{
			name:         "an unset dimension takes min",
			taskRequests: entries("cpu", "500m"),
			defaults:     &task.TaskResourceDefaults{Requests: entries("memory", "256Mi", "storage", "1Gi")},
			want: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("500m"), v1.ResourceMemory: resource.MustParse("256Mi"),
					v1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
				},
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("500m"), v1.ResourceMemory: resource.MustParse("256Mi"),
					v1.ResourceEphemeralStorage: resource.MustParse("1Gi"),
				},
			},
		},
		{
			name:         "a request above max is clamped",
			taskRequests: entries("cpu", "4", "memory", "1Gi"),
			defaults:     &task.TaskResourceDefaults{Max: entries("cpu", "1")},
			want: v1.ResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
				Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
			},
		},
		{
			name:         "a limit above max is clamped",
			taskRequests: entries("cpu", "500m", "memory", "1Gi"),
			taskLimits:   entries("cpu", "4"),
			defaults:     &task.TaskResourceDefaults{Max: entries("cpu", "1")},
			want: v1.ResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("500m"), v1.ResourceMemory: resource.MustParse("1Gi")},
				Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
			},
		},
		{
			name:         "min above max is capped at max",
			taskRequests: entries("memory", "1Gi"),
			defaults:     &task.TaskResourceDefaults{Requests: entries("cpu", "2"), Max: entries("cpu", "1")},
			want: v1.ResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
				Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")},
			},
		},
		{
			name:         "no settings means no change",
			taskRequests: entries("cpu", "4", "memory", "1Gi"),
			want: v1.ResourceRequirements{
				Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("4"), v1.ResourceMemory: resource.MustParse("1Gi")},
				Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("4"), v1.ResourceMemory: resource.MustParse("1Gi")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskTemplate, err := proto.Marshal(&core.TaskTemplate{Target: &core.TaskTemplate_Container{Container: &core.Container{
				Resources: &core.Resources{Requests: tt.taskRequests, Limits: tt.taskLimits},
			}}})
			require.NoError(t, err)
			var defaults []byte
			if tt.defaults != nil {
				defaults, err = proto.Marshal(tt.defaults)
				require.NoError(t, err)
			}

			meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{Spec: flyteorgv1.TaskActionSpec{
				TaskTemplate:         taskTemplate,
				TaskResourceDefaults: defaults,
			}})
			require.NoError(t, err)

			got := flytek8s.ApplyResourceOverrides(*meta.GetOverrides().GetResources(), *meta.GetPlatformResources(), true)
			require.Equal(t, len(tt.want.Requests), len(got.Requests), "requests: %v", got.Requests)
			require.Equal(t, len(tt.want.Limits), len(got.Limits), "limits: %v", got.Limits)
			for name, want := range tt.want.Requests {
				require.Zero(t, want.Cmp(got.Requests[name]), "request %s: got %v, want %v", name, got.Requests[name], want)
			}
			for name, want := range tt.want.Limits {
				require.Zero(t, want.Cmp(got.Limits[name]), "limit %s: got %v, want %v", name, got.Limits[name], want)
			}
		})
	}
}

func TestNewTaskExecutionMetadata_DefaultAccelerator(t *testing.T) {
	defaultAccelerator := &core.GPUAccelerator{Device: "nvidia-a100-80gb"}
	defaults, err := proto.Marshal(&task.TaskResourceDefaults{
		Max:                []*core.Resources_ResourceEntry{{Name: core.Resources_GPU, Value: "2"}},
		DefaultAccelerator: defaultAccelerator,
	})
	require.NoError(t, err)

	gpuRequest := &core.Resources{Requests: []*core.Resources_ResourceEntry{{Name: core.Resources_GPU, Value: "1"}}}
	tests := []struct {
		name string
		tmpl *core.TaskTemplate
		want *core.GPUAccelerator
	}{
		{
			name: "a task requesting GPUs with no accelerator gets the default",
			tmpl: &core.TaskTemplate{Target: &core.TaskTemplate_Container{Container: &core.Container{Resources: gpuRequest}}},
			want: defaultAccelerator,
		},
		{
			name: "a task's own accelerator wins",
			tmpl: &core.TaskTemplate{
				Target:            &core.TaskTemplate_Container{Container: &core.Container{Resources: gpuRequest}},
				ExtendedResources: &core.ExtendedResources{GpuAccelerator: &core.GPUAccelerator{Device: "nvidia-t4"}},
			},
			want: &core.GPUAccelerator{Device: "nvidia-t4"},
		},
		{
			name: "a task requesting no GPUs gets no accelerator",
			tmpl: &core.TaskTemplate{Target: &core.TaskTemplate_Container{Container: &core.Container{}}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskTemplate, err := proto.Marshal(tt.tmpl)
			require.NoError(t, err)

			meta, err := NewTaskExecutionMetadata(&flyteorgv1.TaskAction{Spec: flyteorgv1.TaskActionSpec{
				TaskTemplate:         taskTemplate,
				TaskResourceDefaults: defaults,
			}})
			require.NoError(t, err)

			got := flytek8s.ApplyExtendedResourcesOverrides(
				tt.tmpl.GetExtendedResources(),
				meta.GetOverrides().GetExtendedResources(),
			).GetGpuAccelerator()
			require.True(t, proto.Equal(tt.want, got), "got %v, want %v", got, tt.want)
			// The GPU cap is keyed by the accelerator's resource name, not the generic "gpu".
			_, generic := meta.GetPlatformResources().Limits["gpu"]
			require.False(t, generic)
			require.Contains(t, meta.GetPlatformResources().Limits, v1.ResourceName("nvidia.com/gpu"))
		})
	}
}
