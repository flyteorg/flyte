package plugin

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	_struct "github.com/golang/protobuf/ptypes/struct"

	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

type state int32

const (
	HEALTHY state = iota
	ORPHANED
	REPAIRING
	TOMBSTONED
)

const (
	EXECUTION_ENV_ID   = "execution-env-id"
	EXECUTION_ENV_TYPE = "execution-env-type"
	TTL_SECONDS        = "ttl-seconds"
)

// environment represents a managed fast task environment, including it's definition and current
// state
type environment struct {
	lastAccessedAt time.Time
	extant         *_struct.Struct
	replicas       []string
	spec           *pb.FastTaskEnvironmentSpec
	state          state
}

// InMemoryEnvBuilder is an in-memory implementation of the ExecutionEnvBuilder interface. It is
// used to manage the lifecycle of fast task environments.
type InMemoryEnvBuilder struct {
	environments map[string]*environment
	kubeClient   core.KubeClient
	lock         sync.Mutex
	randSource   *rand.Rand
}

// Get retrieves the environment with the given execution environment ID. If the environment does
// not exist or has been tombstoned, nil is returned.
func (i *InMemoryEnvBuilder) Get(ctx context.Context, executionEnvID string) *_struct.Struct {
	if environment := i.environments[executionEnvID]; environment != nil {
		i.lock.Lock()
		defer i.lock.Unlock()

		if environment.state != TOMBSTONED {
			environment.lastAccessedAt = time.Now()
			return environment.extant
		}
	}
	return nil
}

// Create creates a new fast task environment with the given execution environment ID and
// specification. If the environment already exists, the existing environment is returned.
func (i *InMemoryEnvBuilder) Create(ctx context.Context, executionEnvID string, spec *_struct.Struct) (*_struct.Struct, error) {
	// unmarshall and validate FastTaskEnvironmentSpec
	fastTaskEnvironmentSpec := &pb.FastTaskEnvironmentSpec{}
	if err := utils.UnmarshalStruct(spec, fastTaskEnvironmentSpec); err != nil {
		return nil, err
	}

	if err := isValidEnvironmentSpec(fastTaskEnvironmentSpec); err != nil {
		return nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"detected invalid FastTaskEnvironmentSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.PodTemplateSpec, err)
	}

	logger.Debug(ctx, "creating environment '%s'", executionEnvID)

	// build fastTaskEnvironment extant
	fastTaskEnvironment := &pb.FastTaskEnvironment{
		QueueId: executionEnvID,
	}
	environmentStruct := &_struct.Struct{}
	if err := utils.MarshalStruct(fastTaskEnvironment, environmentStruct); err != nil {
		return nil, fmt.Errorf("unable to marshal ExecutionEnv [%v], Err: [%v]", fastTaskEnvironment, err.Error())
	}

	// create environment
	i.lock.Lock()

	env, exists := i.environments[executionEnvID]
	if exists && env.state != ORPHANED {
		i.lock.Unlock()

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
		lastAccessedAt: time.Now(),
		extant:         environmentStruct,
		replicas:       replicas,
		spec:           fastTaskEnvironmentSpec,
		state:          HEALTHY,
	}

	podNames := make([]string, 0)
	for replica := len(env.replicas); replica < int(fastTaskEnvironmentSpec.ReplicaCount); replica++ {
		nonceBytes := make([]byte, (GetConfig().NonceLength+1)/2)
		if _, err := i.randSource.Read(nonceBytes); err != nil {
			return nil, err
		}

		podName := fmt.Sprintf("%s-%s", executionEnvID, hex.EncodeToString(nonceBytes)[:GetConfig().NonceLength])
		env.replicas = append(env.replicas, podName)
		podNames = append(podNames, podName)
	}

	i.environments[executionEnvID] = env
	i.lock.Unlock()

	// create replicas
	for _, podName := range podNames {
		logger.Debugf(ctx, "creating pod '%s' for environment '%s'", podName, executionEnvID)
		if err := i.createPod(ctx, fastTaskEnvironmentSpec, executionEnvID, podName); err != nil {
			logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, executionEnvID, err)
		}
	}

	logger.Infof(ctx, "created environment '%s'", executionEnvID)
	return env.extant, nil
}

