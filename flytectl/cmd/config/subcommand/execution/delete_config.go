package execution

//go:generate pflags ExecDeleteConfig --default-var DefaultExecDeleteConfig --bind-default-var

var DefaultExecDeleteConfig = &ExecDeleteConfig{}

// ExecutionDeleteConfig stores the flags required by delete execution
type ExecDeleteConfig struct {
	DryRun bool `json:"dryRun" pflag:",execute command without making any modifications."`
}
