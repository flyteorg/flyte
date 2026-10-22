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
	"google.golang.org/protobuf/types/known/timestamppb"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	toolscache "k8s.io/client-go/tools/cache"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/app/internal/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	secretUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/v2/flytestdlib/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/telemetry"
	"github.com/flyteorg/flyte/v2/flytestdlib/utils"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	flytecore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	labelAppManaged = "flyte.org/app-managed"
	labelProject    = "flyte.org/project"
	labelDomain     = "flyte.org/domain"
	labelAppName    = "flyte.org/app-name"

	// labelKnativeService is set on every App Deployment/Pod and Service selector.
	// Its value equals KServiceName(appID).
	labelKnativeService = "serving.knative.dev/service"

	// labelKnativeRevision is optional revision metadata for replica filtering.
	labelKnativeRevision = "serving.knative.dev/revision"

	annotationSpecSHA = "flyte.org/spec-sha"
	annotationAppID   = "flyte.org/app-id"
	annotationAppOrg  = "flyte.org/app-org"
	annotationSpec    = "flyte.org/spec"

	labelAppStopped        = "flyte.org/app-stopped"
	labelKnativeVisibility = "networking.knative.dev/visibility"
	visibilityClusterLocal = "cluster-local"

	scaleZero = "0"

	// maxKServiceNameLen is the Kubernetes DNS label limit.
	maxKServiceNameLen = 63

	defaultAppPort = int32(8080)
)

// AppK8sClientInterface defines the App Deployment/Service lifecycle operations for the App service.
type AppK8sClientInterface interface {
	// Deploy creates or updates the Deployment and Service for the given app. Idempotent — skips
	// the update if the spec SHA annotation is unchanged.
	Deploy(ctx context.Context, app *flyteapp.App) error

	// Stop scales the Deployment to zero and marks it cluster-local.
	// The Deployment is kept so the app can be restarted later.
	Stop(ctx context.Context, appID *flyteapp.Identifier) error

	// GetApp reads the Deployment and returns the full App including reconstructed Spec and live Status.
	// Returns a not-found error (checkable with k8serrors.IsNotFound) if the Deployment does not exist.
	GetApp(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.App, error)

	// List returns apps for the given project/domain scope with optional pagination.
	// limit=0 means no limit. token is the K8s continue token from a previous call.
	// Returns the apps, the continue token for the next page (empty if last page), and any error.
	List(ctx context.Context, project, domain string, limit uint32, token string) ([]*flyteapp.App, string, error)

	// Delete removes the Deployment and Service entirely. The app must be re-created from scratch.
	// Use Stop to scale to zero while preserving the Deployment.
	Delete(ctx context.Context, appID *flyteapp.Identifier) error

	// GetReplicas lists the pods (replicas) currently backing the given app.
	GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error)

	// DeleteReplica force-deletes a specific pod. The Deployment controller will replace it automatically.
	DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error

	// StartWatching starts the Deployment informer and begins dispatching events
	// to subscribers. Must be called before Subscribe.
	StartWatching(ctx context.Context) error

	// StopWatching stops the Deployment informer.
	StopWatching()

	// Subscribe creates a channel that receives WatchResponse events for the
	// given app name. Use empty string to receive events for all apps.
	Subscribe(appName string) chan *flyteapp.WatchResponse

	// Unsubscribe removes a subscription channel previously returned by Subscribe.
	Unsubscribe(appName string, ch chan *flyteapp.WatchResponse)

	// PublicIngress returns the deterministic public Ingress URL for the given app.
	// Returns nil if BaseDomain is not configured.
	PublicIngress(id *flyteapp.Identifier) *flyteapp.Ingress
}

// AppK8sClient implements AppK8sClientInterface using native Kubernetes Deployments and Services.
type AppK8sClient struct {
	k8sClient client.WithWatch
	cache     ctrlcache.Cache
	cfg       *config.InternalAppConfig
	// Kubernetes namespace where all App Deployments and Services are deployed.
	namespace string

	// Watch management
	mu          sync.RWMutex
	subscribers map[string]map[chan *flyteapp.WatchResponse]struct{}
	stopCh      chan struct{}
	watching    bool
}

