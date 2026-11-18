package ray

import (
	"context"
	"testing"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

const testLogsSidecarName = "logs-sidecar"

func rayTestSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsNonRoot:             ptr.To(true),
		RunAsUser:                ptr.To(int64(1001)),
		RunAsGroup:               ptr.To(int64(1002)),
		AllowPrivilegeEscalation: ptr.To(false),
		ReadOnlyRootFilesystem:   ptr.To(true),
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}

func setRaySecurityContextTestConfig(t *testing.T, rayConfig *Config, k8sConfig *config.K8sPluginConfig) {
	t.Helper()
	previousRayConfig := GetConfig()
	previousK8sConfig := config.GetK8sPluginConfig()
	t.Cleanup(func() {
		require.NoError(t, SetConfig(previousRayConfig))
		require.NoError(t, config.SetK8sPluginConfig(previousK8sConfig))
	})
	require.NoError(t, SetConfig(rayConfig))
	require.NoError(t, config.SetK8sPluginConfig(k8sConfig))
}

func rayTestContainer(t *testing.T, pod *corev1.PodSpec, name string) *corev1.Container {
	t.Helper()
	for i := range pod.Containers {
		if pod.Containers[i].Name == name {
			return &pod.Containers[i]
		}
	}
	t.Fatalf("container %q missing from pod", name)
	return nil
}

