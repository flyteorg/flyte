package interfaces

// Handles responding to scheduled workflow execution events and creating executions.
type WorkflowExecutor interface {
	Run()
	Stop() error
}
