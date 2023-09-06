package manager

import (
	"context"
	"fmt"
	"time"

	managerConfig "github.com/flyteorg/flytepropeller/manager/config"
	"github.com/flyteorg/flytepropeller/manager/shardstrategy"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	leader "github.com/flyteorg/flytepropeller/pkg/leaderelection"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	stderrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/imdario/mergo"

	"github.com/prometheus/client_golang/prometheus"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
)

const (
	podTemplateResourceVersion = "podTemplateResourceVersion"
	shardConfigHash            = "shardConfigHash"
)

type metrics struct {
	Scope       promutils.Scope
	RoundTime   promutils.StopWatch
	PodsCreated prometheus.Counter
	PodsDeleted prometheus.Counter
	PodsRunning prometheus.Gauge
}

func newManagerMetrics(scope promutils.Scope) *metrics {
	return &metrics{
		Scope:       scope,
		RoundTime:   scope.MustNewStopWatch("round_time", "Time to perform one round of validating managed pod status'", time.Millisecond),
		PodsCreated: scope.MustNewCounter("pods_created_count", "Total number of pods created"),
		PodsDeleted: scope.MustNewCounter("pods_deleted_count", "Total number of pods deleted"),
		PodsRunning: scope.MustNewGauge("pods_running_count", "Number of managed pods currently running"),
	}
}

// Manager periodically scans k8s to ensure liveness of multiple FlytePropeller controller instances
// and rectifies state based on the configured sharding strategy.
type Manager struct {
	kubeClient               kubernetes.Interface
	leaderElector            *leaderelection.LeaderElector
	metrics                  *metrics
	ownerReferences          []metav1.OwnerReference
	podApplication           string
	podNamespace             string
	podTemplateContainerName string
	podTemplateName          string
	podTemplateNamespace     string
	scanInterval             time.Duration
	shardStrategy            shardstrategy.ShardStrategy
}

func (m *Manager) createPods(ctx context.Context) error {
	t := m.metrics.RoundTime.Start()
	defer t.Stop()

	// retrieve pod metadata
	podTemplate, err := m.kubeClient.CoreV1().PodTemplates(m.podTemplateNamespace).Get(ctx, m.podTemplateName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to retrieve pod template '%s' from namespace '%s' [%v]", m.podTemplateName, m.podTemplateNamespace, err)
	}

	shardConfigHash, err := m.shardStrategy.HashCode()
	if err != nil {
		return err
	}

	podAnnotations := map[string]string{
		"podTemplateResourceVersion": podTemplate.ObjectMeta.ResourceVersion,
		"shardConfigHash":            fmt.Sprintf("%d", shardConfigHash),
	}
	podNames := m.getPodNames()
	podLabels := map[string]string{
		"app": m.podApplication,
	}

	// retrieve existing pods
	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(podLabels).String(),
	}

	pods, err := m.kubeClient.CoreV1().Pods(m.podNamespace).List(ctx, listOptions)
	if err != nil {
		return err
	}

	// note: we are unable to short-circuit if 'len(pods) == len(m.podNames)' because there may be
	// unmanaged flytepropeller pods - which is invalid configuration but will be detected later

	// determine missing managed pods
	podExists := make(map[string]bool)
	for _, podName := range podNames {
		podExists[podName] = false
	}

	podsRunning := 0
	for _, pod := range pods.Items {
		podName := pod.ObjectMeta.Name

		// validate existing pod annotations
		deletePod := false
		for key, value := range podAnnotations {
			if pod.ObjectMeta.Annotations[key] != value {
				logger.Infof(ctx, "detected pod '%s' with stale configuration", podName)
				deletePod = true
				break
			}
		}

		if pod.Status.Phase == v1.PodFailed {
			logger.Warnf(ctx, "detected pod '%s' in 'failed' state", podName)
			deletePod = true
		}

		if deletePod {
			err := m.kubeClient.CoreV1().Pods(m.podNamespace).Delete(ctx, podName, metav1.DeleteOptions{})
			if err != nil {
				return err
			}

			m.metrics.PodsDeleted.Inc()
			logger.Infof(ctx, "deleted pod '%s'", podName)
			continue
		}

		// update podExists to track existing pods
		if _, ok := podExists[podName]; ok {
			podExists[podName] = true

			if pod.Status.Phase == v1.PodRunning {
				podsRunning++
			}
		}
	}

	m.metrics.PodsRunning.Set(float64(podsRunning))

	// create non-existent pods
	errs := stderrors.ErrorCollection{}
	for i, podName := range podNames {
		if exists := podExists[podName]; !exists {
			// initialize pod definition
			baseObjectMeta := podTemplate.Template.ObjectMeta.DeepCopy()
			objectMeta := metav1.ObjectMeta{
				Annotations:     podAnnotations,
				Name:            podName,
				Namespace:       m.podNamespace,
				Labels:          podLabels,
				OwnerReferences: m.ownerReferences,
			}

			err = mergo.Merge(baseObjectMeta, objectMeta, mergo.WithOverride, mergo.WithAppendSlice)
			if err != nil {
				errs.Append(fmt.Errorf("failed to initialize pod ObjectMeta for '%s' [%v]", podName, err))
				continue
			}

			pod := &v1.Pod{
				ObjectMeta: *baseObjectMeta,
				Spec:       *podTemplate.Template.Spec.DeepCopy(),
			}

			err := m.shardStrategy.UpdatePodSpec(&pod.Spec, m.podTemplateContainerName, i)
			if err != nil {
				errs.Append(fmt.Errorf("failed to update pod spec for '%s' [%v]", podName, err))
				continue
			}

			// override leader election namespaced name on managed flytepropeller instances
			container, err := utils.GetContainer(&pod.Spec, m.podTemplateContainerName)
			if err != nil {
				return fmt.Errorf("failed to retrieve flytepropeller container from pod template [%v]", err)
			}

			injectLeaderNameArg := fmt.Sprintf("--propeller.leader-election.lock-config-map.Name=propeller-leader-%d", i)
			container.Args = append(container.Args, injectLeaderNameArg)

			// create pod
			_, err = m.kubeClient.CoreV1().Pods(m.podNamespace).Create(ctx, pod, metav1.CreateOptions{})
			if err != nil {
				errs.Append(fmt.Errorf("failed to create pod '%s' [%v]", podName, err))
				continue
			}

			m.metrics.PodsCreated.Inc()
			logger.Infof(ctx, "created pod '%s'", podName)
		}
	}

	return errs.ErrorOrDefault()
}

