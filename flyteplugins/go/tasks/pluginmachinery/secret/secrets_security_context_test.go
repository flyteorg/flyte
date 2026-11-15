package secret

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	secretUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func regressionSecurityContext(uid int64) *corev1.SecurityContext {
	return &corev1.SecurityContext{
		RunAsUser: ptr.To(uid), RunAsGroup: ptr.To(uid), RunAsNonRoot: ptr.To(true),
		AllowPrivilegeEscalation: ptr.To(false), ReadOnlyRootFilesystem: ptr.To(true),
		Capabilities:   &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}

func regressionSecretPod(t *testing.T, secrets ...*core.Secret) *corev1.Pod {
	t.Helper()
	annotations, err := secretUtils.MarshalSecretsToMapStrings(secrets)
	require.NoError(t, err)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "task", Namespace: "test-namespace", Annotations: annotations,
			Labels: map[string]string{OrganizationLabel: "org", ProjectLabel: "project", DomainLabel: "development"}},
		Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "primary", SecurityContext: regressionSecurityContext(1000)}}},
	}
}

func regressionMutator(kind config.SecretManagerType, injector SecretsInjector) *SecretsPodMutator {
	return NewSecretsMutatorFromInjectors([]config.SecretManagerType{kind}, map[config.SecretManagerType]SecretsInjector{kind: injector})
}

func TestSecretsWebhook_SecurityContextInjectors(t *testing.T) {
	for _, kind := range []config.SecretManagerType{
		config.SecretManagerTypeEmbedded, config.SecretManagerTypeAWS, config.SecretManagerTypeGCP, config.SecretManagerTypeAzure,
	} {
		t.Run(kind.String(), func(t *testing.T) {
			first := &core.Secret{Group: "group", GroupVersion: "version", Key: "first", MountRequirement: core.Secret_FILE}
			second := &core.Secret{Group: "group", GroupVersion: "version", Key: "second", MountRequirement: core.Secret_FILE}
			pod := regressionSecretPod(t, first, second)
			pod.Spec.InitContainers = []corev1.Container{{Name: "existing"}}
			cfg := *config.DefaultConfig
			var injector SecretsInjector
			switch kind {
			case config.SecretManagerTypeAWS:
				injector = NewAWSSecretManagerInjector(cfg.AWSSecretManagerConfig)
			case config.SecretManagerTypeGCP:
				injector = NewGCPSecretManagerInjector(cfg.GCPSecretManagerConfig)
			case config.SecretManagerTypeAzure:
				injector = NewAzureSecretManagerInjector(cfg.AzureSecretManagerConfig)
			case config.SecretManagerTypeEmbedded:
				cfg.EmbeddedSecretManagerConfig.Type = config.EmbeddedSecretManagerTypeK8s
				cfg.EmbeddedSecretManagerConfig.K8sConfig.Namespace = pod.Namespace
				cfg.EmbeddedSecretManagerConfig.ImagePullSecrets.Enabled = false
				stored := []*corev1.Secret{}
				for _, key := range []string{"first", "second"} {
					id := EncodeSecretName("org", "development", "project", key)
					stored = append(stored, &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: EncodeK8sSecretName(id), Namespace: pod.Namespace}, Data: map[string][]byte{id: []byte("dummy")}})
				}
				scheme := runtime.NewScheme()
				require.NoError(t, corev1.AddToScheme(scheme))
				client := clientfake.NewClientBuilder().WithScheme(scheme).WithObjects(stored[0], stored[1]).Build()
				fetcher := NewK8sSecretFetcher(kubefake.NewSimpleClientset(stored[0], stored[1]).CoreV1().Secrets(pod.Namespace))
				injector = NewEmbeddedSecretManagerInjector(cfg.EmbeddedSecretManagerConfig, []SecretFetcher{fetcher}, client, pod.Namespace, newAlwaysMissCache(t), &cfg)
			}
			result, changed, response := regressionMutator(kind, injector).Mutate(context.Background(), pod)
			require.Nil(t, response)
			require.True(t, changed)
			wantCount := 3
			if kind == config.SecretManagerTypeEmbedded {
				wantCount = 2
			}
			require.Len(t, result.Spec.InitContainers, wantCount)
			assert.Nil(t, result.Spec.InitContainers[0].SecurityContext)
			for idx := 1; idx < wantCount; idx++ {
				require.Equal(t, result.Spec.Containers[0].SecurityContext, result.Spec.InitContainers[idx].SecurityContext)
			}
			result.Spec.InitContainers[1].SecurityContext.Capabilities.Drop[0] = "CHANGED"
			assert.Equal(t, corev1.Capability("ALL"), result.Spec.Containers[0].SecurityContext.Capabilities.Drop[0])
			if wantCount == 3 {
				assert.Equal(t, corev1.Capability("ALL"), result.Spec.InitContainers[2].SecurityContext.Capabilities.Drop[0])
			}
		})
	}
}

