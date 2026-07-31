package k8s

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/flyteorg/flyte/v2/app/internal/config"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	flytecoreapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// testScheme builds a runtime.Scheme with Knative and core types registered.
func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(s))
	require.NoError(t, servingv1.AddToScheme(s))
	return s
}

// AppNamespace is the namespace all test clients are configured with.
const AppNamespace = "flyte"

// testRevision builds a Knative Revision object with a given ActualReplicas count.
func testRevision(name, namespace string, actualReplicas int32) *servingv1.Revision {
	return &servingv1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: servingv1.RevisionStatus{
			ActualReplicas: &actualReplicas,
		},
	}
}

// testClient builds an AppK8sClient backed by a fake K8s client.
func testClient(t *testing.T, objs ...client.Object) *AppK8sClient {
	t.Helper()
	s := testScheme(t)
	fc := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(objs...).
		Build()
	cfg := &config.InternalAppConfig{
		DefaultRequestTimeout: 5 * time.Minute,
		MaxRequestTimeout:     time.Hour,
		WatchBufferSize:       100,
	}
	return NewAppK8sClient(fc, nil, AppNamespace, cfg)
}

// testApp builds a minimal flyteapp.App for use in tests.
func testApp(project, domain, name, image string) *flyteapp.App {
	return &flyteapp.App{
		Metadata: &flyteapp.Meta{
			Id: &flyteapp.Identifier{
				Project: project,
				Domain:  domain,
				Name:    name,
			},
		},
		Spec: &flyteapp.Spec{
			AppPayload: &flyteapp.Spec_Container{
				Container: &flytecoreapp.Container{
					Image: image,
				},
			},
		},
	}
}

// testKSvcName returns the KService name an app is created under. Tests derive it
// rather than hardcoding it so the fixtures follow the naming scheme.
func testKSvcName(project, domain, name string) string {
	return KServiceName(&flyteapp.Identifier{Project: project, Domain: domain, Name: name})
}

func TestDeploy_Create(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")

	err := c.Deploy(context.Background(), app)
	require.NoError(t, err)

	ksvc := &servingv1.Service{}
	err = c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc)
	require.NoError(t, err)
	assert.Equal(t, "proj", ksvc.Labels[labelProject])
	assert.Equal(t, "dev", ksvc.Labels[labelDomain])
	assert.Equal(t, "myapp", ksvc.Labels[labelAppName])
	assert.NotEmpty(t, ksvc.Annotations[annotationSpecSHA])
	assert.Equal(t, "proj/dev/myapp", ksvc.Annotations[annotationAppID])
}

func TestDeploy_InjectsInternalAppEndpointPattern(t *testing.T) {
	c := testClient(t)
	c.cfg.NamespacedNameSuffixTemplate = "{{ project }}-{{ domain }}"
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))

	envVars := ksvc.Spec.Template.Spec.Containers[0].Env
	var pattern string
	for _, e := range envVars {
		if e.Name == "INTERNAL_APP_ENDPOINT_PATTERN" {
			pattern = e.Value
			break
		}
	}
	assert.Equal(t, "http://{app_fqdn}-proj-dev.flyte.svc.cluster.local", pattern)
}

func TestDeploy_InjectsExecutionEnvVars(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))

	envVars := ksvc.Spec.Template.Spec.Containers[0].Env
	var gotProject, gotDomain string
	for _, e := range envVars {
		if e.Name == "FLYTE_INTERNAL_EXECUTION_PROJECT" {
			gotProject = e.Value
		}
		if e.Name == "FLYTE_INTERNAL_EXECUTION_DOMAIN" {
			gotDomain = e.Value
		}
	}
	assert.Equal(t, "proj", gotProject)
	assert.Equal(t, "dev", gotDomain)
}

func TestDeploy_InjectsSecretLabelsAndAnnotations(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	app.Spec.SecurityContext = &flyteapp.SecurityContext{
		Secrets: []*flytecoreapp.Secret{
			{Group: "my_group", Key: "my_key", MountRequirement: flytecoreapp.Secret_ENV_VAR},
		},
	}
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))

	tpl := ksvc.Spec.Template
	assert.Equal(t, "flyte", tpl.Labels["organization"])
	assert.Equal(t, "proj", tpl.Labels["project"])
	assert.Equal(t, "dev", tpl.Labels["domain"])
	assert.Equal(t, "true", tpl.Labels["inject-flyte-secrets"])
	assert.Contains(t, tpl.Annotations, "flyte.secrets/s0")
}

func TestDeploy_NoSecretLabelsOrAnnotationsWhenNoSecrets(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))

	tpl := ksvc.Spec.Template
	_, hasLabel := tpl.Labels["inject-flyte-secrets"]
	assert.False(t, hasLabel, "no inject-flyte-secrets label when no secrets configured")
	_, hasAnnotation := tpl.Annotations["flyte.secrets/s0"]
	assert.False(t, hasAnnotation, "no secret annotations when no secrets configured")
}