// NewAppK8sClient creates a new AppK8sClient.
func NewAppK8sClient(k8sClient client.WithWatch, cache ctrlcache.Cache, namespace string, cfg *config.InternalAppConfig) *AppK8sClient {
	return &AppK8sClient{
		k8sClient:   k8sClient,
		cache:       cache,
		cfg:         cfg,
		namespace:   namespace,
		subscribers: make(map[string]map[chan *flyteapp.WatchResponse]struct{}),
	}
}

// defaultOrg is the org returned for apps that have no org persisted on the Deployment.
const defaultOrg = secret.DefaultOrganization

// Deploy creates or updates the Deployment and ClusterIP Service for the given app.
func (c *AppK8sClient) Deploy(ctx context.Context, app *flyteapp.App) error {
	appID := app.GetMetadata().GetId()
	ns := c.namespace
	name := KServiceName(appID)

	if err := k8s.EnsureNamespaceExists(ctx, c.k8sClient, ns); err != nil {
		return fmt.Errorf("failed to ensure namespace %s: %w", ns, err)
	}

	deploy, svc, err := c.buildAppResources(app)
	if err != nil {
		return fmt.Errorf("failed to build resources for app %s: %w", name, err)
	}

	// Ensure the backing ClusterIP Service exists and is up to date.
	existingSvc := &corev1.Service{}
	err = c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, existingSvc)
	if k8serrors.IsNotFound(err) {
		if createErr := c.k8sClient.Create(ctx, svc); createErr != nil && !k8serrors.IsAlreadyExists(createErr) {
			return fmt.Errorf("failed to create Service %s: %w", name, createErr)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get Service %s: %w", name, err)
	} else {
		existingSvc.Spec.Ports = svc.Spec.Ports
		existingSvc.Spec.Selector = svc.Spec.Selector
		if existingSvc.Labels == nil {
			existingSvc.Labels = make(map[string]string)
		}
		for k, v := range svc.Labels {
			existingSvc.Labels[k] = v
		}
		if updateErr := c.k8sClient.Update(ctx, existingSvc); updateErr != nil {
			return fmt.Errorf("failed to update Service %s: %w", name, updateErr)
		}
	}

	existing := &appsv1.Deployment{}
	err = c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, existing)
	if k8serrors.IsNotFound(err) {
		if err := c.k8sClient.Create(ctx, deploy); err != nil {
			return fmt.Errorf("failed to create Deployment %s: %w", name, err)
		}
		logger.Infof(ctx, "Created App Deployment %s/%s", ns, name)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get Deployment %s: %w", name, err)
	}

	existingStopped := existing.Labels != nil && existing.Labels[labelAppStopped] == "true"

	if !existingStopped && existing.Annotations[annotationSpecSHA] == deploy.Annotations[annotationSpecSHA] {
		logger.Debugf(ctx, "App Deployment %s/%s spec unchanged, skipping update", ns, name)
		return nil
	}

	existing.Spec = deploy.Spec
	if existing.Labels == nil {
		existing.Labels = make(map[string]string)
	}
	for k, v := range deploy.Labels {
		existing.Labels[k] = v
	}
	delete(existing.Labels, labelAppStopped)
	delete(existing.Labels, labelKnativeVisibility)
	if existing.Annotations == nil {
		existing.Annotations = make(map[string]string)
	}
	for k, v := range deploy.Annotations {
		existing.Annotations[k] = v
	}
	if err := c.k8sClient.Update(ctx, existing); err != nil {
		return fmt.Errorf("failed to update Deployment %s: %w", name, err)
	}
	logger.Infof(ctx, "Updated App Deployment %s/%s", ns, name)
	return nil
}

