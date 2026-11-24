package flytek8s

const (
	// TaskNamespaceConfigKey is the optional Kubernetes workload namespace in TaskTemplate.config.
	TaskNamespaceConfigKey = "kubernetes.namespace"
	// TaskActionNamespaceAnnotation and TaskActionNameAnnotation locate the controlling
	// TaskAction when a workload cannot have a cross-namespace OwnerReference.
	TaskActionNamespaceAnnotation = "flyte.org/task-action-namespace"
	TaskActionNameAnnotation      = "flyte.org/task-action-name"
)