type regressionAppender struct {
	contexts []*corev1.SecurityContext
	fail     bool
	replace  bool
}

func (regressionAppender) Type() config.SecretManagerType                                  { return config.SecretManagerTypeK8s }
func (regressionAppender) InvalidateCache(context.Context, string, string, string, string) {}
func (i regressionAppender) Inject(_ context.Context, _ *core.Secret, pod *corev1.Pod) (*corev1.Pod, bool, error) {
	if i.fail {
		return pod, false, errors.New("injection failed")
	}
	if i.replace {
		pod = pod.DeepCopy()
	}
	for _, sc := range i.contexts {
		pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{Name: "injected", SecurityContext: sc})
	}
	return pod, true, nil
}

func TestSecretsWebhook_SecurityContextLookup(t *testing.T) {
	const annotationWinner = "flyte"
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		podName     string
		want        string
	}{
		{name: "first fallback", want: "first"},
		{name: "Flyte annotation", annotations: map[string]string{"primary_container_name": "flyte"}, want: "flyte"},
		{name: "kubectl annotation", annotations: map[string]string{"kubectl.kubernetes.io/default-container": "kubectl"}, want: "kubectl"},
		{name: "annotation precedence", annotations: map[string]string{"primary_container_name": "flyte", "kubectl.kubernetes.io/default-container": "kubectl"}, want: annotationWinner},
		{name: "pod name", podName: "flyte", want: "flyte"},
		{name: "annotation beats pod name", podName: "first", annotations: map[string]string{"primary_container_name": "flyte"}, want: "flyte"},
		{name: "empty annotations", annotations: map[string]string{"primary_container_name": "", "kubectl.kubernetes.io/default-container": ""}, want: "first"},
		{name: "missing Flyte declaration", annotations: map[string]string{"primary_container_name": "missing"}},
		{name: "missing kubectl declaration", annotations: map[string]string{"kubectl.kubernetes.io/default-container": "missing"}},
		{name: "invalid higher priority", annotations: map[string]string{"primary_container_name": "missing", "kubectl.kubernetes.io/default-container": "missing"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := regressionSecretPod(t, &core.Secret{Key: "dummy"})
			pod.Name = tc.podName
			for key, value := range tc.annotations {
				pod.Annotations[key] = value
			}
			pod.Spec.Containers = []corev1.Container{
				{Name: "first", SecurityContext: regressionSecurityContext(1000)},
				{Name: "flyte", SecurityContext: regressionSecurityContext(2000)},
				{Name: "kubectl", SecurityContext: regressionSecurityContext(3000)},
			}
			var want *corev1.SecurityContext
			for _, c := range pod.Spec.Containers {
				if c.Name == tc.want {
					want = c.SecurityContext.DeepCopy()
				}
			}
			result, changed, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{contexts: []*corev1.SecurityContext{nil}, replace: true}).Mutate(context.Background(), pod)
			require.Nil(t, response)
			assert.True(t, changed)
			assert.Equal(t, want, result.Spec.InitContainers[0].SecurityContext)
		})
	}
	for _, declarations := range []map[string]string{
		{"primary_container_name": "missing", "kubectl.kubernetes.io/default-container": "flyte"},
	} {
		pod := regressionSecretPod(t, &core.Secret{Key: "dummy"})
		for k, v := range declarations {
			pod.Annotations[k] = v
		}
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: "flyte", SecurityContext: regressionSecurityContext(2000)})
		result, _, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{contexts: []*corev1.SecurityContext{nil}}).Mutate(context.Background(), pod)
		require.Nil(t, response)
		assert.Nil(t, result.Spec.InitContainers[0].SecurityContext)
	}
}

