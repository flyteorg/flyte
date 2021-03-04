package containerwatcher

import (
	"context"
	"os"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/mitchellh/go-ps"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	k8sPauseContainerPid = 1
	k8sAllowedParentPid  = 0
)

// given list of processes returns a list of processes such that,
// the pid does not match any of the given filterPids, this is to filter the /pause and current process.
// and the parentPid is the allowedParentPid. The logic for this is because every process in the shared namespace
// always has a parent pid of 0
func FilterProcessList(procs []ps.Process, filterPids sets.Int, allowedParentPid int) ([]ps.Process, error) {
	var filteredProcs []ps.Process
	for _, p := range procs {
		proc := p
		if proc.PPid() == allowedParentPid {
			if !filterPids.Has(proc.Pid()) {
				filteredProcs = append(filteredProcs, proc)
			}
		}
	}
	return filteredProcs, nil
}

type SharedNamespaceProcessLister struct {
	// PID for the current process
	currentProcessPid int
	pidsToFilter      sets.Int
}

func (s *SharedNamespaceProcessLister) AnyProcessRunning(ctx context.Context) (bool, error) {
	procs, err := s.ListRunningProcesses(ctx)
	if err != nil {
		return false, err
	}
	return len(procs) > 0, nil
}

// Polls all processes and returns a filtered set. Refer to FilterProcessList for understanding the process of filtering
func (s *SharedNamespaceProcessLister) ListRunningProcesses(ctx context.Context) ([]ps.Process, error) {
	procs, err := ps.Processes()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to list processes")
	}
	filteredProcs, err := FilterProcessList(procs, s.pidsToFilter, k8sAllowedParentPid)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to filter processes")
	}
	return filteredProcs, nil
}

// The best option for this is to use https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/
// This is only available as Beta as of 1.16, so we will launch with this feature only as beta
// But this is the most efficient way to monitor the pod
type sharedProcessNSWatcher struct {
	// Rate at which to poll the process list
	pollInterval time.Duration
	// Number of cycles to wait before finalizing exit of container
	cyclesToWait int

	s SharedNamespaceProcessLister
}

func (k sharedProcessNSWatcher) wait(ctx context.Context, cyclesToWait int, f func(ctx context.Context, otherProcessRunning bool) bool) error {
	t := time.NewTicker(k.pollInterval)
	defer t.Stop()
	cyclesOfMissingProcesses := 0
	for {
		select {
		case <-ctx.Done():
			logger.Infof(ctx, "Context canceled")
			return ErrTimeout
		case <-t.C:
			logger.Infof(ctx, "Checking processes to see if any process were started...")
			yes, err := k.s.AnyProcessRunning(ctx)
			if err != nil {
				return err
			}
			if f(ctx, yes) {
				cyclesOfMissingProcesses++
				if cyclesOfMissingProcesses >= cyclesToWait {
					logger.Infof(ctx, "Exiting wait loop")
					return nil
				}
			}
			logger.Infof(ctx, "process not yet started")
		}
	}
}

func (k sharedProcessNSWatcher) WaitToStart(ctx context.Context) error {
	logger.Infof(ctx, "SNPS Watcher waiting for other processes to start")
	defer logger.Infof(ctx, "SNPS Watcher detected process start")
	return k.wait(ctx, 1, func(ctx context.Context, otherProcessRunning bool) bool {
		return otherProcessRunning
	})
}

func (k sharedProcessNSWatcher) WaitToExit(ctx context.Context) error {
	logger.Infof(ctx, "SNPS Watcher waiting for other process to exit")
	defer logger.Infof(ctx, "SNPS Watcher detected process exit")
	return k.wait(ctx, k.cyclesToWait, func(ctx context.Context, otherProcessRunning bool) bool {
		return !otherProcessRunning
	})
}

// c -> clock.Clock allows for injecting a fake clock. The watcher uses a timer
// pollInterval -> time.Duration, wait for this amount of time between successive process checks
// waitNumIntervalsBeforeFinalize -> Number of successive poll intervals of missing processes for the container, before assuming process is complete. 0/1 indicate the first time a process is detected to be missing, the wait if finalized.
// containerStartupTimeout -> Duration for which to wait for the container to start up. If the container has not started up in this time, exit with error.
func NewSharedProcessNSWatcher(ctx context.Context, pollInterval time.Duration, waitNumIntervalsBeforeFinalize int) (Watcher, error) {
	logger.Infof(ctx, "SNPS created with poll interval %s", pollInterval.String())
	currentPid := os.Getpid()
	return sharedProcessNSWatcher{
		pollInterval: pollInterval,
		cyclesToWait: waitNumIntervalsBeforeFinalize,
		s: SharedNamespaceProcessLister{
			currentProcessPid: currentPid,
			pidsToFilter:      sets.NewInt(currentPid, k8sPauseContainerPid),
		},
	}, nil
}