func (m *Manager) getPodNames() []string {
	podCount := m.shardStrategy.GetPodCount()
	var podNames []string
	for i := 0; i < podCount; i++ {
		podNames = append(podNames, fmt.Sprintf("%s-%d", m.podApplication, i))
	}

	return podNames
}

// Run starts the manager instance as either a k8s leader, if configured, or as a standalone process.
func (m *Manager) Run(ctx context.Context) error {
	if m.leaderElector != nil {
		logger.Infof(ctx, "running with leader election")
		m.leaderElector.Run(ctx)
	} else {
		logger.Infof(ctx, "running without leader election")
		if err := m.run(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) run(ctx context.Context) error {
	logger.Infof(ctx, "started manager")
	wait.UntilWithContext(ctx,
		func(ctx context.Context) {
			logger.Debugf(ctx, "validating managed pod(s) state")
			err := m.createPods(ctx)
			if err != nil {
				logger.Errorf(ctx, "failed to create pod(s) [%v]", err)
			}
		},
		m.scanInterval,
	)

	logger.Infof(ctx, "shutting down manager")
	return nil
}

// New creates a new FlytePropeller Manager instance.
func New(ctx context.Context, propellerCfg *propellerConfig.Config, cfg *managerConfig.Config, podNamespace string, ownerReferences []metav1.OwnerReference, kubeClient kubernetes.Interface, scope promutils.Scope) (*Manager, error) {
	shardStrategy, err := shardstrategy.NewShardStrategy(ctx, cfg.ShardConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize shard strategy [%v]", err)
	}

	manager := &Manager{
		kubeClient:               kubeClient,
		metrics:                  newManagerMetrics(scope),
		ownerReferences:          ownerReferences,
		podApplication:           cfg.PodApplication,
		podNamespace:             podNamespace,
		podTemplateContainerName: cfg.PodTemplateContainerName,
		podTemplateName:          cfg.PodTemplateName,
		podTemplateNamespace:     cfg.PodTemplateNamespace,
		scanInterval:             cfg.ScanInterval.Duration,
		shardStrategy:            shardStrategy,
	}

	// configure leader elector
	eventRecorder, err := utils.NewK8sEventRecorder(ctx, kubeClient, "flytepropeller-manager", propellerCfg.PublishK8sEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize k8s event recorder [%v]", err)
	}

	lock, err := leader.NewResourceLock(kubeClient.CoreV1(), kubeClient.CoordinationV1(), eventRecorder, propellerCfg.LeaderElection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resource lock [%v]", err)
	}

	if lock != nil {
		logger.Infof(ctx, "creating leader elector for the controller")
		manager.leaderElector, err = leader.NewLeaderElector(
			lock,
			propellerCfg.LeaderElection,
			func(ctx context.Context) {
				logger.Infof(ctx, "started leading")
				if err := manager.run(ctx); err != nil {
					logger.Error(ctx, err)
				}
			},
			func() {
				// need to check if this elector obtained leadership until k8s client-go api is fixed. currently the
				// OnStoppingLeader func is called as a defer on every elector run, regardless of election status.
				if manager.leaderElector.IsLeader() {
					logger.Info(ctx, "stopped leading")
				}
			})

		if err != nil {
			return nil, fmt.Errorf("failed to initialize leader elector [%v]", err)
		}
	}

	return manager, nil
}