func TestDeploy_DefaultServiceAccount(t *testing.T) {
	c := testClient(t)
	c.cfg.DefaultServiceAccount = "flyte2"
	require.NoError(t, c.Deploy(context.Background(), testApp("proj", "dev", "myapp", "nginx:latest")))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "flyte2", ksvc.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploy_AppServiceAccountOverridesDefault(t *testing.T) {
	c := testClient(t)
	c.cfg.DefaultServiceAccount = "flyte2"
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	app.Spec.SecurityContext = &flyteapp.SecurityContext{
		RunAs: &flytecoreapp.Identity{K8SServiceAccount: "app-requested-sa"},
	}
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "app-requested-sa", ksvc.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploy_NoServiceAccountWhenUnset(t *testing.T) {
	c := testClient(t) // cfg.DefaultServiceAccount is empty
	require.NoError(t, c.Deploy(context.Background(), testApp("proj", "dev", "myapp", "nginx:latest")))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Empty(t, ksvc.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploy_UpdateOnSpecChange(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:1.0")
	require.NoError(t, c.Deploy(context.Background(), app))

	// Change image — spec SHA changes → update should happen.
	app.Spec.GetContainer().Image = "nginx:2.0"
	require.NoError(t, c.Deploy(context.Background(), app))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "nginx:2.0", ksvc.Spec.Template.Spec.Containers[0].Image)
}

func TestDeploy_SkipUpdateWhenUnchanged(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	// Get initial resource version.
	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	initialRV := ksvc.ResourceVersion

	// Deploy same spec — should be a no-op.
	require.NoError(t, c.Deploy(context.Background(), app))

	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, initialRV, ksvc.ResourceVersion, "resource version should not change on no-op deploy")
}

func TestDeploy_AfterStop_ClearsStoppedLabels(t *testing.T) {
	// Regression: Deploy() was skipping the update when the spec SHA was unchanged,
	// even if Stop() had marked the KService as stopped. Clicking "Start App" in the UI sends
	// the same spec, so the SHA matched and the app could never restart.
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "true", ksvc.Labels[labelAppStopped], "app-stopped label should be set after Stop")
	assert.Equal(t, visibilityClusterLocal, ksvc.Labels[labelKnativeVisibility], "service should be cluster-local after Stop")

	// Deploy same spec (as "Start App" would) — must not skip due to SHA match.
	require.NoError(t, c.Deploy(context.Background(), app))

	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	_, stopped := ksvc.Labels[labelAppStopped]
	assert.False(t, stopped, "app-stopped label must be cleared after Deploy following a Stop")
	_, visibility := ksvc.Labels[labelKnativeVisibility]
	assert.False(t, visibility, "visibility label must be cleared after Deploy following a Stop")
}

func TestStop(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "true", ksvc.Labels[labelAppStopped])
	assert.Equal(t, visibilityClusterLocal, ksvc.Labels[labelKnativeVisibility])
	assert.Equal(t, "0", ksvc.Spec.Template.Annotations["autoscaling.knative.dev/min-scale"])
	assert.Equal(t, "0", ksvc.Spec.Template.Annotations["autoscaling.knative.dev/initial-scale"])
}

func TestStop_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	// Should succeed silently — already gone.
	require.NoError(t, c.Stop(context.Background(), id))
}

func TestStop_DeletesLatestReadyRevision(t *testing.T) {
	// When a KService has a LatestReadyRevisionName, Stop() must delete that
	// Revision so its Deployment and pods are immediately terminated.
	// Updating the KService template alone is not sufficient — it does not immediately terminate existing pods.
	// for the autoscaler and does not kill running pods; they only scale down after
	// the stable window (~60s) with no traffic.
	s := testScheme(t)
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKSvcName("proj", "dev", "myapp"),
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "proj",
				labelDomain:     "dev",
				labelAppName:    "myapp",
			},
			Annotations: map[string]string{
				annotationAppID: "proj/dev/myapp",
			},
		},
	}
	ksvc.Status.LatestReadyRevisionName = "myapp-proj-dev-00001"

	rev := &servingv1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-proj-dev-00001",
			Namespace: AppNamespace,
		},
	}

	fc := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ksvc, rev).
		WithStatusSubresource(ksvc).
		Build()
	c := &AppK8sClient{
		k8sClient: fc,
		namespace: AppNamespace,
		cfg:       &config.InternalAppConfig{},
	}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	// KService must be marked stopped so Deploy can reliably clear the stopped state.
	gotKsvc := &servingv1.Service{}
	require.NoError(t, fc.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, gotKsvc))
	assert.Equal(t, "true", gotKsvc.Labels[labelAppStopped],
		"KService must carry the app-stopped label after Stop")
	assert.Equal(t, visibilityClusterLocal, gotKsvc.Labels[labelKnativeVisibility],
		"KService must be cluster-local after Stop")

	// LatestReadyRevision must be deleted so its pods are terminated immediately.
	gotRev := &servingv1.Revision{}
	err := fc.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev-00001", Namespace: AppNamespace}, gotRev)
	assert.True(t, k8serrors.IsNotFound(err), "LatestReadyRevision must be deleted after Stop")
}

