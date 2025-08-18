package plugin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/errorcollector"
	podplugin "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/pod"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

const (
	EXECUTION_ENV_NAME    = "execution-env-name"
	EXECUTION_ENV_TYPE    = "execution-env-type"
	EXECUTION_ENV_VERSION = "execution-env-version"
	TTL_SECONDS           = "ttl-seconds"
	PROJECT_LABEL         = "project"
	DOMAIN_LABEL          = "domain"
	ORGANIZATION_LABEL    = "organization"
)

// builderMetrics is a collection of metrics for the InMemoryEnvBuilder.
type builderMetrics struct {
	environmentsCreated        *prometheus.CounterVec
	environmentsGCed           *prometheus.CounterVec
	environmentOrphansDetected *prometheus.CounterVec
	podsCreated                *prometheus.CounterVec
	podCreationErrors          *prometheus.CounterVec
	podsDeleted                *prometheus.CounterVec
	podsDeletionErrors         *prometheus.CounterVec
	scaleDownBufferSize        prometheus.Gauge
	scaleUpBufferSize          prometheus.Gauge
	scaleDownWorkerEvents      *prometheus.CounterVec
	scaleUpWorkerEvents        *prometheus.CounterVec
}

// newBuilderMetrics creates a new builderMetrics with the given scope.
func newBuilderMetrics(scope promutils.Scope) builderMetrics {
	return builderMetrics{
		environmentsCreated:        scope.MustNewCounterVec("env_created", "The number of environments created", "project", "domain"),
		environmentsGCed:           scope.MustNewCounterVec("env_gced", "The number of environments garbage collected", "project", "domain"),
		environmentOrphansDetected: scope.MustNewCounterVec("env_orphans_detected", "The number of orphaned environments detected", "project", "domain"),
		podsCreated: scope.MustNewCounterVec("pods_created_total",
			"Total number of pods recreated during repair", "project", "domain", "purpose"),
		podCreationErrors: scope.MustNewCounterVec("pod_creation_errors_total",
			"Total number of errors encountered during pod creation", "project", "domain", "purpose"),
		podsDeleted: scope.MustNewCounterVec("pods_deleted_total",
			"Total number of pods deleted", "project", "domain", "purpose"),
		podsDeletionErrors: scope.MustNewCounterVec("pod_deletion_errors_total",
			"Total number of errors encountered during pod deletion", "project", "domain", "purpose"),
		scaleDownBufferSize:   scope.MustNewGauge("scale_down_buffer_size", "Count of environments waiting for scale down"),
		scaleUpBufferSize:     scope.MustNewGauge("scale_up_buffer_size", "Count of environments waiting for scale up"),
		scaleDownWorkerEvents: scope.MustNewCounterVec("scale_down_worker_events", "Count of scale down worker events for environments", "env_id"),
		scaleUpWorkerEvents:   scope.MustNewCounterVec("scale_up_worker_events", "Count of scale up worker events for environments", "env_id"),
	}
}

type environmentBuilderImpl struct {
	kubeClient  core.KubeClient
	randSource  *rand.Rand
	scaleUpChan chan string
	store       interfaces.EnvironmentStore
	metrics     builderMetrics
}

func addObjectMetadata(ctx context.Context, tCtx core.TaskExecutionContext, spec *v1.PodTemplateSpec, cfg *config.K8sPluginConfig) error {
	annotations := tCtx.TaskExecutionMetadata().GetAnnotations()
	// Omit some execution specific labels that don't make sense for a reusable env
	labels := lo.OmitByKeys(tCtx.TaskExecutionMetadata().GetLabels(), []string{
		k8s.ExecutionIDLabel, k8s.WorkflowNameLabel, nodes.NodeIDLabel, nodes.TaskNameLabel})

	tmpl, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to read task template")
	}
	if len(tmpl.GetSecurityContext().GetSecrets()) > 0 {
		secretsMap, err := secrets.MarshalSecretsToMapStrings(tmpl.GetSecurityContext().GetSecrets())
		if err != nil {
			return flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal secrets")
		}
		annotations = utils.UnionMaps(annotations, secretsMap)
		labels[secrets.PodLabel] = secrets.PodLabelValue
	}

	spec.SetAnnotations(utils.UnionMaps(cfg.DefaultAnnotations, spec.GetAnnotations(), annotations))
	spec.SetLabels(utils.UnionMaps(cfg.DefaultLabels, spec.GetLabels(), labels))
	spec.SetNamespace(tCtx.TaskExecutionMetadata().GetNamespace())

	// don't set owner references for fast tasks, as they are intended to outlive a single task execution
	spec.SetOwnerReferences([]metav1.OwnerReference{})

	if cfg.InjectFinalizer { // nolint: staticcheck
		// TODO: add finalizer
	}

	return nil
}

