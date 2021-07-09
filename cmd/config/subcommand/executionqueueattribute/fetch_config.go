package executionqueueattribute

//go:generate pflags AttrFetchConfig --default-var DefaultFetchConfig --bind-default-var

type AttrFetchConfig struct {
	AttrFile string `json:"attrFile" pflag:",attribute file name to be used for generating attribute for the resource type."`
}

var DefaultFetchConfig = &AttrFetchConfig{}