// Stop scales the app Deployment to zero without deleting the Deployment.
func (c *AppK8sClient) Stop(ctx context.Context, appID *flyteapp.Identifier) error {
	ns := c.namespace
	name := KServiceName(appID)

	deploy := &appsv1.Deployment{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, deploy); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get Deployment %s to stop: %w", name, err)
	}

	if deploy.Labels == nil {
		deploy.Labels = make(map[string]string)
	}
	deploy.Labels[labelAppStopped] = "true"
	deploy.Labels[labelKnativeVisibility] = visibilityClusterLocal

	if deploy.Spec.Template.Annotations == nil {
		deploy.Spec.Template.Annotations = make(map[string]string)
	}
	deploy.Spec.Template.Annotations["autoscaling.knative.dev/min-scale"] = scaleZero
	deploy.Spec.Template.Annotations["autoscaling.knative.dev/initial-scale"] = scaleZero

	zero := int32(0)
	deploy.Spec.Replicas = &zero

	if err := c.k8sClient.Update(ctx, deploy); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to update Deployment %s to stop: %w", name, err)
	}

	// Delete any running pods immediately so stop takes effect without waiting.
	podList := &corev1.PodList{}
	if err := c.k8sClient.List(ctx, podList, client.InNamespace(ns), client.MatchingLabels{labelKnativeService: name}); err == nil {
		for i := range podList.Items {
			_ = c.k8sClient.Delete(ctx, &podList.Items[i])
		}
	}

	logger.Infof(ctx, "Stopped App Deployment %s/%s (scaled to zero)", ns, name)
	return nil
}

// Delete removes the Deployment and Service for the given app entirely.
func (c *AppK8sClient) Delete(ctx context.Context, appID *flyteapp.Identifier) error {
	ns := c.namespace
	name := KServiceName(appID)

	deploy := &appsv1.Deployment{}
	deploy.Name = name
	deploy.Namespace = ns
	if err := c.k8sClient.Delete(ctx, deploy); err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment %s: %w", name, err)
	}

	svc := &corev1.Service{}
	svc.Name = name
	svc.Namespace = ns
	if err := c.k8sClient.Delete(ctx, svc); err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service %s: %w", name, err)
	}

	logger.Infof(ctx, "Deleted App Deployment and Service %s/%s", ns, name)
	return nil
}

// StartWatching starts the Deployment informer and begins dispatching events
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
		return fmt.Errorf("shared cache is required for App Deployment informer")
	}

	return c.setupInformer(ctx)
}

func (c *AppK8sClient) setupInformer(ctx context.Context) error {
	informer, err := c.cache.GetInformer(ctx, &appsv1.Deployment{})
	if err != nil {
		return fmt.Errorf("failed to get Deployment informer: %w", err)
	}

	_, err = informer.AddEventHandler(toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok || !isManagedDeployment(deploy) {
				return
			}
			c.handleDeploymentEvent(ctx, deploy, k8swatch.Added)
		},
		UpdateFunc: func(_, newObj interface{}) {
			deploy, ok := newObj.(*appsv1.Deployment)
			if !ok || !isManagedDeployment(deploy) {
				return
			}
			c.handleDeploymentEvent(ctx, deploy, k8swatch.Modified)
		},
		DeleteFunc: func(obj interface{}) {
			if tombstone, ok := obj.(toolscache.DeletedFinalStateUnknown); ok {
				obj = tombstone.Obj
			}
			deploy, ok := obj.(*appsv1.Deployment)
			if !ok || !isManagedDeployment(deploy) {
				return
			}
			c.handleDeploymentEvent(ctx, deploy, k8swatch.Deleted)
		},
	})
	if err != nil {
		return fmt.Errorf("failed to add Deployment informer handler: %w", err)
	}

	return nil
}

// isManagedDeployment returns true if the Deployment has the flyte.org/app-managed label.
func isManagedDeployment(deploy *appsv1.Deployment) bool {
	return deploy.Labels[labelAppManaged] == "true"
}

// handleDeploymentEvent converts a Deployment event into a WatchResponse and notifies subscribers.
func (c *AppK8sClient) handleDeploymentEvent(ctx context.Context, deploy *appsv1.Deployment, eventType k8swatch.EventType) {
	app, err := c.deploymentToApp(ctx, deploy)
	if err != nil {
		return
	}

	appID := app.GetMetadata().GetId()

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

	c.notifySubscribers(ctx, appID.GetName(), resp)
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

// StopWatching stops the Deployment informer.
func (c *AppK8sClient) StopWatching() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watching && c.stopCh != nil {
		close(c.stopCh)
		c.watching = false
	}
}