func TestDelete(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Delete(context.Background(), id))

	ksvc := &servingv1.Service{}
	err := c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace}, ksvc)
	assert.True(t, k8serrors.IsNotFound(err))
}

func TestDelete_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	require.NoError(t, c.Delete(context.Background(), id))
}

func TestGetApp_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	app, err := c.GetApp(context.Background(), id)
	require.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))
	assert.Nil(t, app)
}

func TestGetApp_Stopped(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	result, err := c.GetApp(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, result.Status.Conditions, 1)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_STOPPED, result.Status.Conditions[0].DeploymentStatus)
}

func TestGetApp_CurrentReplicas(t *testing.T) {
	s := testScheme(t)
	// Pre-populate a KService with LatestReadyRevisionName already set in status,
	// and the corresponding Revision with ActualReplicas=4.
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testKSvcName("proj", "dev", "myapp"),
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "proj",
				labelDomain:     "dev",
				labelAppName:    "myapp",
			},
			Annotations: map[string]string{
				annotationAppID: "proj/dev/myapp",
			},
		},
	}
	ksvc.Status.LatestReadyRevisionName = "myapp-00001"

	rev := testRevision("myapp-00001", AppNamespace, 4)

	fc := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ksvc, rev).
		WithStatusSubresource(ksvc).
		Build()
	c := &AppK8sClient{
		k8sClient: fc,
		namespace: AppNamespace,
		cfg:       &config.InternalAppConfig{},
	}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	result, err := c.GetApp(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, uint32(4), result.Status.CurrentReplicas)
}

func TestGetApp_SpecRoundTrip(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	app.Spec.Profile = &flyteapp.Profile{
		Type:             "FastAPI",
		Name:             "My App",
		ShortDescription: "A test app",
	}
	app.Spec.Autoscaling = &flyteapp.AutoscalingConfig{
		Replicas: &flyteapp.Replicas{Min: 1, Max: 5},
	}
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	result, err := c.GetApp(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, result.Spec)
	assert.Equal(t, "FastAPI", result.Spec.Profile.Type)
	assert.Equal(t, "My App", result.Spec.Profile.Name)
	assert.Equal(t, uint32(1), result.Spec.Autoscaling.Replicas.Min)
	assert.Equal(t, uint32(5), result.Spec.Autoscaling.Replicas.Max)
}

func TestList(t *testing.T) {
	s := testScheme(t)
	// Pre-populate two KServices with different project labels.
	ksvc1 := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app1",
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "proj",
				labelDomain:     "dev",
				labelAppName:    "app1",
			},
			Annotations: map[string]string{
				annotationAppID: "proj/dev/app1",
			},
		},
	}
	ksvc2 := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app2",
			Namespace: "other-dev",
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "other",
				labelDomain:     "dev",
				labelAppName:    "app2",
			},
			Annotations: map[string]string{
				annotationAppID: "other/dev/app2",
			},
		},
	}

	fc := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ksvc1, ksvc2).
		Build()
	c := &AppK8sClient{
		k8sClient: fc,
		namespace: AppNamespace,
		cfg: &config.InternalAppConfig{
			DefaultRequestTimeout: 5 * time.Minute,
			MaxRequestTimeout:     time.Hour,
		},
	}

	apps, nextToken, err := c.List(context.Background(), "proj", "dev", 0, "")
	require.NoError(t, err)
	assert.Empty(t, nextToken)
	require.Len(t, apps, 1)
	assert.Equal(t, "proj", apps[0].Metadata.Id.Project)
	assert.Equal(t, "app1", apps[0].Metadata.Id.Name)
}

func TestGetReplicas(t *testing.T) {
	s := testScheme(t)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-abc",
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelKnativeService: testKSvcName("proj", "dev", "myapp"),
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{Ready: true},
			},
		},
	}
	fc := fake.NewClientBuilder().WithScheme(s).WithObjects(pod).Build()
	c := &AppK8sClient{
		k8sClient: fc,
		namespace: AppNamespace,
		cfg:       &config.InternalAppConfig{},
	}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	replicas, err := c.GetReplicas(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, replicas, 1)
	assert.Equal(t, "myapp-abc", replicas[0].Metadata.Id.Name)
	assert.Equal(t, "ACTIVE", replicas[0].Status.DeploymentStatus)
}

