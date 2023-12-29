package processor

import (
	"context"

	"github.com/NYTimes/gizmo/pubsub"
	gizmoAWS "github.com/NYTimes/gizmo/pubsub/aws"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/configuration"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	configCommon "github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func NewBackgroundProcessor(ctx context.Context, processorConfiguration configuration.EventProcessorConfiguration, service artifact.ArtifactRegistryServer, created chan<- artifact.Artifact, scope promutils.Scope) EventsProcessorInterface {

	// TODO: Add retry logic
	var sub pubsub.Subscriber
	switch processorConfiguration.CloudProvider {
	case configCommon.CloudDeploymentAWS:
		ff := false
		sqsConfig := gizmoAWS.SQSConfig{
			QueueName:           processorConfiguration.Subscriber.QueueName,
			QueueOwnerAccountID: processorConfiguration.Subscriber.AccountID,
			// The AWS configuration type uses SNS to SQS for notifications.
			// Gizmo by default will decode the SQS message using Base64 decoding.
			// However, the message body of SQS is the SNS message format which isn't Base64 encoded.
			// gatepr: Understand this when we do sqs testing
			ConsumeBase64: &ff,
		}
		sqsConfig.Region = processorConfiguration.Region
		var err error
		// gatepr: wrap this in retry
		sub, err = gizmoAWS.NewSubscriber(sqsConfig)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config [%+v] and err: %v", sqsConfig, err)
		}

		if err != nil {
			panic(err)
		}
		return NewPubSubProcessor(sub, scope)
	case configCommon.CloudDeploymentGCP:
		panic("Artifacts not implemented for GCP")
	case configCommon.CloudDeploymentSandbox:
		//
		return NewSandboxCloudEventProcessor()
	case configCommon.CloudDeploymentLocal:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Not using a background events processor, cfg is [%s]", processorConfiguration)
		return nil
	}
}
