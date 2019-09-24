package executions

import (
	"context"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
)

type project = string
type domain = string
type workflowName = string

type tag = string

type singleQueueConfiguration struct {
	PrimaryQueue string
	DynamicQueue string
}

type queueConfigSet = map[singleQueueConfiguration]bool

type queues = []singleQueueConfiguration

type queueConfig = map[tag]queues

// Catch-all queues for a project
type defaultProjectQueueAssignment = map[project]singleQueueConfiguration

/**
Catch-all queues for a project + domain combo
project: {
    domain: queue
}
*/
type defaultProjectDomainQueueAssignment = map[project]map[domain]singleQueueConfiguration

/**
Expanded workflowConfig structure:
  project: {
    domain: {
      workflow: queue
    }
  }
*/
// Stores an execution queue (when it exists) that matches all tags specified by a workflow config
type workflowQueueAssignment = map[project]map[domain]map[workflowName]singleQueueConfiguration

type QueueAllocator interface {
	GetQueue(ctx context.Context, identifier core.Identifier) singleQueueConfiguration
}

type queueAllocatorImpl struct {
	queueConfigMap                         queueConfig
	defaultQueue                           singleQueueConfiguration
	defaultProjectQueueAssignmentMap       defaultProjectQueueAssignment
	defaultProjectDomainQueueAssignmentMap defaultProjectDomainQueueAssignment
	workflowQueueAssignmentMap             workflowQueueAssignment
	config                                 runtimeInterfaces.Configuration
}

// Returns an arbitrary map entry's key from the input map. Used when a workflow can be run on multiple queues.
func getAnyMapKey(input queueConfigSet) singleQueueConfiguration {
	for key := range input {
		return key
	}
	// Nothing can be returned.
	logger.Error(context.Background(), "can't find any map key for empty map")
	return singleQueueConfiguration{}
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
				PrimaryQueue: queue.Primary,
				DynamicQueue: queue.Dynamic,
			})
		}
	}
	q.queueConfigMap = queueConfigMap
}

func (q *queueAllocatorImpl) findQueueCandidates(
	config runtimeInterfaces.WorkflowConfig) queueConfigSet {
	// go through and find queues that match *all* specified tags
	queueCandidates := make(queueConfigSet)
	for _, queue := range q.queueConfigMap[config.Tags[0]] {
		queueCandidates[queue] = true
	}
	for i := 1; i < len(config.Tags); i++ {
		filteredQueueCandidates := make(queueConfigSet)
		for _, queue := range q.queueConfigMap[config.Tags[i]] {
			if ok := queueCandidates[queue]; ok {
				filteredQueueCandidates[queue] = true
			}
		}
		if len(filteredQueueCandidates) == 0 {
			break
		}
		queueCandidates = filteredQueueCandidates
	}
	return queueCandidates
}

func (q *queueAllocatorImpl) refreshWorkflowQueueMap(workflowConfigs []runtimeInterfaces.WorkflowConfig) {
	logger.Debug(context.Background(), "refreshing workflow configs")
	var workflowQueueMap = make(workflowQueueAssignment)
	var projectQueueMap = make(defaultProjectQueueAssignment)
	var projectDomainQueueMap = make(defaultProjectDomainQueueAssignment)
	for _, config := range workflowConfigs {
		var queue singleQueueConfiguration
		// go through and find queues that match *all* specified tags
		queueCandidates := q.findQueueCandidates(config)

		if len(queueCandidates) > 0 {
			queue = getAnyMapKey(queueCandidates)
		}

		if config.Project == "" {
			// This is a default queue assignment
			q.defaultQueue = queue
			continue
		}

		// Now assign the queue to the most-specific configuration that is possible.
		projectSubMap, ok := workflowQueueMap[config.Project]
		if !ok {
			projectSubMap = make(map[domain]map[workflowName]singleQueueConfiguration)
		}
		// This queue applies to *all* workflows in this project
		if config.Domain == "" {
			projectQueueMap[config.Project] = queue
			continue
		}

		defaultProjectDomainMap, ok := projectDomainQueueMap[config.Project]
		if !ok {
			defaultProjectDomainMap = make(map[domain]singleQueueConfiguration)
			projectDomainQueueMap[config.Project] = defaultProjectDomainMap
		}

		// This queue applies to *all* workflows in this project + domain combo
		if config.WorkflowName == "" {
			defaultProjectDomainMap[config.Domain] = queue
			continue
		}

		// This queue applies to individual workflows with this project + domain + workflowName combo
		domainSubMap, ok := projectSubMap[config.Domain]
		if !ok {
			domainSubMap = make(map[workflowName]singleQueueConfiguration)
		}

		domainSubMap[config.WorkflowName] = queue
		projectSubMap[config.Domain] = domainSubMap
		workflowQueueMap[config.Project] = projectSubMap
	}
	q.defaultProjectQueueAssignmentMap = projectQueueMap
	q.defaultProjectDomainQueueAssignmentMap = projectDomainQueueMap
	q.workflowQueueAssignmentMap = workflowQueueMap
}