func TestGetReplicas_FiltersToLatestRevision(t *testing.T) {
	s := testScheme(t)
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: testKSvcName("proj", "dev", "myapp"), Namespace: AppNamespace},
		Status: servingv1.ServiceStatus{
			ConfigurationStatusFields: servingv1.ConfigurationStatusFields{
				LatestReadyRevisionName: "myapp-00002",
			},
		},
	}
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-new",
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelKnativeService:  testKSvcName("proj", "dev", "myapp"),
				labelKnativeRevision: "myapp-00002",
			},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, ContainerStatuses: []corev1.ContainerStatus{{Ready: true}}},
	}
	oldPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-old",
			Namespace: AppNamespace,
			Labels: map[string]string{
				labelKnativeService:  testKSvcName("proj", "dev", "myapp"),
				labelKnativeRevision: "myapp-00001",
			},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	fc := fake.NewClientBuilder().WithScheme(s).WithObjects(ksvc, newPod, oldPod).Build()
	c := &AppK8sClient{k8sClient: fc, namespace: AppNamespace, cfg: &config.InternalAppConfig{}}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	replicas, err := c.GetReplicas(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, replicas, 1)
	assert.Equal(t, "myapp-new", replicas[0].Metadata.Id.Name)
}

func TestDeleteReplica(t *testing.T) {
	s := testScheme(t)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-abc",
			Namespace: AppNamespace,
		},
	}
	fc := fake.NewClientBuilder().WithScheme(s).WithObjects(pod).Build()
	c := &AppK8sClient{
		k8sClient: fc,
		namespace: AppNamespace,
		cfg:       &config.InternalAppConfig{},
	}

	replicaID := &flyteapp.ReplicaIdentifier{
		AppId: &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"},
		Name:  "myapp-abc",
	}
	require.NoError(t, c.DeleteReplica(context.Background(), replicaID))

	err := fc.Get(context.Background(),
		client.ObjectKey{Name: "myapp-abc", Namespace: AppNamespace}, &corev1.Pod{})
	assert.True(t, k8serrors.IsNotFound(err))
}

func TestHandleKServiceEvent(t *testing.T) {
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp",
			Namespace: AppNamespace,
			Annotations: map[string]string{
				annotationAppID: "proj/dev/myapp",
			},
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "proj",
				labelDomain:     "dev",
				labelAppName:    "myapp",
			},
		},
	}

	tests := []struct {
		eventType    k8swatch.EventType
		wantEventKey string
	}{
		{k8swatch.Added, "create"},
		{k8swatch.Modified, "update"},
		{k8swatch.Deleted, "delete"},
	}

	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			c := testClient(t)
			ch := c.Subscribe("myapp")
			c.handleKServiceEvent(context.Background(), ksvc, tt.eventType)

			select {
			case resp := <-ch:
				require.NotNil(t, resp)
				switch tt.wantEventKey {
				case "create":
					assert.NotNil(t, resp.GetCreateEvent())
					assert.Equal(t, "proj", resp.GetCreateEvent().App.Metadata.Id.Project)
				case "update":
					assert.NotNil(t, resp.GetUpdateEvent())
					assert.Equal(t, "myapp", resp.GetUpdateEvent().UpdatedApp.Metadata.Id.Name)
				case "delete":
					assert.NotNil(t, resp.GetDeleteEvent())
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("expected event not received")
			}
		})
	}
}

func TestKServiceName(t *testing.T) {
	tests := []struct {
		desc string
		id   *flyteapp.Identifier
		want string
	}{
		{
			desc: "standard identifier",
			id:   &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"},
			want: "k-myapp-nyzyrk24yeye56darerfyox46y",
		},
		{
			// Case is part of the identity — nothing else in Flyte folds it — so this
			// must not collide with the lowercase identifier above. Only the readable
			// prefix is lowercased, because a DNS label has to be.
			desc: "case-variant identity hashes distinctly",
			id:   &flyteapp.Identifier{Project: "PROJ", Domain: "Dev", Name: "MyApp"},
			want: "k-myapp-eaq2xqk7hypngwmksa3d7vfyk4",
		},
		{
			desc: "app names differing only by suffix stay distinct",
			id:   &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "my-long-service-name-v1"},
			want: "k-my-long-service-name-v1-lx7ylbnzddt2iphutzu2zfo62a",
		},
		{
			desc: "long app name is truncated to fit the DNS label limit",
			id: &flyteapp.Identifier{
				Project: "proj",
				Domain:  "dev",
				Name:    "this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit",
			},
			want: "k-this-is-a-very-long-app-name-that-6z6rtjpjidce5tijzvreclx75e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := KServiceName(tt.id)
			assert.Equal(t, tt.want, got)
			assert.LessOrEqual(t, len(got), maxKServiceNameLen)
			// Must be a valid DNS-1035 label — leading letter, lowercase
			// alphanumerics and hyphens, trailing alphanumeric — and end in the
			// digest, with the readable prefix optional.
			assert.Regexp(t, `^k-([a-z0-9]([-a-z0-9]*[a-z0-9])?-)?[a-z2-7]{26}$`, got)
		})
	}
}

// TestKServiceName_DistinguishesAmbiguousIdentities is the regression test for
// issue #7622. The previous derivation joined the fields with "-", so these two
// identities both flattened to "svc-team-prod-x" and the second app to deploy
// would overwrite the first app's spec.
func TestKServiceName_DistinguishesAmbiguousIdentities(t *testing.T) {
	first := &flyteapp.Identifier{Name: "svc", Project: "team", Domain: "prod-x"}
	second := &flyteapp.Identifier{Name: "svc", Project: "team-prod", Domain: "x"}

	assert.NotEqual(t, KServiceName(first), KServiceName(second))
}

