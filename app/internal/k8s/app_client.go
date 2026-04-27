package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	toolscache "k8s.io/client-go/tools/cache"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/app/internal/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	flytecore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	labelAppManaged = "flyte.org/app-managed"
	labelProject    = "flyte.org/project"
	labelDomain     = "flyte.org/domain"
	labelAppName    = "flyte.org/app-name"

	// labelKnativeService is set by Knative on every Deployment/Pod it creates
	// from a KService. Its value equals the KService name. We use this to find
	// replica pods, since our own labels on the KService aren't propagated
	// onto Knative-generated pods.
	labelKnativeService = "serving.knative.dev/service"

	// labelKnativeRevision is set by Knative on every pod and identifies the
	// Revision that pod belongs to. Used to restrict log queries to pods of
	// the latest revision during/after a rollout.
	labelKnativeRevision = "serving.knative.dev/revision"

	annotationSpecSHA = "flyte.org/spec-sha"
	annotationAppID   = "flyte.org/app-id"
	annotationAppOrg  = "flyte.org/app-org"
	annotationSpec    = "flyte.org/spec"

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

	// GetApp reads the KService and returns the full App including reconstructed Spec and live Status.
	// Returns a not-found error (checkable with k8serrors.IsNotFound) if the KService does not exist.
	GetApp(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.App, error)

	// List returns apps for the given project/domain scope with optional pagination.
	// limit=0 means no limit. token is the K8s continue token from a previous call.
	// Returns the apps, the continue token for the next page (empty if last page), and any error.
	List(ctx context.Context, project, domain string, limit uint32, token string) ([]*flyteapp.App, string, error)

	// Delete removes the KService CRD entirely. The app must be re-created from scratch.
	// Use Stop to scale to zero while preserving the KService.
	Delete(ctx context.Context, appID *flyteapp.Identifier) error

	// GetReplicas lists the pods (replicas) currently backing the given app.
	GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error)

	// DeleteReplica force-deletes a specific pod. Knative will replace it automatically.
	DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error

	// StartWatching starts the KService informer and begins dispatching events
	// to subscribers. Must be called before Subscribe.
	StartWatching(ctx context.Context) error

	// StopWatching stops the KService informer.
	StopWatching()

	// Subscribe creates a channel that receives WatchResponse events for the
	// given app name. Use empty string to receive events for all apps.
	Subscribe(appName string) chan *flyteapp.WatchResponse

	// Unsubscribe removes a subscription channel previously returned by Subscribe.
	Unsubscribe(appName string, ch chan *flyteapp.WatchResponse)
}

// AppK8sClient implements AppK8sClientInterface using controller-runtime.
type AppK8sClient struct {
	k8sClient client.WithWatch
	cache     ctrlcache.Cache
	cfg       *config.InternalAppConfig

	// Watch management
	mu          sync.RWMutex
	subscribers map[string]map[chan *flyteapp.WatchResponse]struct{}
	stopCh      chan struct{}
	watching    bool
}

// NewAppK8sClient creates a new AppK8sClient.
func NewAppK8sClient(k8sClient client.WithWatch, cache ctrlcache.Cache, cfg *config.InternalAppConfig) *AppK8sClient {
	return &AppK8sClient{
		k8sClient:   k8sClient,
		cache:       cache,
		cfg:         cfg,
		subscribers: make(map[string]map[chan *flyteapp.WatchResponse]struct{}),
	}
}

// AppNamespace is the fixed Kubernetes namespace where all KService objects are deployed.
const AppNamespace = "flyte"

// defaultOrg is the org returned for apps that have no org persisted on the KService.
// we always surface a non-empty value for callers (e.g. the UI) that expect one.
const defaultOrg = "flyte"