func getFastTaskEnvironmentSpec(ctx context.Context, tCtx core.TaskExecutionContext,
	executionEnv *idlcore.ExecutionEnv) (*pb.FastTaskEnvironmentSpec, error) {

	// retrieve environment
	fastTaskEnvironmentSpec := &pb.FastTaskEnvironmentSpec{}
	switch executionEnv.GetEnvironment().(type) {
	case *idlcore.ExecutionEnv_Spec:
		if err := utils.UnmarshalStruct(executionEnv.GetSpec(), fastTaskEnvironmentSpec); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment spec")
		}

		// if `podTemplateSpec` is provided, then we use it. otherwise, we generate the pod spec
		// from the task template.
		var podTemplateSpec v1.PodTemplateSpec
		if len(fastTaskEnvironmentSpec.GetPodTemplateSpec()) > 0 {
			if err := json.Unmarshal(fastTaskEnvironmentSpec.GetPodTemplateSpec(), &podTemplateSpec); err != nil {
				return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal pod template spec")
			}
		} else {
			podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, tCtx)
			if err != nil {
				return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to create environment")
			}

			podTemplateSpec = v1.PodTemplateSpec{
				ObjectMeta: *objectMeta,
				Spec:       *podSpec,
			}
			fastTaskEnvironmentSpec.PrimaryContainerName = primaryContainerName
		}

		if err := addObjectMetadata(ctx, tCtx, &podTemplateSpec, config.GetK8sPluginConfig()); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to add object metadata")
		}
		podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
		if err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal pod template spec")
		}

		fastTaskEnvironmentSpec.PodTemplateSpec = podTemplateSpecBytes
	case *idlcore.ExecutionEnv_Extant:
		// TODO - this execution path is not implemented. it was originally designed to
		// support fasttask workers that are not managed by the fasttask plugin. for example, a
		// user manually starting a worker and executing a task against it. one way to get this
		// worker is to:
		// (1) create an additional environment state (ex. `EXTANT`) that is used to identify
		// environments that are not managed by the fasttask plugin.
		// (2) allow workers to create environments (ex. in `EXTANT` state) and register themselves
		// when connecting rather than waiting for an existing environment.
		// (3) exclude `EXTANT` environments from orphan detection, scaleUp, and scaleDown operations.
		return nil, errors.New("executing fasttask from extant is not implemented")
	}

	return fastTaskEnvironmentSpec, nil
}

// getMetricLabels returns the metric labels for the given environment ID.
func (e *environmentBuilderImpl) getMetricLabels(environmentID string) []string {
	environment := e.store.Get(environmentID)
	if environment == nil {
		return nil
	}
	return []string{environment.EnvID().Project, environment.EnvID().Domain}
}

