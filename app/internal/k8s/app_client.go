package k8s

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8swatch "k8s.io/apimachinery/pkg/watch"
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

	annotationSpecSHA = "flyte.org/spec-sha"
	annotationAppID   = "flyte.org/app-id"

	maxScaleZero = "0"

	// maxKServiceNameLen is the Kubernetes DNS label limit.
	maxKServiceNameLen = 63
)

var (
	traefikIngressRouteGVK = schema.GroupVersionKind{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute"}
	traefikMiddlewareGVK   = schema.GroupVersionKind{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware"}
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
	// Returns a not-found error (checkable with k8serrors.IsNotFound) if the KService does not exist.
	GetStatus(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.Status, error)

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

	// Watch returns a channel of WatchResponse events for KServices matching the
	// given project/domain scope. If appName is non-empty, only events for that
	// specific app are returned. The channel is closed when ctx is cancelled.
	Watch(ctx context.Context, project, domain, appName string) (<-chan *flyteapp.WatchResponse, error)
}

// AppK8sClient implements AppK8sClientInterface using controller-runtime.
type AppK8sClient struct {
	k8sClient client.WithWatch
	cache     ctrlcache.Cache
	cfg       *config.InternalAppConfig
}

// NewAppK8sClient creates a new AppK8sClient.
func NewAppK8sClient(k8sClient client.WithWatch, cache ctrlcache.Cache, cfg *config.InternalAppConfig) *AppK8sClient {
	return &AppK8sClient{
		k8sClient: k8sClient,
		cache:     cache,
		cfg:       cfg,
	}
}

// appNamespace returns the K8s namespace for a given project/domain pair.
// Follows the same convention as the Actions and Secret services: "{project}-{domain}".
func appNamespace(project, domain string) string {
	return fmt.Sprintf("%s-%s", project, domain)
}

// Deploy creates or updates the KService for the given app.
func (c *AppK8sClient) Deploy(ctx context.Context, app *flyteapp.App) error {
	appID := app.GetMetadata().GetId()
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	name := kserviceName(appID)

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
		return c.deployIngress(ctx, app)
	}
	if err != nil {
		return fmt.Errorf("failed to get KService %s: %w", name, err)
	}

	// Skip KService update if spec has not changed, but still ensure ingress exists
	// (it may be missing if the app was deployed before ingress support was added).
	if existing.Annotations[annotationSpecSHA] == ksvc.Annotations[annotationSpecSHA] {
		logger.Debugf(ctx, "KService %s/%s spec unchanged, skipping update", ns, name)
		return c.deployIngress(ctx, app)
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
	return c.deployIngress(ctx, app)
}

// Stop sets max-scale=0 on the KService, scaling it to zero without deleting it.
func (c *AppK8sClient) Stop(ctx context.Context, appID *flyteapp.Identifier) error {
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	name := kserviceName(appID)
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
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	name := kserviceName(appID)
	ksvc := &servingv1.Service{}
	ksvc.Name = name
	ksvc.Namespace = ns
	if err := c.k8sClient.Delete(ctx, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			c.deleteIngress(ctx, appID)
			return nil
		}
		return fmt.Errorf("failed to delete KService %s: %w", name, err)
	}
	logger.Infof(ctx, "Deleted KService %s/%s", ns, name)
	c.deleteIngress(ctx, appID)
	return nil
}

// watchBackoff controls the reconnect timing for Watch. Declared as vars so
// tests can override them without sleeping.
var (
	watchBackoffInitial = 1 * time.Second
	watchBackoffMax     = 30 * time.Second
	watchBackoffFactor  = 2.0
)

// watchState holds mutable reconnect state for a single Watch call.
// It is goroutine-local — no mutex needed.
type watchState struct {
	// lastResourceVersion is the most recent RV seen from any event or Bookmark.
	// Passed to openWatch on reconnect so K8s resumes from exactly where we left off.
	lastResourceVersion string
	backoff             time.Duration
	consecutiveErrors   int
}

func (s *watchState) nextBackoff() time.Duration {
	d := s.backoff
	if d == 0 {
		d = watchBackoffInitial
	}
	s.backoff = time.Duration(math.Min(float64(d)*watchBackoffFactor, float64(watchBackoffMax)))
	return d
}

func (s *watchState) resetBackoff() {
	s.backoff = watchBackoffInitial
	s.consecutiveErrors = 0
}

// Watch returns a channel of WatchResponse events for KServices in the given
// project/domain scope. If appName is non-empty, only events for that specific
// app are returned. The channel is closed only when ctx is cancelled.
//
// The goroutine reconnects transparently when the underlying K8s watch closes
// unexpectedly, tracking resourceVersion to resume without gaps or replays.
func (c *AppK8sClient) Watch(ctx context.Context, project, domain, appName string) (<-chan *flyteapp.WatchResponse, error) {
	ns := appNamespace(project, domain)
	labels := map[string]string{labelAppManaged: "true"}
	if appName != "" {
		labels[labelAppName] = strings.ToLower(appName)
	}

	// Open the first watcher eagerly so initial errors (RBAC, missing CRD) are
	// returned synchronously before spawning the goroutine.
	watcher, err := c.openWatch(ctx, ns, labels, "")
	if err != nil {
		return nil, err
	}

	ch := make(chan *flyteapp.WatchResponse, 64)
	go c.watchLoop(ctx, ns, labels, watcher, ch)
	return ch, nil
}

// openWatch starts a K8s watch from resourceVersion (empty = watch from now).
func (c *AppK8sClient) openWatch(ctx context.Context, ns string, labels map[string]string, resourceVersion string) (k8swatch.Interface, error) {
	opts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels(labels),
	}
	if resourceVersion != "" {
		opts = append(opts, &client.ListOptions{
			Raw: &metav1.ListOptions{
				ResourceVersion:     resourceVersion,
				AllowWatchBookmarks: true,
			},
		})
	}
	watcher, err := c.k8sClient.Watch(ctx, &servingv1.ServiceList{}, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to start KService watch in namespace %s: %w", ns, err)
	}
	return watcher, nil
}