// createPod creates a new pod for the given execution environment ID and pod name. The pod is
// created using the given FastTaskEnvironmentSpec.
func (i *InMemoryEnvBuilder) createPod(ctx context.Context, fastTaskEnvironmentSpec *pb.FastTaskEnvironmentSpec, executionEnvID, podName string) error {
	podTemplateSpec := &v1.PodTemplateSpec{}
	if err := json.Unmarshal(fastTaskEnvironmentSpec.PodTemplateSpec, podTemplateSpec); err != nil {
		return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
			"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", fastTaskEnvironmentSpec.PodTemplateSpec, err.Error())
	}

	podSpec := &podTemplateSpec.Spec
	objectMeta := &podTemplateSpec.ObjectMeta

	// identify the primary container
	primaryContainerIndex := -1
	if len(fastTaskEnvironmentSpec.PrimaryContainerName) > 0 {
		for index, container := range podSpec.Containers {
			if container.Name == fastTaskEnvironmentSpec.PrimaryContainerName {
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
	objectMeta.Labels[EXECUTION_ENV_ID] = executionEnvID
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = make(map[string]string, 0)
	}
	objectMeta.Annotations[TTL_SECONDS] = fmt.Sprintf("%d", fastTaskEnvironmentSpec.GetTtlSeconds())

	// create new volume 'workdir'
	podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
		Name: "workdir",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	})

	// add init container to copy worker binary to volume
	podSpec.InitContainers = append(podSpec.InitContainers, v1.Container{
		Name:            "worker",
		Image:           GetConfig().Image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         []string{"cp", "/usr/local/bin/worker", "/tmp/worker"},
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    GetConfig().InitContainerCPU,
				v1.ResourceMemory: GetConfig().InitContainerMemory,
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:    GetConfig().InitContainerCPU,
				v1.ResourceMemory: GetConfig().InitContainerMemory,
			},
		},
		VolumeMounts: []v1.VolumeMount{
			{
				Name:      "workdir",
				MountPath: "/tmp",
			},
		},
	})

	// update primary container arguments and volume mounts
	container := &podSpec.Containers[primaryContainerIndex]
	container.Args = []string{
		"/tmp/worker",
		"bridge",
		"--queue-id",
		executionEnvID,
		"--fasttask-url",
		GetConfig().CallbackURI,
	}

	if fastTaskEnvironmentSpec.BacklogLength > 0 {
		container.Args = append(container.Args, "--backlog-length", fmt.Sprintf("%d", fastTaskEnvironmentSpec.BacklogLength))
	}
	if fastTaskEnvironmentSpec.Parallelism > 0 {
		container.Args = append(container.Args, "--parallelism", fmt.Sprintf("%d", fastTaskEnvironmentSpec.Parallelism))
	}

	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:		"workdir",
		MountPath:	"/tmp",
	})

	// use kubeclient to create worker
	return i.kubeClient.GetClient().Create(ctx, &v1.Pod{
		ObjectMeta: *objectMeta,
		Spec: *podSpec,
	})
}

// deletePod deletes the pod with the given name and namespace.
func (i *InMemoryEnvBuilder) deletePod(ctx context.Context, name types.NamespacedName) error {
	objectMeta := metav1.ObjectMeta{
		Name: name.Name,
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
	now_seconds := time.Now().Unix()
	environmentReplicas := make(map[string][]types.NamespacedName, 0)

	i.lock.Lock()
	for environmentID, environment := range i.environments {
		if environment.state == REPAIRING {
			continue
		}

		// if the environment has a ttlSeconds termination criteria then check if it has expired
		if ttlCriteria, ok := environment.spec.GetTerminationCriteria().(*pb.FastTaskEnvironmentSpec_TtlSeconds); ok {
			if environment.state == TOMBSTONED || now_seconds - environment.lastAccessedAt.Unix() >= int64(ttlCriteria.TtlSeconds) {
				environment.state = TOMBSTONED

				podTemplateSpec := &v1.PodTemplateSpec{}
				if err := json.Unmarshal(environment.spec.PodTemplateSpec, podTemplateSpec); err != nil {
					return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
						"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", environment.spec.PodTemplateSpec, err.Error())
				}

				podNames := make([]types.NamespacedName, 0)
				for _, podName := range environment.replicas {
					podNames = append(podNames,
						types.NamespacedName{
							Name: podName,
							Namespace: podTemplateSpec.Namespace,
						})
				}

				logger.Infof(ctx, "tombstoning environment '%s'", environmentID)
				environmentReplicas[environmentID] = podNames
			}
		}
	}
	i.lock.Unlock()

	// delete environments
	deletedEnvironments := make([]string, 0)
	for environmentID, podNames := range environmentReplicas {
		deleted := true
		for _, podName := range podNames {
			logger.Debugf(ctx, "deleting pod '%s' for environment '%s'", podName, environmentID)
			err := i.deletePod(ctx, podName)
			if err != nil && !k8serrors.IsNotFound(err) {
				logger.Warnf(ctx, "failed to gc pod '%s' for environment '%s' [%v]", podName, environmentID, err)
				deleted = false
			}
		}

		if deleted {
			deletedEnvironments = append(deletedEnvironments, environmentID)
		}
	}

	// remove deleted environments
	i.lock.Lock()
	for _, environmentID := range deletedEnvironments {
		logger.Infof(ctx, "garbage collected environment '%s'", environmentID)
		delete(i.environments, environmentID)
	}
	i.lock.Unlock()

	return nil
}

