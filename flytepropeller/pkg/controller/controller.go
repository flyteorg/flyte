// Package controller contains the K8s controller logic. This does not contain the actual workflow re-conciliation.
// It is then entrypoint into the K8s based Flyte controller.
package controller

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flyteK8sConfig "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"

	"github.com/flyteorg/flytepropeller/events"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	clientset "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	informers "github.com/flyteorg/flytepropeller/pkg/client/informers/externalversions"
	lister "github.com/flyteorg/flytepropeller/pkg/client/listers/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/catalog"
	errors3 "github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/factory"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/recovery"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flytepropeller/pkg/controller/workflow"
	"github.com/flyteorg/flytepropeller/pkg/controller/workflowstore"
	leader "github.com/flyteorg/flytepropeller/pkg/leaderelection"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	"github.com/flyteorg/flytestdlib/contextutils"
	stdErrs "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	"github.com/pkg/errors"

	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin" //nolint

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/clock"

	k8sInformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/record"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	resourceLevelMonitorCycleDuration = 5 * time.Second
	missing                           = "missing"
	podDefaultNamespace               = "flyte"
	podNamespaceEnvVar                = "POD_NAMESPACE"
)

type metrics struct {
	Scope            promutils.Scope
	EnqueueCountWf   prometheus.Counter
	EnqueueCountTask prometheus.Counter
}

// Controller is the controller implementation for FlyteWorkflow resources
type Controller struct {
	workerPool          *WorkerPool
	flyteworkflowSynced cache.InformerSynced
	workQueue           CompositeWorkQueue
	gc                  *GarbageCollector
	numWorkers          int
	workflowStore       workflowstore.FlyteWorkflow
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder      record.EventRecorder
	metrics       *metrics
	leaderElector *leaderelection.LeaderElector
	levelMonitor  *ResourceLevelMonitor
}

// Run either as a leader -if configured- or as a standalone process.
func (c *Controller) Run(ctx context.Context) error {
	if c.leaderElector == nil {
		logger.Infof(ctx, "Running without leader election.")
		return c.run(ctx)
	}

	logger.Infof(ctx, "Attempting to acquire leader lease and act as leader.")
	go c.leaderElector.Run(ctx)
	<-ctx.Done()
	return nil
}

// Start the actual work of controller (e.g. GC, consume and process queue items... etc.)
func (c *Controller) run(ctx context.Context) error {
	// Initializing WorkerPool
	logger.Info(ctx, "Initializing controller")
	if err := c.workerPool.Initialize(ctx); err != nil {
		return err
	}

	// Start the GC
	if err := c.gc.StartGC(ctx); err != nil {
		logger.Errorf(ctx, "failed to start background GC")
		return err
	}

	// Start the collector process
	c.levelMonitor.RunCollector(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info(ctx, "Starting FlyteWorkflow controller")
	return c.workerPool.Run(ctx, c.numWorkers, c.flyteworkflowSynced)
}

// Called from leader elector -if configured- to start running as the leader.
func (c *Controller) onStartedLeading(ctx context.Context) {
	backgroundCtx, cancelNow := context.WithCancel(ctx)
	logger.Infof(ctx, "Acquired leader lease.")
	go func() {
		if err := c.run(backgroundCtx); err != nil {
			logger.Panic(backgroundCtx, err)
		}
	}()

	<-backgroundCtx.Done()
	logger.Infof(ctx, "Lost leader lease.")
	cancelNow()
}

// enqueueFlyteWorkflow takes a FlyteWorkflow resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than FlyteWorkflow.
func (c *Controller) enqueueFlyteWorkflow(obj interface{}) {
	ctx := context.TODO()
	wf, ok := obj.(*v1alpha1.FlyteWorkflow)
	if !ok {
		logger.Errorf(ctx, "Received a non Workflow object")
		return
	}
	key := wf.GetK8sWorkflowID()
	logger.Infof(ctx, "==> Enqueueing workflow [%v]", key)
	c.workQueue.AddRateLimited(key.String())
}

func (c *Controller) enqueueWorkflowForNodeUpdates(workflowID v1alpha1.WorkflowID) {
	ctx := context.TODO()

	// validate workflowID
	_, _, err := cache.SplitMetaNamespaceKey(workflowID)
	if err != nil {
		logger.Warnf(ctx, "failed to add incorrectly formatted workflowID '%s' to subqueue", workflowID)
		return
	}

	// add workflowID to subqueue
	c.workQueue.AddToSubQueue(workflowID)
	c.metrics.EnqueueCountTask.Inc()
	logger.Debugf(ctx, "added workflowID '%s' to subqueue", workflowID)
}

func (c *Controller) getWorkflowUpdatesHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueFlyteWorkflow,
		UpdateFunc: func(old, new interface{}) {
			// TODO we might need to handle updates to the workflow itself.
			// Initially maybe we should not support it at all
			c.enqueueFlyteWorkflow(new)
		},
		DeleteFunc: func(obj interface{}) {
			// There is a corner case where the obj is not in fact a valid resource (it sends a DeletedFinalStateUnknown
			// object instead) -it has to do with missing some event that leads to not knowing the final state of the
			// resource. In which case, we can't use the regular metaAccessor to read obj name/namespace but should
			// instead use cache.DeletionHandling* helper functions that know how to deal with that.

			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				logger.Errorf(context.TODO(), "Unable to get key for deleted obj. Error[%v]", err)
				return
			}

			_, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				logger.Errorf(context.TODO(), "Unable to split enqueued key into namespace/execId. Error[%v]", err)
				return
			}

			logger.Infof(context.TODO(), "Deletion triggered for %v", name)
		},
	}
}