// watchLoop is the reconnect loop. It drains watcher until it closes, then
// reopens with exponential backoff. Closes ch only when ctx is cancelled.
func (c *AppK8sClient) watchLoop(
	ctx context.Context,
	ns string,
	labels map[string]string,
	watcher k8swatch.Interface,
	ch chan<- *flyteapp.WatchResponse,
) {
	defer close(ch)
	defer watcher.Stop()

	state := &watchState{backoff: watchBackoffInitial}

	for {
		reconnect := c.drainWatcher(ctx, watcher, ch, state)
		if !reconnect {
			return // ctx cancelled
		}

		watcher.Stop()
		state.consecutiveErrors++
		delay := state.nextBackoff()
		logger.Warnf(ctx, "KService watch in namespace %s closed unexpectedly (attempt %d); reconnecting in %v",
			ns, state.consecutiveErrors, delay)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		newWatcher, err := c.openWatch(ctx, ns, labels, state.lastResourceVersion)
		if err != nil {
			logger.Errorf(ctx, "Failed to reopen KService watch in namespace %s: %v", ns, err)
			// Use an immediately-closed watcher so the loop retries with further backoff.
			watcher = k8swatch.NewEmptyWatch()
			continue
		}
		watcher = newWatcher
	}
}

// drainWatcher processes events from watcher until the channel closes or ctx is done.
// Returns true if reconnect is needed, false if ctx was cancelled.
func (c *AppK8sClient) drainWatcher(
	ctx context.Context,
	watcher k8swatch.Interface,
	ch chan<- *flyteapp.WatchResponse,
	state *watchState,
) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return true
			}

			c.updateResourceVersion(event, state)

			switch event.Type {
			case k8swatch.Error:
				if status, ok := event.Object.(*metav1.Status); ok {
					logger.Warnf(ctx, "KService watch received error event (code=%d reason=%s): %s; will reconnect",
						status.Code, status.Reason, status.Message)
				} else {
					logger.Warnf(ctx, "KService watch received error event (type %T); will reconnect", event.Object)
				}
				return true
			case k8swatch.Bookmark:
				// resourceVersion already updated — nothing to forward.
				state.resetBackoff()
			default:
				resp := c.kserviceEventToWatchResponse(ctx, event)
				if resp == nil {
					continue
				}
				state.resetBackoff()
				select {
				case ch <- resp:
				case <-ctx.Done():
					return false
				}
			}
		}
	}
}

// updateResourceVersion extracts and stores the latest resourceVersion from a watch event.
// Called before event type dispatch so both normal events and Bookmarks checkpoint the position.
func (c *AppK8sClient) updateResourceVersion(event k8swatch.Event, state *watchState) {
	switch event.Type {
	case k8swatch.Added, k8swatch.Modified, k8swatch.Deleted, k8swatch.Bookmark:
		if ksvc, ok := event.Object.(*servingv1.Service); ok {
			if rv := ksvc.GetResourceVersion(); rv != "" {
				state.lastResourceVersion = rv
			}
		}
	}
}

