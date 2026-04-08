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
	return &AppK8sClient{
		k8sClient: fc,
		cfg: &config.InternalAppConfig{
			DefaultRequestTimeout: 5 * time.Minute,
			MaxRequestTimeout:     time.Hour,
		},
	}
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
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc)
	require.NoError(t, err)
	assert.Equal(t, "proj", ksvc.Labels[labelProject])
	assert.Equal(t, "dev", ksvc.Labels[labelDomain])
	assert.Equal(t, "myapp", ksvc.Labels[labelAppName])
	assert.NotEmpty(t, ksvc.Annotations[annotationSpecSHA])
	assert.Equal(t, "proj/dev/myapp", ksvc.Annotations[annotationAppID])
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
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc))
	assert.Equal(t, "nginx:2.0", ksvc.Spec.Template.Spec.Containers[0].Image)
}

func TestDeploy_SkipUpdateWhenUnchanged(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	// Get initial resource version.
	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc))
	initialRV := ksvc.ResourceVersion

	// Deploy same spec — should be a no-op.
	require.NoError(t, c.Deploy(context.Background(), app))

	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc))
	assert.Equal(t, initialRV, ksvc.ResourceVersion, "resource version should not change on no-op deploy")
}

func TestStop(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	ksvc := &servingv1.Service{}
	require.NoError(t, c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc))
	assert.Equal(t, "0", ksvc.Spec.Template.Annotations["autoscaling.knative.dev/max-scale"])
}

func TestStop_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	// Should succeed silently — already gone.
	require.NoError(t, c.Stop(context.Background(), id))
}

func TestDelete(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Delete(context.Background(), id))

	ksvc := &servingv1.Service{}
	err := c.k8sClient.Get(context.Background(),
		client.ObjectKey{Name: "myapp", Namespace: "proj-dev"}, ksvc)
	assert.True(t, k8serrors.IsNotFound(err))
}

func TestDelete_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	require.NoError(t, c.Delete(context.Background(), id))
}

func TestGetStatus_NotFound(t *testing.T) {
	c := testClient(t)
	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "missing"}
	status, err := c.GetStatus(context.Background(), id)
	require.Error(t, err)
	assert.True(t, k8serrors.IsNotFound(err))
	assert.Nil(t, status)
}

func TestGetStatus_Stopped(t *testing.T) {
	c := testClient(t)
	app := testApp("proj", "dev", "myapp", "nginx:latest")
	require.NoError(t, c.Deploy(context.Background(), app))

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	require.NoError(t, c.Stop(context.Background(), id))

	status, err := c.GetStatus(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, status.Conditions, 1)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_STOPPED, status.Conditions[0].DeploymentStatus)
}

func TestGetStatus_CurrentReplicas(t *testing.T) {
	s := testScheme(t)
	// Pre-populate a KService with LatestReadyRevisionName already set in status,
	// and the corresponding Revision with ActualReplicas=4.
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp",
			Namespace: "proj-dev",
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

	rev := testRevision("myapp-00001", "proj-dev", 4)

	fc := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(ksvc, rev).
		WithStatusSubresource(ksvc).
		Build()
	c := &AppK8sClient{
		k8sClient: fc,
		cfg:       &config.InternalAppConfig{},
	}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	status, err := c.GetStatus(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, uint32(4), status.CurrentReplicas)
}

func TestList(t *testing.T) {
	s := testScheme(t)
	// Pre-populate two KServices with different project labels.
	ksvc1 := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app1",
			Namespace: "proj-dev",
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
			Namespace: "proj-dev",
			Labels: map[string]string{
				labelAppName: "myapp",
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
		cfg:       &config.InternalAppConfig{},
	}

	id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
	replicas, err := c.GetReplicas(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, replicas, 1)
	assert.Equal(t, "myapp-abc", replicas[0].Metadata.Id.Name)
	assert.Equal(t, "ACTIVE", replicas[0].Status.DeploymentStatus)
}

func TestDeleteReplica(t *testing.T) {
	s := testScheme(t)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp-abc",
			Namespace: "proj-dev",
		},
	}
	fc := fake.NewClientBuilder().WithScheme(s).WithObjects(pod).Build()
	c := &AppK8sClient{
		k8sClient: fc,
		cfg:       &config.InternalAppConfig{},
	}

	replicaID := &flyteapp.ReplicaIdentifier{
		AppId: &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"},
		Name:  "myapp-abc",
	}
	require.NoError(t, c.DeleteReplica(context.Background(), replicaID))

	err := fc.Get(context.Background(),
		client.ObjectKey{Name: "myapp-abc", Namespace: "proj-dev"}, &corev1.Pod{})
	assert.True(t, k8serrors.IsNotFound(err))
}

func TestKserviceEventToWatchResponse(t *testing.T) {
	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "myapp",
			Namespace: "proj-dev",
			Annotations: map[string]string{
				annotationAppID: "proj/dev/myapp",
			},
		},
	}

	tests := []struct {
		eventType    k8swatch.EventType
		wantNil      bool
		wantEventKey string
	}{
		{k8swatch.Added, false, "create"},
		{k8swatch.Modified, false, "update"},
		{k8swatch.Deleted, false, "delete"},
		{k8swatch.Error, true, ""},
		{k8swatch.Bookmark, true, ""},
	}

	c := testClient(t)
	for _, tt := range tests {
		t.Run(string(tt.eventType), func(t *testing.T) {
			resp := c.kserviceEventToWatchResponse(context.Background(), k8swatch.Event{
				Type:   tt.eventType,
				Object: ksvc,
			})
			if tt.wantNil {
				assert.Nil(t, resp)
				return
			}
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
		})
	}
}

func TestKserviceName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"myapp", "myapp"},
		{"MyApp", "myapp"},
		// v1 and v2 variants stay distinct — no truncation collision.
		{"my-long-service-name-v1", "my-long-service-name-v1"},
		{"my-long-service-name-v2", "my-long-service-name-v2"},
		// Names over 63 chars get a hash suffix instead of blind truncation.
		{
			"this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit",
			func() string {
				name := "this-is-a-very-long-app-name-that-exceeds-the-kubernetes-dns-label-limit"
				sum := sha256.Sum256([]byte(name))
				return name[:54] + "-" + hex.EncodeToString(sum[:4])
			}(),
		},
	}
	for _, tt := range tests {
		id := &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: tt.name}
		got := kserviceName(id)
		assert.Equal(t, tt.want, got)
		assert.LessOrEqual(t, len(got), maxKServiceNameLen)
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
