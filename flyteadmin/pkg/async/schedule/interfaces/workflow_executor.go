package interfaces

//go:generate mockery-v2 --name=WorkflowExecutor --output=../mocks --case=underscore --with-expecter

// Handles responding to scheduled workflow execution events and creating executions.
type WorkflowExecutor interface {
	Run()
	Stop() error
}