// Deploy creates or updates the KService for the given app.
func (c *AppK8sClient) Deploy(ctx context.Context, app *flyteapp.App) error {
	appID := app.GetMetadata().GetId()
	ns := AppNamespace
	name := KServiceName(appID)

	if err := k8s.EnsureNamespaceExists(ctx, c.k8sClient, ns); err != nil {
		return fmt.Errorf("failed to ensure namespace %s: %w", ns, err)
	}

	ksvc, err := c.buildKService(app)
	if err != nil {
		return fmt.Errorf("failed to build KService for app %s: %w", name, err)
	}

	existing := &servingv1.Service{}
	err = c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, existing)
	if k8serrors.IsNotFound(err) {
		if err := c.k8sClient.Create(ctx, ksvc); err != nil {
			return fmt.Errorf("failed to create KService %s: %w", name, err)
		}
		logger.Infof(ctx, "Created KService %s/%s", ns, name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get KService %s: %w", name, err)
	}

	if existing.Annotations[annotationSpecSHA] == ksvc.Annotations[annotationSpecSHA] {
		logger.Debugf(ctx, "KService %s/%s spec unchanged, skipping update", ns, name)
		return nil
	}

	existing.Spec = ksvc.Spec
	// Merge labels and annotations rather than replacing them wholesale.
	// Knative sets immutable annotations (e.g. serving.knative.dev/creator)
	// on creation; overwriting them causes the admission webhook to reject the update.
	if existing.Labels == nil {
		existing.Labels = make(map[string]string)
	}
	for k, v := range ksvc.Labels {
		existing.Labels[k] = v
	}
	if existing.Annotations == nil {
		existing.Annotations = make(map[string]string)
	}
	for k, v := range ksvc.Annotations {
		existing.Annotations[k] = v
	}
	if err := c.k8sClient.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update KService %s: %w", name, err)
	}
	logger.Infof(ctx, "Updated KService %s/%s", ns, name)
	return nil
}

// Stop sets max-scale=0 on the KService, scaling it to zero without deleting it.
func (c *AppK8sClient) Stop(ctx context.Context, appID *flyteapp.Identifier) error {
	ns := AppNamespace
	name := KServiceName(appID)
	patch := []byte(`{"spec":{"template":{"metadata":{"annotations":{"autoscaling.knative.dev/max-scale":"0"}}}}}`)
	ksvc := &servingv1.Service{}
	ksvc.Name = name
	ksvc.Namespace = ns
	if err := c.k8sClient.Patch(ctx, ksvc, client.RawPatch(types.MergePatchType, patch)); err != nil {
		if k8serrors.IsNotFound(err) {
			// Already stopped/deleted — treat as success.
			return nil
		}
		return fmt.Errorf("failed to patch KService %s to stop: %w", name, err)
	}
	logger.Infof(ctx, "Stopped KService %s/%s (max-scale=0)", ns, name)
	return nil
}

// Delete removes the KService CRD for the given app entirely.
func (c *AppK8sClient) Delete(ctx context.Context, appID *flyteapp.Identifier) error {
	ns := AppNamespace
	name := KServiceName(appID)
	ksvc := &servingv1.Service{}
	ksvc.Name = name
	ksvc.Namespace = ns
	if err := c.k8sClient.Delete(ctx, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete KService %s: %w", name, err)
	}
	logger.Infof(ctx, "Deleted KService %s/%s", ns, name)
	return nil
}

// StartWatching starts the KService informer and begins dispatching events
// to subscriber channels. Must be called once during setup.
func (c *AppK8sClient) StartWatching(ctx context.Context) error {
	c.mu.Lock()
	if c.watching {
		c.mu.Unlock()
		return nil
	}
	c.watching = true
	c.stopCh = make(chan struct{})
	c.mu.Unlock()

	if c.cache == nil {
		return fmt.Errorf("shared cache is required for KService informer")
	}

	return c.setupInformer(ctx)
}

