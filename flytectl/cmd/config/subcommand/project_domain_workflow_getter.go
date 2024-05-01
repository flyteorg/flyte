package subcommand

import rootConfig "github.com/flyteorg/flytectl/cmd/config"

// ProjectDomainWorkflowGetter defines a interface for getting the project domain workflow.
type ProjectDomainWorkflowGetter interface {
	GetProject() string
	GetDomain() string
	GetWorkflow() string
}

// PDWGetterCommandLine implements the command line way of getting project domain and workflow
type PDWGetterCommandLine struct {
	Config *rootConfig.Config
	Args   []string
}

// GetProject returns the cobra parsed Project from the Config
func (g PDWGetterCommandLine) GetProject() string {
	if g.Config == nil {
		return ""
	}
	return g.Config.Project
}

// GetDomain returns the cobra parsed Domain from the Config
func (g PDWGetterCommandLine) GetDomain() string {
	if g.Config == nil {
		return ""
	}
	return g.Config.Domain
}

// GetWorkflow returns the first argument from the commandline
func (g PDWGetterCommandLine) GetWorkflow() string {
	if g.Args == nil || len(g.Args) == 0 {
		return ""
	}
	return g.Args[0]
}
