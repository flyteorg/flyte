package plugin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

type state int32

const (
	HEALTHY state = iota
	INITIALIZING
	ORPHANED
	REPAIRING
	TOMBSTONED
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
	environmentsRepaired       *prometheus.CounterVec
	podsCreated                *prometheus.CounterVec
	podCreationErrors          *prometheus.CounterVec
	podsDeleted                *prometheus.CounterVec
	podsDeletionErrors         *prometheus.CounterVec
}

// newBuilderMetrics creates a new builderMetrics with the given scope.
func newBuilderMetrics(scope promutils.Scope) builderMetrics {
	return builderMetrics{
		environmentsCreated:        scope.MustNewCounterVec("env_created", "The number of environments created", "project", "domain"),
		environmentsGCed:           scope.MustNewCounterVec("env_gced", "The number of environments garbage collected", "project", "domain"),
		environmentOrphansDetected: scope.MustNewCounterVec("env_orphans_detected", "The number of orphaned environments detected", "project", "domain"),
		environmentsRepaired: scope.MustNewCounterVec("environments_repaired_total",
			"Total number of environments successfully repaired", "project", "domain"),
		podsCreated: scope.MustNewCounterVec("pods_created_total",
			"Total number of pods recreated during repair", "project", "domain", "purpose"),
		podCreationErrors: scope.MustNewCounterVec("pod_creation_errors_total",
			"Total number of errors encountered during pod creation", "project", "domain", "purpose"),
		podsDeleted: scope.MustNewCounterVec("pods_deleted_total",
			"Total number of pods deleted", "project", "domain", "purpose"),
		podsDeletionErrors: scope.MustNewCounterVec("pod_deletion_errors_total",
			"Total number of errors encountered during pod deletion", "project", "domain", "purpose"),
	}
}

// environment represents a managed fast task environment, including it's definition and current
// state
type environment struct {
	extant         *_struct.Struct
	lastAccessedAt time.Time
	name           string
	replicas       []string
	spec           *pb.FastTaskEnvironmentSpec
	state          state
	// Adds the execution environment ID to the environment struct as well for easier lookup
	executionEnvID core.ExecutionEnvID
}

// InMemoryEnvBuilder is an in-memory implementation of the ExecutionEnvBuilder interface. It is
// used to manage the lifecycle of fast task environments.
type InMemoryEnvBuilder struct {
	environments map[string]*environment
	kubeClient   core.KubeClient
	lock         sync.Mutex
	metrics      builderMetrics
	randSource   *rand.Rand
}

// Get retrieves the environment with the given execution environment ID. If the environment does
// not exist or has been tombstoned, nil is returned.
func (i *InMemoryEnvBuilder) Get(ctx context.Context, executionEnvID core.ExecutionEnvID) *_struct.Struct {
	if environment := i.environments[executionEnvID.String()]; environment != nil {
		i.lock.Lock()
		defer i.lock.Unlock()

		if environment.state != INITIALIZING && environment.state != TOMBSTONED {
			environment.lastAccessedAt = time.Now()
			return environment.extant
		}
	}
	logger.Debugf(ctx, "environment '%s' not found", executionEnvID)
	return nil
}