func (e *environmentBuilderImpl) createPod(ctx context.Context, podName string, env interfaces.Environment) error {
	fastTaskEnvironmentSpec := env.FastTaskEnvironmentSpec()
	envID := env.EnvID()

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(fastTaskEnvironmentSpec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.GetPodTemplateSpec(), err.Error())
	}

	podSpec := &podTemplateSpec.Spec
	objectMeta := &podTemplateSpec.ObjectMeta

	// identify the primary container
	primaryContainerIndex := -1
	if len(fastTaskEnvironmentSpec.GetPrimaryContainerName()) > 0 {
		for index, container := range podSpec.Containers {
			if container.Name == fastTaskEnvironmentSpec.GetPrimaryContainerName() {
				primaryContainerIndex = index
				break
			}
		}
	} else if len(podSpec.Containers) == 1 {
		primaryContainerIndex = 0
	}

	if primaryContainerIndex == -1 {
		return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to identify primary container in the environment spec [%v]", podSpec)
	}

	// add execution environment labels and annotations
	objectMeta.Name = podName
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string, 0)
	}
	objectMeta.Labels[EXECUTION_ENV_TYPE] = fastTaskType
	objectMeta.Labels[EXECUTION_ENV_NAME] = envID.Name
	objectMeta.Labels[EXECUTION_ENV_VERSION] = envID.Version
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string, 0)
	}
	objectMeta.Annotations[TTL_SECONDS] = fmt.Sprintf("%d", int(getEnvironmentTTLOrDefault(fastTaskEnvironmentSpec)))

	// update primaryContainer
	container := &podSpec.Containers[primaryContainerIndex]
	container.Name = podName
	if _, exists := objectMeta.Annotations[flytek8s.PrimaryContainerKey]; exists {
		objectMeta.Annotations[flytek8s.PrimaryContainerKey] = container.Name
	}

	container.Args = []string{
		"unionai-actor-bridge",
	}

	// append additional worker args before plugin args to ensure they are overridden
	container.Args = append(container.Args, GetConfig().AdditionalWorkerArgs...)
	if fastTaskEnvironmentSpec.GetParallelism() > 0 {
		container.Args = append(container.Args, "--parallelism", fmt.Sprintf("%d", fastTaskEnvironmentSpec.GetParallelism()))
	}

	container.Args = append(container.Args,
		"--backlog-length",
		fmt.Sprintf("%d", fastTaskEnvironmentSpec.GetBacklogLength()),
		"--queue-id",
		envID.String(),
		"--worker-id",
		podName,
		"--fasttask-url",
		GetConfig().CallbackURI,
	)

	// set rust log level
	logLevel := GetConfig().WorkerLogLevel
	if !slices.Contains(logLevels, logLevel) {
		logger.Warnf(ctx, "invalid worker log level [%s], defaulting to info", logLevel)
		logLevel = logLevelInfo
	}

	container.Env = append(container.Env, v1.EnvVar{
		Name:  "RUST_LOG",
		Value: logLevel,
	})

	logger.Debugf(ctx, "creating pod '%s' for environment '%s'", podName, envID)

	// use kubeclient to create worker
	err := e.kubeClient.GetClient().Create(ctx, &v1.Pod{
		ObjectMeta: *objectMeta,
		Spec:       *podSpec,
	})
	metricLabels := append(e.getMetricLabels(envID.String()), "create")

	if err != nil {
		logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, envID, err)
		e.metrics.podCreationErrors.WithLabelValues(metricLabels...).Inc()
		return err
	}

	e.metrics.podsCreated.WithLabelValues(metricLabels...).Inc()

	return nil
}

// detectOrphanedEnvironments detects orphaned environments by identifying pods with the fast task
// execution type label.
func (e *environmentBuilderImpl) detectOrphanedEnvironments(ctx context.Context, k8sReader client.Reader) error {
	// retrieve all pods with fast task execution type label
	matchingLabelsOption := client.MatchingLabels{}
	matchingLabelsOption[EXECUTION_ENV_TYPE] = fastTaskType

	podList := &v1.PodList{}
	if err := k8sReader.List(ctx, podList, matchingLabelsOption); err != nil {
		return err
	}

	// detect orphaned environments
	for _, pod := range podList.Items {
		envID, err := parseExectionEnvID(pod.GetLabels())
		if err != nil {
			logger.Warnf(ctx, "failed to parse ExecutionEnvID [%v]", err)
			continue
		}

		// check if environment exists in store
		env := e.store.Get(envID.String())
		if env == nil {
			logger.Infof(ctx, "detected orphaned environment '%s'", env)

			// serialize podTemplateSpec with Namespace so we can use it to delete TOMBSTONED environments
			podTemplateSpec := &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: pod.Namespace,
				},
			}

			podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
			if err != nil {
				logger.Warnf(ctx, "failed to marshal PodTemplateSpec [%v]", err)
				continue
			}

			// parse ttl seconds from annotations
			ttlSeconds := 0
			if ttlSecondsStr, exists := pod.Annotations[TTL_SECONDS]; exists {
				ttlSeconds, err = strconv.Atoi(ttlSecondsStr)
				if err != nil {
					// this should be unreachable because we are serializing the integer which
					// sets this annotation. if parsing errors then we leave ttlSeconds as 0 to
					// ensure the orphaned environment is garbage collected immediately
					ttlSeconds = 0
					logger.Warnf(ctx, "failed to parse TTL_SECONDS [%s] for pod '%s' [%v]", ttlSecondsStr, pod.Name, err)
				}
			}

			// create orphaned environment
			now := time.Now().Unix()
			env = &environmentImpl{
				createdAt: now,
				envID:     envID,
				fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
					PodTemplateSpec: podTemplateSpecBytes,
					TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
						TtlSeconds: int32(ttlSeconds),
					},
				},
				lastScaledDownAt: now,
				lock:             sync.RWMutex{},
				state:            interfaces.ORPHANED,
				workers:          &sync.Map{},
			}

			// add environment to store
			e.store.GetOrCreate(envID.String(), env)
			metricsLabels := e.getMetricLabels(envID.String())
			e.metrics.environmentOrphansDetected.WithLabelValues(metricsLabels...).Inc()
		}

		logger.Infof(ctx, "detected orphaned pod '%v' for environment '%s'", pod, envID)
		// if worker does not exist then we create it
		env.GetOrCreateWorker(pod.Name)
	}

	return nil
}

