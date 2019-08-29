package configs

//go:generate pflags DataCatalogConfig

// This configuration is the base configuration to start admin
type DataCatalogConfig struct {
	StoragePrefix string `json:"storage-prefix" pflag:",StoragePrefix specifies the prefix where DataCatalog stores offloaded ArtifactData in CloudStorage. If not specified, the data will be stored in the base container directly."`
	MetricsScope  string `json:"metrics-scope" pflag:",Scope that the metrics will record under."`
	ProfilerPort  int    `json:"profiler-port" pflag:",Port that the profiling service is listening on."`
}
