package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/app/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
)

const (
	labelAppManaged = "flyte.org/app-managed"
	labelProject    = "flyte.org/project"
	labelDomain     = "flyte.org/domain"
	labelAppName    = "flyte.org/app-name"

	annotationSpecSHA = "flyte.org/spec-sha"
	annotationAppID   = "flyte.org/app-id"

	maxScaleZero = "0"

	// maxKServiceNameLen is the Kubernetes DNS label limit.
	maxKServiceNameLen = 63
)

// AppK8sClientInterface defines the KService lifecycle operations for the App service.
type AppK8sClientInterface interface {
	// Deploy creates or updates the KService for the given app. Idempotent — skips
	// the update if the spec SHA annotation is unchanged.
	Deploy(ctx context.Context, app *flyteapp.App) error

	// Stop scales the KService to zero by setting max-scale=0. The KService CRD
	// is kept so the app can be restarted later.
	Stop(ctx context.Context, appID *flyteapp.Identifier) error

	// GetStatus reads the KService and maps its conditions to a DeploymentStatus.
	// Returns a Status with STOPPED if the KService does not exist.
	GetStatus(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.Status, error)

	// List returns all apps (spec + live status) for the given project/domain scope.
	List(ctx context.Context, project, domain string) ([]*flyteapp.App, error)

	// Delete removes the KService CRD entirely. The app must be re-created from scratch.
	// Use Stop to scale to zero while preserving the KService.
	Delete(ctx context.Context, appID *flyteapp.Identifier) error

	// GetReplicas lists the pods (replicas) currently backing the given app.
	GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error)

	// DeleteReplica force-deletes a specific pod. Knative will replace it automatically.
	DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error

	// Watch returns a channel of WatchResponse events for KServices matching the
	// given project/domain scope. The channel is closed when ctx is cancelled.
	Watch(ctx context.Context, project, domain string) (<-chan *flyteapp.WatchResponse, error)
}

// AppK8sClient implements AppK8sClientInterface using controller-runtime.
type AppK8sClient struct {
	k8sClient client.WithWatch
	cache     ctrlcache.Cache
	namespace string
	cfg       *config.AppConfig
}

// NewAppK8sClient creates a new AppK8sClient.
func NewAppK8sClient(k8sClient client.WithWatch, cache ctrlcache.Cache, cfg *config.AppConfig) *AppK8sClient {
	return &AppK8sClient{
		k8sClient: k8sClient,
		cache:     cache,
		namespace: cfg.Namespace,
		cfg:       cfg,
	}
}

// Deploy creates or updates the KService for the given app.
func (c *AppK8sClient) Deploy(ctx context.Context, app *flyteapp.App) error {
	appID := app.GetMetadata().GetId()
	name := kserviceName(appID)

	ksvc, err := c.buildKService(app)
	if err != nil {
		return fmt.Errorf("failed to build KService for app %s: %w", name, err)
	}

	existing := &servingv1.Service{}
	err = c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: c.namespace}, existing)
	if k8serrors.IsNotFound(err) {
		if err := c.k8sClient.Create(ctx, ksvc); err != nil {
			return fmt.Errorf("failed to create KService %s: %w", name, err)
		}
		logger.Infof(ctx, "Created KService %s/%s", c.namespace, name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get KService %s: %w", name, err)
	}

	// Skip update if spec has not changed.
	if existing.Annotations[annotationSpecSHA] == ksvc.Annotations[annotationSpecSHA] {
		logger.Debugf(ctx, "KService %s/%s spec unchanged, skipping update", c.namespace, name)
		return nil
	}

	existing.Spec = ksvc.Spec
	existing.Labels = ksvc.Labels
	existing.Annotations = ksvc.Annotations
	if err := c.k8sClient.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update KService %s: %w", name, err)
	}
	logger.Infof(ctx, "Updated KService %s/%s", c.namespace, name)
	return nil
}