// kserviceEventToWatchResponse maps a K8s watch event to a flyteapp.WatchResponse.
// Returns nil for event types that should not be forwarded (Error, Bookmark).
func (c *AppK8sClient) kserviceEventToWatchResponse(ctx context.Context, event k8swatch.Event) *flyteapp.WatchResponse {
	ksvc, ok := event.Object.(*servingv1.Service)
	if !ok {
		return nil
	}
	app, err := c.kserviceToApp(ctx, ksvc)
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
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	name := kserviceName(appID)
	ksvc := &servingv1.Service{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, ksvc); err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("KService %s not found: %w", name, err)
		}
		return nil, fmt.Errorf("failed to get KService %s: %w", name, err)
	}
	return c.kserviceToStatus(ctx, ksvc), nil
}

// List returns apps for the given project/domain scope with optional pagination.
func (c *AppK8sClient) List(ctx context.Context, project, domain string, limit uint32, token string) ([]*flyteapp.App, string, error) {
	ns := appNamespace(project, domain)

	listOpts := []client.ListOption{
		client.InNamespace(ns),
		client.MatchingLabels{labelAppManaged: "true"},
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
// logic as the service layer so GetStatus/List/Watch are consistent with Create.
func (c *AppK8sClient) publicIngress(id *flyteapp.Identifier) *flyteapp.Ingress {
	if c.cfg.IngressEnabled && c.cfg.IngressBaseURL != "" {
		return &flyteapp.Ingress{
			PublicUrl: fmt.Sprintf("%s/%s/%s/%s",
				strings.TrimRight(c.cfg.IngressBaseURL, "/"),
				id.GetProject(), id.GetDomain(), id.GetName(),
			),
		}
	}
	if c.cfg.BaseDomain == "" {
		return nil
	}
	scheme := c.cfg.Scheme
	if scheme == "" {
		scheme = "https"
	}
	host := strings.ToLower(fmt.Sprintf("%s-%s-%s.%s",
		id.GetName(), id.GetProject(), id.GetDomain(), c.cfg.BaseDomain))
	return &flyteapp.Ingress{PublicUrl: scheme + "://" + host}
}

// deployIngress creates or updates the Traefik resources for the given app:
//   - strip-<name>  Middleware: strips /<project>/<domain>/<app> prefix
//   - host-<name>   Middleware: rewrites Host to the Knative hostname so Kourier
//     can route the request (Knative's K8s service is ExternalName → Kourier)
//   - app-<name>    IngressRoute: matches PathPrefix and chains both middlewares
//
// No-ops when IngressEnabled is false.
func (c *AppK8sClient) deployIngress(ctx context.Context, app *flyteapp.App) error {
	if !c.cfg.IngressEnabled {
		return nil
	}
	appID := app.GetMetadata().GetId()
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	ksvcName := kserviceName(appID)
	stripName := "strip-" + ksvcName
	hostName := "host-" + ksvcName
	routeName := "app-" + ksvcName
	pathPrefix := fmt.Sprintf("/%s/%s/%s", appID.GetProject(), appID.GetDomain(), appID.GetName())
	entryPoint := c.cfg.IngressEntryPoint
	if entryPoint == "" {
		entryPoint = "web"
	}

	// Knative's K8s service is ExternalName → kourier-internal. Kourier routes by
	// Host header, so we must rewrite it to the deterministic Knative hostname.
	// hostname pattern matches the configured domain-template: {name}-{namespace}.{domain}
	knativeHost := fmt.Sprintf("%s-%s.%s", ksvcName, ns, c.cfg.BaseDomain)

	stripMW := &unstructured.Unstructured{}
	stripMW.SetGroupVersionKind(traefikMiddlewareGVK)
	stripMW.SetName(stripName)
	stripMW.SetNamespace(ns)
	stripMW.Object["spec"] = map[string]interface{}{
		"stripPrefix": map[string]interface{}{
			"prefixes": []interface{}{pathPrefix},
		},
	}

	hostMW := &unstructured.Unstructured{}
	hostMW.SetGroupVersionKind(traefikMiddlewareGVK)
	hostMW.SetName(hostName)
	hostMW.SetNamespace(ns)
	hostMW.Object["spec"] = map[string]interface{}{
		"headers": map[string]interface{}{
			"customRequestHeaders": map[string]interface{}{
				"Host": knativeHost,
			},
		},
	}

	ir := &unstructured.Unstructured{}
	ir.SetGroupVersionKind(traefikIngressRouteGVK)
	ir.SetName(routeName)
	ir.SetNamespace(ns)
	ir.Object["spec"] = map[string]interface{}{
		"entryPoints": []interface{}{entryPoint},
		"routes": []interface{}{
			map[string]interface{}{
				"match": fmt.Sprintf("PathPrefix(`%s`)", pathPrefix),
				"kind":  "Rule",
				"middlewares": []interface{}{
					map[string]interface{}{"name": stripName},
					map[string]interface{}{"name": hostName},
				},
				"services": []interface{}{
					map[string]interface{}{
						"name": ksvcName,
						"port": int64(80),
					},
				},
			},
		},
	}

	for _, obj := range []*unstructured.Unstructured{stripMW, hostMW, ir} {
		existing := &unstructured.Unstructured{}
		existing.SetGroupVersionKind(obj.GroupVersionKind())
		err := c.k8sClient.Get(ctx, client.ObjectKey{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)
		if k8serrors.IsNotFound(err) {
			if createErr := c.k8sClient.Create(ctx, obj); createErr != nil {
				return fmt.Errorf("failed to create %s %s/%s: %w", obj.GetKind(), ns, obj.GetName(), createErr)
			}
			logger.Infof(ctx, "Created %s %s/%s", obj.GetKind(), ns, obj.GetName())
		} else if err != nil {
			return fmt.Errorf("failed to get %s %s/%s: %w", obj.GetKind(), ns, obj.GetName(), err)
		} else {
			obj.SetResourceVersion(existing.GetResourceVersion())
			if updateErr := c.k8sClient.Update(ctx, obj); updateErr != nil {
				return fmt.Errorf("failed to update %s %s/%s: %w", obj.GetKind(), ns, obj.GetName(), updateErr)
			}
			logger.Infof(ctx, "Updated %s %s/%s", obj.GetKind(), ns, obj.GetName())
		}
	}
	return nil
}

// deleteIngress removes the Traefik IngressRoute and both Middlewares for the given app.
// Errors are logged as warnings — ingress cleanup failure does not block app deletion.
func (c *AppK8sClient) deleteIngress(ctx context.Context, appID *flyteapp.Identifier) {
	if !c.cfg.IngressEnabled {
		return
	}
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	ksvcName := kserviceName(appID)

	type resource struct {
		gvk  schema.GroupVersionKind
		name string
	}
	resources := []resource{
		{traefikIngressRouteGVK, "app-" + ksvcName},
		{traefikMiddlewareGVK, "strip-" + ksvcName},
		{traefikMiddlewareGVK, "host-" + ksvcName},
	}
	for _, r := range resources {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(r.gvk)
		obj.SetName(r.name)
		obj.SetNamespace(ns)
		if err := c.k8sClient.Delete(ctx, obj); err != nil && !k8serrors.IsNotFound(err) {
			logger.Warnf(ctx, "Failed to delete %s %s/%s: %v", r.gvk.Kind, ns, r.name, err)
		} else {
			logger.Infof(ctx, "Deleted %s %s/%s", r.gvk.Kind, ns, r.name)
		}
	}
}

// --- Helpers ---

// kserviceName returns the KService name for an app. Since each app is deployed
// to its own project/domain namespace, the name only needs to be unique within
// that namespace — the app name alone suffices.
// Names are lower-cased and capped at 63 chars (K8s DNS label limit). For names
// that exceed 63 chars, the first 54 chars are kept and an 8-char SHA256 suffix
// is appended to avoid collisions between names with a long common prefix.
func kserviceName(id *flyteapp.Identifier) string {
	name := strings.ToLower(id.GetName())
	if len(name) <= maxKServiceNameLen {
		return name
	}
	sum := sha256.Sum256([]byte(name))
	suffix := hex.EncodeToString(sum[:4]) // 4 bytes = 8 hex chars
	return name[:maxKServiceNameLen-9] + "-" + suffix
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
	ns := appNamespace(appID.GetProject(), appID.GetDomain())

	sha, err := specSHA(spec)
	if err != nil {
		return nil, err
	}

	podSpec, err := buildPodSpec(spec)
	if err != nil {
		return nil, err
	}
	// Inject cluster-level default env vars (e.g. _U_EP_OVERRIDE) before user vars
	// so they can be overridden by app-specific env vars if needed.
	if len(c.cfg.DefaultEnvVars) > 0 && len(podSpec.Containers) > 0 {
		defaults := make([]corev1.EnvVar, 0, len(c.cfg.DefaultEnvVars))
		for _, e := range c.cfg.DefaultEnvVars {
			defaults = append(defaults, corev1.EnvVar{Name: e.Name, Value: e.Value})
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

	return status
}

// GetReplicas lists the pods currently backing the given app.
func (c *AppK8sClient) GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error) {
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
	podList := &corev1.PodList{}
	if err := c.k8sClient.List(ctx, podList,
		client.InNamespace(ns),
		client.MatchingLabels{labelAppName: appID.GetName()},
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
	appID := replicaID.GetAppId()
	ns := appNamespace(appID.GetProject(), appID.GetDomain())
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

	appID := &flyteapp.Identifier{
		Project: parts[0],
		Domain:  parts[1],
		Name:    parts[2],
	}

	return &flyteapp.App{
		Metadata: &flyteapp.Meta{
			Id: appID,
		},
		Status: c.kserviceToStatus(ctx, ksvc),
	}, nil
}