// Create creates a new fast task environment with the given execution environment ID and
// specification. If the environment already exists, the existing environment is returned.
func (i *InMemoryEnvBuilder) Create(ctx context.Context, executionEnvID core.ExecutionEnvID, spec *_struct.Struct) (*_struct.Struct, error) {
	// unmarshall and validate FastTaskEnvironmentSpec
	fastTaskEnvironmentSpec := &pb.FastTaskEnvironmentSpec{}
	if err := utils.UnmarshalStruct(spec, fastTaskEnvironmentSpec); err != nil {
		return nil, err
	}

	if err := isValidEnvironmentSpec(executionEnvID, fastTaskEnvironmentSpec); err != nil {
		return nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"detected invalid EnvironmentSpec for environment '%s', Err: [%v]", executionEnvID, err)
	}

	logger.Debugf(ctx, "creating environment '%s'", executionEnvID)

	// build fastTaskEnvironment extant
	fastTaskEnvironment := &pb.FastTaskEnvironment{
		QueueId: executionEnvID.String(),
	}
	environmentStruct := &_struct.Struct{}
	if err := utils.MarshalStruct(fastTaskEnvironment, environmentStruct); err != nil {
		return nil, fmt.Errorf("unable to marshal ExecutionEnv [%v], Err: [%v]", fastTaskEnvironment, err.Error())
	}

	// create environment
	i.lock.Lock()

	env, exists := i.environments[executionEnvID.String()]
	if exists && env.state != ORPHANED {
		i.lock.Unlock()
		logger.Debugf(ctx, "environment '%s' already exists", executionEnvID)
		// if exists we created from another task in race condition between `Get` and `Create`
		return env.extant, nil
	}

	var replicas []string
	if env != nil {
		// if environment already exists then copy existing replicas
		replicas = env.replicas
	} else {
		replicas = make([]string, 0)
	}

	env = &environment{
		extant:         environmentStruct,
		lastAccessedAt: time.Now(),
		name:           executionEnvID.Name,
		replicas:       replicas,
		spec:           fastTaskEnvironmentSpec,
		state:          INITIALIZING,
		executionEnvID: executionEnvID,
	}

	podNames := make([]string, 0)
	for replica := len(env.replicas); replica < int(fastTaskEnvironmentSpec.GetReplicaCount()); replica++ {
		podName := i.getPodName(env.name)
		env.replicas = append(env.replicas, podName)
		podNames = append(podNames, podName)
	}

	i.environments[executionEnvID.String()] = env
	metricLabels := []string{executionEnvID.Project, executionEnvID.Domain}
	i.metrics.environmentsCreated.WithLabelValues(metricLabels...).Inc()

	i.lock.Unlock()

	metricLabels = append(metricLabels, "create")
	// create replicas
	var errorMessages []string
	for _, podName := range podNames {
		logger.Debugf(ctx, "creating pod '%s' for environment '%s'", podName, executionEnvID)
		err := i.createPod(ctx, fastTaskEnvironmentSpec, executionEnvID, podName)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, executionEnvID, err)
			errorMessages = append(errorMessages, fmt.Sprintf("pod '%s': %v", podName, err))
			i.metrics.podCreationErrors.WithLabelValues(metricLabels...).Inc()
		} else {
			i.metrics.podsCreated.WithLabelValues(metricLabels...).Inc()
		}
	}

	i.lock.Lock()
	defer i.lock.Unlock()

	// return error only if all pods failed to create
	// NOTE: this is a temporary solution to ensure that the task fails if all pods fail to create.
	// This will be removed when persistent replica metadata handles observability issues. @pvditt @hamersaw
	if len(errorMessages) == len(podNames) {
		delete(i.environments, executionEnvID.String())

		logger.Errorf(ctx, "failed to create any pods for environment '%s': %s", executionEnvID, strings.Join(errorMessages, "; "))
		return nil, fmt.Errorf("failed to create any pods for environment '%s': %s", executionEnvID, strings.Join(errorMessages, "; "))
	}

	env.state = HEALTHY
	logger.Infof(ctx, "created environment '%s'", executionEnvID)

	return env.extant, nil
}

func (i *InMemoryEnvBuilder) getPodName(envName string) string {
	nonceBytes := make([]byte, (GetConfig().NonceLength+1)/2)
	if _, err := i.randSource.Read(nonceBytes); err != nil {
		return ""
	}

	return fmt.Sprintf("%s-%s", sanitizeEnvName(envName), hex.EncodeToString(nonceBytes)[:GetConfig().NonceLength])
}

// Status returns the status of the environment with the given execution environment ID. This
// includes the details of each pod in the environment replica set.
func (i *InMemoryEnvBuilder) Status(ctx context.Context, executionEnvID core.ExecutionEnvID) (interface{}, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	statuses := make(map[string]*v1.Pod, 0)

	// check if environment exists
	environment, exists := i.environments[executionEnvID.String()]
	if !exists {
		logger.Debugf(ctx, "environment '%s' not found", executionEnvID)
		return statuses, nil
	}

	// retrieve pod details from kubeclient cache
	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(environment.spec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
		return nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", environment.spec.GetPodTemplateSpec(), err.Error())
	}

	for _, podName := range environment.replicas {
		pod := v1.Pod{}
		err := i.kubeClient.GetCache().Get(ctx, types.NamespacedName{
			Name:      podName,
			Namespace: podTemplateSpec.Namespace,
		}, &pod)

		if isPodNotFoundErr(err) {
			statuses[podName] = nil
		} else {
			statuses[podName] = &pod
		}
	}

	return statuses, nil
}

