// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package config

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (Config) elemValueOrNil(v interface{}) interface{} {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return reflect.Zero(t.Elem()).Interface()
		} else {
			return reflect.ValueOf(v).Interface()
		}
	} else if v == nil {
		return reflect.Zero(t).Interface()
	}

	return v
}

func (Config) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (Config) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in Config and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg Config) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("Config", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "kube-config"), defaultConfig.KubeConfigPath, "Path to kubernetes client config file.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "master"), defaultConfig.MasterURL, "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "workers"), defaultConfig.Workers, "Number of threads to process workflows")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "workflow-reeval-duration"), defaultConfig.WorkflowReEval.String(), "Frequency of re-evaluating workflows")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "downstream-eval-duration"), defaultConfig.DownstreamEval.String(), "Frequency of re-evaluating downstream tasks")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "limit-namespace"), defaultConfig.LimitNamespace, "Namespaces to watch for this propeller")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "prof-port"), defaultConfig.ProfilerPort.String(), "Profiler port")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "metadata-prefix"), defaultConfig.MetadataPrefix, "MetadataPrefix should be used if all the metadata for Flyte executions should be stored under a specific prefix in CloudStorage. If not specified,  the data will be stored in the base container directly.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "rawoutput-prefix"), defaultConfig.DefaultRawOutputPrefix, "a fully qualified storage path of the form s3://flyte/abc/...,  where all data sandboxes should be stored.")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "rawoutput-suffix"), defaultConfig.DefaultRawOutputSuffix, "path parts for a storage suffix that gets added to the raw output dir in the form foo/bar/baz...,  where all data sandboxes should be stored.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.type"), defaultConfig.Queue.Type, "Type of composite queue to use for the WorkQueue")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.queue.type"), defaultConfig.Queue.Queue.Type, "Type of RateLimiter to use for the WorkQueue")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.queue.base-delay"), defaultConfig.Queue.Queue.BaseDelay.String(), "base backoff delay for failure")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.queue.max-delay"), defaultConfig.Queue.Queue.MaxDelay.String(), "Max backoff delay for failure")
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "queue.queue.rate"), defaultConfig.Queue.Queue.Rate, "Bucket Refill rate per second")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "queue.queue.capacity"), defaultConfig.Queue.Queue.Capacity, "Bucket capacity as number of items")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.sub-queue.type"), defaultConfig.Queue.Sub.Type, "Type of RateLimiter to use for the WorkQueue")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.sub-queue.base-delay"), defaultConfig.Queue.Sub.BaseDelay.String(), "base backoff delay for failure")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.sub-queue.max-delay"), defaultConfig.Queue.Sub.MaxDelay.String(), "Max backoff delay for failure")
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "queue.sub-queue.rate"), defaultConfig.Queue.Sub.Rate, "Bucket Refill rate per second")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "queue.sub-queue.capacity"), defaultConfig.Queue.Sub.Capacity, "Bucket capacity as number of items")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "queue.batching-interval"), defaultConfig.Queue.BatchingInterval.String(), "Duration for which downstream updates are buffered")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "queue.batch-size"), defaultConfig.Queue.BatchSize, "Number of downstream triggered top-level objects to re-enqueue every duration. -1 indicates all available.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "metrics-prefix"), defaultConfig.MetricsPrefix, "An optional prefix for all published metrics.")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "metrics-keys"), defaultConfig.MetricKeys, "Metrics labels applied to prometheus metrics emitted by the service.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "enable-admin-launcher"), defaultConfig.EnableAdminLauncher, "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "max-workflow-retries"), defaultConfig.MaxWorkflowRetries, "")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "max-ttl-hours"), defaultConfig.MaxTTLInHours, "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "gc-interval"), defaultConfig.GCInterval.String(), "")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "leader-election.enabled"), defaultConfig.LeaderElection.Enabled, "Enables/Disables leader election.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "leader-election.lock-config-map.Namespace"), defaultConfig.LeaderElection.LockConfigMap.Namespace, "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "leader-election.lock-config-map.Name"), defaultConfig.LeaderElection.LockConfigMap.Name, "")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "leader-election.lease-duration"), defaultConfig.LeaderElection.LeaseDuration.String(), "Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "leader-election.renew-deadline"), defaultConfig.LeaderElection.RenewDeadline.String(), "Duration that the acting master will retry refreshing leadership before giving up.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "leader-election.retry-period"), defaultConfig.LeaderElection.RetryPeriod.String(), "Duration the LeaderElector clients should wait between tries of actions.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "publish-k8s-events"), defaultConfig.PublishK8sEvents, "Enable events publishing to K8s events API.")
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "max-output-size-bytes"), defaultConfig.MaxDatasetSizeBytes, "Deprecated! Use storage.limits.maxDownloadMBs instead")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "enable-grpc-latency-metrics"), defaultConfig.EnableGrpcLatencyMetrics, "Enable grpc latency metrics. Note Histograms metrics can be expensive on Prometheus servers.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "kube-client-config.burst"), defaultConfig.KubeConfig.Burst, "Max burst rate for throttle. 0 defaults to 10")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "kube-client-config.timeout"), defaultConfig.KubeConfig.Timeout.String(), "Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "node-config.default-deadlines.node-execution-deadline"), defaultConfig.NodeConfig.DefaultDeadlines.DefaultNodeExecutionDeadline.String(), "Default value of node execution timeout that includes the time spent to run the node/workflow")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "node-config.default-deadlines.node-active-deadline"), defaultConfig.NodeConfig.DefaultDeadlines.DefaultNodeActiveDeadline.String(), "Default value of node timeout that includes the time spent queued.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "node-config.default-deadlines.workflow-active-deadline"), defaultConfig.NodeConfig.DefaultDeadlines.DefaultWorkflowActiveDeadline.String(), "Default value of workflow timeout that includes the time spent queued.")
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "node-config.max-node-retries-system-failures"), defaultConfig.NodeConfig.MaxNodeRetriesOnSystemFailures, "Maximum number of retries per node for node failure due to infra issues")
	cmdFlags.Int32(fmt.Sprintf("%v%v", prefix, "node-config.interruptible-failure-threshold"), defaultConfig.NodeConfig.InterruptibleFailureThreshold, "number of failures for a node to be still considered interruptible. Negative numbers are treated as complementary (ex. -1 means last attempt is non-interruptible).'")
	cmdFlags.Int32(fmt.Sprintf("%v%v", prefix, "node-config.default-max-attempts"), defaultConfig.NodeConfig.DefaultMaxAttempts, "Default maximum number of attempts for a node")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "node-config.ignore-retry-cause"), defaultConfig.NodeConfig.IgnoreRetryCause, "Ignore retry cause and count all attempts toward a node's max attempts")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "node-config.enable-cr-debug-metadata"), defaultConfig.NodeConfig.EnableCRDebugMetadata, "Collapse node on any terminal state,  not just successful terminations. This is useful to reduce the size of workflow state in etcd.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "max-streak-length"), defaultConfig.MaxStreakLength, "Maximum number of consecutive rounds that one propeller worker can use for one workflow - >1 => turbo-mode is enabled.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "event-config.raw-output-policy"), defaultConfig.EventConfig.RawOutputPolicy, "How output data should be passed along in execution events.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "event-config.fallback-to-output-reference"), defaultConfig.EventConfig.FallbackToOutputReference, "Whether output data should be sent by reference when it is too large to be sent inline in execution events.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "event-config.error-on-already-exists"), defaultConfig.EventConfig.ErrorOnAlreadyExists, "Whether to return an error when an event already exists.")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "include-shard-key-label"), defaultConfig.IncludeShardKeyLabel, "Include the specified shard key label in the k8s FlyteWorkflow CRD label selector")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "exclude-shard-key-label"), defaultConfig.ExcludeShardKeyLabel, "Exclude the specified shard key label from the k8s FlyteWorkflow CRD label selector")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "include-project-label"), defaultConfig.IncludeProjectLabel, "Include the specified project label in the k8s FlyteWorkflow CRD label selector")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "exclude-project-label"), defaultConfig.ExcludeProjectLabel, "Exclude the specified project label from the k8s FlyteWorkflow CRD label selector")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "include-domain-label"), defaultConfig.IncludeDomainLabel, "Include the specified domain label in the k8s FlyteWorkflow CRD label selector")
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "exclude-domain-label"), defaultConfig.ExcludeDomainLabel, "Exclude the specified domain label from the k8s FlyteWorkflow CRD label selector")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "cluster-id"), defaultConfig.ClusterID, "Unique cluster id running this flytepropeller instance with which to annotate execution events")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "create-flyteworkflow-crd"), defaultConfig.CreateFlyteWorkflowCRD, "Enable creation of the FlyteWorkflow CRD on startup")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "node-execution-worker-count"), defaultConfig.NodeExecutionWorkerCount, "Number of workers to evaluate node executions,  currently only used for array nodes")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "accelerated-inputs.enabled"), defaultConfig.AcceleratedInputs.Enabled, "Enabled accelerated inputs feature which overwrites remote artifacts path to local disk paths")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "accelerated-inputs.remote-path-prefix"), defaultConfig.AcceleratedInputs.RemotePathPrefix, "Remote path prefix that should be replaced with local path prefix")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "accelerated-inputs.local-path-prefix"), defaultConfig.AcceleratedInputs.LocalPathPrefix, "Path to locally mounted directory k8s pod")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "accelerated-inputs.volume-path"), defaultConfig.AcceleratedInputs.VolumePath, "Path to locally mounted directory on k8s host node")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "array-node.event-version"), defaultConfig.ArrayNode.EventVersion, "ArrayNode eventing version. 0 => legacy (drop-in replacement for maptask),  1 => new")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "array-node.default-parallelism-behavior"), defaultConfig.ArrayNode.DefaultParallelismBehavior, "Default parallelism behavior for array nodes")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "array-node.use-map-plugin-logs"), defaultConfig.ArrayNode.UseMapPluginLogs, "Override subNode log links with those configured for the map plugin logs")
	return cmdFlags
}
