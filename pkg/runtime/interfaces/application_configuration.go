package interfaces

import (
	"github.com/flyteorg/flytestdlib/config"

	"golang.org/x/time/rate"
)

// DbConfig is used to for initiating the database connection with the store that holds registered
// entities (e.g. workflows, tasks, launch plans...)
type DbConfig struct {
	DeprecatedHost         string `json:"host" pflag:",deprecated"`
	DeprecatedPort         int    `json:"port" pflag:",deprecated"`
	DeprecatedDbName       string `json:"dbname" pflag:",deprecated"`
	DeprecatedUser         string `json:"username" pflag:",deprecated"`
	DeprecatedPassword     string `json:"password" pflag:",deprecated"`
	DeprecatedPasswordPath string `json:"passwordPath" pflag:",deprecated"`
	DeprecatedExtraOptions string `json:"options" pflag:",deprecated"`
	DeprecatedDebug        bool   `json:"debug" pflag:",deprecated"`

	EnableForeignKeyConstraintWhenMigrating bool            `json:"enableForeignKeyConstraintWhenMigrating" pflag:",Whether to enable gorm foreign keys when migrating the db"`
	MaxIdleConnections                      int             `json:"maxIdleConnections" pflag:",maxIdleConnections sets the maximum number of connections in the idle connection pool."`
	MaxOpenConnections                      int             `json:"maxOpenConnections" pflag:",maxOpenConnections sets the maximum number of open connections to the database."`
	ConnMaxLifeTime                         config.Duration `json:"connMaxLifeTime" pflag:",sets the maximum amount of time a connection may be reused"`
	PostgresConfig                          PostgresConfig  `json:"postgres"`
}

// PostgresConfig includes specific config options for opening a connection to a postgres database.
type PostgresConfig struct {
	Host   string `json:"host" pflag:",The host name of the database server"`
	Port   int    `json:"port" pflag:",The port name of the database server"`
	DbName string `json:"dbname" pflag:",The database name"`
	User   string `json:"username" pflag:",The database user who is connecting to the server."`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password" pflag:",The database password."`
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the database password."`
	ExtraOptions string `json:"options" pflag:",See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above."`
	Debug        bool   `json:"debug" pflag:" Whether or not to start the database connection with debug mode enabled."`
}

// This configuration is the base configuration to start admin
type ApplicationConfig struct {
	// The RoleName key inserted as an annotation (https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)
	// in Flyte Workflow CRDs created in the CreateExecution flow. The corresponding role value is defined in the
	// launch plan that is used to create the execution.
	RoleNameKey string `json:"roleNameKey"`
	// Top-level name applied to all metrics emitted by the application.
	MetricsScope string `json:"metricsScope"`
	// Determines which port the profiling server used for admin monitoring and application debugging uses.
	ProfilerPort int `json:"profilerPort"`
	// This defines the nested path on the configured external storage provider where workflow closures are remotely
	// offloaded.
	MetadataStoragePrefix []string `json:"metadataStoragePrefix"`
	// Event version to be used for Flyte workflows
	EventVersion int `json:"eventVersion"`
	// Specifies the shared buffer size which is used to queue asynchronous event writes.
	AsyncEventsBufferSize int `json:"asyncEventsBufferSize"`
	// Controls the maximum number of task nodes that can be run in parallel for the entire workflow.
	// This is useful to achieve fairness. Note: MapTasks are regarded as one unit,
	// and parallelism/concurrency of MapTasks is independent from this.
	MaxParallelism int32 `json:"maxParallelism"`
}

func (a *ApplicationConfig) GetRoleNameKey() string {
	return a.RoleNameKey
}

func (a *ApplicationConfig) GetMetricsScope() string {
	return a.MetricsScope
}

func (a *ApplicationConfig) GetProfilerPort() int {
	return a.ProfilerPort
}

func (a *ApplicationConfig) GetMetadataStoragePrefix() []string {
	return a.MetadataStoragePrefix
}

func (a *ApplicationConfig) GetEventVersion() int {
	return a.EventVersion
}

func (a *ApplicationConfig) GetAsyncEventsBufferSize() int {
	return a.AsyncEventsBufferSize
}

func (a *ApplicationConfig) GetMaxParallelism() int32 {
	return a.MaxParallelism
}