// GetApp reads the Deployment and returns the full App including reconstructed Spec and live Status.
func (c *AppK8sClient) GetApp(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.App, error) {
	ns := c.namespace
	name := KServiceName(appID)
	deploy := &appsv1.Deployment{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, deploy); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("App Deployment %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get App Deployment %s: %w", name, err)
	}
	return c.deploymentToApp(ctx, deploy)
}

// specFromAnnotation deserializes the Spec stored in the flyte.org/spec annotation.
// Returns nil if the annotation is absent or malformed.
func specFromAnnotation(obj metav1.Object) *flyteapp.Spec {
	b64, ok := obj.GetAnnotations()[annotationSpec]
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
	ns := c.namespace

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

	list := &appsv1.DeploymentList{}
	if err := c.k8sClient.List(ctx, list, listOpts...); err != nil {
		return nil, "", fmt.Errorf("failed to list App Deployments for %s/%s: %w", project, domain, err)
	}

	apps := make([]*flyteapp.App, 0, len(list.Items))
	for i := range list.Items {
		a, err := c.deploymentToApp(ctx, &list.Items[i])
		if err != nil {
			logger.Warnf(ctx, "Skipping Deployment %s: failed to convert to app: %v", list.Items[i].Name, err)
			continue
		}
		apps = append(apps, a)
	}
	return apps, list.Continue, nil
}

// PublicIngress returns the deterministic public URL for an app.
// When BaseDomain contains a path prefix (e.g. "flyte-console.caml.calicolabs.com/apps"),
// routes are formatted as "{scheme}://{base_domain}/{kservice_name}/" so all apps can be
// served behind a single IAP-protected ingress host without wildcard DNS certificates.
func (c *AppK8sClient) PublicIngress(id *flyteapp.Identifier) *flyteapp.Ingress {
	if c.cfg.BaseDomain == "" {
		return nil
	}
	scheme := c.cfg.Scheme
	if scheme == "" {
		scheme = "https"
	}
	ksvcName := KServiceName(id)
	if strings.Contains(c.cfg.BaseDomain, "/") {
		base := strings.TrimRight(c.cfg.BaseDomain, "/")
		return &flyteapp.Ingress{PublicUrl: fmt.Sprintf("%s://%s/%s/", scheme, base, ksvcName)}
	}
	host := strings.ToLower(fmt.Sprintf("%s.%s", ksvcName, c.cfg.BaseDomain))
	url := scheme + "://" + host
	if c.cfg.IngressAppsPort != 0 {
		url += fmt.Sprintf(":%d", c.cfg.IngressAppsPort)
	}
	return &flyteapp.Ingress{PublicUrl: url}
}

// --- Helpers ---

// KServiceName returns the K8s Deployment/Service name for an app.
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

// renderNamespacedSuffix substitutes {{ project }} and {{ domain }}
// in a template string to produce the name suffix used in INTERNAL_APP_ENDPOINT_PATTERN.
func renderNamespacedSuffix(tmpl, project, domain string) string {
	return strings.NewReplacer(
		"{{ project }}", strings.ToLower(project),
		"{{ domain }}", strings.ToLower(domain),
	).Replace(tmpl)
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
	return hex.EncodeToString(sum[:8])
}

func primaryAppPort(podSpec *corev1.PodSpec) int32 {
	for _, c := range podSpec.Containers {
		for _, p := range c.Ports {
			if p.ContainerPort > 0 {
				return p.ContainerPort
			}
		}
	}
	return defaultAppPort
}

// AppleSiliconWorkloadID scopes an Apple Silicon app's host workload to one revision of its spec. During a
// rollout the new revision's pods then get their own host process instead of the old revision's, and an old
// pod shutting down can't tear down the new revision's. Replicas of one revision share the host process.
func AppleSiliconWorkloadID(name, specSHA string) string {
	return "app-" + name + "-" + specSHA
}