func (c *AppK8sClient) setupInformer(ctx context.Context) error {
	informer, err := c.cache.GetInformer(ctx, &servingv1.Service{})
	if err != nil {
		return fmt.Errorf("failed to get KService informer: %w", err)
	}

	_, err = informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ksvc, ok := obj.(*servingv1.Service)
			if !ok || !isManagedKService(ksvc) {
				return
			}
			c.handleKServiceEvent(ctx, ksvc, k8swatch.Added)
		},
		UpdateFunc: func(_, newObj interface{}) {
			ksvc, ok := newObj.(*servingv1.Service)
			if !ok || !isManagedKService(ksvc) {
				return
			}
			c.handleKServiceEvent(ctx, ksvc, k8swatch.Modified)
		},
		DeleteFunc: func(obj interface{}) {
			// The informer may deliver a DeletedFinalStateUnknown tombstone
			// when a delete event was missed; unwrap it first.
			if tombstone, ok := obj.(toolscache.DeletedFinalStateUnknown); ok {
				obj = tombstone.Obj
			}
			ksvc, ok := obj.(*servingv1.Service)
			if !ok {
				return
			}
			c.handleKServiceEvent(ctx, ksvc, k8swatch.Deleted)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add KService informer handler: %w", err)
	}

	return nil
}

// isManagedKService returns true if the KService has the flyte.org/app-managed label.
func isManagedKService(ksvc *servingv1.Service) bool {
	return ksvc.Labels[labelAppManaged] == "true"
}

// handleKServiceEvent converts a KService event into a WatchResponse and
// notifies all matching subscribers.
func (c *AppK8sClient) handleKServiceEvent(ctx context.Context, ksvc *servingv1.Service, eventType k8swatch.EventType) {
	app, err := c.kserviceToApp(ctx, ksvc)
	if err != nil {
		return
	}

	var resp *flyteapp.WatchResponse
	switch eventType {
	case k8swatch.Added:
		resp = &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_CreateEvent{
				CreateEvent: &flyteapp.CreateEvent{App: app},
			},
		}
	case k8swatch.Modified:
		resp = &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_UpdateEvent{
				UpdateEvent: &flyteapp.UpdateEvent{UpdatedApp: app},
			},
		}
	case k8swatch.Deleted:
		resp = &flyteapp.WatchResponse{
			Event: &flyteapp.WatchResponse_DeleteEvent{
				DeleteEvent: &flyteapp.DeleteEvent{App: app},
			},
		}
	default:
		return
	}

	appName := app.GetMetadata().GetId().GetName()
	c.notifySubscribers(ctx, appName, resp)
}

// notifySubscribers sends a WatchResponse to all subscribers for the given app name.
func (c *AppK8sClient) notifySubscribers(ctx context.Context, appName string, resp *flyteapp.WatchResponse) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for ch := range c.subscribers[appName] {
		select {
		case ch <- resp:
		default:
			logger.Warnf(ctx, "subscriber channel full, dropping update for app: %s", appName)
		}
	}
}

// Subscribe creates a channel that receives WatchResponse events for the given app name.
func (c *AppK8sClient) Subscribe(appName string) chan *flyteapp.WatchResponse {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan *flyteapp.WatchResponse, c.cfg.WatchBufferSize)
	if c.subscribers[appName] == nil {
		c.subscribers[appName] = make(map[chan *flyteapp.WatchResponse]struct{})
	}
	c.subscribers[appName][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the given channel from the subscription list.
func (c *AppK8sClient) Unsubscribe(appName string, ch chan *flyteapp.WatchResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if channels, ok := c.subscribers[appName]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(c.subscribers, appName)
		}
	}
}

// StopWatching stops the KService informer.
func (c *AppK8sClient) StopWatching() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watching && c.stopCh != nil {
		close(c.stopCh)
		c.watching = false
	}
}

// GetApp reads the KService and returns the full App including reconstructed Spec and live Status.
func (c *AppK8sClient) GetApp(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.App, error) {
	ns := AppNamespace
	name := KServiceName(appID)
	ksvc := &servingv1.Service{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("KService %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get KService %s: %w", name, err)
	}
	return c.kserviceToApp(ctx, ksvc)
}