// This section holds common config for AWS
type AWSConfig struct {
	Region string `json:"region"`
}

// This section holds common config for GCP
type GCPConfig struct {
	ProjectID string `json:"projectId"`
}

// This section holds configuration for the event scheduler used to schedule workflow executions.
type EventSchedulerConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Scheme string `json:"scheme"`

	// Deprecated : Some cloud providers require a region to be set.
	Region string `json:"region"`
	// Deprecated : The role assumed to register and activate schedules.
	ScheduleRole string `json:"scheduleRole"`
	// Deprecated : The name of the queue for which scheduled events should enqueue.
	TargetName string `json:"targetName"`
	// Deprecated : Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix   string                `json:"scheduleNamePrefix"`
	AWSSchedulerConfig   *AWSSchedulerConfig   `json:"aws"`
	FlyteSchedulerConfig *FlyteSchedulerConfig `json:"local"`
}

func (e *EventSchedulerConfig) GetScheme() string {
	return e.Scheme
}

func (e *EventSchedulerConfig) GetRegion() string {
	return e.Region
}

func (e *EventSchedulerConfig) GetScheduleRole() string {
	return e.ScheduleRole
}

func (e *EventSchedulerConfig) GetTargetName() string {
	return e.TargetName
}

func (e *EventSchedulerConfig) GetScheduleNamePrefix() string {
	return e.ScheduleNamePrefix
}

func (e *EventSchedulerConfig) GetAWSSchedulerConfig() *AWSSchedulerConfig {
	return e.AWSSchedulerConfig
}

func (e *EventSchedulerConfig) GetFlyteSchedulerConfig() *FlyteSchedulerConfig {
	return e.FlyteSchedulerConfig
}

type AWSSchedulerConfig struct {
	// Some cloud providers require a region to be set.
	Region string `json:"region"`
	// The role assumed to register and activate schedules.
	ScheduleRole string `json:"scheduleRole"`
	// The name of the queue for which scheduled events should enqueue.
	TargetName string `json:"targetName"`
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string `json:"scheduleNamePrefix"`
}

func (a *AWSSchedulerConfig) GetRegion() string {
	return a.Region
}

func (a *AWSSchedulerConfig) GetScheduleRole() string {
	return a.ScheduleRole
}

func (a *AWSSchedulerConfig) GetTargetName() string {
	return a.TargetName
}

func (a *AWSSchedulerConfig) GetScheduleNamePrefix() string {
	return a.ScheduleNamePrefix
}

// FlyteSchedulerConfig is the config for native or default flyte scheduler
type FlyteSchedulerConfig struct {
}

// This section holds configuration for the executor that processes workflow scheduled events fired.
type WorkflowExecutorConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Scheme string `json:"scheme"`
	// Deprecated : Some cloud providers require a region to be set.
	Region string `json:"region"`
	// Deprecated : The name of the queue onto which scheduled events will enqueue.
	ScheduleQueueName string `json:"scheduleQueueName"`
	// Deprecated : The account id (according to whichever cloud provider scheme is used) that has permission to read from the above
	// queue.
	AccountID                   string                       `json:"accountId"`
	AWSWorkflowExecutorConfig   *AWSWorkflowExecutorConfig   `json:"aws"`
	FlyteWorkflowExecutorConfig *FlyteWorkflowExecutorConfig `json:"local"`
}

func (w *WorkflowExecutorConfig) GetScheme() string {
	return w.Scheme
}

func (w *WorkflowExecutorConfig) GetRegion() string {
	return w.Region
}

func (w *WorkflowExecutorConfig) GetScheduleScheduleQueueName() string {
	return w.ScheduleQueueName
}

func (w *WorkflowExecutorConfig) GetAccountID() string {
	return w.AccountID
}

func (w *WorkflowExecutorConfig) GetAWSWorkflowExecutorConfig() *AWSWorkflowExecutorConfig {
	return w.AWSWorkflowExecutorConfig
}

func (w *WorkflowExecutorConfig) GetFlyteWorkflowExecutorConfig() *FlyteWorkflowExecutorConfig {
	return w.FlyteWorkflowExecutorConfig
}