// TestKServiceName_TrimsHyphensAroundAppName covers app names that the proto
// pattern forbids but nothing currently enforces. The readable segment is tidied so
// the name never contains "--", and the result stays a valid DNS-1035 label.
func TestKServiceName_TrimsHyphensAroundAppName(t *testing.T) {
	// The label rule Knative applies: leading letter, lowercase alphanumerics and
	// hyphens, trailing alphanumeric.
	dns1035 := regexp.MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

	names := []string{"myapp", "-foo", "foo-", "-foo-", "---", "-", "",
		// 35+ characters, so the truncation path runs before the trim.
		"-----------------------------------x",
		"this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit"}

	seen := make(map[string]string, len(names))
	for _, name := range names {
		got := KServiceName(&flyteapp.Identifier{Project: "proj", Domain: "dev", Name: name})

		assert.Regexp(t, dns1035, got, "name %q produced an invalid DNS-1035 label", name)
		assert.NotContains(t, got, "--", "name %q produced a doubled hyphen", name)
		assert.LessOrEqual(t, len(got), maxKServiceNameLen)
		// Trimming must not merge distinct identities — the digest keeps them apart.
		if prev, dup := seen[got]; dup {
			t.Errorf("names %q and %q collided on %s", prev, name, got)
		}
		seen[got] = name
	}
}

func TestKServiceName_OrgIsPartOfIdentity(t *testing.T) {
	base := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	otherOrg := &flyteapp.Identifier{Org: "acme", Project: "proj", Domain: "dev", Name: "myapp"}

	assert.NotEqual(t, KServiceName(base), KServiceName(otherOrg),
		"apps in different orgs must not share a KService")

	// An unset org and an explicit default org are the same app: Get returns the
	// default, so a client round-tripping an app must not land on a second KService.
	defaultedOrg := &flyteapp.Identifier{Org: defaultOrg, Project: "proj", Domain: "dev", Name: "myapp"}
	assert.Equal(t, KServiceName(base), KServiceName(defaultedOrg))
}

func TestIdentifierFromKService(t *testing.T) {
	ksvc := func(labels, annotations map[string]string) *servingv1.Service {
		return &servingv1.Service{ObjectMeta: metav1.ObjectMeta{
			Name: "k-myapp-abc", Labels: labels, Annotations: annotations,
		}}
	}

	tests := []struct {
		desc        string
		labels      map[string]string
		annotations map[string]string
		want        *flyteapp.Identifier
		wantErrOn   string
	}{
		{
			desc:   "unset org reads back as the default org",
			labels: map[string]string{labelProject: "proj", labelDomain: "dev", labelAppName: "myapp"},
			want:   &flyteapp.Identifier{Org: defaultOrg, Project: "proj", Domain: "dev", Name: "myapp"},
		},
		{
			desc:        "org annotation is honored",
			labels:      map[string]string{labelProject: "proj", labelDomain: "dev", labelAppName: "myapp"},
			annotations: map[string]string{annotationAppOrg: "acme"},
			want:        &flyteapp.Identifier{Org: "acme", Project: "proj", Domain: "dev", Name: "myapp"},
		},
		{
			// Labels store the identity verbatim, and case is part of it, so an app
			// round-trips as exactly the identity it was deployed under.
			desc:   "case is preserved",
			labels: map[string]string{labelProject: "Proj", labelDomain: "Dev", labelAppName: "MyApp"},
			want:   &flyteapp.Identifier{Org: defaultOrg, Project: "Proj", Domain: "Dev", Name: "MyApp"},
		},
		{
			// Nothing validates app names today, so a blank one deploys and has to
			// read back. Present-but-empty is not the same as absent.
			desc:   "empty label value is a valid identity",
			labels: map[string]string{labelProject: "proj", labelDomain: "dev", labelAppName: ""},
			want:   &flyteapp.Identifier{Org: defaultOrg, Project: "proj", Domain: "dev", Name: ""},
		},
		{
			desc:      "missing project label is not an identity",
			labels:    map[string]string{labelDomain: "dev", labelAppName: "myapp"},
			wantErrOn: labelProject,
		},
		{
			desc:      "missing domain label is not an identity",
			labels:    map[string]string{labelProject: "proj", labelAppName: "myapp"},
			wantErrOn: labelDomain,
		},
		{
			desc:      "missing app-name label is not an identity",
			labels:    map[string]string{labelProject: "proj", labelDomain: "dev"},
			wantErrOn: labelAppName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := identifierFromKService(ksvc(tt.labels, tt.annotations))
			if tt.wantErrOn != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrOn)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want.GetOrg(), got.GetOrg())
			assert.Equal(t, tt.want.GetProject(), got.GetProject())
			assert.Equal(t, tt.want.GetDomain(), got.GetDomain())
			assert.Equal(t, tt.want.GetName(), got.GetName())
		})
	}
}