// specFromAnnotation deserializes the Spec stored in the flyte.org/spec annotation.
// Returns nil if the annotation is absent or malformed.
func specFromAnnotation(ksvc *servingv1.Service) *flyteapp.Spec {
	b64, ok := ksvc.Annotations[annotationSpec]
	if !ok {
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil
	}
	spec := &flyteapp.Spec{}
	if err := proto.Unmarshal(b, spec); err != nil {
		return nil
	}
	return spec
}

// List returns apps for the given project/domain scope with optional pagination.
func (c *AppK8sClient) List(ctx context.Context, project, domain string, limit uint32, token string) ([]*flyteapp.App, string, error) {
	ns := AppNamespace

	matchLabels := map[string]string{labelAppManaged: "true"}
	if project != "" {
		matchLabels[labelProject] = project
	}
	if domain != "" {
		matchLabels[labelDomain] = domain
	}

	listOpts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels(matchLabels),
	}
	if limit > 0 {
		listOpts = append(listOpts, client.Limit(int64(limit)))
	}
	if token != "" {
		listOpts = append(listOpts, client.Continue(token))
	}

	list := &servingv1.ServiceList{}
	if err := c.k8sClient.List(ctx, list, listOpts...); err != nil {
		return nil, "", fmt.Errorf("failed to list KServices for %s/%s: %w", project, domain, err)
	}

	apps := make([]*flyteapp.App, 0, len(list.Items))
	for i := range list.Items {
		a, err := c.kserviceToApp(ctx, &list.Items[i])
		if err != nil {
			logger.Warnf(ctx, "Skipping KService %s: failed to convert to app: %v", list.Items[i].Name, err)
			continue
		}
		apps = append(apps, a)
	}
	return apps, list.Continue, nil
}

// publicIngress returns the deterministic public URL for an app using the same
// logic as the service layer so GetApp/List/Watch are consistent with Create.
func (c *AppK8sClient) publicIngress(id *flyteapp.Identifier) *flyteapp.Ingress {
	if c.cfg.BaseDomain == "" {
		return nil
	}
	scheme := c.cfg.Scheme
	if scheme == "" {
		scheme = "https"
	}
	host := strings.ToLower(fmt.Sprintf("%s.%s",
		KServiceName(id), c.cfg.BaseDomain))
	url := scheme + "://" + host
	if c.cfg.IngressAppsPort != 0 {
		url += fmt.Sprintf(":%d", c.cfg.IngressAppsPort)
	}
	return &flyteapp.Ingress{PublicUrl: url}
}

// --- Helpers ---

// KServiceName returns the KService name for an app. All apps share the
// "flyte" namespace, so the name must be unique across all (project, domain, name)
// triples — we encode all three in the name. Lower-cased and capped at 63 chars
// (K8s DNS label limit); names exceeding the cap fall back to a deterministic
// 8-char SHA256 suffix to guarantee uniqueness without exceeding the limit.
func KServiceName(id *flyteapp.Identifier) string {
	raw := strings.ToLower(fmt.Sprintf("%s-%s-%s", id.GetName(), id.GetProject(), id.GetDomain()))
	if len(raw) <= maxKServiceNameLen {
		return raw
	}
	sum := sha256.Sum256([]byte(id.GetProject() + "/" + id.GetDomain() + "/" + id.GetName()))
	suffix := hex.EncodeToString(sum[:4])
	prefix := raw
	if len(prefix) > maxKServiceNameLen-9 {
		prefix = prefix[:maxKServiceNameLen-9]
	}
	return prefix + "-" + suffix
}

// marshalSpec serializes the App Spec proto and returns the raw bytes.
func marshalSpec(spec *flyteapp.Spec) ([]byte, error) {
	b, err := proto.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal spec: %w", err)
	}
	return b, nil
}

// specSHA computes a SHA256 digest of the serialized App Spec proto.
func specSHA(specBytes []byte) string {
	sum := sha256.Sum256(specBytes)
	return hex.EncodeToString(sum[:8]) // 8 bytes = 16 hex chars, enough for change detection
}

