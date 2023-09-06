package containerwatcher

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/logger"
	v13 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type ContainerInformation struct {
	Namespace string
	PodName   string
	Name      string
}

type kubeAPIWatcher struct {
	coreClient v1.CoreV1Interface
	info       ContainerInformation
}

func (k kubeAPIWatcher) wait(ctx context.Context, info ContainerInformation, f func(ctx context.Context, pod *v13.Pod) (bool, error)) error {
	s := fields.OneTermEqualSelector("metadata.name", info.PodName)
	watcher, err := k.coreClient.Pods(info.Namespace).Watch(ctx, v12.ListOptions{
		Watch:         true,
		FieldSelector: s.String(),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()
	for {
		select {
		case v := <-watcher.ResultChan():
			p := v.Object.(*v13.Pod)
			if stop, err := f(ctx, p); err != nil {
				return err
			} else if stop {
				return nil
			}
		case <-ctx.Done():
			logger.Infof(ctx, "Pod [%s/%s] watcher canceled", info.Namespace, info.PodName)
			return ErrTimeout
		}
	}
}

func (k kubeAPIWatcher) WaitToStart(ctx context.Context) error {
	f := func(ctx context.Context, p *v13.Pod) (bool, error) {
		for _, c := range p.Status.ContainerStatuses {
			if c.Name == k.info.Name {
				if c.Ready {
					return true, nil
				}
			}
		}
		return false, nil
	}
	return k.wait(ctx, k.info, f)
}

func (k kubeAPIWatcher) WaitToExit(ctx context.Context) error {
	f := func(ctx context.Context, p *v13.Pod) (bool, error) {
		for _, c := range p.Status.ContainerStatuses {
			if c.Name == k.info.Name {
				// TODO we may want to check failure reason and return that?
				if c.State.Terminated != nil {
					if c.State.Terminated.ExitCode != 0 {
						return true, fmt.Errorf("Container exited with Exit code [%d]. Reason [%s]%s. ", c.State.Terminated.ExitCode, c.State.Terminated.Reason, c.State.Terminated.Message)
					}
					return true, nil
				}
			}
		}
		return false, nil
	}
	return k.wait(ctx, k.info, f)
}

func NewKubeAPIWatcher(_ context.Context, clientConfig clientcmd.ClientConfig, info ContainerInformation) (Watcher, error) {
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return kubeAPIWatcher{info: info}, err
	}
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return kubeAPIWatcher{info: info}, err
	}
	return kubeAPIWatcher{coreClient: kubeClient.CoreV1(), info: info}, nil
}