func resolveRawOutputDataPrefix(appID *flyteapp.Identifier, podSpec *corev1.PodSpec, cfg *config.InternalAppConfig) string {
	for _, c := range podSpec.Containers {
		for _, e := range c.Env {
			if (e.Name == "DUNE_ARTIFACTS_BUCKET" || e.Name == "DUNE_PRIMARY_BUCKET") && e.Value != "" {
				return fmt.Sprintf("gs://%s/flyte-apps/%s/%s/%s", e.Value, appID.GetProject(), appID.GetDomain(), appID.GetName())
			}
		}
	}
	if cfg != nil && cfg.DefaultEnvVars != nil {
		if base := cfg.DefaultEnvVars["RAW_OUTPUT_DATA_PREFIX"]; base != "" {
			return fmt.Sprintf("%s/%s/%s/%s", strings.TrimRight(base, "/"), appID.GetProject(), appID.GetDomain(), appID.GetName())
		}
	}
	return ""
}

// buildAppResources constructs a Kubernetes Deployment and ClusterIP Service from an App proto.
func (c *AppK8sClient) buildAppResources(app *flyteapp.App) (*appsv1.Deployment, *corev1.Service, error) {
	appID := app.GetMetadata().GetId()
	spec := app.GetSpec()
	name := KServiceName(appID)
	ns := c.namespace

	specBytes, err := marshalSpec(spec)
	if err != nil {
		return nil, nil, err
	}
	sha := specSHA(specBytes)

	podSpec, err := buildPodSpec(spec)
	if err != nil {
		return nil, nil, err
	}
	// Apple Silicon apps run their ML work on the devbox host's Metal GPU via the Apple Silicon
	// bridge; on any backend without one this fails the deploy with APPLE_SILICON_NOT_AVAILABLE.
	if flytek8s.IsAppleSilicon(spec.GetExtendedResources().GetGpuAccelerator()) && len(podSpec.Containers) > 0 {
		if err := flytek8s.ApplyAppleSiliconBridge(&podSpec, primaryContainerName(spec, podSpec), AppleSiliconWorkloadID(name, sha)); err != nil {
			return nil, nil, err
		}
	}
	if sa := spec.GetSecurityContext().GetRunAs().GetK8SServiceAccount(); sa != "" {
		podSpec.ServiceAccountName = sa
	} else if podSpec.ServiceAccountName == "" && c.cfg.DefaultServiceAccount != "" {
		podSpec.ServiceAccountName = c.cfg.DefaultServiceAccount
	}

	if len(c.cfg.DefaultEnvVars) > 0 && len(podSpec.Containers) > 0 {
		defaults := make([]corev1.EnvVar, 0, len(c.cfg.DefaultEnvVars))
		for k, v := range c.cfg.DefaultEnvVars {
			defaults = append(defaults, corev1.EnvVar{Name: k, Value: v})
		}
		podSpec.Containers[0].Env = append(defaults, podSpec.Containers[0].Env...)
	}

	if len(podSpec.Containers) > 0 {
		suffix := renderNamespacedSuffix(c.cfg.NamespacedNameSuffixTemplate, appID.GetProject(), appID.GetDomain())
		podSpec.Containers[0].Env = append(podSpec.Containers[0].Env, corev1.EnvVar{
			Name:  "INTERNAL_APP_ENDPOINT_PATTERN",
			Value: fmt.Sprintf("http://{app_fqdn}-%s.%s.svc.cluster.local", suffix, ns),
		})
		podSpec.Containers[0].Env = append(podSpec.Containers[0].Env,
			corev1.EnvVar{Name: "FLYTE_INTERNAL_EXECUTION_PROJECT", Value: appID.GetProject()},
			corev1.EnvVar{Name: "FLYTE_INTERNAL_EXECUTION_DOMAIN", Value: appID.GetDomain()},
		)
		// A deployment that opted out of telemetry opts out its apps too. Last,
		// replacing any value the app's own env set, so it cannot undo it.
		if telemetry.Disabled() {
			env := podSpec.Containers[0].Env[:0]
			for _, ev := range podSpec.Containers[0].Env {
				if ev.Name != telemetry.DisableEnvVar {
					env = append(env, ev)
				}
			}
			podSpec.Containers[0].Env = append(env, corev1.EnvVar{Name: telemetry.DisableEnvVar, Value: "1"})
		}
	}

	rawDataPrefix := resolveRawOutputDataPrefix(appID, &podSpec, c.cfg)
	for i := range podSpec.Containers {
		for j, arg := range podSpec.Containers[i].Command {
			if strings.Contains(arg, "{{.rawOutputDataPrefix}}") {
				podSpec.Containers[i].Command[j] = strings.ReplaceAll(arg, "{{.rawOutputDataPrefix}}", rawDataPrefix)
			}
		}
		for j, arg := range podSpec.Containers[i].Args {
			if strings.Contains(arg, "{{.rawOutputDataPrefix}}") {
				podSpec.Containers[i].Args[j] = strings.ReplaceAll(arg, "{{.rawOutputDataPrefix}}", rawDataPrefix)
			}
		}
	}

	templateAnnotations := buildAutoscalingAnnotations(spec, c.cfg)
	templateLabels := map[string]string{
		labelKnativeService:       name,
		labelAppManaged:           "true",
		flytek8s.ManagedLabelKey:  flytek8s.ManagedLabelValue,
		labelProject:              appID.GetProject(),
		labelDomain:               appID.GetDomain(),
		labelAppName:              appID.GetName(),
	}

	if p := spec.GetPod(); p != nil && p.GetMetadata() != nil {
		for k, v := range p.GetMetadata().GetLabels() {
			if k == "kueue.x-k8s.io/queue-name" {
				continue
			}
			if _, exists := templateLabels[k]; !exists {
				templateLabels[k] = v
			}
		}
		for k, v := range p.GetMetadata().GetAnnotations() {
			templateAnnotations[k] = v
		}
	}

	if securityContext := spec.GetSecurityContext(); securityContext != nil && len(securityContext.GetSecrets()) > 0 {
		templateLabels[secret.OrganizationLabel] = secret.DefaultOrganization
		templateLabels[secret.ProjectLabel] = appID.GetProject()
		templateLabels[secret.DomainLabel] = appID.GetDomain()
		templateLabels[secretUtils.PodLabel] = secretUtils.PodLabelValue
		secretsMap, err := secretUtils.MarshalSecretsToMapStrings(securityContext.GetSecrets())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal secrets: %w", err)
		}
		for k, v := range secretsMap {
			templateAnnotations[k] = v
		}
	}

	replicas := int32(1)
	if r := spec.GetAutoscaling().GetReplicas(); r != nil && r.GetMin() > 0 {
		replicas = int32(r.GetMin())
	}
	if spec.GetDesiredState() == flyteapp.Spec_DESIRED_STATE_STOPPED {
		replicas = 0
	}

	objLabels := map[string]string{
		labelAppManaged: "true",
		labelProject:    appID.GetProject(),
		labelDomain:     appID.GetDomain(),
		labelAppName:    appID.GetName(),
	}
	if spec.GetDesiredState() == flyteapp.Spec_DESIRED_STATE_STOPPED {
		objLabels[labelAppStopped] = "true"
		objLabels[labelKnativeVisibility] = visibilityClusterLocal
	}

	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    objLabels,
			Annotations: map[string]string{
				annotationSpecSHA: sha,
				annotationAppID:   fmt.Sprintf("%s/%s/%s", appID.GetProject(), appID.GetDomain(), appID.GetName()),
				annotationAppOrg:  appID.GetOrg(),
				annotationSpec:    base64.StdEncoding.EncodeToString(specBytes),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKnativeService: name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      templateLabels,
					Annotations: templateAnnotations,
				},
				Spec: podSpec,
			},
		},
	}

	appPort := primaryAppPort(&podSpec)
	svcPorts := []corev1.ServicePort{
		{
			Name:       "http",
			Port:       80,
			TargetPort: intstr.FromInt32(appPort),
			Protocol:   corev1.ProtocolTCP,
		},
		{
			Name:       "http-app",
			Port:       8080,
			TargetPort: intstr.FromInt32(appPort),
			Protocol:   corev1.ProtocolTCP,
		},
	}
	if appPort != 80 && appPort != 8080 {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:       "http-custom",
			Port:       appPort,
			TargetPort: intstr.FromInt32(appPort),
			Protocol:   corev1.ProtocolTCP,
		})
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels: map[string]string{
				labelAppManaged: "true",
				labelProject:    appID.GetProject(),
				labelDomain:     appID.GetDomain(),
				labelAppName:    appID.GetName(),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				labelKnativeService: name,
			},
			Ports: svcPorts,
		},
	}

	return deploy, svc, nil
}