// The reported URL and the KService's own name must derive from the same identity.
// Any path that rebuilds the identity without the org computes a host for a
// different app, and the link resolves to nothing.
func TestGetApp_IngressUsesAppOrg(t *testing.T) {
	c := testClient(t)
	c.cfg.BaseDomain = "example.com"
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	app.Metadata.Id.Org = "acme"
	require.NoError(t, c.Deploy(context.Background(), app))

	got, err := c.GetApp(context.Background(), app.GetMetadata().GetId())
	require.NoError(t, err)

	assert.Equal(t, "https://"+KServiceName(app.GetMetadata().GetId())+".example.com",
		got.GetStatus().GetIngress().GetPublicUrl(),
		"the reported URL must address the KService the app actually lives in")
}

// TestKServiceToApp_RoundTripsIdentity pins the full loop: the identity written by
// Deploy is the identity Get reads back, org included.
func TestKServiceToApp_RoundTripsIdentity(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	app.Metadata.Id.Org = "acme"
	require.NoError(t, c.Deploy(context.Background(), app))

	got, err := c.GetApp(context.Background(), app.GetMetadata().GetId())
	require.NoError(t, err)

	gotID := got.GetMetadata().GetId()
	assert.Equal(t, "acme", gotID.GetOrg())
	assert.Equal(t, "proj", gotID.GetProject())
	assert.Equal(t, "dev", gotID.GetDomain())
	assert.Equal(t, "myapp", gotID.GetName())
	assert.Equal(t, KServiceName(app.GetMetadata().GetId()), KServiceName(gotID),
		"the round-tripped identity must resolve to the same KService")
}

func TestRenderNamespacedSuffix(t *testing.T) {
	tests := []struct {
		tmpl    string
		project string
		domain  string
		want    string
	}{
		{"{{ project }}-{{ domain }}", "myproject", "dev", "myproject-dev"},
		{"{{ project }}-{{ domain }}", "MyProject", "Dev", "myproject-dev"},
		{"{{ project }}-{{ domain }}", "proj", "prod", "proj-prod"},
		{"custom-{{ domain }}", "proj", "dev", "custom-dev"},
		{"", "proj", "dev", ""},
	}
	for _, tt := range tests {
		got := renderNamespacedSuffix(tt.tmpl, tt.project, tt.domain)
		assert.Equal(t, tt.want, got)
	}
}

func TestPodDeploymentStatus(t *testing.T) {
	tests := []struct {
		name       string
		pod        corev1.Pod
		wantStatus string
		wantReason string
	}{
		{
			name: "running and ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase:             corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{{Ready: true}},
				},
			},
			wantStatus: "ACTIVE",
		},
		{
			name: "running but container not ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: false, State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"},
						}},
					},
				},
			},
			wantStatus: "DEPLOYING",
			wantReason: "ContainerCreating",
		},
		{
			name: "pending with waiting reason",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{Reason: "ImagePullBackOff"},
						}},
					},
				},
			},
			wantStatus: "PENDING",
			wantReason: "ImagePullBackOff",
		},
		{
			name: "failed",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase:  corev1.PodFailed,
					Reason: "OOMKilled",
				},
			},
			wantStatus: "FAILED",
			wantReason: "OOMKilled",
		},
		{
			name: "succeeded",
			pod: corev1.Pod{
				Status: corev1.PodStatus{Phase: corev1.PodSucceeded},
			},
			wantStatus: "STOPPED",
			wantReason: "pod completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, reason := podDeploymentStatus(&tt.pod)
			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantReason, reason)
		})
	}
}

// --- Informer subscribe/unsubscribe tests ---

func TestSubscribe_ReceivesEvent(t *testing.T) {
	c := testClient(t)
	ch := c.Subscribe("myapp")
	defer c.Unsubscribe("myapp", ch)

	ksvc := testKsvc("myapp", AppNamespace, "100")
	c.handleKServiceEvent(context.Background(), ksvc, k8swatch.Added)

	select {
	case resp := <-ch:
		require.NotNil(t, resp.GetCreateEvent())
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected event not received")
	}
}

func TestSubscribe_AppSpecificDoesNotReceiveOtherApps(t *testing.T) {
	c := testClient(t)
	ch := c.Subscribe("app1")
	defer c.Unsubscribe("app1", ch)

	// Event for app2 should not be delivered to app1 subscriber.
	c.handleKServiceEvent(context.Background(), testKsvc("app2", AppNamespace, "1"), k8swatch.Added)

	select {
	case <-ch:
		t.Fatal("received unexpected event for a different app")
	case <-time.After(30 * time.Millisecond):
		// Correct: no event delivered.
	}
}

func TestUnsubscribe_ClosesChannel(t *testing.T) {
	c := testClient(t)
	ch := c.Subscribe("myapp")
	c.Unsubscribe("myapp", ch)

	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after Unsubscribe")
}