// createPod creates a new pod for the given execution environment ID and pod name. The pod is
// created using the given FastTaskEnvironmentSpec.
func (i *InMemoryEnvBuilder) createPod(ctx context.Context, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec,
	executionEnvID core.ExecutionEnvID, podName string) error {

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
	objectMeta.Labels[EXECUTION_ENV_NAME] = executionEnvID.Name
	objectMeta.Labels[EXECUTION_ENV_VERSION] = executionEnvID.Version
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string, 0)
	}
	objectMeta.Annotations[TTL_SECONDS] = fmt.Sprintf("%d", int(getTTLOrDefault(fastTaskEnvironmentSpec).Seconds()))

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
		executionEnvID.String(),
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

	// use kubeclient to create worker
	return i.kubeClient.GetClient().Create(ctx, &v1.Pod{
		ObjectMeta: *objectMeta,
		Spec:       *podSpec,
	})
}

// deletePod deletes the pod with the given name and namespace.
func (i *InMemoryEnvBuilder) deletePod(ctx context.Context, name types.NamespacedName) error {
	objectMeta := metav1.ObjectMeta{
		Name:      name.Name,
		Namespace: name.Namespace,
	}

	return i.kubeClient.GetClient().Delete(ctx, &v1.Pod{
		ObjectMeta: objectMeta,
	}, client.GracePeriodSeconds(0))
}

// Start starts the environment garbage collector, orphan detection, and environment repair
// processes.
func (i *InMemoryEnvBuilder) Start(ctx context.Context) error {
	// detect orphaned environments using k8s client
	if err := i.detectOrphanedEnvironments(ctx, i.kubeClient.GetClient()); err != nil {
		return err
	}

	// start environment garbage collector
	go func() {
		wait.UntilWithContext(ctx,
			func(ctx context.Context) {
				if err := i.gcEnvironments(ctx); err != nil {
					logger.Warnf(ctx, "failed to gc environment(s) [%v]", err)
				}
			},
			GetConfig().EnvGCInterval.Duration,
		)
	}()

	// start environment repair
	go func() {
		wait.UntilWithContext(ctx,
			func(ctx context.Context) {
				if err := i.repairEnvironments(ctx); err != nil {
					logger.Warnf(ctx, "failed to repair environment(s) [%v]", err)
				}
			},
			GetConfig().EnvRepairInterval.Duration,
		)
	}()

	// start orphan detection
	go func() {
		wait.UntilWithContext(ctx,
			func(ctx context.Context) {
				if err := i.detectOrphanedEnvironments(ctx, i.kubeClient.GetCache()); err != nil {
					logger.Warnf(ctx, "failed to detect orphaned environment(s) [%v]", err)
				}
			},
			GetConfig().EnvDetectOrphanInterval.Duration,
		)
	}()

	return nil
}