func TestSecretsWebhook_SecurityContextPreservation(t *testing.T) {
	pod := regressionSecretPod(t, &core.Secret{Key: "dummy"})
	explicit := regressionSecurityContext(4000)
	empty := &corev1.SecurityContext{}
	pod.Spec.InitContainers = []corev1.Container{{Name: "existing-nil"}, {Name: "existing-explicit", SecurityContext: explicit.DeepCopy()}}
	result, _, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{contexts: []*corev1.SecurityContext{nil, explicit.DeepCopy(), empty}}).Mutate(context.Background(), pod)
	require.Nil(t, response)
	require.Len(t, result.Spec.InitContainers, 5)
	assert.Nil(t, result.Spec.InitContainers[0].SecurityContext)
	assert.Equal(t, explicit, result.Spec.InitContainers[1].SecurityContext)
	assert.Equal(t, result.Spec.Containers[0].SecurityContext, result.Spec.InitContainers[2].SecurityContext)
	assert.Equal(t, explicit, result.Spec.InitContainers[3].SecurityContext)
	assert.Equal(t, empty, result.Spec.InitContainers[4].SecurityContext)
}

func TestSecretsWebhook_SecurityContextUnchangedPaths(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*corev1.Pod)
	}{
		{"primary without context", func(p *corev1.Pod) { p.Spec.Containers[0].SecurityContext = nil }},
		{"no containers", func(p *corev1.Pod) { p.Spec.Containers = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pod := regressionSecretPod(t, &core.Secret{Key: "dummy"})
			tc.mutate(pod)
			result, changed, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{contexts: []*corev1.SecurityContext{nil}}).Mutate(context.Background(), pod)
			require.Nil(t, response)
			assert.True(t, changed)
			assert.Nil(t, result.Spec.InitContainers[0].SecurityContext)
		})
	}
	t.Run("no secrets", func(t *testing.T) {
		pod := regressionSecretPod(t)
		pod.Spec.InitContainers = []corev1.Container{{Name: "existing"}}
		before := pod.DeepCopy()
		result, changed, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{contexts: []*corev1.SecurityContext{nil}}).Mutate(context.Background(), pod)
		require.Nil(t, response)
		assert.False(t, changed)
		assert.Equal(t, before, result)
	})
	t.Run("injection failure", func(t *testing.T) {
		pod := regressionSecretPod(t, &core.Secret{Key: "dummy"})
		_, changed, response := regressionMutator(config.SecretManagerTypeK8s, regressionAppender{fail: true}).Mutate(context.Background(), pod)
		require.NotNil(t, response)
		assert.False(t, changed)
		assert.Contains(t, response.Result.Message, NotFoundAcrossAllScopesMsg)
	})
	t.Run("malformed annotation", func(t *testing.T) {
		pod := regressionSecretPod(t)
		pod.Annotations["flyte.secrets/s0"] = "invalid!"
		_, changed, response := (&SecretsPodMutator{}).Mutate(context.Background(), pod)
		require.NotNil(t, response)
		assert.False(t, changed)
		assert.EqualValues(t, 400, response.Result.Code)
	})
}
