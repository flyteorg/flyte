package executionqueueattribute

//go:generate pflags AttrDeleteConfig --default-var DefaultDelConfig --bind-default-var

// AttrDeleteConfig Matchable resource attributes configuration passed from command line
type AttrDeleteConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for delete attribute for the resource type."`
	DryRun   bool   `json:"dryRun" pflag:",execute command without making any modifications."`
}

var DefaultDelConfig = &AttrDeleteConfig{}
