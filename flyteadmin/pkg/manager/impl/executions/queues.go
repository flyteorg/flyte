package executions

import (
	"context"
	"math/rand"

	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/resources"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
)

type tag = string

type singleQueueConfiguration struct {
	DynamicQueue string
}

type queues = []singleQueueConfiguration

type queueConfig = map[tag]queues

type QueueAllocator interface {
	GetQueue(ctx context.Context, identifier core.Identifier) singleQueueConfiguration
}

type queueAllocatorImpl struct {
	queueConfigMap  queueConfig
	config          runtimeInterfaces.Configuration
	db              repoInterfaces.Repository
	resourceManager interfaces.ResourceInterface
}

func (q *queueAllocatorImpl) refreshExecutionQueues(executionQueues []runtimeInterfaces.ExecutionQueue) {
	logger.Debug(context.Background(), "refreshing execution queues")
	var queueConfigMap = make(queueConfig)
	for _, queue := range executionQueues {
		for _, tag := range queue.Attributes {
			queuesForTag, ok := queueConfigMap[tag]
			if !ok {
				queuesForTag = make(queues, 0, 1)
			}
			queueConfigMap[tag] = append(queuesForTag, singleQueueConfiguration{
				DynamicQueue: queue.Dynamic,
			})
		}
	}
	q.queueConfigMap = queueConfigMap
}

func (q *queueAllocatorImpl) GetQueue(ctx context.Context, identifier core.Identifier) singleQueueConfiguration {
	// NOTE: If refreshing the execution queues & workflow configs on every call to GetQueue becomes too slow we should
	// investigate caching the computed queue assignments.
	executionQueues := q.config.QueueConfiguration().GetExecutionQueues()
	q.refreshExecutionQueues(executionQueues)

	resource, err := q.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Workflow:     identifier.Name,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	})

	if err != nil {
		logger.Warningf(ctx, "Failed to fetch override values when assigning execution queue for [%+v] with err: %v",
			identifier, err)
	}

	if resource != nil && resource.Attributes != nil && resource.Attributes.GetExecutionQueueAttributes() != nil {
		for _, tag := range resource.Attributes.GetExecutionQueueAttributes().Tags {
			matches, ok := q.queueConfigMap[tag]
			if !ok {
				continue
			}
			/* #nosec */
			return matches[rand.Intn(len(matches))]
		}
	}
	var tags []string
	var defaultTags []string
	// If we've made it this far, check to see if a domain-specific default workflow config exists for this particular domain.
	for _, workflowConfig := range q.config.QueueConfiguration().GetWorkflowConfigs() {
		if workflowConfig.Domain == identifier.Domain {
			tags = workflowConfig.Tags
		} else if len(workflowConfig.Domain) == 0 {
			defaultTags = workflowConfig.Tags
		}
	}
	if len(tags) == 0 {
		// Use the uber-default queue
		tags = defaultTags
	}
	for _, tag := range tags {
		matches, ok := q.queueConfigMap[tag]
		if !ok {
			continue
		}
		/* #nosec */
		return matches[rand.Intn(len(matches))]
	}
	logger.Infof(ctx, "found no matching queue for [%+v]", identifier)
	return singleQueueConfiguration{}
}

func NewQueueAllocator(config runtimeInterfaces.Configuration, db repoInterfaces.Repository) QueueAllocator {
	queueAllocator := queueAllocatorImpl{
		config:          config,
		db:              db,
		resourceManager: resources.NewResourceManager(db, config.ApplicationConfiguration()),
	}
	return &queueAllocator
}
