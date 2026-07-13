package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

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

func TestDeploy_Create(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")

	err := c.Deploy(context.Background(), app)
	require.NoError(t, err)

	ksvc := &servingv1.Service{}
	err = c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc)
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))

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

func TestDeploy_DefaultServiceAccount(t *testing.T) {
	c := testClient(t)
	c.cfg.DefaultServiceAccount = "flyte2"
	require.NoError(t, c.Deploy(context.Background(), testApp("proj", "dev", "myapp", "nginx:latest")))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "app-requested-sa", ksvc.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploy_NoServiceAccountWhenUnset(t *testing.T) {
	c := testClient(t) // cfg.DefaultServiceAccount is empty
	require.NoError(t, c.Deploy(context.Background(), testApp("proj", "dev", "myapp", "nginx:latest")))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "nginx:2.0", ksvc.Spec.Template.Spec.Containers[0].Image)
}

func TestDeploy_SkipUpdateWhenUnchanged(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	// Get initial resource version.
	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
	initialRV := ksvc.ResourceVersion

	// Deploy same spec — should be a no-op.
	require.NoError(t, c.Deploy(context.Background(), app))

	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
	assert.Equal(t, "true", ksvc.Labels[labelAppStopped], "app-stopped label should be set after Stop")
	assert.Equal(t, visibilityClusterLocal, ksvc.Labels[labelKnativeVisibility], "service should be cluster-local after Stop")

	// Deploy same spec (as "Start App" would) — must not skip due to SHA match.
	require.NoError(t, c.Deploy(context.Background(), app))

	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc))
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
			Name:      "myapp-proj-dev",
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, gotKsvc))
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
		client.ObjectKey{Name: "myapp-proj-dev", Namespace: AppNamespace}, ksvc)
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
			Name:      "myapp-proj-dev",
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
				labelKnativeService: "myapp-proj-dev",
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
		ObjectMeta: metav1.ObjectMeta{Name: "myapp-proj-dev", Namespace: AppNamespace},
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
				labelKnativeService:  "myapp-proj-dev",
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
				labelKnativeService:  "myapp-proj-dev",
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
			Labels: map[string]string{labelAppManaged: "true"},
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
		name string
		want string
	}{
		{"myapp", "myapp-proj-dev"},
		{"MyApp", "myapp-proj-dev"},
		{"my-long-service-name-v1", "my-long-service-name-v1-proj-dev"},
		{"my-long-service-name-v2", "my-long-service-name-v2-proj-dev"},
		// Names whose {name}-{project}-{domain} exceeds 63 chars get a hash
		// suffix instead of blind truncation.
		{
			"this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit",
			func() string {
				raw := "this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit-proj-dev"
				sum := sha256.Sum256([]byte("proj/dev/this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit"))
				return raw[:54] + "-" + hex.EncodeToString(sum[:4])
			}(),
		},
	}
	for _, tt := range tests {
		id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: tt.name}
		got := KServiceName(id)
		assert.Equal(t, tt.want, got)
		assert.LessOrEqual(t, len(got), maxKServiceNameLen)
	}
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
			Labels:          map[string]string{labelAppManaged: "true"},
		},
	}
}

// --- Status message format tests ---

func TestKserviceToApp_StoppedDesiredState(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")

	require.NoError(t, c.Deploy(context.Background(), app))
	require.NoError(t, c.Stop(context.Background(), app.Metadata.Id))

	got, err := c.GetApp(context.Background(), app.Metadata.Id)
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Spec_DESIRED_STATE_STOPPED, got.GetSpec().GetDesiredState(),
		"stopped app should have DesiredState=STOPPED in the returned spec")
}

func TestKserviceToStatus_Messages(t *testing.T) {
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
			status := c.kserviceToStatus(context.Background(), tt.ksvc())
			require.NotNil(t, status)
			require.Len(t, status.Conditions, len(tt.wantConditions))
			for i, want := range tt.wantConditions {
				assert.Equal(t, want.phase, status.Conditions[i].DeploymentStatus, "condition[%d] phase", i)
				assert.Equal(t, want.message, status.Conditions[i].Message, "condition[%d] message", i)
			}
		})
	}
}
