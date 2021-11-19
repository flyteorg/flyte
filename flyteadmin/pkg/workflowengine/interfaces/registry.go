package interfaces

// WorkflowExecutorRegistry is a singleton provider of a WorkflowExecutor implementation to use for
// creating and deleting Flyte workflow CRD objects.
type WorkflowExecutorRegistry interface {
	// Register registers a new WorkflowExecutor to handle creating and aborting Flyte workflow executions.
	Register(executor WorkflowExecutor)
	// RegisterDefault registers the default WorkflowExecutor to handle creating and aborting Flyte workflow executions.
	RegisterDefault(executor WorkflowExecutor)
	// GetExecutor resolves the definitive WorkflowExecutor implementation to be used for creating and aborting Flyte workflow executions.
	GetExecutor() WorkflowExecutor
}
