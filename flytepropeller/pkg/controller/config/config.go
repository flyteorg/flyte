// Package config contains the core configuration for FlytePropeller. This configuration can be added under the “propeller“ section.
//
//	Example config:
//
// ----------------
//
//	propeller:
//	   rawoutput-prefix: s3://my-container/test/
//	   metadata-prefix: metadata/propeller/sandbox
//	   workers: 4
//	   workflow-reeval-duration: 10s
//	   downstream-eval-duration: 5s
//	   limit-namespace: "all"
//	   prof-port: 11254
//	   metrics-prefix: flyte
//	   enable-admin-launcher: true
//	   max-ttl-hours: 1
//	   gc-interval: 500m
//	   queue:
//	     type: batch
//	     queue:
//	       type: bucket
//	       rate: 1000
//	       capacity: 10000
//	     sub-queue:
//	       type: bucket
//	       rate: 1000
//	       capacity: 10000
//	   # This config assumes using `make start` in flytesnacks repo to startup a DinD k3s container
//	   kube-config: "$HOME/kubeconfig/k3s/k3s.yaml"
//	   publish-k8s-events: true
//	   workflowStore:
//	     policy: "ResourceVersionCache"
package config

import (
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/contextutils"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "propeller"

var (
	configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

	defaultConfig = &Config{
		Workers: 20,
		WorkflowReEval: config.Duration{
			Duration: 10 * time.Second,
		},
		DownstreamEval: config.Duration{
			Duration: 30 * time.Second,
		},
		MaxWorkflowRetries: 10,
		MaxTTLInHours:      23,
		GCInterval: config.Duration{
			Duration: 30 * time.Minute,
		},
		MaxDatasetSizeBytes: 10 * 1024 * 1024,
		Queue: CompositeQueueConfig{
			Type: CompositeQueueBatch,
			BatchingInterval: config.Duration{
				Duration: time.Second,
			},
			BatchSize: -1,
			Queue: WorkqueueConfig{
				Type:      WorkqueueTypeMaxOfRateLimiter,
				BaseDelay: config.Duration{Duration: time.Second * 0},
				MaxDelay:  config.Duration{Duration: time.Second * 60},
				Rate:      1000,
				Capacity:  10000,
			},
			Sub: WorkqueueConfig{
				Type:     WorkqueueTypeBucketRateLimiter,
				Rate:     1000,
				Capacity: 10000,
			},
		},
		KubeConfig: KubeClientConfig{
			QPS:     100,
			Burst:   25,
			Timeout: config.Duration{Duration: 30 * time.Second},
		},
		LeaderElection: LeaderElectionConfig{
			Enabled:       false,
			LeaseDuration: config.Duration{Duration: time.Second * 15},
			RenewDeadline: config.Duration{Duration: time.Second * 10},
			RetryPeriod:   config.Duration{Duration: time.Second * 2},
		},
		NodeConfig: NodeConfig{
			MaxNodeRetriesOnSystemFailures: 3,
			InterruptibleFailureThreshold:  1,
		},
		MaxStreakLength: 8, // Turbo mode is enabled by default
		ProfilerPort: config.Port{
			Port: 10254,
		},
		LimitNamespace:      "all",
		MetadataPrefix:      "metadata/propeller",
		EnableAdminLauncher: true,
		MetricsPrefix:       "flyte",
		MetricKeys: []string{contextutils.ProjectKey.String(), contextutils.DomainKey.String(),
			contextutils.WorkflowIDKey.String(), contextutils.TaskIDKey.String()},
		EventConfig: EventConfig{
			RawOutputPolicy: RawOutputPolicyReference,
		},
		ClusterID:              "propeller",
		CreateFlyteWorkflowCRD: false,
	}
)

// Config that uses the flytestdlib Config module to generate commandline and load config files. This configuration is
// the base configuration to start propeller
// NOTE: when adding new fields, do not mark them as "omitempty" if it's desirable to read the value from env variables.
type Config struct {
	KubeConfigPath           string               `json:"kube-config" pflag:",Path to kubernetes client config file."`
	MasterURL                string               `json:"master"`
	Workers                  int                  `json:"workers" pflag:",Number of threads to process workflows"`
	WorkflowReEval           config.Duration      `json:"workflow-reeval-duration" pflag:",Frequency of re-evaluating workflows"`
	DownstreamEval           config.Duration      `json:"downstream-eval-duration" pflag:",Frequency of re-evaluating downstream tasks"`
	LimitNamespace           string               `json:"limit-namespace" pflag:",Namespaces to watch for this propeller"`
	ProfilerPort             config.Port          `json:"prof-port" pflag:",Profiler port"`
	MetadataPrefix           string               `json:"metadata-prefix,omitempty" pflag:",MetadataPrefix should be used if all the metadata for Flyte executions should be stored under a specific prefix in CloudStorage. If not specified, the data will be stored in the base container directly."`
	DefaultRawOutputPrefix   string               `json:"rawoutput-prefix" pflag:",a fully qualified storage path of the form s3://flyte/abc/..., where all data sandboxes should be stored."`
	Queue                    CompositeQueueConfig `json:"queue,omitempty" pflag:",Workflow workqueue configuration, affects the way the work is consumed from the queue."`
	MetricsPrefix            string               `json:"metrics-prefix" pflag:",An optional prefix for all published metrics."`
	MetricKeys               []string             `json:"metrics-keys" pflag:",Metrics labels applied to prometheus metrics emitted by the service."`
	EnableAdminLauncher      bool                 `json:"enable-admin-launcher" pflag:"Enable remote Workflow launcher to Admin"`
	MaxWorkflowRetries       int                  `json:"max-workflow-retries" pflag:"Maximum number of retries per workflow"`
	MaxTTLInHours            int                  `json:"max-ttl-hours" pflag:"Maximum number of hours a completed workflow should be retained. Number between 1-23 hours"`
	GCInterval               config.Duration      `json:"gc-interval" pflag:"Run periodic GC every 30 minutes"`
	LeaderElection           LeaderElectionConfig `json:"leader-election,omitempty" pflag:",Config for leader election."`
	PublishK8sEvents         bool                 `json:"publish-k8s-events" pflag:",Enable events publishing to K8s events API."`
	MaxDatasetSizeBytes      int64                `json:"max-output-size-bytes" pflag:",Maximum size of outputs per task"`
	EnableGrpcLatencyMetrics bool                 `json:"enable-grpc-latency-metrics" pflag:",Enable grpc latency metrics. Note Histograms metrics can be expensive on Prometheus servers."`
	KubeConfig               KubeClientConfig     `json:"kube-client-config" pflag:",Configuration to control the Kubernetes client"`
	NodeConfig               NodeConfig           `json:"node-config,omitempty" pflag:",config for a workflow node"`
	MaxStreakLength          int                  `json:"max-streak-length" pflag:",Maximum number of consecutive rounds that one propeller worker can use for one workflow - >1 => turbo-mode is enabled."`
	EventConfig              EventConfig          `json:"event-config,omitempty" pflag:",Configures execution event behavior."`
	IncludeShardKeyLabel     []string             `json:"include-shard-key-label" pflag:",Include the specified shard key label in the k8s FlyteWorkflow CRD label selector"`
	ExcludeShardKeyLabel     []string             `json:"exclude-shard-key-label" pflag:",Exclude the specified shard key label from the k8s FlyteWorkflow CRD label selector"`
	IncludeProjectLabel      []string             `json:"include-project-label" pflag:",Include the specified project label in the k8s FlyteWorkflow CRD label selector"`
	ExcludeProjectLabel      []string             `json:"exclude-project-label" pflag:",Exclude the specified project label from the k8s FlyteWorkflow CRD label selector"`
	IncludeDomainLabel       []string             `json:"include-domain-label" pflag:",Include the specified domain label in the k8s FlyteWorkflow CRD label selector"`
	ExcludeDomainLabel       []string             `json:"exclude-domain-label" pflag:",Exclude the specified domain label from the k8s FlyteWorkflow CRD label selector"`
	ClusterID                string               `json:"cluster-id" pflag:",Unique cluster id running this flytepropeller instance with which to annotate execution events"`
	CreateFlyteWorkflowCRD   bool                 `json:"create-flyteworkflow-crd" pflag:",Enable creation of the FlyteWorkflow CRD on startup"`
}

// KubeClientConfig contains the configuration used by flytepropeller to configure its internal Kubernetes Client.
type KubeClientConfig struct {
	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS float32 `json:"qps" pflag:"-,Max QPS to the master for requests to KubeAPI. 0 defaults to 5."`
	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int `json:"burst" pflag:",Max burst rate for throttle. 0 defaults to 10"`
	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout config.Duration `json:"timeout" pflag:",Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout."`
}

type CompositeQueueType = string

const (
	CompositeQueueSimple CompositeQueueType = "simple"
	CompositeQueueBatch  CompositeQueueType = "batch"
)

// CompositeQueueConfig contains configuration for the controller queue and the downstream resource queue
type CompositeQueueConfig struct {
	Type             CompositeQueueType `json:"type" pflag:",Type of composite queue to use for the WorkQueue"`
	Queue            WorkqueueConfig    `json:"queue,omitempty" pflag:",Workflow workqueue configuration, affects the way the work is consumed from the queue."`
	Sub              WorkqueueConfig    `json:"sub-queue,omitempty" pflag:",SubQueue configuration, affects the way the nodes cause the top-level Work to be re-evaluated."`
	BatchingInterval config.Duration    `json:"batching-interval" pflag:",Duration for which downstream updates are buffered"`
	BatchSize        int                `json:"batch-size" pflag:"-1,Number of downstream triggered top-level objects to re-enqueue every duration. -1 indicates all available."`
}

type WorkqueueType = string

const (
	WorkqueueTypeDefault                       WorkqueueType = "default"
	WorkqueueTypeBucketRateLimiter             WorkqueueType = "bucket"
	WorkqueueTypeExponentialFailureRateLimiter WorkqueueType = "expfailure"
	WorkqueueTypeMaxOfRateLimiter              WorkqueueType = "maxof"
)

// WorkqueueConfig has the configuration to configure a workqueue. We may want to generalize this in a package like k8sutils
type WorkqueueConfig struct {
	// Refer to https://github.com/kubernetes/client-go/tree/master/util/workqueue
	Type      WorkqueueType   `json:"type" pflag:",Type of RateLimiter to use for the WorkQueue"`
	BaseDelay config.Duration `json:"base-delay" pflag:",base backoff delay for failure"`
	MaxDelay  config.Duration `json:"max-delay" pflag:",Max backoff delay for failure"`
	Rate      int64           `json:"rate" pflag:",Bucket Refill rate per second"`
	Capacity  int             `json:"capacity" pflag:",Bucket capacity as number of items"`
}

// NodeConfig contains configuration that is useful for every node execution
type NodeConfig struct {
	DefaultDeadlines               DefaultDeadlines `json:"default-deadlines,omitempty" pflag:",Default value for timeouts"`
	MaxNodeRetriesOnSystemFailures int64            `json:"max-node-retries-system-failures" pflag:"2,Maximum number of retries per node for node failure due to infra issues"`
	InterruptibleFailureThreshold  int64            `json:"interruptible-failure-threshold" pflag:"1,number of failures for a node to be still considered interruptible'"`
}

// DefaultDeadlines contains default values for timeouts
type DefaultDeadlines struct {
	DefaultNodeExecutionDeadline  config.Duration `json:"node-execution-deadline" pflag:",Default value of node execution timeout that includes the time spent to run the node/workflow"`
	DefaultNodeActiveDeadline     config.Duration `json:"node-active-deadline" pflag:",Default value of node timeout that includes the time spent queued."`
	DefaultWorkflowActiveDeadline config.Duration `json:"workflow-active-deadline" pflag:",Default value of workflow timeout that includes the time spent queued."`
}

// LeaderElectionConfig Contains leader election configuration.
type LeaderElectionConfig struct {
	// Enable or disable leader election.
	Enabled bool `json:"enabled" pflag:",Enables/Disables leader election."`

	// Determines the name of the configmap that leader election will use for holding the leader lock.
	LockConfigMap types.NamespacedName `json:"lock-config-map" pflag:",ConfigMap namespace/name to use for resource lock."`

	// Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last
	// observed ack
	LeaseDuration config.Duration `json:"lease-duration" pflag:",Duration that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack."`

	// RenewDeadline is the duration that the acting master will retry refreshing leadership before giving up.
	RenewDeadline config.Duration `json:"renew-deadline" pflag:",Duration that the acting master will retry refreshing leadership before giving up."`

	// RetryPeriod is the duration the LeaderElector clients should wait between tries of actions.
	RetryPeriod config.Duration `json:"retry-period" pflag:",Duration the LeaderElector clients should wait between tries of actions."`
}

// Defines how output data should be passed along in execution events.
type RawOutputPolicy = string

const (
	// Only send output data as a URI referencing where outputs have been uploaded
	RawOutputPolicyReference RawOutputPolicy = "reference"
	// Send raw output data in events.
	RawOutputPolicyInline RawOutputPolicy = "inline"
)

type EventConfig struct {
	RawOutputPolicy           RawOutputPolicy `json:"raw-output-policy" pflag:",How output data should be passed along in execution events."`
	FallbackToOutputReference bool            `json:"fallback-to-output-reference" pflag:",Whether output data should be sent by reference when it is too large to be sent inline in execution events."`
}

// GetConfig extracts the Configuration from the global config module in flytestdlib and returns the corresponding type-casted object.
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

// MustRegisterSubSection can be used to configure any subsections the the propeller configuration
func MustRegisterSubSection(subSectionKey string, section config.Config) config.Section {
	return configSection.MustRegisterSection(subSectionKey, section)
}