// Stop sets max-scale=0 on the KService, scaling it to zero without deleting it.
func (c *AppK8sClient) Stop(ctx context.Context, appID *flyteapp.Identifier) error {
	name := kserviceName(appID)
	patch := []byte(`{"spec":{"template":{"metadata":{"annotations":{"autoscaling.knative.dev/max-scale":"0"}}}}}`)
	ksvc := &servingv1.Service{}
	ksvc.Name = name
	ksvc.Namespace = c.namespace
	if err := c.k8sClient.Patch(ctx, ksvc, client.RawPatch(types.MergePatchType, patch)); err != nil {
		if k8serrors.IsNotFound(err) {
			// Already stopped/deleted — treat as success.
			return nil
		}
		return fmt.Errorf("failed to patch KService %s to stop: %w", name, err)
	}
	logger.Infof(ctx, "Stopped KService %s/%s (max-scale=0)", c.namespace, name)
	return nil
}

// Delete removes the KService CRD for the given app entirely.
func (c *AppK8sClient) Delete(ctx context.Context, appID *flyteapp.Identifier) error {
	name := kserviceName(appID)
	ksvc := &servingv1.Service{}
	ksvc.Name = name
	ksvc.Namespace = c.namespace
	if err := c.k8sClient.Delete(ctx, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete KService %s: %w", name, err)
	}
	logger.Infof(ctx, "Deleted KService %s/%s", c.namespace, name)
	return nil
}

// Watch returns a channel of WatchResponse events for KServices in the given
// project/domain scope. Pass empty strings to watch all managed KServices.
// The channel is closed when ctx is cancelled or the underlying watch terminates.
func (c *AppK8sClient) Watch(ctx context.Context, project, domain string) (<-chan *flyteapp.WatchResponse, error) {
	labels := client.MatchingLabels{labelAppManaged: "true"}
	if project != "" {
		labels[labelProject] = project
	}
	if domain != "" {
		labels[labelDomain] = domain
	}

	watcher, err := c.k8sClient.Watch(ctx, &servingv1.ServiceList{},
		client.InNamespace(c.namespace),
		labels,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start KService watch for %s/%s: %w", project, domain, err)
	}

	ch := make(chan *flyteapp.WatchResponse, 64)
	go func() {
		defer close(ch)
		defer watcher.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}
				resp := kserviceEventToWatchResponse(event)
				if resp == nil {
					continue
				}
				select {
				case ch <- resp:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch, nil
}

// kserviceEventToWatchResponse maps a K8s watch event to a flyteapp.WatchResponse.
// Returns nil for event types that should not be forwarded (Error, Bookmark).
func kserviceEventToWatchResponse(event k8swatch.Event) *flyteapp.WatchResponse {
	ksvc, ok := event.Object.(*servingv1.Service)
	if !ok {
		return nil
	}
	app, err := kserviceToApp(ksvc)
	if err != nil {
		// KService is not managed by us — skip it.
		return nil
	}
	switch event.Type {
	case k8swatch.Added:
		return &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_CreateEvent{
				CreateEvent: &flyteapp.CreateEvent{App: app},
			},
		}
	case k8swatch.Modified:
		return &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_UpdateEvent{
				UpdateEvent: &flyteapp.UpdateEvent{UpdatedApp: app},
			},
		}
	case k8swatch.Deleted:
		return &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_DeleteEvent{
				DeleteEvent: &flyteapp.DeleteEvent{App: app},
			},
		}
	default:
		return nil
	}
}

// GetStatus reads the KService and maps its conditions to a flyteapp.Status proto.
func (c *AppK8sClient) GetStatus(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.Status, error) {
	name := kserviceName(appID)
	ksvc := &servingv1.Service{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: c.namespace}, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			return statusWithPhase(flyteapp.Status_DEPLOYMENT_STATUS_STOPPED, "KService not found"), nil
		}
		return nil, fmt.Errorf("failed to get KService %s: %w", name, err)
	}
	return kserviceToStatus(ksvc), nil
}