type AWSWorkflowExecutorConfig struct {
	// Some cloud providers require a region to be set.
	Region string `json:"region"`
	// The name of the queue onto which scheduled events will enqueue.
	ScheduleQueueName string `json:"scheduleQueueName"`
	// The account id (according to whichever cloud provider scheme is used) that has permission to read from the above
	// queue.
	AccountID string `json:"accountId"`
}

func (a *AWSWorkflowExecutorConfig) GetRegion() string {
	return a.Region
}

func (a *AWSWorkflowExecutorConfig) GetScheduleScheduleQueueName() string {
	return a.ScheduleQueueName
}

func (a *AWSWorkflowExecutorConfig) GetAccountID() string {
	return a.AccountID
}

// FlyteWorkflowExecutorConfig specifies the workflow executor configuration for the native flyte scheduler
type FlyteWorkflowExecutorConfig struct {
	// This allows to control the number of TPS that hit admin using the scheduler.
	// eg : 100 TPS will send at the max 100 schedule requests to admin per sec.
	// Burst specifies burst traffic count
	AdminRateLimit *AdminRateLimit `json:"adminRateLimit"`
}

func (f *FlyteWorkflowExecutorConfig) GetAdminRateLimit() *AdminRateLimit {
	return f.AdminRateLimit
}

type AdminRateLimit struct {
	Tps   rate.Limit `json:"tps"`
	Burst int        `json:"burst"`
}

func (f *AdminRateLimit) GetTps() rate.Limit {
	return f.Tps
}

func (f *AdminRateLimit) GetBurst() int {
	return f.Burst
}

// This configuration is the base configuration for all scheduler-related set-up.
type SchedulerConfig struct {
	// Determines which port the profiling server used for scheduler monitoring and application debugging uses.
	ProfilerPort           config.Port            `json:"profilerPort"`
	EventSchedulerConfig   EventSchedulerConfig   `json:"eventScheduler"`
	WorkflowExecutorConfig WorkflowExecutorConfig `json:"workflowExecutor"`
	// Specifies the number of times to attempt recreating a workflow executor client should there be any disruptions.
	ReconnectAttempts int `json:"reconnectAttempts"`
	// Specifies the time interval to wait before attempting to reconnect the workflow executor client.
	ReconnectDelaySeconds int `json:"reconnectDelaySeconds"`
}

func (s *SchedulerConfig) GetEventSchedulerConfig() EventSchedulerConfig {
	return s.EventSchedulerConfig
}

func (s *SchedulerConfig) GetWorkflowExecutorConfig() WorkflowExecutorConfig {
	return s.WorkflowExecutorConfig
}

func (s *SchedulerConfig) GetReconnectAttempts() int {
	return s.ReconnectAttempts
}

func (s *SchedulerConfig) GetReconnectDelaySeconds() int {
	return s.ReconnectDelaySeconds
}

// Configuration specific to setting up signed urls.
type SignedURL struct {
	// Whether signed urls should even be returned with GetExecutionData, GetNodeExecutionData and GetTaskExecutionData
	// response objects.
	Enabled bool `json:"enabled" pflag:",Whether signed urls should even be returned with GetExecutionData, GetNodeExecutionData and GetTaskExecutionData response objects."`
	// The amount of time for which a signed URL is valid.
	DurationMinutes int `json:"durationMinutes"`
	// The principal that signs the URL. This is only applicable to GCS URL.
	SigningPrincipal string `json:"signingPrincipal"`
}

//go:generate enumer -type=InlineEventDataPolicy -trimprefix=InlineEventDataPolicy
type InlineEventDataPolicy int

const (
	// InlineEventDataPolicyOffload specifies that inline execution event data (e.g. outputs) should be offloaded to the
	// configured cloud blob store.
	InlineEventDataPolicyOffload InlineEventDataPolicy = iota
	// InlineEventDataPolicyStoreInline specifies that inline execution event data should be saved inline with execution
	// database entries.
	InlineEventDataPolicyStoreInline
)

