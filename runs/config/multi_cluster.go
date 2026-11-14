package config

import "github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"

// MultiClusterConfig places each run on one cluster, picked by the run's queue, as Flyte v1's
// labelClusterMap does with execution cluster labels. It is the runs.multiCluster config section,
// and it is off while QueueClusterMap is empty.
type MultiClusterConfig struct {
	// HomeCluster names the cluster reached through runs.actionsService. It gets runs with no queue.
	HomeCluster string `json:"homeCluster"`
	// Clusters holds the actions service of every other cluster, by cluster name.
	Clusters map[string]serviceclient.ServiceConfig `json:"clusters"`
	// QueueClusterMap lists the clusters each queue runs on.
	QueueClusterMap map[string][]ClusterEntity `json:"queueClusterMap"`
}

// ClusterEntity is one cluster a queue runs on. Weights are between 0 and 1, as in v1:
// zero-weight clusters are skipped, unless all of a queue's weights are 0, which picks
// its clusters equally.
type ClusterEntity struct {
	ID     string  `json:"id"`
	Weight float32 `json:"weight"`
}

var multiClusterSection = configSection.MustRegisterSection("multiCluster", &MultiClusterConfig{HomeCluster: "home"})

// GetMultiClusterConfig returns the parsed runs.multiCluster configuration.
func GetMultiClusterConfig() *MultiClusterConfig {
	return multiClusterSection.GetConfig().(*MultiClusterConfig)
}