// Returns a queue specifically matching identifier project, domain, and name
// Barring a match for that, a queue matching a combination of project + domain will be returned.
// And if there is no existing match for that, a queue matching the project will be returned if it exists.
func (q *queueAllocatorImpl) getQueueForIdentifier(identifier core.Identifier) *singleQueueConfiguration {
	projectSubMap, ok := q.workflowQueueAssignmentMap[identifier.Project]
	if !ok {
		return nil
	}
	domainSubMap, ok := projectSubMap[identifier.Domain]
	if !ok {
		return nil
	}
	queue, ok := domainSubMap[identifier.Name]
	if !ok {
		return nil
	}
	return &queue
}

func (q *queueAllocatorImpl) getQueueForProjectAndDomain(identifier core.Identifier) *singleQueueConfiguration {
	domainSubMap, ok := q.defaultProjectDomainQueueAssignmentMap[identifier.Project]
	if !ok {
		return nil
	}
	defaultDomainQueue, ok := domainSubMap[identifier.Domain]
	if !ok {
		return nil
	}
	return &defaultDomainQueue
}

func (q *queueAllocatorImpl) getQueueForProject(identifier core.Identifier) *singleQueueConfiguration {
	queue, ok := q.defaultProjectQueueAssignmentMap[identifier.Project]
	if !ok {
		return nil
	}
	return &queue
}

func (q *queueAllocatorImpl) GetQueue(ctx context.Context, identifier core.Identifier) singleQueueConfiguration {
	// NOTE: If refreshing the execution queues & workflow configs on every call to GetQueue becomes too slow we should
	// investigate caching the computed queue assignments.
	executionQueues := q.config.QueueConfiguration().GetExecutionQueues()
	q.refreshExecutionQueues(executionQueues)

	workflowConfigs := q.config.QueueConfiguration().GetWorkflowConfigs()
	q.refreshWorkflowQueueMap(workflowConfigs)

	logger.Debugf(ctx,
		"Evaluating execution queue for [%+v] with available queues [%+v] and available workflow configs [%+v]",
		identifier, executionQueues, workflowConfigs)

	queue := q.getQueueForIdentifier(identifier)
	if queue != nil {
		logger.Debugf(ctx, "Found queue for identifier [%+v]: %v", identifier, queue)
		return *queue
	}
	queue = q.getQueueForProjectAndDomain(identifier)
	if queue != nil {
		logger.Debugf(ctx, "Found queue for project+domain [%s/%s]: %v", identifier.Project, identifier.Domain, queue)
		return *queue
	}
	queue = q.getQueueForProject(identifier)
	if queue != nil {
		logger.Debugf(ctx, "Found queue for project [%s]: %v", identifier.Project, queue)
		return *queue
	}
	return q.defaultQueue
}

func NewQueueAllocator(config runtimeInterfaces.Configuration) QueueAllocator {
	queueAllocator := queueAllocatorImpl{
		config: config,
	}
	return &queueAllocator
}
