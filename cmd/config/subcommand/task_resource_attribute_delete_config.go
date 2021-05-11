package subcommand

//go:generate pflags TaskResourceAttrDeleteConfig --default-var DefaultTaskResourceDelConfig

// TaskResourceAttrDeleteConfig Matchable resource attributes configuration passed from command line
type TaskResourceAttrDeleteConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for delete attribute for the resource type."`
}

var DefaultTaskResourceDelConfig = &TaskResourceAttrDeleteConfig{}
