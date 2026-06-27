package clustered

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"

	clusteredpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/plugins"
)

// injectTorchRunEnv upserts the env vars that _entrypoint.py and torchrun need.
// Plugin-owned names overwrite any user-provided entries with the same name —
// duplicate Env entries have undefined precedence in Kubernetes and would
// confuse torchrun/_entrypoint.
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

	pluginEnv := []corev1.EnvVar{
		{Name: "NNODES", Value: strconv.Itoa(int(spec.GetReplicas()))},
		{Name: "NPROC_PER_NODE", Value: strconv.Itoa(int(spec.GetNprocPerNode()))},
		{Name: "MASTER_PORT", Value: "29500"},
		{Name: "RDZV_BACKEND", Value: rdzvBackend},
		// Restart budget so the SDK runtime can detect the terminal attempt and only write
		// error.pb once the JobSet has exhausted its restarts. Same source as failure.go's
		// buildFailurePolicy, so it always matches the JobSet FailurePolicy.MaxRestarts.
		{Name: "JOBSET_MAX_RESTARTS", Value: strconv.Itoa(int(spec.GetFailurePolicy().GetMaxRestarts()))},
		{
			Name: "JOBSET_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.labels['jobset.sigs.k8s.io/jobset-name']",
				},
			},
		},
		{
			Name: "JOBSET_RESTART_ATTEMPT",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.annotations['jobset.sigs.k8s.io/restart-attempt']",
				},
			},
		},
		{
			Name: "POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	container.Env = upsertEnv(container.Env, pluginEnv)
}

// upsertEnv replaces existing entries with the same name and appends new ones.
func upsertEnv(existing []corev1.EnvVar, updates []corev1.EnvVar) []corev1.EnvVar {
	idx := make(map[string]int, len(existing))
	for i, e := range existing {
		idx[e.Name] = i
	}
	for _, u := range updates {
		if i, ok := idx[u.Name]; ok {
			existing[i] = u
		} else {
			existing = append(existing, u)
			idx[u.Name] = len(existing) - 1
		}
	}
	return existing
}