func TestSubscribe_MultipleSubscribers(t *testing.T) {
	c := testClient(t)
	ch1 := c.Subscribe("myapp")
	ch2 := c.Subscribe("myapp")
	defer c.Unsubscribe("myapp", ch1)
	defer c.Unsubscribe("myapp", ch2)

	c.handleKServiceEvent(context.Background(), testKsvc("myapp", AppNamespace, "1"), k8swatch.Added)

	for _, ch := range []chan *flyteapp.WatchResponse{ch1, ch2} {
		select {
		case resp := <-ch:
			require.NotNil(t, resp.GetCreateEvent())
		case <-time.After(100 * time.Millisecond):
			t.Fatal("expected event not received by subscriber")
		}
	}
}

// testKsvc builds a minimal KService that kserviceToApp can parse.
func testKsvc(name, ns, rv string) *servingv1.Service {
	return &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       ns,
			ResourceVersion: rv,
			Annotations:     map[string]string{annotationAppID: "proj/dev/" + name},
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    "proj",
				labelDomain:     "dev",
				labelAppName:    name,
			},
		},
	}
}

// --- Status message format tests ---

func TestKServiceToApp_StoppedDesiredState(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")

	require.NoError(t, c.Deploy(context.Background(), app))
	require.NoError(t, c.Stop(context.Background(), app.Metadata.Id))

	got, err := c.GetApp(context.Background(), app.Metadata.Id)
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Spec_DESIRED_STATE_STOPPED, got.GetSpec().GetDesiredState(),
		"stopped app should have DesiredState=STOPPED in the returned spec")
}