// buildPodSpec constructs a corev1.PodSpec from an App Spec.
// Supports Container and K8sPod payloads.
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
		pod := p.Pod
		if pod == nil || pod.GetPodSpec() == nil {
			return corev1.PodSpec{}, fmt.Errorf("K8sPod app payload has no pod spec")
		}
		var podSpec corev1.PodSpec
		if err := utils.UnmarshalStructToObj(pod.GetPodSpec(), &podSpec); err != nil {
			return corev1.PodSpec{}, fmt.Errorf("failed to unmarshal K8sPod spec: %w", err)
		}
		if podSpec.EnableServiceLinks == nil {
			podSpec.EnableServiceLinks = boolPtr(false)
		}
		return podSpec, nil

	default:
		return corev1.PodSpec{}, fmt.Errorf("app spec has no payload (container or pod required)")
	}
}

// primaryContainerName returns the app's serving container: the one a K8sPod payload names, or
// the first container otherwise.
func primaryContainerName(spec *flyteapp.Spec, podSpec corev1.PodSpec) string {
	if name := spec.GetPod().GetPrimaryContainerName(); name != "" {
		return name
	}
	return podSpec.Containers[0].Name
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

// buildAutoscalingAnnotations returns the autoscaling annotations for the pod template.
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

// deploymentToStatus maps a Deployment's conditions and replica counts to a flyteapp.Status proto.
func (c *AppK8sClient) deploymentToStatus(ctx context.Context, deploy *appsv1.Deployment) *flyteapp.Status {
	if deploy.Labels != nil && deploy.Labels[labelAppStopped] == "true" {
		return &flyteapp.Status{
			Conditions: []*flyteapp.Condition{{
				DeploymentStatus:   flyteapp.Status_DEPLOYMENT_STATUS_STOPPED,
				Message:            "App scaled to zero",
				LastTransitionTime: timestamppb.Now(),
			}},
		}
	}

	var availableCond *appsv1.DeploymentCondition
	var failureCond *appsv1.DeploymentCondition
	var progressingCond *appsv1.DeploymentCondition
	for i := range deploy.Status.Conditions {
		cond := &deploy.Status.Conditions[i]
		switch cond.Type {
		case appsv1.DeploymentAvailable:
			availableCond = cond
		case appsv1.DeploymentReplicaFailure:
			failureCond = cond
		case appsv1.DeploymentProgressing:
			progressingCond = cond
		}
	}

	serviceReady := (availableCond != nil && availableCond.Status == corev1.ConditionTrue) || deploy.Status.ReadyReplicas > 0
	serviceFailed := (failureCond != nil && failureCond.Status == corev1.ConditionTrue) || (progressingCond != nil && progressingCond.Status == corev1.ConditionFalse)

	var conditions []*flyteapp.Condition
	if serviceFailed {
		msg := "Deployment failed"
		ts := timestamppb.Now()
		if failureCond != nil && failureCond.Status == corev1.ConditionTrue {
			if failureCond.Reason != "" && failureCond.Message != "" {
				msg = fmt.Sprintf("%s: %s", failureCond.Reason, failureCond.Message)
			} else if failureCond.Message != "" {
				msg = failureCond.Message
			}
			if !failureCond.LastTransitionTime.IsZero() {
				ts = timestamppb.New(failureCond.LastTransitionTime.Time)
			}
		} else if progressingCond != nil {
			if progressingCond.Reason != "" && progressingCond.Message != "" {
				msg = fmt.Sprintf("%s: %s", progressingCond.Reason, progressingCond.Message)
			} else if progressingCond.Message != "" {
				msg = progressingCond.Message
			}
			if !progressingCond.LastTransitionTime.IsZero() {
				ts = timestamppb.New(progressingCond.LastTransitionTime.Time)
			}
		}
		conditions = append(conditions, &flyteapp.Condition{
			DeploymentStatus:   flyteapp.Status_DEPLOYMENT_STATUS_FAILED,
			Message:            msg,
			LastTransitionTime: ts,
		})
	} else if serviceReady {
		msg := "Service is ready"
		ts := timestamppb.Now()
		if availableCond != nil {
			if availableCond.Message != "" {
				msg = availableCond.Message
			}
			if !availableCond.LastTransitionTime.IsZero() {
				ts = timestamppb.New(availableCond.LastTransitionTime.Time)
			}
		}
		conditions = append(conditions, &flyteapp.Condition{
			DeploymentStatus:   flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE,
			Message:            msg,
			LastTransitionTime: ts,
		})
	} else if len(deploy.Status.Conditions) > 0 {
		msg := "Deployment is progressing"
		ts := timestamppb.Now()
		if progressingCond != nil {
			if progressingCond.Reason != "" && progressingCond.Message != "" {
				msg = fmt.Sprintf("%s: %s", progressingCond.Reason, progressingCond.Message)
			} else if progressingCond.Message != "" {
				msg = progressingCond.Message
			}
			if !progressingCond.LastTransitionTime.IsZero() {
				ts = timestamppb.New(progressingCond.LastTransitionTime.Time)
			}
		}
		conditions = append(conditions, &flyteapp.Condition{
			DeploymentStatus:   flyteapp.Status_DEPLOYMENT_STATUS_DEPLOYING,
			Message:            msg,
			LastTransitionTime: ts,
		})
	}

	status := &flyteapp.Status{Conditions: conditions}

	if appIDStr := deploy.Annotations[annotationAppID]; appIDStr != "" {
		parts := strings.SplitN(appIDStr, "/", 3)
		if len(parts) == 3 {
			appID := &flyteapp.Identifier{Project: parts[0], Domain: parts[1], Name: parts[2]}
			status.Ingress = c.PublicIngress(appID)
		}
	}

	if deploy.Status.ReadyReplicas > 0 {
		status.CurrentReplicas = uint32(deploy.Status.ReadyReplicas)
	} else if deploy.Status.Replicas > 0 {
		status.CurrentReplicas = uint32(deploy.Status.Replicas)
	}

	status.K8SMetadata = &flyteapp.K8SMetadata{
		Namespace: deploy.Namespace,
	}

	if !deploy.CreationTimestamp.IsZero() {
		status.CreatedAt = timestamppb.New(deploy.CreationTimestamp.Time)
	}

	return status
}

// GetReplicas lists the pods currently backing the given app.
func (c *AppK8sClient) GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error) {
	ns := c.namespace
	name := KServiceName(appID)

	labels := client.MatchingLabels{labelKnativeService: name}
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

// DeleteReplica force-deletes a specific pod. The Deployment controller will schedule a replacement automatically.
func (c *AppK8sClient) DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error {
	ns := c.namespace
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

// deploymentToApp reconstructs a flyteapp.App from a Deployment by reading the
// app identifier from annotations and the live status from Deployment conditions.
func (c *AppK8sClient) deploymentToApp(ctx context.Context, deploy *appsv1.Deployment) (*flyteapp.App, error) {
	appIDStr, ok := deploy.Annotations[annotationAppID]
	if !ok {
		return nil, fmt.Errorf("Deployment %s missing %s annotation", deploy.Name, annotationAppID)
	}

	parts := strings.SplitN(appIDStr, "/", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("Deployment %s has malformed %s annotation: %q", deploy.Name, annotationAppID, appIDStr)
	}

	org := deploy.Annotations[annotationAppOrg]
	if org == "" {
		org = defaultOrg
	}
	appID := &flyteapp.Identifier{
		Org:     org,
		Project: parts[0],
		Domain:  parts[1],
		Name:    parts[2],
	}

	spec := specFromAnnotation(deploy)
	if spec != nil && deploy.Labels[labelAppStopped] == "true" {
		spec.DesiredState = flyteapp.Spec_DESIRED_STATE_STOPPED
	}

	return &flyteapp.App{
		Metadata: &flyteapp.Meta{
			Id: appID,
		},
		Spec:   spec,
		Status: c.deploymentToStatus(ctx, deploy),
	}, nil
}
