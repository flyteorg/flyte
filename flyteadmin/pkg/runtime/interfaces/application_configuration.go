package interfaces

type DbConfigSection struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	DbName string `json:"dbname"`
	User   string `json:"username"`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password"`
	PasswordPath string `json:"passwordPath"`
	// See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.
	ExtraOptions string `json:"options"`
}

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
	RoleNameKey           string   `json:"roleNameKey"`
	KubeConfig            string   `json:"kubeconfig"`
	MetricsScope          string   `json:"metricsScope"`
	ProfilerPort          int      `json:"profilerPort"`
	MetadataStoragePrefix []string `json:"metadataStoragePrefix"`
}

type EventSchedulerConfig struct {
	Scheme       string `json:"scheme"`
	Region       string `json:"region"`
	ScheduleRole string `json:"scheduleRole"`
	TargetName   string `json:"targetName"`
}

type WorkflowExecutorConfig struct {
	Scheme            string `json:"scheme"`
	Region            string `json:"region"`
	ScheduleQueueName string `json:"scheduleQueueName"`
	AccountID         string `json:"accountId"`
}

// This configuration is the base configuration for all scheduler-related set-up.
type SchedulerConfig struct {
	EventSchedulerConfig   EventSchedulerConfig   `json:"eventScheduler"`
	WorkflowExecutorConfig WorkflowExecutorConfig `json:"workflowExecutor"`
}

// Configuration specific to setting up signed urls.
type SignedURL struct {
	DurationMinutes int `json:"durationMinutes"`
}

// This configuration handles all requests to get remote data such as execution inputs & outputs.
type RemoteDataConfig struct {
	Scheme    string    `json:"scheme"`
	Region    string    `json:"region"`
	SignedURL SignedURL `json:"signedUrls"`
}

type NotificationsPublisherConfig struct {
	TopicName string `json:"topicName"`
}

type NotificationsProcessorConfig struct {
	QueueName string `json:"queueName"`
	AccountID string `json:"accountId"`
}

type NotificationsEmailerConfig struct {
	Subject string `json:"subject"`
	Sender  string `json:"sender"`
	Body    string `json:"body"`
}

// Configuration specific to notifications handling
type NotificationsConfig struct {
	Type                         string                       `json:"type"`
	Region                       string                       `json:"region"`
	NotificationsPublisherConfig NotificationsPublisherConfig `json:"publisher"`
	NotificationsProcessorConfig NotificationsProcessorConfig `json:"processor"`
	NotificationsEmailerConfig   NotificationsEmailerConfig   `json:"emailer"`
}

type Domain struct {
	ID   string `json:"id"`
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