func TestKServiceToStatus_Messages(t *testing.T) {
	tests := []struct {
		name           string
		ksvc           func() *servingv1.Service
		wantConditions []struct {
			phase   flyteapp.Status_DeploymentStatus
			message string
		}
	}{
		{
			name: "active — single Ready=True condition",
			ksvc: func() *servingv1.Service {
				ksvc := testKsvc("myapp", AppNamespace, "1")
				ksvc.Status.Status = duckv1.Status{
					Conditions: duckv1.Conditions{{
						Type:   servingv1.ServiceConditionReady,
						Status: corev1.ConditionTrue,
					}},
				}
				return ksvc
			},
			wantConditions: []struct {
				phase   flyteapp.Status_DeploymentStatus
				message string
			}{
				{flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, "Service is ready"},
			},
		},
		{
			name: "active — all three sub-conditions True",
			ksvc: func() *servingv1.Service {
				ksvc := testKsvc("myapp", AppNamespace, "1")
				ksvc.Status.Status = duckv1.Status{
					Conditions: duckv1.Conditions{
						{Type: servingv1.ServiceConditionConfigurationsReady, Status: corev1.ConditionTrue},
						{Type: servingv1.ServiceConditionRoutesReady, Status: corev1.ConditionTrue},
						{Type: servingv1.ServiceConditionReady, Status: corev1.ConditionTrue},
					},
				}
				return ksvc
			},
			wantConditions: []struct {
				phase   flyteapp.Status_DeploymentStatus
				message string
			}{
				{flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, "Configuration is ready"},
				{flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, "Routes are ready"},
				{flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, "Service is ready"},
			},
		},
		{
			name: "deploying — RoutesReady=Unknown skipped, only Ready=Unknown emitted",
			ksvc: func() *servingv1.Service {
				ksvc := testKsvc("myapp", AppNamespace, "1")
				ksvc.Status.LatestCreatedRevisionName = "myapp-00002"
				ksvc.Status.LatestReadyRevisionName = "myapp-00001"
				ksvc.Status.MarkRouteNotYetReady()
				return ksvc
			},
			wantConditions: []struct {
				phase   flyteapp.Status_DeploymentStatus
				message string
			}{
				{flyteapp.Status_DEPLOYMENT_STATUS_PENDING, "TrafficNotMigrated: Traffic is not yet migrated to the latest revision."},
			},
		},
		{
			name: "stopped",
			ksvc: func() *servingv1.Service {
				ksvc := testKsvc("myapp", AppNamespace, "1")
				if ksvc.Labels == nil {
					ksvc.Labels = map[string]string{}
				}
				ksvc.Labels[labelAppStopped] = "true"
				return ksvc
			},
			wantConditions: []struct {
				phase   flyteapp.Status_DeploymentStatus
				message string
			}{
				{flyteapp.Status_DEPLOYMENT_STATUS_STOPPED, "App scaled to zero"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := testClient(t)
			status := c.kServiceToStatus(context.Background(), tt.ksvc(),
				&flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"})
			require.NotNil(t, status)
			require.Len(t, status.Conditions, len(tt.wantConditions))
			for i, want := range tt.wantConditions {
				assert.Equal(t, want.phase, status.Conditions[i].DeploymentStatus, "condition[%d] phase", i)
				assert.Equal(t, want.message, status.Conditions[i].Message, "condition[%d] message", i)
			}
		})
	}
}

// transformStructToStructPB converts an arbitrary Go object into a *structpb.Struct
// by round-tripping through JSON. It fails the test on any marshaling error.
func transformStructToStructPB(t *testing.T, obj interface{}) *structpb.Struct {
	t.Helper()
	data, err := json.Marshal(obj)
	require.NoError(t, err)
	m := make(map[string]interface{})
	err = json.Unmarshal(data, &m)
	require.NoError(t, err)
	s, err := structpb.NewStruct(m)
	require.NoError(t, err)
	return s
}

func TestBuildPodSpec_Container(t *testing.T) {
	spec := &flyteapp.Spec{
		AppPayload: &flyteapp.Spec_Container{
			Container: &flytecoreapp.Container{
				Image:   "nginx:latest",
				Command: []string{"nginx"},
				Args:    []string{"-g", "daemon off;"},
				Env: []*flytecoreapp.KeyValuePair{
					{Key: "FOO", Value: "bar"},
				},
				Ports: []*flytecoreapp.ContainerPort{
					{ContainerPort: 8080, Name: "http"},
				},
				Resources: &flytecoreapp.Resources{
					Requests: []*flytecoreapp.Resources_ResourceEntry{
						{Name: flytecoreapp.Resources_CPU, Value: "100m"},
						{Name: flytecoreapp.Resources_MEMORY, Value: "128Mi"},
					},
				},
			},
		},
	}

	podSpec, err := buildPodSpec(spec)
	require.NoError(t, err)
	require.Len(t, podSpec.Containers, 1)
	assert.Equal(t, "app", podSpec.Containers[0].Name)
	assert.Equal(t, "nginx:latest", podSpec.Containers[0].Image)
	assert.Equal(t, []string{"nginx"}, podSpec.Containers[0].Command)
	assert.Equal(t, []string{"-g", "daemon off;"}, podSpec.Containers[0].Args)
	assert.Equal(t, []corev1.EnvVar{{Name: "FOO", Value: "bar"}}, podSpec.Containers[0].Env)
	assert.Equal(t, []corev1.ContainerPort{{ContainerPort: 8080, Name: "http"}}, podSpec.Containers[0].Ports)
	assert.Equal(t, "100m", podSpec.Containers[0].Resources.Requests.Cpu().String())
	assert.Equal(t, "128Mi", podSpec.Containers[0].Resources.Requests.Memory().String())
	assert.NotNil(t, podSpec.EnableServiceLinks)
	assert.False(t, *podSpec.EnableServiceLinks)
}

func TestBuildPodSpec_Pod(t *testing.T) {
	podSpecMap := map[string]interface{}{
		"containers": []map[string]interface{}{
			{
				"name":  "app",
				"image": "my-image:v1",
				"ports": []map[string]interface{}{
					{"containerPort": float64(80), "name": "http"},
				},
				"resources": map[string]interface{}{
					"requests": map[string]interface{}{
						"cpu":    "250m",
						"memory": "256Mi",
					},
				},
			},
		},
		"restartPolicy": "Always",
	}

	spec := &flyteapp.Spec{
		AppPayload: &flyteapp.Spec_Pod{
			Pod: &flytecoreapp.K8SPod{
				PodSpec: transformStructToStructPB(t, podSpecMap),
			},
		},
	}

	podSpec, err := buildPodSpec(spec)
	require.NoError(t, err)
	require.Len(t, podSpec.Containers, 1)
	assert.Equal(t, "app", podSpec.Containers[0].Name)
	assert.Equal(t, "my-image:v1", podSpec.Containers[0].Image)
	assert.Len(t, podSpec.Containers[0].Ports, 1)
	assert.Equal(t, int32(80), podSpec.Containers[0].Ports[0].ContainerPort)
	assert.Equal(t, "250m", podSpec.Containers[0].Resources.Requests.Cpu().String())
	assert.Equal(t, "256Mi", podSpec.Containers[0].Resources.Requests.Memory().String())
	assert.Equal(t, corev1.RestartPolicyAlways, podSpec.RestartPolicy)
	assert.NotNil(t, podSpec.EnableServiceLinks)
	assert.False(t, *podSpec.EnableServiceLinks)
}

func TestBuildPodSpec_Pod_NilPodSpec(t *testing.T) {
	spec := &flyteapp.Spec{
		AppPayload: &flyteapp.Spec_Pod{
			Pod: &flytecoreapp.K8SPod{},
		},
	}

	_, err := buildPodSpec(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "has no pod spec")
}

func TestBuildPodSpec_Pod_InvalidJSON(t *testing.T) {
	// Create a Struct that cannot be unmarshaled into corev1.PodSpec.
	// "containers" must be an array, not a string.
	s := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"containers": {Kind: &structpb.Value_StringValue{StringValue: "not-an-array"}},
		},
	}

	spec := &flyteapp.Spec{
		AppPayload: &flyteapp.Spec_Pod{
			Pod: &flytecoreapp.K8SPod{
				PodSpec: s,
			},
		},
	}

	_, err := buildPodSpec(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal K8sPod spec")
}

func TestBuildPodSpec_NoPayload(t *testing.T) {
	spec := &flyteapp.Spec{}
	_, err := buildPodSpec(spec)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app spec has no payload")
}
