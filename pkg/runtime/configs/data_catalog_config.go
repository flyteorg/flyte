package configs

//go:generate pflags DataCatalogConfig

// This configuration is the base configuration to start admin
type DataCatalogConfig struct {
	StoragePrefix string `json:"storage-prefix" pflag:",StoragePrefix specifies the prefix where DataCatalog stores offloaded ArtifactData in CloudStorage. If not specified, the data will be stored in the base container directly."`
}