// gcEnvironments garbage collects environments that have expired based on their termination
// criteria.
func (i *InMemoryEnvBuilder) gcEnvironments(ctx context.Context) error {
	// identify environments that have expired
	environmentReplicas := make(map[string][]types.NamespacedName, 0)

	i.lock.Lock()
	for environmentID, environment := range i.environments {
		if environment.state == REPAIRING {
			continue
		}

		// currently ttlSeconds is the only supported termination criteria for environments. if more are
		// added this logic will need to be updated.
		if environment.state == TOMBSTONED || time.Since(environment.lastAccessedAt) >= getTTLOrDefault(environment.spec) {
			environment.state = TOMBSTONED

			podTemplateSpec := &v1.PodTemplateSpec{}
			if err := json.Unmarshal(environment.spec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
				return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
					"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", environment.spec.GetPodTemplateSpec(), err.Error())
			}

			podNames := make([]types.NamespacedName, 0)
			for _, podName := range environment.replicas {
				podNames = append(podNames,
					types.NamespacedName{
						Name:      podName,
						Namespace: podTemplateSpec.Namespace,
					})
			}

			logger.Infof(ctx, "tombstoning environment '%s'", environmentID)
			environmentReplicas[environmentID] = podNames
		}
	}
	i.lock.Unlock()

	// delete environments
	deletedEnvironments := make([]string, 0)
	for environmentID, podNames := range environmentReplicas {
		deleted := true
		metricLabels := append(i.getMetricLabels(environmentID), "gc")
		for _, podName := range podNames {
			logger.Debugf(ctx, "deleting pod '%s' for environment '%s'", podName, environmentID)
			err := i.deletePod(ctx, podName)
			if err != nil && !k8serrors.IsNotFound(err) {
				logger.Warnf(ctx, "failed to gc pod '%s' for environment '%s' [%v]", podName, environmentID, err)
				deleted = false
				i.metrics.podsDeletionErrors.WithLabelValues(metricLabels...).Inc()
			} else {
				i.metrics.podsDeleted.WithLabelValues(metricLabels...).Inc()
			}
		}

		if deleted {
			deletedEnvironments = append(deletedEnvironments, environmentID)
		}
	}

	// remove deleted environments
	i.lock.Lock()
	for _, environmentID := range deletedEnvironments {
		metricLabels := i.getMetricLabels(environmentID)
		logger.Infof(ctx, "garbage collected environment '%s'", environmentID)
		i.metrics.environmentsGCed.WithLabelValues(metricLabels...).Inc()
		delete(i.environments, environmentID)
	}
	i.lock.Unlock()

	return nil
}

// getMetricLabels returns the metric labels for the given environment ID.
func (i *InMemoryEnvBuilder) getMetricLabels(environmentID string) []string {
	environment, exists := i.environments[environmentID]
	if !exists {
		return nil
	}
	return []string{environment.executionEnvID.Project, environment.executionEnvID.Domain}
}

// repairEnvironments repairs environments that have been externally modified (ie. pod deletion).
func (i *InMemoryEnvBuilder) repairEnvironments(ctx context.Context) error {
	environmentSpecs := make(map[string]pb.FastTaskEnvironmentSpec, 0)
	environmentReplicas := make(map[string][]string, 0)

	// identify environments in need of repair
	i.lock.Lock()
	pod := &v1.Pod{}
	for environmentID, env := range i.environments {
		// check if environment is repairable (ie. HEALTHY or REPAIRING state)
		if env.state != HEALTHY && env.state != REPAIRING {
			continue
		}

		podTemplateSpec := &v1.PodTemplateSpec{}
		if err := json.Unmarshal(env.spec.GetPodTemplateSpec(), podTemplateSpec); err != nil {
			return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
				"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", env.spec.GetPodTemplateSpec(), err.Error())
		}

		podNames := make([]string, 0)
		for index, podName := range env.replicas {
			err := i.kubeClient.GetCache().Get(ctx, types.NamespacedName{
				Name:      podName,
				Namespace: podTemplateSpec.Namespace,
			}, pod)

			if isPodNotFoundErr(err) {
				newPodName := i.getPodName(env.name)
				/*nonceBytes := make([]byte, (GetConfig().NonceLength+1)/2)
				if _, err := i.randSource.Read(nonceBytes); err != nil {
					return err
				}

				newPodName := fmt.Sprintf("%s-%s", env.name, hex.EncodeToString(nonceBytes)[:GetConfig().NonceLength])*/
				env.replicas[index] = newPodName
				podNames = append(podNames, newPodName)
			}
		}

		if len(podNames) > 0 {
			logger.Infof(ctx, "repairing environment '%s'", environmentID)
			env.state = REPAIRING
			environmentSpecs[environmentID] = *env.spec
			environmentReplicas[environmentID] = podNames
		}
	}
	i.lock.Unlock()

	// attempt to repair replicas
	for environmentID, environmentSpec := range environmentSpecs {
		metricLabels := append(i.getMetricLabels(environmentID), "repair")
		executionEnvID := i.environments[environmentID].executionEnvID

		for _, podName := range environmentReplicas[environmentID] {
			logger.Debugf(ctx, "creating pod '%s' for environment '%s'", podName, environmentID)
			if err := i.createPod(ctx, &environmentSpec, executionEnvID, podName); err != nil {
				logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, environmentID, err)
				i.metrics.podCreationErrors.WithLabelValues(metricLabels...).Inc()
			} else {
				i.metrics.podsCreated.WithLabelValues(metricLabels...).Inc()
			}
		}
	}

	// transition repaired environments to HEALTHY
	i.lock.Lock()
	for environmentID := range environmentSpecs {
		environment, exists := i.environments[environmentID]
		if !exists {
			// this should be unreachable as repair / gc operations use the same lock to ensure
			// concurrent operations do not interfere with each other
			logger.Warnf(ctx, "environment '%s' was deleted during repair operation", environmentID)
			continue
		}
		logger.Infof(ctx, "repaired environment '%s'", environmentID)
		i.metrics.environmentsRepaired.WithLabelValues(i.getMetricLabels(environmentID)...).Inc()

		environment.state = HEALTHY
	}
	i.lock.Unlock()

	return nil
}