func TestBuildResourceRaySecurityContext(t *testing.T) {
	securityContext := rayTestSecurityContext()
	headOverride := &corev1.SecurityContext{RunAsUser: ptr.To(int64(2001))}
	mergedHeadContext := securityContext.DeepCopy()
	mergedHeadContext.RunAsUser = ptr.To(int64(2001))
	logsContext := &corev1.SecurityContext{RunAsUser: ptr.To(int64(3001))}
	logsOverride := &corev1.SecurityContext{AllowPrivilegeEscalation: ptr.To(false)}
	mergedLogsContext := logsContext.DeepCopy()
	mergedLogsContext.AllowPrivilegeEscalation = ptr.To(false)

	for _, test := range []struct {
		name                  string
		defaultContext        *corev1.SecurityContext
		taskContext           *corev1.SecurityContext
		headContext           *corev1.SecurityContext
		podContext            *corev1.PodSecurityContext
		logsContext           *corev1.SecurityContext
		headLogsContext       *corev1.SecurityContext
		autoscalerOptions     *plugins.AutoscalerOptions
		disableAutoscaling    bool
		withoutLogs           bool
		primaryNotFirst       bool
		wantHeadContext       *corev1.SecurityContext
		wantAutoscalerContext *corev1.SecurityContext
		wantLogsContext       *corev1.SecurityContext
	}{
		{
			name: "default context with no autoscaler options", defaultContext: securityContext,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: securityContext,
		},
		{
			name: "default context with autoscaler options", defaultContext: securityContext,
			autoscalerOptions: &plugins.AutoscalerOptions{IdleTimeoutSeconds: 30},
			wantHeadContext:   securityContext, wantAutoscalerContext: securityContext, wantLogsContext: securityContext,
		},
		{
			name: "task context overrides default", defaultContext: logsContext, taskContext: securityContext,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: securityContext,
		},
		{
			name: "head context overrides task context", taskContext: securityContext, headContext: headOverride,
			wantHeadContext: mergedHeadContext, wantAutoscalerContext: mergedHeadContext, wantLogsContext: mergedHeadContext,
		},
		{
			name: "head context without base context", headContext: securityContext,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: securityContext,
		},
		{name: "no container context"},
		{
			name:              "no context with autoscaler options",
			autoscalerOptions: &plugins.AutoscalerOptions{IdleTimeoutSeconds: 30},
		},
		{
			name:       "pod context does not become a container context",
			podContext: &corev1.PodSecurityContext{RunAsNonRoot: ptr.To(true), RunAsUser: ptr.To(int64(1001))},
		},
		{
			name: "explicit empty head context", headContext: &corev1.SecurityContext{},
			wantHeadContext: &corev1.SecurityContext{}, wantAutoscalerContext: &corev1.SecurityContext{}, wantLogsContext: &corev1.SecurityContext{},
		},
		{
			name: "explicit logs context", defaultContext: securityContext, logsContext: logsContext,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: logsContext,
		},
		{
			name: "explicit empty logs context", defaultContext: securityContext, logsContext: &corev1.SecurityContext{},
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: &corev1.SecurityContext{},
		},
		{
			name: "head template sets logs context", defaultContext: securityContext, headLogsContext: logsOverride,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: logsOverride,
		},
		{
			name: "head template merges explicit logs context", defaultContext: securityContext,
			logsContext: logsContext, headLogsContext: logsOverride,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: mergedLogsContext,
		},
		{
			name: "autoscaling disabled without options", defaultContext: securityContext, disableAutoscaling: true,
			wantHeadContext: securityContext, wantLogsContext: securityContext,
		},
		{
			name: "autoscaling disabled with options", defaultContext: securityContext, disableAutoscaling: true,
			autoscalerOptions: &plugins.AutoscalerOptions{IdleTimeoutSeconds: 30},
			wantHeadContext:   securityContext, wantLogsContext: securityContext,
		},
		{
			name: "no logs sidecar configured", defaultContext: securityContext, withoutLogs: true,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext,
		},
		{
			name: "primary container is not first", taskContext: securityContext, primaryNotFirst: true,
			wantHeadContext: securityContext, wantAutoscalerContext: securityContext, wantLogsContext: securityContext,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			rayConfig := &Config{}
			if !test.withoutLogs {
				rayConfig.LogsSidecar = &corev1.Container{
					Name: testLogsSidecarName, Image: testImage, SecurityContext: test.logsContext.DeepCopy(),
				}
			}
			setRaySecurityContextTestConfig(t, rayConfig, &config.K8sPluginConfig{
				DefaultSecurityContext: test.defaultContext.DeepCopy(), DefaultPodSecurityContext: test.podContext.DeepCopy(),
			})

			input := dummyRayCustomObj()
			input.RayCluster.EnableAutoscaling = !test.disableAutoscaling
			input.RayCluster.AutoscalerOptions = test.autoscalerOptions
			if test.headContext != nil || test.headLogsContext != nil {
				headPodSpec := &corev1.PodSpec{}
				if test.headContext != nil {
					headPodSpec.Containers = append(headPodSpec.Containers, corev1.Container{
						Name: RayHeadContainerName, SecurityContext: test.headContext.DeepCopy(),
					})
				}
				if test.headLogsContext != nil {
					headPodSpec.Containers = append(headPodSpec.Containers, corev1.Container{
						Name: testLogsSidecarName, SecurityContext: test.headLogsContext.DeepCopy(),
					})
				}
				input.RayCluster.HeadGroupSpec.K8SPod = &core.K8SPod{PodSpec: transformStructToStructPB(t, headPodSpec)}
			}

			task := dummyRayTaskTemplate("security-context", input)
			if test.taskContext != nil || test.primaryNotFirst {
				pod := &corev1.PodSpec{Containers: []corev1.Container{{
					Name: "main", Image: testImage, Args: testArgs, SecurityContext: test.taskContext.DeepCopy(),
				}}}
				if test.primaryNotFirst {
					pod.Containers = append([]corev1.Container{{Name: "other", Image: testImage, SecurityContext: logsContext.DeepCopy()}}, pod.Containers...)
				}
				task.Target = transformPodSpecToTaskTemplateTarget(pod)
				task.Config = map[string]string{flytek8s.PrimaryContainerKey: "main"}
			}
			taskContext := dummyRayTaskContext(task, resourceRequirements, nil, "", serviceAccount)
			resource, err := (rayJobResourceHandler{}).BuildResource(context.Background(), taskContext)
			require.NoError(t, err)
			job := resource.(*rayv1.RayJob)
			headPod := &job.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
			assert.Equal(t, test.wantHeadContext, rayTestContainer(t, headPod, RayHeadContainerName).SecurityContext)
			assert.Equal(t, test.podContext, headPod.SecurityContext)
			assert.Equal(t, !test.disableAutoscaling, *job.Spec.RayClusterSpec.EnableInTreeAutoscaling)
			if test.primaryNotFirst {
				assert.Equal(t, "other", headPod.Containers[0].Name)
				assert.Equal(t, logsContext, headPod.Containers[0].SecurityContext)
			}
			if test.withoutLogs {
				for _, container := range headPod.Containers {
					assert.NotEqual(t, testLogsSidecarName, container.Name)
				}
			} else {
				assert.Equal(t, test.wantLogsContext, rayTestContainer(t, headPod, testLogsSidecarName).SecurityContext)
			}
			autoscaler := job.Spec.RayClusterSpec.AutoscalerOptions
			if test.wantAutoscalerContext != nil || test.autoscalerOptions != nil {
				require.NotNil(t, autoscaler)
				assert.Equal(t, test.wantAutoscalerContext, autoscaler.SecurityContext)
				if test.autoscalerOptions != nil {
					require.NotNil(t, autoscaler.IdleTimeoutSeconds)
					assert.Equal(t, int32(30), *autoscaler.IdleTimeoutSeconds)
				}
			} else {
				assert.Nil(t, autoscaler)
			}
		})
	}
}

