package taskresourceattribute

//go:generate pflags AttrDeleteConfig --default-var DefaultDelConfig

// AttrDeleteConfig Matchable resource attributes configuration passed from command line
type AttrDeleteConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for delete attribute for the resource type."`
}

var DefaultDelConfig = &AttrDeleteConfig{}
