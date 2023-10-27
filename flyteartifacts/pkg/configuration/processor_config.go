package configuration

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const artifactsProcessor = "artifactsProcessor"

// QueueSubscriberConfig is a generic config object for reading from things like SQS
type QueueSubscriberConfig struct {
	// The name of the queue onto which workflow notifications will enqueue.
	QueueName string `json:"queueName"`
	// The account id (according to whichever cloud provider scheme is used) that has permission to read from the above
	// queue.
	AccountID string `json:"accountId"`
}

type EventProcessorConfiguration struct {
	CloudProvider config.CloudDeployment `json:"cloudProvider" pflag:",Your deployment environment."`

	Region string `json:"region" pflag:",The region if applicable for the cloud provider."`

	Subscriber QueueSubscriberConfig `json:"subscriber" pflag:",The configuration for the subscriber."`
}

var defaultEventProcessorConfiguration = EventProcessorConfiguration{}

var EventsProcessorConfig = config.MustRegisterSection(artifactsProcessor, &defaultEventProcessorConfiguration)

func GetEventsProcessorConfig() *EventProcessorConfiguration {
	return EventsProcessorConfig.GetConfig().(*EventProcessorConfiguration)
}