func TestBuildResourceRaySecurityContextIsolation(t *testing.T) {
	contexts := func(t *testing.T, job *rayv1.RayJob) map[string]*corev1.SecurityContext {
		headPod := &job.Spec.RayClusterSpec.HeadGroupSpec.Template.Spec
		require.NotNil(t, job.Spec.RayClusterSpec.AutoscalerOptions)
		return map[string]*corev1.SecurityContext{
			"head":       rayTestContainer(t, headPod, RayHeadContainerName).SecurityContext,
			"logs":       rayTestContainer(t, headPod, testLogsSidecarName).SecurityContext,
			"autoscaler": job.Spec.RayClusterSpec.AutoscalerOptions.SecurityContext,
		}
	}
	for _, modified := range []string{"head", "logs", "autoscaler"} {
		t.Run(modified, func(t *testing.T) {
			securityContext := rayTestSecurityContext()
			logsSidecar := &corev1.Container{Name: testLogsSidecarName, Image: testImage}
			setRaySecurityContextTestConfig(t, &Config{LogsSidecar: logsSidecar}, &config.K8sPluginConfig{
				DefaultSecurityContext: securityContext,
			})
			input := dummyRayCustomObj()
			task := dummyRayTaskTemplate("security-context-isolation", input)
			build := func() *rayv1.RayJob {
				taskContext := dummyRayTaskContext(task, resourceRequirements, nil, "", serviceAccount)
				resource, err := (rayJobResourceHandler{}).BuildResource(context.Background(), taskContext)
				require.NoError(t, err)
				return resource.(*rayv1.RayJob)
			}
			first, second := contexts(t, build()), contexts(t, build())
			for _, securityContext := range first {
				require.Equal(t, rayTestSecurityContext(), securityContext)
			}
			*first[modified].RunAsUser = 0
			first[modified].Capabilities.Drop[0] = "NET_ADMIN"
			first[modified].SeccompProfile.Type = corev1.SeccompProfileTypeUnconfined
			for name, securityContext := range first {
				if name != modified {
					assert.Equal(t, rayTestSecurityContext(), securityContext, name)
				}
			}
			for name, securityContext := range second {
				assert.Equal(t, rayTestSecurityContext(), securityContext, name)
			}
			assert.Equal(t, rayTestSecurityContext(), securityContext)
			assert.Equal(t, &corev1.Container{Name: testLogsSidecarName, Image: testImage}, logsSidecar)
		})
	}
}