// buildKService constructs a Knative Service manifest from an App proto.
func (c *AppK8sClient) buildKService(app *flyteapp.App) (*servingv1.Service, error) {
	appID := app.GetMetadata().GetId()
	spec := app.GetSpec()
	name := KServiceName(appID)
	ns := AppNamespace

	specBytes, err := marshalSpec(spec)
	if err != nil {
		return nil, err
	}
	sha := specSHA(specBytes)

	podSpec, err := buildPodSpec(spec)
	if err != nil {
		return nil, err
	}
	// Inject cluster-level default env vars (e.g. _U_EP_OVERRIDE) before user vars
	// so they can be overridden by app-specific env vars if needed.
	if len(c.cfg.DefaultEnvVars) > 0 && len(podSpec.Containers) > 0 {
		defaults := make([]corev1.EnvVar, 0, len(c.cfg.DefaultEnvVars))
		for k, v := range c.cfg.DefaultEnvVars {
			defaults = append(defaults, corev1.EnvVar{Name: k, Value: v})
		}
		podSpec.Containers[0].Env = append(defaults, podSpec.Containers[0].Env...)
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
			Namespace: ns,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    appID.GetProject(),
				labelDomain:     appID.GetDomain(),
				labelAppName:    appID.GetName(),
			},
			Annotations: map[string]string{
				annotationSpecSHA: sha,
				annotationAppID:   fmt.Sprintf("%s/%s/%s", appID.GetProject(), appID.GetDomain(), appID.GetName()),
				annotationAppOrg:  appID.GetOrg(),
				annotationSpec:    base64.StdEncoding.EncodeToString(specBytes),
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
			Name:    "app",
			Image:   c.GetImage(),
			Command: c.GetCommand(),
			Args:    c.GetArgs(),
		}
		for _, e := range c.GetEnv() {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  e.GetKey(),
				Value: e.GetValue(),
			})
		}
		for _, p := range c.GetPorts() {
			container.Ports = append(container.Ports, corev1.ContainerPort{
				ContainerPort: int32(p.GetContainerPort()),
				Name:          p.GetName(),
			})
		}
		container.Resources = buildResourceRequirements(c.GetResources())
		return corev1.PodSpec{
			Containers:         []corev1.Container{container},
			EnableServiceLinks: boolPtr(false),
		}, nil

	case *flyteapp.Spec_Pod:
		// K8sPod payloads are not yet supported — the pod spec serialization
		// from flyteplugins is needed for a complete implementation.
		return corev1.PodSpec{}, fmt.Errorf("K8sPod app payload is not yet supported")

	default:
		return corev1.PodSpec{}, fmt.Errorf("app spec has no payload (container or pod required)")
	}
}

// buildResourceRequirements maps flyteidl2 Resources to corev1.ResourceRequirements.
func buildResourceRequirements(res *flytecore.Resources) corev1.ResourceRequirements {
	if res == nil {
		return corev1.ResourceRequirements{}
	}
	reqs := corev1.ResourceRequirements{}
	if len(res.GetRequests()) > 0 {
		reqs.Requests = make(corev1.ResourceList)
		for _, e := range res.GetRequests() {
			if name, ok := protoResourceName(e.GetName()); ok {
				reqs.Requests[name] = k8sresource.MustParse(e.GetValue())
			}
		}
	}
	if len(res.GetLimits()) > 0 {
		reqs.Limits = make(corev1.ResourceList)
		for _, e := range res.GetLimits() {
			if name, ok := protoResourceName(e.GetName()); ok {
				reqs.Limits[name] = k8sresource.MustParse(e.GetValue())
			}
		}
	}
	return reqs
}

// protoResourceName maps a flyteidl2 ResourceName to the equivalent corev1.ResourceName.
func protoResourceName(name flytecore.Resources_ResourceName) (corev1.ResourceName, bool) {
	switch name {
	case flytecore.Resources_CPU:
		return corev1.ResourceCPU, true
	case flytecore.Resources_MEMORY:
		return corev1.ResourceMemory, true
	case flytecore.Resources_STORAGE:
		return corev1.ResourceStorage, true
	case flytecore.Resources_EPHEMERAL_STORAGE:
		return corev1.ResourceEphemeralStorage, true
	default:
		return "", false
	}
}

