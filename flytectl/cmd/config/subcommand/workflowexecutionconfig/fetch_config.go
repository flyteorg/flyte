package workflowexecutionconfig

//go:generate pflags AttrFetchConfig --default-var DefaultFetchConfig --bind-default-var

type AttrFetchConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for generating attribute for the resource type."`
	Gen      bool   `json:"gen" pflag:",generates an empty workflow execution config file with conformance to the api format."`
}

var DefaultFetchConfig = &AttrFetchConfig{}