// ResourceLevelMonitor is responsible for emitting metrics that show the current number of Flyte workflows,
// by project and domain. It needs to be kicked off. The periodicity is not currently configurable because it seems
// unnecessary. It will also a timer measuring how long it takes to run each measurement cycle.
type ResourceLevelMonitor struct {
	Scope promutils.Scope

	// Meta timer - this times each collection cycle to measure how long it takes to collect the levels GaugeVec below
	CollectorTimer promutils.StopWatch

	// System Observability: This is a labeled gauge that emits the current number of FlyteWorkflow objects in the informer. It is used
	// to monitor current levels. It currently only splits by project/domain, not workflow status.
	levels labeled.Gauge

	// The thing that we want to measure the current levels of
	lister lister.FlyteWorkflowLister
}

func (r *ResourceLevelMonitor) countList(ctx context.Context, workflows []*v1alpha1.FlyteWorkflow) map[string]map[string]int {
	// Map of Projects to Domains to counts
	counts := map[string]map[string]int{}

	// Collect all workflow metrics
	for _, wf := range workflows {
		execID := wf.GetExecutionID()
		var project string
		var domain string
		if execID.WorkflowExecutionIdentifier == nil {
			logger.Warningf(ctx, "Workflow does not have an execution identifier! [%v]", wf)
			project = missing
			domain = missing
		} else {
			project = wf.ExecutionID.Project
			domain = wf.ExecutionID.Domain
		}
		if _, ok := counts[project]; !ok {
			counts[project] = map[string]int{}
		}
		counts[project][domain]++
	}

	return counts
}

func (r *ResourceLevelMonitor) collect(ctx context.Context) {
	// Emit gauges at both the project/domain level - aggregation to be handled by Prometheus
	workflows, err := r.lister.List(labels.Everything())
	if err != nil {
		logger.Errorf(ctx, "Error listing workflows when attempting to collect data for gauges %s", err)
	}

	counts := r.countList(ctx, workflows)

	// Emit labeled metrics, for each project/domain combination. This can be aggregated later with Prometheus queries.
	for project, val := range counts {
		for domain, num := range val {
			tempContext := contextutils.WithProjectDomain(ctx, project, domain)
			r.levels.Set(tempContext, float64(num))
		}
	}
}

func (r *ResourceLevelMonitor) RunCollector(ctx context.Context) {
	ticker := time.NewTicker(resourceLevelMonitorCycleDuration)
	collectorCtx := contextutils.WithGoroutineLabel(ctx, "resource-level-monitor")

	go func() {
		pprof.SetGoroutineLabels(collectorCtx)
		for {
			select {
			case <-collectorCtx.Done():
				return
			case <-ticker.C:
				t := r.CollectorTimer.Start()
				r.collect(collectorCtx)
				t.Stop()
			}
		}
	}()
}

func NewResourceLevelMonitor(scope promutils.Scope, lister lister.FlyteWorkflowLister) *ResourceLevelMonitor {
	return &ResourceLevelMonitor{
		Scope:          scope,
		CollectorTimer: scope.MustNewStopWatch("collection_cycle", "Measures how long it takes to run a collection", time.Millisecond),
		levels:         labeled.NewGauge("flyteworkflow", "Current FlyteWorkflow levels per instance of propeller", scope),
		lister:         lister,
	}
}

func newControllerMetrics(scope promutils.Scope) *metrics {
	c := scope.MustNewCounterVec("wf_enqueue", "workflow enqueue count.", "type")
	return &metrics{
		Scope:            scope,
		EnqueueCountWf:   c.WithLabelValues("wf"),
		EnqueueCountTask: c.WithLabelValues("task"),
	}
}