// List returns all apps for the given project/domain by listing KServices with label selectors.
func (c *AppK8sClient) List(ctx context.Context, project, domain string) ([]*flyteapp.App, error) {
	list := &servingv1.ServiceList{}
	if err := c.k8sClient.List(ctx, list,
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			labelProject: project,
			labelDomain:  domain,
		},
	); err != nil {
		return nil, fmt.Errorf("failed to list KServices for %s/%s: %w", project, domain, err)
	}

	apps := make([]*flyteapp.App, 0, len(list.Items))
	for i := range list.Items {
		a, err := kserviceToApp(&list.Items[i])
		if err != nil {
			logger.Warnf(ctx, "Skipping KService %s: failed to convert to app: %v", list.Items[i].Name, err)
			continue
		}
		apps = append(apps, a)
	}
	return apps, nil
}

// --- Helpers ---

// kserviceName builds the KService name from an app identifier.
// Format: "{project}-{domain}-{name}", truncated to 63 chars.
func kserviceName(id *flyteapp.Identifier) string {
	name := fmt.Sprintf("%s-%s-%s", id.GetProject(), id.GetDomain(), id.GetName())
	if len(name) > maxKServiceNameLen {
		name = name[:maxKServiceNameLen]
	}
	return strings.ToLower(name)
}

// specSHA computes a SHA256 digest of the serialized App Spec proto.
func specSHA(spec *flyteapp.Spec) (string, error) {
	b, err := proto.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal spec: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8]), nil // 8 bytes = 16 hex chars, enough for change detection
}

// buildKService constructs a Knative Service manifest from an App proto.
func (c *AppK8sClient) buildKService(app *flyteapp.App) (*servingv1.Service, error) {
	appID := app.GetMetadata().GetId()
	spec := app.GetSpec()
	name := kserviceName(appID)

	sha, err := specSHA(spec)
	if err != nil {
		return nil, err
	}

	podSpec, err := buildPodSpec(spec)
	if err != nil {
		return nil, err
	}

	templateAnnotations := buildAutoscalingAnnotations(spec, c.cfg)

	timeoutSecs := c.cfg.DefaultRequestTimeout.Seconds()
	if t := spec.GetTimeouts().GetRequestTimeout(); t != nil {
		timeoutSecs = t.AsDuration().Seconds()
		if timeoutSecs > c.cfg.MaxRequestTimeout.Seconds() {
			timeoutSecs = c.cfg.MaxRequestTimeout.Seconds()
		}
	}
	timeoutSecsInt := int64(timeoutSecs)

	ksvc := &servingv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: c.namespace,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    appID.GetProject(),
				labelDomain:     appID.GetDomain(),
				labelAppName:    appID.GetName(),
			},
			Annotations: map[string]string{
				annotationSpecSHA: sha,
				annotationAppID:   fmt.Sprintf("%s/%s/%s", appID.GetProject(), appID.GetDomain(), appID.GetName()),
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: templateAnnotations,
					},
					Spec: servingv1.RevisionSpec{
						PodSpec:        podSpec,
						TimeoutSeconds: &timeoutSecsInt,
					},
				},
			},
		},
	}
	return ksvc, nil
}

// buildPodSpec constructs a corev1.PodSpec from an App Spec.
// Supports Container payload only for now; K8sPod support can be added in a follow-up.
func buildPodSpec(spec *flyteapp.Spec) (corev1.PodSpec, error) {
	switch p := spec.GetAppPayload().(type) {
	case *flyteapp.Spec_Container:
		c := p.Container
		container := corev1.Container{
			Name:  "app",
			Image: c.GetImage(),
			Args:  c.GetArgs(),
		}
		for _, e := range c.GetEnv() {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  e.GetKey(),
				Value: e.GetValue(),
			})
		}
		return corev1.PodSpec{Containers: []corev1.Container{container}}, nil

	case *flyteapp.Spec_Pod:
		// K8sPod payloads are not yet supported — the pod spec serialization
		// from flyteplugins is needed for a complete implementation.
		return corev1.PodSpec{}, fmt.Errorf("K8sPod app payload is not yet supported")

	default:
		return corev1.PodSpec{}, fmt.Errorf("app spec has no payload (container or pod required)")
	}
}