func (e *environmentBuilderImpl) GetOrCreateEnvironment(ctx context.Context, tCtx core.TaskExecutionContext,
	envID interfaces.ExecutionEnvID, executionEnv *idlcore.ExecutionEnv) (interfaces.Environment, error) {
	env := e.store.Get(envID.String())
	if env != nil {
		state := env.State()
		if state == interfaces.HEALTHY || state == interfaces.TOMBSTONED {
			return env, nil
		}
	}

	fastTaskEnvironmentSpec, err := getFastTaskEnvironmentSpec(ctx, tCtx, executionEnv)
	if err != nil {
		return nil, err
	}

	if env == nil {
		// if environment did not exist then create it
		now := time.Now().Unix()
		env = &environmentImpl{
			createdAt:               now,
			envID:                   envID,
			fastTaskEnvironmentSpec: fastTaskEnvironmentSpec,
			lastScaledDownAt:        now,
			lock:                    sync.RWMutex{},
			state:                   interfaces.INITIALIZING,
			workers:                 &sync.Map{},
		}

		// ensure multiple processes don't create pods for a new environment
		loadedEnv := e.store.GetOrCreate(envID.String(), env)
		if loadedEnv != env {
			return loadedEnv, nil
		}

		metricLabels := []string{envID.Project, envID.Domain}
		e.metrics.environmentsCreated.WithLabelValues(metricLabels...).Inc()

		// attempt to create a worker pod
		var errorMessages []string
		workerCount := 0
		minReplicaCount := getMinReplicaCount(fastTaskEnvironmentSpec)
		for i := 0; i < minReplicaCount; i++ {
			podName := e.getPodName(envID.Name)
			env.GetOrCreateWorker(podName)
			workerCount++

			err := e.createPod(ctx, podName, env)
			if err != nil {
				logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, envID.String(), err)
				errorMessages = append(errorMessages, fmt.Sprintf("pod '%s': %v", podName, err))
			} else {
				break
			}
		}

		if minReplicaCount != 0 && len(errorMessages) == minReplicaCount {
			// if we fail to create all workers then we set the environment to TOMBSTONED. this
			// is a non-recoverable state and will cause all future task executions to immediately
			// fail. the environment will be removed from the store after a TTL.
			env.SetFailureMessage(strings.Join(errorMessages, "; "))
			env.SetState(interfaces.TOMBSTONED)
		} else {
			env.SetState(interfaces.HEALTHY)
		}

		// if we started fewer workers than the number of configured replicas then we initiate a
		// scale up in the background
		if workerCount < minReplicaCount {
			e.ScaleUp(ctx, envID.String())
		}
	} else if env.State() == interfaces.ORPHANED {
		// update orphaned environment
		env.Recover(envID, fastTaskEnvironmentSpec)
		env.SetState(interfaces.HEALTHY)
	}

	return env, nil
}