func getAdminClient(ctx context.Context) (client service.AdminServiceClient, signalClient service.SignalServiceClient, opt []grpc.DialOption, err error) {
	cfg := admin.GetConfig(ctx)
	clients, err := admin.NewClientsetBuilder().WithConfig(cfg).Build(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to initialize clientset. Error: %w", err)
	}

	credentialsFuture := admin.NewPerRPCCredentialsFuture()
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(admin.NewAuthInterceptor(cfg, nil, credentialsFuture)),
		grpc.WithPerRPCCredentials(credentialsFuture),
	}

	return clients.AdminClient(), clients.SignalServiceClient(), opts, nil
}

// New returns a new FlyteWorkflow controller
func New(ctx context.Context, cfg *config.Config, kubeclientset kubernetes.Interface, flytepropellerClientset clientset.Interface,
	flyteworkflowInformerFactory informers.SharedInformerFactory, informerFactory k8sInformers.SharedInformerFactory,
	kubeClient executors.Client, scope promutils.Scope) (*Controller, error) {

	adminClient, signalClient, authOpts, err := getAdminClient(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to initialize Admin client, err :%s", err.Error())
		return nil, err
	}
	var launchPlanActor launchplan.FlyteAdmin
	if cfg.EnableAdminLauncher {
		launchPlanActor, err = launchplan.NewAdminLaunchPlanExecutor(ctx, adminClient, cfg.DownstreamEval.Duration,
			launchplan.GetAdminConfig(), scope.NewSubScope("admin_launcher"))
		if err != nil {
			logger.Errorf(ctx, "failed to create Admin workflow Launcher, err: %v", err.Error())
			return nil, err
		}

		if err := launchPlanActor.Initialize(ctx); err != nil {
			logger.Errorf(ctx, "failed to initialize Admin workflow Launcher, err: %v", err.Error())
			return nil, err
		}
	} else {
		launchPlanActor = launchplan.NewFailFastLaunchPlanExecutor()
	}

	logger.Info(ctx, "Setting up event sink and recorder")
	eventSink, err := events.ConstructEventSink(ctx, events.GetConfig(ctx), scope.NewSubScope("event_sink"))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create EventSink [%v], error %v", events.GetConfig(ctx).Type, err)
	}
	gc, err := NewGarbageCollector(cfg, scope, clock.RealClock{}, kubeclientset.CoreV1().Namespaces(), flytepropellerClientset.FlyteworkflowV1alpha1())
	if err != nil {
		logger.Errorf(ctx, "failed to initialize GC for workflows")
		return nil, errors.Wrapf(err, "failed to initialize WF GC")
	}

	eventRecorder, err := utils.NewK8sEventRecorder(ctx, kubeclientset, controllerAgentName, cfg.PublishK8sEvents)
	if err != nil {
		logger.Errorf(ctx, "failed to event recorder %v", err)
		return nil, errors.Wrapf(err, "failed to initialize resource lock.")
	}
	controller := &Controller{
		metrics:    newControllerMetrics(scope),
		recorder:   eventRecorder,
		gc:         gc,
		numWorkers: cfg.Workers,
	}

	lock, err := leader.NewResourceLock(kubeclientset.CoreV1(), kubeclientset.CoordinationV1(), eventRecorder, cfg.LeaderElection)
	if err != nil {
		logger.Errorf(ctx, "failed to initialize resource lock.")
		return nil, errors.Wrapf(err, "failed to initialize resource lock.")
	}

	if lock != nil {
		logger.Infof(ctx, "Creating leader elector for the controller.")
		controller.leaderElector, err = leader.NewLeaderElector(lock, cfg.LeaderElection, controller.onStartedLeading, func() {
			logger.Fatal(ctx, "Lost leader state. Shutting down.")
		})

		if err != nil {
			logger.Errorf(ctx, "failed to initialize leader elector.")
			return nil, errors.Wrapf(err, "failed to initialize leader elector.")
		}
	}

	// WE are disabling this as the metrics have high cardinality. Metrics seem to be emitted per pod and this has problems
	// when we create new pods
	// Set Client Metrics Provider
	// setClientMetricsProvider(scope.NewSubScope("k8s_client"))

	// obtain references to shared index informers for FlyteWorkflow.
	flyteworkflowInformer := flyteworkflowInformerFactory.Flyteworkflow().V1alpha1().FlyteWorkflows()
	controller.flyteworkflowSynced = flyteworkflowInformer.Informer().HasSynced

	podTemplateInformer := informerFactory.Core().V1().PodTemplates()

	// set default namespace for pod template store
	podNamespace, found := os.LookupEnv(podNamespaceEnvVar)
	if !found {
		podNamespace = podDefaultNamespace
	}

	flytek8s.DefaultPodTemplateStore.SetDefaultNamespace(podNamespace)

	sCfg := storage.GetConfig()
	if sCfg == nil {
		logger.Errorf(ctx, "Storage configuration missing.")
	}

	store, err := storage.NewDataStore(sCfg, scope.NewSubScope("metastore"))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Metadata storage")
	}

	logger.Info(ctx, "Setting up Catalog client.")
	catalogClient, err := catalog.NewCatalogClient(ctx, authOpts...)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create datacatalog client")
	}

	workQ, err := NewCompositeWorkQueue(ctx, cfg.Queue, scope)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create WorkQueue [%v]", scope.CurrentScope())
	}
	controller.workQueue = workQ

	controller.workflowStore, err = workflowstore.NewWorkflowStore(ctx, workflowstore.GetConfig(), flyteworkflowInformer.Lister(), flytepropellerClientset.FlyteworkflowV1alpha1(), scope)
	if err != nil {
		return nil, stdErrs.Wrapf(errors3.CausedByError, err, "failed to initialize workflow store")
	}

	controller.levelMonitor = NewResourceLevelMonitor(scope.NewSubScope("collector"), flyteworkflowInformer.Lister())

	recoveryClient := recovery.NewClient(adminClient)
	nodeHandlerFactory, err := factory.NewHandlerFactory(ctx, launchPlanActor, launchPlanActor,
		kubeClient, catalogClient, recoveryClient, &cfg.EventConfig, cfg.ClusterID, signalClient, scope)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create node handler factory")
	}

	nodeExecutor, err := nodes.NewExecutor(ctx, cfg.NodeConfig, store, controller.enqueueWorkflowForNodeUpdates, eventSink,
		launchPlanActor, launchPlanActor, cfg.MaxDatasetSizeBytes, storage.DataReference(cfg.DefaultRawOutputPrefix), kubeClient,
		catalogClient, recoveryClient, &cfg.EventConfig, cfg.ClusterID, signalClient, nodeHandlerFactory, scope)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create Controller.")
	}

	workflowExecutor, err := workflow.NewExecutor(ctx, store, controller.enqueueWorkflowForNodeUpdates, eventSink, controller.recorder, cfg.MetadataPrefix, nodeExecutor, &cfg.EventConfig, cfg.ClusterID, scope)
	if err != nil {
		return nil, err
	}

	handler := NewPropellerHandler(ctx, cfg, store, controller.workflowStore, workflowExecutor, scope)
	controller.workerPool = NewWorkerPool(ctx, scope, workQ, handler)

	if cfg.EnableGrpcLatencyMetrics {
		grpc_prometheus.EnableClientHandlingTimeHistogram()
	}

	logger.Info(ctx, "Setting up event handlers")
	// Set up an event handler for when FlyteWorkflow resources change
	flyteworkflowInformer.Informer().AddEventHandler(controller.getWorkflowUpdatesHandler())

	updateHandler := flytek8s.GetPodTemplateUpdatesHandler(&flytek8s.DefaultPodTemplateStore)
	podTemplateInformer.Informer().AddEventHandler(updateHandler)
	return controller, nil
}