// buildAutoscalingAnnotations returns the Knative autoscaling annotations for the revision template.
func buildAutoscalingAnnotations(spec *flyteapp.Spec, cfg *config.AppConfig) map[string]string {
	annotations := map[string]string{}
	autoscaling := spec.GetAutoscaling()
	if autoscaling == nil {
		return annotations
	}

	if r := autoscaling.GetReplicas(); r != nil {
		annotations["autoscaling.knative.dev/min-scale"] = fmt.Sprintf("%d", r.GetMin())
		annotations["autoscaling.knative.dev/max-scale"] = fmt.Sprintf("%d", r.GetMax())
	}

	if m := autoscaling.GetScalingMetric(); m != nil {
		switch metric := m.GetMetric().(type) {
		case *flyteapp.ScalingMetric_RequestRate:
			annotations["autoscaling.knative.dev/metric"] = "rps"
			annotations["autoscaling.knative.dev/target"] = fmt.Sprintf("%d", metric.RequestRate.GetTargetValue())
		case *flyteapp.ScalingMetric_Concurrency:
			annotations["autoscaling.knative.dev/metric"] = "concurrency"
			annotations["autoscaling.knative.dev/target"] = fmt.Sprintf("%d", metric.Concurrency.GetTargetValue())
		}
	}

	if p := autoscaling.GetScaledownPeriod(); p != nil {
		annotations["autoscaling.knative.dev/window"] = p.AsDuration().String()
	}

	return annotations
}

// statusWithPhase builds a flyteapp.Status with a single Condition set to the given phase.
func statusWithPhase(phase flyteapp.Status_DeploymentStatus, message string) *flyteapp.Status {
	return &flyteapp.Status{
		Conditions: []*flyteapp.Condition{
			{
				DeploymentStatus:   phase,
				Message:            message,
				LastTransitionTime: timestamppb.Now(),
			},
		},
	}
}

// kserviceToStatus maps a KService's conditions to a flyteapp.Status proto.
func kserviceToStatus(ksvc *servingv1.Service) *flyteapp.Status {
	var phase flyteapp.Status_DeploymentStatus
	var message string

	// Check if max-scale=0 is set — explicitly stopped by the control plane.
	if ann := ksvc.Spec.Template.Annotations; ann != nil {
		if ann["autoscaling.knative.dev/max-scale"] == maxScaleZero {
			phase = flyteapp.Status_DEPLOYMENT_STATUS_STOPPED
			message = "App scaled to zero"
		}
	}

	if phase == flyteapp.Status_DEPLOYMENT_STATUS_UNSPECIFIED {
		switch {
		case ksvc.IsReady():
			phase = flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE
		case ksvc.IsFailed():
			phase = flyteapp.Status_DEPLOYMENT_STATUS_FAILED
			if c := ksvc.Status.GetCondition(servingv1.ServiceConditionReady); c != nil {
				message = c.Message
			}
		case ksvc.Status.LatestCreatedRevisionName != ksvc.Status.LatestReadyRevisionName:
			phase = flyteapp.Status_DEPLOYMENT_STATUS_DEPLOYING
		default:
			phase = flyteapp.Status_DEPLOYMENT_STATUS_PENDING
		}
	}

	status := statusWithPhase(phase, message)

	// Populate ingress URL from KService route status.
	if url := ksvc.Status.URL; url != nil {
		status.Ingress = &flyteapp.Ingress{
			PublicUrl: url.String(),
		}
	}

	// Populate current replica count and K8s namespace metadata.
	status.CurrentReplicas = uint32(len(ksvc.Status.Traffic))
	status.K8SMetadata = &flyteapp.K8SMetadata{
		Namespace: ksvc.Namespace,
	}

	return status
}

