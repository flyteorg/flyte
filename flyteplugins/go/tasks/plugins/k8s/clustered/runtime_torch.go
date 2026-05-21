package clustered

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// injectTorchRunEnv appends the env vars that _entrypoint.py and torchrun need.
// JOB_COMPLETION_INDEX is injected automatically by kubelet for Indexed Jobs.
func injectTorchRunEnv(container *corev1.Container, spec *clusteredpb.ClusteredTaskSpec) {
	rdzvBackend := "static"
	if tr := spec.GetRuntime().GetTorchrun(); tr != nil {
		switch tr.GetRdzvBackend() {
		case clusteredpb.RdzvBackend_C10D:
			rdzvBackend = "c10d"
		default:
			rdzvBackend = "static"
		}
	}

	container.Env = append(container.Env,
		corev1.EnvVar{Name: "NNODES", Value: strconv.Itoa(int(spec.GetReplicas()))},
		corev1.EnvVar{Name: "NPROC_PER_NODE", Value: strconv.Itoa(int(spec.GetNprocPerNode()))},
		corev1.EnvVar{Name: "MASTER_PORT", Value: "29500"},
		corev1.EnvVar{Name: "RDZV_BACKEND", Value: rdzvBackend},
		corev1.EnvVar{
			Name: "JOBSET_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.labels['jobset.sigs.k8s.io/jobset-name']",
				},
			},
		},
		corev1.EnvVar{
			Name: "JOBSET_RESTART_ATTEMPT",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.annotations['jobset.sigs.k8s.io/restart-attempt']",
				},
			},
		},
		corev1.EnvVar{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	)
}
