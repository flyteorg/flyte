package subcommand

//go:generate pflags TaskResourceAttrFetchConfig --default-var DefaultTaskResourceFetchConfig

type TaskResourceAttrFetchConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for generating attribute for the resource type."`
}

var DefaultTaskResourceFetchConfig = &TaskResourceAttrFetchConfig{}
