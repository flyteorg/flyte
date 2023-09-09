package google

type TokenSourceFactoryType = string

const (
	TokenSourceTypeDefault                 = "default"
	TokenSourceTypeGkeTaskWorkloadIdentity = "gke-task-workload-identity" // #nosec
)

type TokenSourceFactoryConfig struct {
	// Type is type of TokenSourceFactory, possible values are 'default' or 'gke-task-workload-identity'.
	// - 'default' uses default credentials, see https://cloud.google.com/iam/docs/service-accounts#default
	Type TokenSourceFactoryType `json:"type" pflag:",Defines type of TokenSourceFactory, possible values are 'default' and 'gke-task-workload-identity'"`

	// Configuration for GKE task workload identity token source factory
	GkeTaskWorkloadIdentityTokenSourceFactoryConfig GkeTaskWorkloadIdentityTokenSourceFactoryConfig `json:"gke-task-workload-identity" pflag:"Extra configuration for GKE task workload identity token source factory"`
}

func GetDefaultConfig() TokenSourceFactoryConfig {
	return TokenSourceFactoryConfig{
		Type: "default",
	}
}
