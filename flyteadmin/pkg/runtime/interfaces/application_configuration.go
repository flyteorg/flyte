package interfaces

// This configuration section is used to for initiating the database connection with the store that holds registered
// entities (e.g. workflows, tasks, launch plans...)
// This struct specifically maps to the flyteadmin config yaml structure.
type DbConfigSection struct {
	// The host name of the database server
	Host string `json:"host"`
	// The port name of the database server
	Port int `json:"port"`
	// The database name
	DbName string `json:"dbname"`
	// The database user who is connecting to the server.
	User string `json:"username"`
	// Either Password or PasswordPath must be set.
	// The Password resolves to the database password.
	Password     string `json:"password"`
	PasswordPath string `json:"passwordPath"`
	// See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.
	ExtraOptions string `json:"options"`
}

// This represents a configuration used for initiating database connections much like DbConfigSection, however the
// password is *resolved* in this struct and therefore it is used as the value the runtime provider returns to callers
// requesting the database config.
type DbConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DbName       string `json:"dbname"`
	User         string `json:"username"`
	Password     string `json:"password"`
	ExtraOptions string `json:"options"`
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
	// Some cloud providers require a region to be set.
	Region string `json:"region"`
	// The role assumed to register and activate schedules.
	ScheduleRole string `json:"scheduleRole"`
	// The name of the queue for which scheduled events should enqueue.
	TargetName string `json:"targetName"`
	// Optional: The application-wide prefix to be applied for schedule names.
	ScheduleNamePrefix string `json:"scheduleNamePrefix"`
}

// This section holds configuration for the executor that processes workflow scheduled events fired.
type WorkflowExecutorConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Scheme string `json:"scheme"`
	// Some cloud providers require a region to be set.
	Region string `json:"region"`
	// The name of the queue onto which scheduled events will enqueue.
	ScheduleQueueName string `json:"scheduleQueueName"`
	// The account id (according to whichever cloud provider scheme is used) that has permission to read from the above
	// queue.
	AccountID string `json:"accountId"`
}

// This configuration is the base configuration for all scheduler-related set-up.
type SchedulerConfig struct {
	EventSchedulerConfig   EventSchedulerConfig   `json:"eventScheduler"`
	WorkflowExecutorConfig WorkflowExecutorConfig `json:"workflowExecutor"`
	// Specifies the number of times to attempt recreating a workflow executor client should there be any disruptions.
	ReconnectAttempts int `json:"reconnectAttempts"`
	// Specifies the time interval to wait before attempting to reconnect the workflow executor client.
	ReconnectDelaySeconds int `json:"reconnectDelaySeconds"`
}

// Configuration specific to setting up signed urls.
type SignedURL struct {
	// The amount of time for which a signed URL is valid.
	DurationMinutes int `json:"durationMinutes"`
}

// This configuration handles all requests to get remote data such as execution inputs & outputs.
type RemoteDataConfig struct {
	// Defines the cloud provider that backs the scheduler. In the absence of a specification the no-op, 'local'
	// scheme is used.
	Scheme string `json:"scheme"`
	// Some cloud providers require a region to be set.
	Region    string    `json:"region"`
	SignedURL SignedURL `json:"signedUrls"`
	// Specifies the max size in bytes for which execution data such as inputs and outputs will be populated in line.
	MaxSizeInBytes int64 `json:"maxSizeInBytes"`
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

// This section handles the configuration of notifications emails.
type NotificationsEmailerConfig struct {
	// The optionally templatized subject used in notification emails.
	Subject string `json:"subject"`
	// The optionally templatized sender used in notification emails.
	Sender string `json:"sender"`
	// The optionally templatized body the sender used in notification emails.
	Body string `json:"body"`
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
	GetDbConfig() DbConfig
	GetTopLevelConfig() *ApplicationConfig
	GetSchedulerConfig() *SchedulerConfig
	GetRemoteDataConfig() *RemoteDataConfig
	GetNotificationsConfig() *NotificationsConfig
	GetDomainsConfig() *DomainsConfig
}