// repairEnvironments repairs environments that have been externally modified (ie. pod deletion).
func (i *InMemoryEnvBuilder) repairEnvironments(ctx context.Context) error {
	environmentSpecs := make(map[string]pb.FastTaskEnvironmentSpec, 0)
	environmentReplicas := make(map[string][]string, 0)

	// identify environments in need of repair
	i.lock.Lock()
	pod := &v1.Pod{}
	for environmentID, environment := range i.environments {
		// check if environment is repairable (ie. HEALTHY or REPAIRING state)
		if environment.state != HEALTHY && environment.state != REPAIRING {
			continue
		}

		podTemplateSpec := &v1.PodTemplateSpec{}
		if err := json.Unmarshal(environment.spec.PodTemplateSpec, podTemplateSpec); err != nil {
			return flyteerrors.Errorf(flyteerrors.BadTaskSpecification,
				"unable to unmarshal PodTemplateSpec [%v], Err: [%v]", environment.spec.PodTemplateSpec, err.Error())
		}

		podNames := make([]string, 0)
		for index, podName := range environment.replicas {
			err := i.kubeClient.GetCache().Get(ctx, types.NamespacedName{
				Name: podName,
				Namespace: podTemplateSpec.Namespace,
			}, pod)

			if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) {
				nonceBytes := make([]byte, (GetConfig().NonceLength+1)/2)
				if _, err := i.randSource.Read(nonceBytes); err != nil {
					return err
				}

				newPodName := fmt.Sprintf("%s-%s", environmentID, hex.EncodeToString(nonceBytes)[:GetConfig().NonceLength])
				environment.replicas[index] = newPodName
				podNames = append(podNames, newPodName)
			}
		}

		if len(podNames) > 0 {
			logger.Infof(ctx, "repairing environment '%s'", environmentID)
			environment.state = REPAIRING
			environmentSpecs[environmentID] = *environment.spec
			environmentReplicas[environmentID] = podNames
		}
	}
	i.lock.Unlock()

	// attempt to repair replicas
	for environmentID, environmentSpec := range environmentSpecs {
		for _, podName := range environmentReplicas[environmentID] {
			logger.Debugf(ctx, "creating pod '%s' for environment '%s'", podName, environmentID)
			if err := i.createPod(ctx, &environmentSpec, environmentID, podName); err != nil {
				logger.Warnf(ctx, "failed to create pod '%s' for environment '%s' [%v]", podName, environmentID, err)
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
		environment.state = HEALTHY
	}
	i.lock.Unlock()

	return nil
}

// detectOrphanedEnvironments detects orphaned environments by identifying pods with the fast task
// execution type label.
func (i *InMemoryEnvBuilder) detectOrphanedEnvironments(ctx context.Context, k8sReader client.Reader) error {
	// retrieve all pods with fast task execution type label
	var matchingLabelsOption client.MatchingLabels
	matchingLabelsOption = make(map[string]string)
	matchingLabelsOption[EXECUTION_ENV_TYPE] = fastTaskType

	podList := &v1.PodList{}
	if err := k8sReader.List(ctx, podList, matchingLabelsOption); err != nil {
		return err
	}

	// detect orphaned environments
	i.lock.Lock()
	defer i.lock.Unlock()

	orphanedEnvironments := make(map[string]*environment, 0)
	for _, pod := range podList.Items {
		// if environment exists we do not need to process
		environmentID, labelExists := pod.Labels[EXECUTION_ENV_ID]
		if !labelExists {
			continue
		}

		_, environmentExists := i.environments[environmentID]
		if environmentExists {
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
				lastAccessedAt: time.Now(),
				extant:         nil,
				replicas:       make([]string, 0),
				spec:           &pb.FastTaskEnvironmentSpec{
					PodTemplateSpec: podTemplateSpecBytes,
					TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
						TtlSeconds: int32(ttlSeconds),
					},
				},
				state: ORPHANED,
			}

			orphanedEnvironments[environmentID] = orphanedEnvironment
		}

		orphanedEnvironment.replicas = append(orphanedEnvironment.replicas, pod.Name)
	}

	// copy orphaned environments to env builder
	for environmentID, orphanedEnvironment := range orphanedEnvironments {
		logger.Infof(ctx, "detected orphaned environment '%s'", environmentID)
		i.environments[environmentID] = orphanedEnvironment
	}

	return nil
}

// NewEnvironmentBuilder creates a new InMemoryEnvBuilder with the given kube client.
func NewEnvironmentBuilder(kubeClient core.KubeClient) *InMemoryEnvBuilder {
	return &InMemoryEnvBuilder{
		environments: make(map[string]*environment),
		kubeClient:   kubeClient,
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}
