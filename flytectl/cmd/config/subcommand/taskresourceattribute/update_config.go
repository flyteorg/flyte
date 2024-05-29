package taskresourceattribute

//go:generate pflags AttrUpdateConfig --default-var DefaultUpdateConfig --bind-default-var

// AttrUpdateConfig Matchable resource attributes configuration passed from command line
type AttrUpdateConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for updating attribute for the resource type."`
	DryRun   bool   `json:"dryRun" pflag:",execute command without making any modifications."`
	Force    bool   `json:"force" pflag:",do not ask for an acknowledgement during updates."`
}

var DefaultUpdateConfig = &AttrUpdateConfig{}