// This configuration handles all requests to get and write remote data such as execution inputs & outputs.
type RemoteDataConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Scheme string `json:"scheme"`
	// Some cloud providers require a region to be set.
	Region    string    `json:"region"`
	SignedURL SignedURL `json:"signedUrls"`
	// Specifies the max size in bytes for which execution data such as inputs and outputs will be populated in line.
	MaxSizeInBytes int64 `json:"maxSizeInBytes"`
	// Specifies how inline execution event data should be saved in the backend
	InlineEventDataPolicy InlineEventDataPolicy `json:"inlineEventDataPolicy" pflag:",Specifies how inline execution event data should be saved in the backend"`
}

// This section handles configuration for the workflow notifications pipeline.
type NotificationsPublisherConfig struct {
	// The topic which notifications use, e.g. AWS SNS topics.
	TopicName string `json:"topicName"`
}

// This section handles configuration for processing workflow events.
type NotificationsProcessorConfig struct {
	// The name of the queue onto which workflow notifications will enqueue.
	QueueName string `json:"queueName"`
	// The account id (according to whichever cloud provider scheme is used) that has permission to read from the above
	// queue.
	AccountID string `json:"accountId"`
}

type EmailServerConfig struct {
	ServiceName string `json:"serviceName"`
	// Only one of these should be set.
	APIKeyEnvVar   string `json:"apiKeyEnvVar"`
	APIKeyFilePath string `json:"apiKeyFilePath"`
}

// This section handles the configuration of notifications emails.
type NotificationsEmailerConfig struct {
	// For use with external email services (mailchimp/sendgrid)
	EmailerConfig EmailServerConfig `json:"emailServerConfig"`
	// The optionally templatized subject used in notification emails.
	Subject string `json:"subject"`
	// The optionally templatized sender used in notification emails.
	Sender string `json:"sender"`
	// The optionally templatized body the sender used in notification emails.
	Body string `json:"body"`
}

// This section handles configuration for the workflow notifications pipeline.
type EventsPublisherConfig struct {
	// The topic which events should be published, e.g. node, task, workflow
	TopicName string `json:"topicName"`
	// Event types: task, node, workflow executions
	EventTypes []string `json:"eventTypes"`
}

type ExternalEventsConfig struct {
	Enable bool `json:"enable"`
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Type      string    `json:"type"`
	AWSConfig AWSConfig `json:"aws"`
	GCPConfig GCPConfig `json:"gcp"`
	// Publish events to a pubsub tops
	EventsPublisherConfig EventsPublisherConfig `json:"eventsPublisher"`
	// Number of times to attempt recreating a notifications processor client should there be any disruptions.
	ReconnectAttempts int `json:"reconnectAttempts"`
	// Specifies the time interval to wait before attempting to reconnect the notifications processor client.
	ReconnectDelaySeconds int `json:"reconnectDelaySeconds"`
}

// Configuration specific to notifications handling
type NotificationsConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Type string `json:"type"`
	//  Deprecated: Please use AWSConfig instead.
	Region                       string                       `json:"region"`
	AWSConfig                    AWSConfig                    `json:"aws"`
	GCPConfig                    GCPConfig                    `json:"gcp"`
	NotificationsPublisherConfig NotificationsPublisherConfig `json:"publisher"`
	NotificationsProcessorConfig NotificationsProcessorConfig `json:"processor"`
	NotificationsEmailerConfig   NotificationsEmailerConfig   `json:"emailer"`
	// Number of times to attempt recreating a notifications processor client should there be any disruptions.
	ReconnectAttempts int `json:"reconnectAttempts"`
	// Specifies the time interval to wait before attempting to reconnect the notifications processor client.
	ReconnectDelaySeconds int `json:"reconnectDelaySeconds"`
}

// Domains are always globally set in the application config, whereas individual projects can be individually registered.
type Domain struct {
	// Unique identifier for a domain.
	ID string `json:"id"`
	// Human readable name for a domain.
	Name string `json:"name"`
}

type DomainsConfig = []Domain

// Defines the interface to return top-level config structs necessary to start up a flyteadmin application.
type ApplicationConfiguration interface {
	GetDbConfig() *DbConfig
	GetTopLevelConfig() *ApplicationConfig
	GetSchedulerConfig() *SchedulerConfig
	GetRemoteDataConfig() *RemoteDataConfig
	GetNotificationsConfig() *NotificationsConfig
	GetDomainsConfig() *DomainsConfig
	GetExternalEventsConfig() *ExternalEventsConfig
}
