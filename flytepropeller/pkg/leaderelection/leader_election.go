package controller

import (
	"context"
	"fmt"
	"os"

	v12 "k8s.io/client-go/kubernetes/typed/coordination/v1"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"k8s.io/apimachinery/pkg/util/rand"

	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	// Env var to lookup pod name in. In pod spec, you will have to specify it like this:
	//  env:
	//  - name: POD_NAME
	//    valueFrom:
	//      fieldRef:
	//        fieldPath: metadata.name
	podNameEnvVar = "POD_NAME"
)

// NewResourceLock creates a new config map resource lock for use in a leader election loop
func NewResourceLock(corev1 v1.CoreV1Interface, coordinationV1 v12.CoordinationV1Interface, eventRecorder record.EventRecorder, options config.LeaderElectionConfig) (
	resourcelock.Interface, error) {

	if !options.Enabled {
		return nil, nil
	}

	// Default the LeaderElectionID
	if len(options.LockConfigMap.String()) == 0 {
		return nil, fmt.Errorf("to enable leader election, a config map must be provided")
	}

	// Leader id, needs to be unique
	return resourcelock.New(resourcelock.ConfigMapsLeasesResourceLock,
		options.LockConfigMap.Namespace,
		options.LockConfigMap.Name,
		corev1,
		coordinationV1,
		resourcelock.ResourceLockConfig{
			Identity:      getUniqueLeaderID(),
			EventRecorder: eventRecorder,
		})
}

func getUniqueLeaderID() string {
	val, found := os.LookupEnv(podNameEnvVar)
	if found {
		return val
	}

	id, err := os.Hostname()
	if err != nil {
		id = ""
	}

	return fmt.Sprintf("%v_%v", id, rand.String(10))
}

func NewLeaderElector(lock resourcelock.Interface, cfg config.LeaderElectionConfig,
	leaderFn func(ctx context.Context), leaderStoppedFn func()) (*leaderelection.LeaderElector, error) {
	return leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: cfg.LeaseDuration.Duration,
		RenewDeadline: cfg.RenewDeadline.Duration,
		RetryPeriod:   cfg.RetryPeriod.Duration,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: leaderFn,
			OnStoppedLeading: leaderStoppedFn,
		},
	})
}
