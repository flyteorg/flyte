package subcommand

//go:generate pflags TaskResourceAttrUpdateConfig --default-var DefaultTaskResourceUpdateConfig

// TaskResourceAttrUpdateConfig Matchable resource attributes configuration passed from command line
type TaskResourceAttrUpdateConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for updating attribute for the resource type."`
}

var DefaultTaskResourceUpdateConfig = &TaskResourceAttrUpdateConfig{}
