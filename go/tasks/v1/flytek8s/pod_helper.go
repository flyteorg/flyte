package flytek8s

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/logger"
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

const PodKind = "pod"

func ToK8sPod(ctx context.Context, taskCtx types.TaskContext, taskContainer *core.Container, inputs *core.LiteralMap) (*v1.PodSpec, error) {
	c, err := ToK8sContainer(ctx, taskCtx, taskContainer, inputs)
	if err != nil {
		return nil, err
	}

	containers := []v1.Container{
		*c,
	}
	return &v1.PodSpec{
		// We could specify Scheduler, Affinity, nodename etc
		RestartPolicy:      v1.RestartPolicyNever,
		Containers:         containers,
		Tolerations:        GetTolerationsForResources(c.Resources),
		ServiceAccountName: taskCtx.GetK8sServiceAccount(),
	}, nil
}

func BuildPodWithSpec(podSpec *v1.PodSpec) *v1.Pod {
	pod := v1.Pod{
		TypeMeta: v12.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		Spec: *podSpec,
	}

	return &pod
}

func BuildIdentityPod() *v1.Pod {
	return &v1.Pod{
		TypeMeta: v12.TypeMeta{
			Kind:       PodKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}
}

// Important considerations.
// Pending Status in Pod could be for various reasons and sometimes could signal a problem
// Case I: Pending because the Image pull is failing and it is backing off
//         This could be transient. So we can actually rely on the failure reason.
//         The failure transitions from ErrImagePull -> ImagePullBackoff
// Case II: Not enough resources are available. This is tricky. It could be that the total number of
//          resources requested is beyond the capability of the system. for this we will rely on configuration
//          and hence input gates. We should not allow bad requests that request for large number of resource through.
//          In the case it makes through, we will fail after timeout
func DemystifyPending(status v1.PodStatus) (types.TaskStatus, error) {
	// Search over the difference conditions in the status object.  Note that the 'Pending' this function is
	// demystifying is the 'phase' of the pod status. This is different than the PodReady condition type also used below
	for _, c := range status.Conditions {
		switch c.Type {
		case v1.PodScheduled:
			if c.Status == v1.ConditionFalse {
				// Waiting to be scheduled. This usually refers to inability to acquire resources.
				return types.TaskStatusQueued, nil
			}

		case v1.PodReasonUnschedulable:
			// We Ignore case in which we are unable to find resources on the cluster. This is because
			// - The resources may be not available at the moment, but may become available eventually
			//   The pod scheduler will keep on looking at this pod and trying to satisfy it.
			//
			//  Pod status looks like this:
			// 	message: '0/1 nodes are available: 1 Insufficient memory.'
			//  reason: Unschedulable
			// 	status: "False"
			// 	type: PodScheduled
			return types.TaskStatusQueued, nil

		case v1.PodReady:
			if c.Status == v1.ConditionFalse {
				// This happens in the case the image is having some problems. In the following example, K8s is having
				// problems downloading an image. To ensure that, we will have to iterate over all the container statuses and
				// find if some container has imagepull failure
				// e.g.
				//     - lastProbeTime: null
				//      lastTransitionTime: 2018-12-18T00:57:30Z
				//      message: 'containers with unready status: [myapp-container]'
				//      reason: ContainersNotReady
				//      status: "False"
				//      type: Ready
				//
				// e.g. Container status
				//     - image: blah
				//      imageID: ""
				//      lastState: {}
				//      name: myapp-container
				//      ready: false
				//      restartCount: 0
				//      state:
				//        waiting:
				//          message: Back-off pulling image "blah"
				//          reason: ImagePullBackOff
				for _, containerStatus := range status.ContainerStatuses {
					if !containerStatus.Ready {
						if containerStatus.State.Waiting != nil {
							// There are a variety of reasons that can cause a pod to be in this waiting state.
							// Waiting state may be legitimate when the container is being downloaded, started or init containers are running
							reason := containerStatus.State.Waiting.Reason
							switch reason {
							case "ErrImagePull", "ContainerCreating", "PodInitializing":
								// But, there are only two "reasons" when a pod is successfully being created and hence it is in
								// waiting state
								// Refer to https://github.com/kubernetes/kubernetes/blob/master/pkg/kubelet/kubelet_pods.go
								// and look for the default waiting states
								// We also want to allow Image pulls to be retried, so ErrImagePull will be ignored
								// as it eventually enters into ImagePullBackOff
								// ErrImagePull -> Transitionary phase to ImagePullBackOff
								// ContainerCreating -> Image is being downloaded
								// PodInitializing -> Init containers are running
								return types.TaskStatusQueued, nil

							case "ImagePullBackOff":
								return types.TaskStatusRetryableFailure(errors.Errorf(reason,
									containerStatus.State.Waiting.Message)), nil

							case "CreateContainerError":
								// This happens if for instance the command to the container is incorrect, ie doesn't run
								return types.TaskStatusPermanentFailure(errors.Errorf(
									"CreateContainerError", containerStatus.State.Waiting.Reason)), nil

							default:
								// Since we are not checking for all error states, we may end up perpetually
								// in the queued state returned at the bottom of this function, until the Pod is reaped
								// by K8s and we get elusive 'pod not found' errors
								// So be default if the container is not waiting with the PodInitializing/ContainerCreating
								// reasons, then we will assume a failure reason, and fail instantly
								logger.Errorf(context.TODO(), "Pod pending with Waiting container with unhandled reason %s", reason)
								return types.TaskStatusRetryableFailure(errors.Errorf(reason, containerStatus.State.Waiting.Message)), nil
							}

						}
					}
				}
			}
		}
	}

	return types.TaskStatusQueued, nil
}