// deletePod deletes the pod with the given name and namespace.
func (e *environmentBuilderImpl) deletePod(ctx context.Context, env interfaces.Environment, name types.NamespacedName) error {
	logger.Debugf(ctx, "deleting pod '%s' for environment '%s'", name.Name, env.EnvID())

	objectMeta := metav1.ObjectMeta{
		Name:      name.Name,
		Namespace: name.Namespace,
	}

	err := e.kubeClient.GetClient().Delete(ctx, &v1.Pod{
		ObjectMeta: objectMeta,
	}, client.GracePeriodSeconds(0))

	metricLabels := append(e.getMetricLabels(env.EnvID().String()), "gc")

	if err != nil {
		logger.Warnf(ctx, "failed to gc pod '%s' for environment '%s' [%v]", name.Name, env.EnvID(), err)
		e.metrics.podsDeletionErrors.WithLabelValues(metricLabels...).Inc()
		return err
	}

	e.metrics.podsDeleted.WithLabelValues(metricLabels...).Inc()
	return nil
}

func (e *environmentBuilderImpl) getPodName(envName string) string {
	nonceBytes := make([]byte, (GetConfig().NonceLength+1)/2)
	if _, err := e.randSource.Read(nonceBytes); err != nil {
		return ""
	}

	return fmt.Sprintf("%s-%s", sanitizeEnvName(envName), hex.EncodeToString(nonceBytes)[:GetConfig().NonceLength])
}

func (e *environmentBuilderImpl) deleteEnvironment(ctx context.Context, env interfaces.Environment) error {
	// attempt to delete all workers
	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), err.Error())
	}

	deletedWorkers := make([]string, 0)
	allDeleted := true
	env.RangeWorkers(func(workerID string, _ interfaces.Worker) bool {
		namespacedName := types.NamespacedName{
			Name:      workerID,
			Namespace: podTemplateSpec.Namespace,
		}

		err := e.deletePod(ctx, env, namespacedName)
		if err != nil && !k8serrors.IsNotFound(err) {
			logger.Warnf(ctx, "failed to gc pod '%s' for environment '%s' [%v]", workerID, env.EnvID().String(), err)
			allDeleted = false
		} else {
			deletedWorkers = append(deletedWorkers, workerID)
		}

		return true
	})

	if allDeleted {
		metricLabels := e.getMetricLabels(env.EnvID().String())
		logger.Infof(ctx, "garbage collected environment '%s'", env.EnvID())
		e.metrics.environmentsGCed.WithLabelValues(metricLabels...).Inc()
		e.store.Delete(env.EnvID().String())
	} else {
		for _, worker := range deletedWorkers {
			env.DeleteWorker(worker)
		}
	}

	return nil
}

func (e *environmentBuilderImpl) scaleDown(ctx context.Context, env interfaces.Environment) error {
	now := time.Now().Unix()
	orphanedWorkers := make([]string, 0)
	expiredWorkers := make([]string, 0)
	orphanedTTL := GetConfig().OrphanedWorkerTTL.Seconds()
	workerTTL := getReplicaTTLOrDefault(env.FastTaskEnvironmentSpec())

	// identify latest accessed timestamp
	lastAccessedAt := int64(0)
	workerCount := 0
	env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
		workerLastAccessedAt := worker.LastAccessedAt()
		if workerLastAccessedAt > lastAccessedAt {
			lastAccessedAt = workerLastAccessedAt
		}

		if worker.State() == interfaces.ORPHANED {
			if float64(now-workerLastAccessedAt) > orphanedTTL {
				orphanedWorkers = append(orphanedWorkers, worker.ID())
			}
		} else {
			if float64(now-workerLastAccessedAt) > workerTTL {
				// expired will only get scaled down if we have more than the minimum number of replicas
				expiredWorkers = append(expiredWorkers, worker.ID())
			}
		}

		workerCount++
		return true
	})

	if lastAccessedAt == 0 {
		// if no workers exist then we use the createdAt timestamp
		lastAccessedAt = env.CreatedAt()
	}

	// check if environment TTL has expired
	if float64(time.Now().Unix()-lastAccessedAt) > getEnvironmentTTLOrDefault(env.FastTaskEnvironmentSpec()) {
		return e.deleteEnvironment(ctx, env)
	}

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), err.Error())
	}

	// delete orphaned workers disregarding min replica count to get to the minimum number of healthy replicas as soon as possible
	for _, worker := range orphanedWorkers {
		e.metrics.scaleDownWorkerEvents.WithLabelValues(env.EnvID().String()).Inc()

		namespacedName := types.NamespacedName{
			Name:      worker,
			Namespace: podTemplateSpec.Namespace,
		}

		err := e.deletePod(ctx, env, namespacedName)
		if err != nil && !k8serrors.IsNotFound(err) {
			logger.Warnf(ctx, "failed to gc orphaned pod '%s' for environment '%s' [%v]", worker, env.EnvID().String(), err)
		} else {
			env.DeleteWorker(worker)
			workerCount--
		}
	}

	minReplicaCount := getMinReplicaCount(env.FastTaskEnvironmentSpec())
	for _, worker := range expiredWorkers {
		if workerCount <= minReplicaCount {
			break
		}
		e.metrics.scaleDownWorkerEvents.WithLabelValues(env.EnvID().String()).Inc()

		namespacedName := types.NamespacedName{
			Name:      worker,
			Namespace: podTemplateSpec.Namespace,
		}

		err := e.deletePod(ctx, env, namespacedName)
		if err != nil && !k8serrors.IsNotFound(err) {
			logger.Warnf(ctx, "failed to gc pod '%s' for environment '%s' [%v]", worker, env.EnvID().String(), err)
		} else {
			env.DeleteWorker(worker)
			workerCount--
		}
	}

	return nil
}