// detectOrphanedEnvironments detects orphaned environments by identifying pods with the fast task
// execution type label.
func (i *InMemoryEnvBuilder) detectOrphanedEnvironments(ctx context.Context, k8sReader client.Reader) error {
	// retrieve all pods with fast task execution type label
	matchingLabelsOption := client.MatchingLabels{}
	matchingLabelsOption[EXECUTION_ENV_TYPE] = fastTaskType

	podList := &v1.PodList{}
	if err := k8sReader.List(ctx, podList, matchingLabelsOption); err != nil {
		return err
	}

	// detect orphaned environments
	i.lock.Lock()
	defer i.lock.Unlock()

	orphanedEnvironments := make(map[string]*environment, 0)
	orphanedPods := make(map[string][]string, 0)
	for _, pod := range podList.Items {
		// check if environment already exists or pod is accounted for
		executionEnvID, err := parseExectionEnvID(pod.GetLabels())
		if err != nil {
			logger.Warnf(ctx, "failed to parse ExecutionEnvID [%v]", err)
			continue
		}

		environmentID := executionEnvID.String()
		if environment, environmentExists := i.environments[environmentID]; environmentExists {
			found := false
			for _, replica := range environment.replicas {
				if replica == pod.Name {
					found = true
					break
				}
			}

			if !found {
				orphanedPods[environmentID] = append(orphanedPods[environmentID], pod.Name)
			}

			continue
		}

		// create or add pod to orphaned environment
		orphanedEnvironment, exists := orphanedEnvironments[environmentID]
		if !exists {
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
			orphanedEnvironment = &environment{
				extant:         nil,
				lastAccessedAt: time.Now(),
				name:           "orphaned",
				replicas:       make([]string, 0),
				spec: &pb.FastTaskEnvironmentSpec{
					PodTemplateSpec: podTemplateSpecBytes,
					TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
						TtlSeconds: int32(ttlSeconds),
					},
				},
				state:          ORPHANED,
				executionEnvID: executionEnvID,
			}

			orphanedEnvironments[environmentID] = orphanedEnvironment
		}

		orphanedEnvironment.replicas = append(orphanedEnvironment.replicas, pod.Name)
	}

	// copy orphaned environments to env builder
	for environmentID, orphanedEnvironment := range orphanedEnvironments {
		logger.Infof(ctx, "detected orphaned environment '%s'", environmentID)
		i.environments[environmentID] = orphanedEnvironment
		// Get the labels after adding to environments map.
		metricsLabels := i.getMetricLabels(environmentID)
		i.metrics.environmentOrphansDetected.WithLabelValues(metricsLabels...).Inc()
	}

	for environmentID, podNames := range orphanedPods {
		logger.Infof(ctx, "detected orphaned pod(s) '%v' for environment '%s'", podNames, environmentID)
		i.environments[environmentID].replicas = append(i.environments[environmentID].replicas, podNames...)
	}

	return nil
}

// NewEnvironmentBuilder creates a new InMemoryEnvBuilder with the given kube client.
func NewEnvironmentBuilder(kubeClient core.KubeClient, scope promutils.Scope) *InMemoryEnvBuilder {
	return &InMemoryEnvBuilder{
		environments: make(map[string]*environment),
		kubeClient:   kubeClient,
		metrics:      newBuilderMetrics(scope),
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
