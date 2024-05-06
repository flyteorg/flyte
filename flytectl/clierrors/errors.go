package clierrors

var (
	ErrInvalidStateUpdate     = "invalid state passed. Specify either activate or archive\n"
	ErrInvalidBothStateUpdate = "invalid state passed. Specify either activate or deactivate\n"

	ErrProjectNotPassed     = "project id wasn't passed\n" // #nosec
	ErrProjectIDBothPassed  = "both project and id are passed\n"
	ErrProjectNameNotPassed = "project name is a required flag"
	ErrFailedProjectUpdate  = "Project %v failed to update due to %w\n"

	ErrLPNotPassed        = "launch plan name wasn't passed\n"
	ErrLPVersionNotPassed = "launch plan version wasn't passed\n" //nolint
	ErrFailedLPUpdate     = "launch plan %v failed to update due to %w\n"

	ErrExecutionNotPassed    = "execution name wasn't passed\n"
	ErrFailedExecutionUpdate = "execution %v failed to update due to %v\n"

	ErrWorkflowNotPassed    = "workflow name wasn't passed\n"
	ErrFailedWorkflowUpdate = "workflow %v failed to update to due to %v\n"

	ErrTaskNotPassed    = "task name wasn't passed\n" // #nosec
	ErrFailedTaskUpdate = "task %v failed to update to due to %v\n"

	ErrSandboxExists = "sandbox already exists!\n"
)