// getShardedLabelSelectorRequirements computes the collection of LabelSelectorReuquirements to
// satisfy sharding criteria.
func getShardedLabelSelectorRequirements(cfg *config.Config) []v1.LabelSelectorRequirement {
	selectors := []struct {
		label     string
		operation v1.LabelSelectorOperator
		values    []string
	}{
		{k8s.ShardKeyLabel, v1.LabelSelectorOpIn, cfg.IncludeShardKeyLabel},
		{k8s.ShardKeyLabel, v1.LabelSelectorOpNotIn, cfg.ExcludeShardKeyLabel},
		{k8s.ProjectLabel, v1.LabelSelectorOpIn, cfg.IncludeProjectLabel},
		{k8s.ProjectLabel, v1.LabelSelectorOpNotIn, cfg.ExcludeProjectLabel},
		{k8s.DomainLabel, v1.LabelSelectorOpIn, cfg.IncludeDomainLabel},
		{k8s.DomainLabel, v1.LabelSelectorOpNotIn, cfg.ExcludeDomainLabel},
	}

	var labelSelectorRequirements []v1.LabelSelectorRequirement
	for _, selector := range selectors {
		if len(selector.values) > 0 {
			labelSelectorRequirement := v1.LabelSelectorRequirement{
				Key:      selector.label,
				Operator: selector.operation,
				Values:   selector.values,
			}

			labelSelectorRequirements = append(labelSelectorRequirements, labelSelectorRequirement)
		}
	}

	return labelSelectorRequirements
}