func boolPtr(b bool) *bool { return &b }

// buildAutoscalingAnnotations returns the Knative autoscaling annotations for the revision template.
func buildAutoscalingAnnotations(spec *flyteapp.Spec, cfg *config.InternalAppConfig) map[string]string {
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
// It fetches the latest ready Revision to read the accurate ActualReplicas count.
func (c *AppK8sClient) kserviceToStatus(ctx context.Context, ksvc *servingv1.Service) *flyteapp.Status {
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

	// Populate ingress URL from the app annotation so the URL is consistent
	// with the Create response regardless of Knative route readiness.
	if appIDStr := ksvc.Annotations[annotationAppID]; appIDStr != "" {
		parts := strings.SplitN(appIDStr, "/", 3)
		if len(parts) == 3 {
			appID := &flyteapp.Identifier{Project: parts[0], Domain: parts[1], Name: parts[2]}
			status.Ingress = c.publicIngress(appID)
		}
	}

	// Populate current replica count from the latest ready Revision.
	if revName := ksvc.Status.LatestReadyRevisionName; revName != "" {
		rev := &servingv1.Revision{}
		if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: revName, Namespace: ksvc.Namespace}, rev); err == nil {
			if rev.Status.ActualReplicas != nil {
				status.CurrentReplicas = uint32(*rev.Status.ActualReplicas)
			}
		}
	}
	status.K8SMetadata = &flyteapp.K8SMetadata{
		Namespace: ksvc.Namespace,
	}

	if !ksvc.CreationTimestamp.IsZero() {
		status.CreatedAt = timestamppb.New(ksvc.CreationTimestamp.Time)
	}

	return status
}

// GetReplicas lists the pods currently backing the given app. When the KService
// has a latest ready revision, only pods from that revision are returned — old
// revision pods terminating during a rollout are filtered out. If the KService
// has no ready revision yet (initial rollout), all pods for the service are
// returned.
func (c *AppK8sClient) GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error) {
	ns := AppNamespace
	name := KServiceName(appID)

	labels := client.MatchingLabels{labelKnativeService: name}
	ksvc := &servingv1.Service{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, ksvc); err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get KService for app %s/%s/%s: %w",
				appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
		}
	} else if rev := ksvc.Status.LatestReadyRevisionName; rev != "" {
		labels[labelKnativeRevision] = rev
	}

	podList := &corev1.PodList{}
	if err := c.k8sClient.List(ctx, podList,
		client.InNamespace(ns),
		labels,
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
	ns := AppNamespace
	pod := &corev1.Pod{}
	pod.Name = replicaID.GetName()
	pod.Namespace = ns
	if err := c.k8sClient.Delete(ctx, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete pod %s/%s: %w", ns, replicaID.GetName(), err)
	}
	logger.Infof(ctx, "Deleted replica pod %s/%s", ns, replicaID.GetName())
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
func (c *AppK8sClient) kserviceToApp(ctx context.Context, ksvc *servingv1.Service) (*flyteapp.App, error) {
	appIDStr, ok := ksvc.Annotations[annotationAppID]
	if !ok {
		return nil, fmt.Errorf("KService %s missing %s annotation", ksvc.Name, annotationAppID)
	}

	// annotation format: "{project}/{domain}/{name}"
	parts := strings.SplitN(appIDStr, "/", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("KService %s has malformed %s annotation: %q", ksvc.Name, annotationAppID, appIDStr)
	}

	org := ksvc.Annotations[annotationAppOrg]
	if org == "" {
		org = defaultOrg
	}
	appID := &flyteapp.Identifier{
		Org:     org,
		Project: parts[0],
		Domain:  parts[1],
		Name:    parts[2],
	}

	return &flyteapp.App{
		Metadata: &flyteapp.Meta{
			Id: appID,
		},
		Spec:   specFromAnnotation(ksvc),
		Status: c.kserviceToStatus(ctx, ksvc),
	}, nil
}
