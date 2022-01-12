package clierrors

var (
	ErrInvalidStateUpdate = "Invalid state passed. Specify either activate or archive\n"

	ErrProjectNotPassed    = "Project not passed\n"
	ErrFailedProjectUpdate = "Project %v failed to get updated to %v state due to %v\n"

	ErrLPNotPassed        = "Launch plan name not passed\n"
	ErrLPVersionNotPassed = "Launch plan version not passed\n" //nolint
	ErrFailedLPUpdate     = "Launch plan %v failed to get updated due to %v\n"

	ErrExecutionNotPassed    = "Execution name not passed\n"
	ErrFailedExecutionUpdate = "Execution %v failed to get updated due to %v\n"

	ErrWorkflowNotPassed    = "Workflow name not passed\n"
	ErrFailedWorkflowUpdate = "Workflow %v failed to get updated to due to %v\n"

	ErrTaskNotPassed    = "Task name not passed\n" // #nosec
	ErrFailedTaskUpdate = "Task %v failed to get updated to due to %v\n"

	ErrSandboxExists = "Sandbox Exist\n"
)