// SharedInformerOptions creates informer options to work with FlytePropeller Sharding
func SharedInformerOptions(cfg *config.Config, defaultNamespace string) []informers.SharedInformerOption {
	labelSelector := IgnoreCompletedWorkflowsLabelSelector()

	shardedLabelSelectorRequirements := getShardedLabelSelectorRequirements(cfg)
	if len(shardedLabelSelectorRequirements) != 0 {
		labelSelector.MatchExpressions = append(labelSelector.MatchExpressions, shardedLabelSelectorRequirements...)
	}

	opts := []informers.SharedInformerOption{
		informers.WithTweakListOptions(func(options *v1.ListOptions) {
			options.LabelSelector = v1.FormatLabelSelector(labelSelector)
		}),
	}

	if cfg.LimitNamespace != defaultNamespace {
		opts = append(opts, informers.WithNamespace(cfg.LimitNamespace))
	}
	return opts
}

func CreateControllerManager(ctx context.Context, cfg *config.Config, options manager.Options) (*manager.Manager, error) {

	_, kubecfg, err := utils.GetKubeConfig(ctx, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "error building Kubernetes Clientset")
	}

	mgr, err := manager.New(kubecfg, options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to initialize controller-runtime manager")
	}
	return &mgr, nil
}

// StartControllerManager Start controller runtime manager to start listening to resource changes.
// K8sPluginManager uses controller runtime to create informers for the CRDs being monitored by plugins. The informer
// EventHandler enqueues the owner workflow for reevaluation. These informer events allow propeller to detect
// workflow changes faster than the default sync interval for workflow CRDs.
func StartControllerManager(ctx context.Context, mgr *manager.Manager) error {
	ctx = contextutils.WithGoroutineLabel(ctx, "controller-runtime-manager")
	pprof.SetGoroutineLabels(ctx)
	logger.Infof(ctx, "Starting controller-runtime manager")
	return (*mgr).Start(ctx)
}

// StartController creates a new FlytePropeller Controller and starts it
func StartController(ctx context.Context, cfg *config.Config, defaultNamespace string, mgr *manager.Manager, scope *promutils.Scope) error {
	// Setup cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	kubeClient, kubecfg, err := utils.GetKubeConfig(ctx, cfg)
	if err != nil {
		return errors.Wrapf(err, "error building Kubernetes Clientset")
	}

	flyteworkflowClient, err := clientset.NewForConfig(kubecfg)
	if err != nil {
		return errors.Wrapf(err, "error building FlyteWorkflow clientset")
	}

	// Create FlyteWorkflow CRD if it does not exist
	if cfg.CreateFlyteWorkflowCRD {
		logger.Infof(ctx, "creating FlyteWorkflow CRD")
		apiextensionsClient, err := apiextensionsclientset.NewForConfig(kubecfg)
		if err != nil {
			return errors.Wrapf(err, "error building apiextensions clientset")
		}

		_, err = apiextensionsClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &flyteworkflow.CRD, v1.CreateOptions{})
		if err != nil {
			if apierrors.IsAlreadyExists(err) {
				logger.Warnf(ctx, "FlyteWorkflow CRD already exists")
			} else {
				return errors.Wrapf(err, "failed to create FlyteWorkflow CRD")
			}
		}
	}

	opts := SharedInformerOptions(cfg, defaultNamespace)
	flyteworkflowInformerFactory := informers.NewSharedInformerFactoryWithOptions(flyteworkflowClient, cfg.WorkflowReEval.Duration, opts...)

	informerFactory := k8sInformers.NewSharedInformerFactoryWithOptions(kubeClient, flyteK8sConfig.GetK8sPluginConfig().DefaultPodTemplateResync.Duration)

	c, err := New(ctx, cfg, kubeClient, flyteworkflowClient, flyteworkflowInformerFactory, informerFactory, *mgr, *scope)
	if err != nil {
		return errors.Wrap(err, "failed to start FlytePropeller")
	} else if c == nil {
		return errors.Errorf("Failed to create a new instance of FlytePropeller")
	}

	go flyteworkflowInformerFactory.Start(ctx.Done())
	if flyteK8sConfig.GetK8sPluginConfig().DefaultPodTemplateName != "" {
		go informerFactory.Start(ctx.Done())
	}

	if err = c.Run(ctx); err != nil {
		return errors.Wrapf(err, "Error running FlytePropeller.")
	}
	return nil
}