// GetReplicas lists the pods currently backing the given app by matching
// the flyte.org/project, flyte.org/domain, and flyte.org/app-name labels.
func (c *AppK8sClient) GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error) {
	podList := &corev1.PodList{}
	if err := c.k8sClient.List(ctx, podList,
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			labelProject: appID.GetProject(),
			labelDomain:  appID.GetDomain(),
			labelAppName: appID.GetName(),
		},
	); err != nil {
		return nil, fmt.Errorf("failed to list pods for app %s/%s/%s: %w",
			appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	replicas := make([]*flyteapp.Replica, 0, len(podList.Items))
	for i := range podList.Items {
		replicas = append(replicas, podToReplica(appID, &podList.Items[i]))
	}
	return replicas, nil
}

// DeleteReplica force-deletes a specific pod. Knative will schedule a replacement automatically.
func (c *AppK8sClient) DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error {
	pod := &corev1.Pod{}
	pod.Name = replicaID.GetName()
	pod.Namespace = c.namespace
	if err := c.k8sClient.Delete(ctx, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete pod %s/%s: %w", c.namespace, replicaID.GetName(), err)
	}
	logger.Infof(ctx, "Deleted replica pod %s/%s", c.namespace, replicaID.GetName())
	return nil
}

// podToReplica maps a corev1.Pod to a flyteapp.Replica proto.
func podToReplica(appID *flyteapp.Identifier, pod *corev1.Pod) *flyteapp.Replica {
	status, reason := podDeploymentStatus(pod)
	return &flyteapp.Replica{
		Metadata: &flyteapp.ReplicaMeta{
			Id: &flyteapp.ReplicaIdentifier{
				AppId: appID,
				Name:  pod.Name,
			},
		},
		Status: &flyteapp.ReplicaStatus{
			DeploymentStatus: status,
			Reason:           reason,
		},
	}
}

// podDeploymentStatus maps a pod's phase and conditions to a status string and reason.
func podDeploymentStatus(pod *corev1.Pod) (string, string) {
	switch pod.Status.Phase {
	case corev1.PodRunning:
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				if cs.State.Waiting != nil {
					return "DEPLOYING", cs.State.Waiting.Reason
				}
				return "DEPLOYING", "container not ready"
			}
		}
		return "ACTIVE", ""
	case corev1.PodPending:
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
				return "PENDING", cs.State.Waiting.Reason
			}
		}
		return "PENDING", string(pod.Status.Phase)
	case corev1.PodFailed:
		reason := pod.Status.Reason
		if reason == "" && len(pod.Status.ContainerStatuses) > 0 {
			if t := pod.Status.ContainerStatuses[0].State.Terminated; t != nil {
				reason = t.Reason
			}
		}
		return "FAILED", reason
	case corev1.PodSucceeded:
		return "STOPPED", "pod completed"
	default:
		return "PENDING", string(pod.Status.Phase)
	}
}

// kserviceToApp reconstructs a flyteapp.App from a KService by reading the
// app identifier from annotations and the live status from KService conditions.
func kserviceToApp(ksvc *servingv1.Service) (*flyteapp.App, error) {
	appIDStr, ok := ksvc.Annotations[annotationAppID]
	if !ok {
		return nil, fmt.Errorf("KService %s missing %s annotation", ksvc.Name, annotationAppID)
	}

	// annotation format: "{project}/{domain}/{name}"
	parts := strings.SplitN(appIDStr, "/", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("KService %s has malformed %s annotation: %q", ksvc.Name, annotationAppID, appIDStr)
	}

	appID := &flyteapp.Identifier{
		Project: parts[0],
		Domain:  parts[1],
		Name:    parts[2],
	}

	return &flyteapp.App{
		Metadata: &flyteapp.Meta{
			Id: appID,
		},
		Status: kserviceToStatus(ksvc),
	}, nil
}