func (e *environmentBuilderImpl) scaleUp(ctx context.Context, env interfaces.Environment) error {

	if env.State() != interfaces.HEALTHY {
		return nil
	}

	// verify we should perform scale up operation
	workerCount := 0
	env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
		workerCount++
		return true
	})

	if workerCount >= int(env.FastTaskEnvironmentSpec().GetReplicaCount()) {
		return nil
	}

	// scale up to the minimum number of replicas
	minReplicaCount := getMinReplicaCount(env.FastTaskEnvironmentSpec())
	if workerCount < minReplicaCount {
		for i := workerCount; i < minReplicaCount; i++ {
			podName := e.getPodName(env.EnvID().Name)
			env.GetOrCreateWorker(podName)
			e.metrics.scaleUpWorkerEvents.WithLabelValues(env.EnvID().String()).Inc()

			err := e.createPod(ctx, podName, env)
			if err != nil {
				logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, env.EnvID().String(), err)
			}
		}
		return nil
	}

	// scale up by 1 replica
	podName := e.getPodName(env.EnvID().Name)
	env.GetOrCreateWorker(podName)
	e.metrics.scaleUpWorkerEvents.WithLabelValues(env.EnvID().String()).Inc()

	err := e.createPod(ctx, podName, env)
	if err != nil {
		logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, env.EnvID().String(), err)
	}

	return nil
}

func (e *environmentBuilderImpl) ScaleUp(ctx context.Context, executionEnvID string) {
	select {
	case e.scaleUpChan <- executionEnvID:
		e.metrics.scaleUpBufferSize.Set(float64(len(e.scaleUpChan)))
	default:
		logger.Warnf(ctx, "scaleUpChan full, dropping scale up request for environment '%s'", executionEnvID)
	}
}

func (e *environmentBuilderImpl) start(ctx context.Context, store interfaces.EnvironmentStore, detectOrphansChan, scaleDownChan, scaleUpChan chan string) {
	if err := e.detectOrphanedEnvironments(ctx, e.kubeClient.GetClient()); err != nil {
		logger.Warnf(ctx, "failed to detect orphaned environments [%v]", err)
	}

	lastOrphanDetectionAt := time.Now().Unix()
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-detectOrphansChan:
				if err := e.detectOrphanedEnvironments(ctx, e.kubeClient.GetCache()); err != nil {
					logger.Warnf(ctx, "failed to detect orphaned environments [%v]", err)
				}
			case environmentID := <-scaleDownChan:
				if env := store.Get(environmentID); env != nil {
					if err := e.scaleDown(ctx, env); err != nil {
						logger.Warnf(ctx, "failed to scale down environment '%s' [%v]", environmentID, err)
					}
				}
			case environmentID := <-scaleUpChan:
				if env := store.Get(environmentID); env != nil {
					if err := e.scaleUp(ctx, env); err != nil {
						logger.Warnf(ctx, "failed to scale up environment '%s' [%v]", environmentID, err)
					}
				}
			case <-ticker.C:
				now := time.Now().Unix()

				if float64(now-lastOrphanDetectionAt) >= GetConfig().EnvDetectOrphanInterval.Seconds() {
					select {
					case detectOrphansChan <- "":
					default:
						// This should never happen
						logger.Warnf(ctx, "detectOrphansChan full")
					}

					lastOrphanDetectionAt = now
				}

				for _, env := range e.store.List() {
					if float64(now-env.LastScaledDownAt()) >= GetConfig().EnvScaleDownInterval.Seconds() {
						select {
						case scaleDownChan <- env.EnvID().String():
							env.SetLastScaledDownAt(now)
							e.metrics.scaleDownBufferSize.Set(float64(len(scaleDownChan)))
						default:
							logger.Warnf(ctx, "scaleDownChan full, unable to scaleDown environment '%s'", env.EnvID().String())
						}
					}
				}
			}
		}
	}()
}

func (e *environmentBuilderImpl) GetWorkerPod(ctx context.Context, executionEnvID, workerID string) (*v1.Pod, error) {
	// retrieve worker
	env := e.store.Get(executionEnvID)
	if env == nil {
		return nil, fmt.Errorf("environment '%s' not found", executionEnvID)
	}

	worker := env.GetWorker(workerID)
	if worker == nil {
		return nil, fmt.Errorf("worker '%s' not found in environment '%s'", workerID, executionEnvID)
	}

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), err.Error())
	}

	// retrieve pod
	namespacedName := types.NamespacedName{
		Name:      workerID,
		Namespace: podTemplateSpec.Namespace,
	}

	pod := &v1.Pod{}
	if err := e.kubeClient.GetClient().Get(ctx, namespacedName, pod); err != nil {
		return nil, err
	}

	return pod, nil
}

func (e *environmentBuilderImpl) ValidateWorkerPods(ctx context.Context, executionEnvID string, taskInfo *core.TaskInfo) (string, error) {
	// retrieve environment
	env := e.store.Get(executionEnvID)
	if env == nil {
		return "", fmt.Errorf("environment '%s' not found", executionEnvID)
	}

	// retrieve worker pod names
	var podNames []string
	env.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
		podNames = append(podNames, worker.ID())
		return true
	})

	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return "", flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", env.FastTaskEnvironmentSpec().GetPodTemplateSpec(), err.Error())
	}

	// validate that all worker pads are not in a failure state
	allReplicasFailed := true
	messageCollector := errorcollector.NewErrorMessageCollector()
	for i, podName := range podNames {
		namespacedName := types.NamespacedName{
			Name:      podName,
			Namespace: podTemplateSpec.Namespace,
		}

		pod := &v1.Pod{}
		if err := e.kubeClient.GetCache().Get(ctx, namespacedName, pod); err != nil {
			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) || k8serrors.IsResourceExpired(err) {
				// pod does not exist because it has not yet been populated in the kubeclient
				// cache or was deleted. to be safe, we treat both as a non-failure state.
				allReplicasFailed = false
				break
			}

			return "", err
		}

		phaseInfo, err := podplugin.DemystifyPodStatus(ctx, pod, *taskInfo)
		if err != nil {
			return "", err
		}

		switch phaseInfo.Phase() {
		case core.PhasePermanentFailure, core.PhaseRetryableFailure:
			if phaseInfo.Err() != nil {
				messageCollector.Collect(i, phaseInfo.Err().GetMessage())
			} else {
				messageCollector.Collect(i, phaseInfo.Reason())
			}
		default:
			allReplicasFailed = false
		}
	}

	if allReplicasFailed {
		// an optimization would be to transition the environment to a TOMBSTONED state here so
		// that subsequent task executions immediately fail. if this becomes a common occurrence
		// then we should consider adding this optimization.
		return messageCollector.Summary(maxErrorMessageLength), nil
	}

	return "", nil
}

func newEnvironmentBuilder(ctx context.Context, kubeClient core.KubeClient, store interfaces.EnvironmentStore, scope promutils.Scope, config *Config) interfaces.EnvironmentBuilder {
	detectOrphansChan := make(chan string, 1)
	// If the scaleDownChan reaches capacity, then the fast task state loop will be blocked
	scaleDownChan := make(chan string, config.ScalingBufferSize)
	scaleUpChan := make(chan string, config.ScalingBufferSize)
	builder := &environmentBuilderImpl{
		kubeClient:  kubeClient,
		randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
		scaleUpChan: scaleUpChan,
		store:       store,
		metrics:     newBuilderMetrics(scope),
	}

	builder.start(ctx, store, detectOrphansChan, scaleDownChan, scaleUpChan)
	return builder
}
