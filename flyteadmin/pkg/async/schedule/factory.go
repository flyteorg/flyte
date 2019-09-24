package schedule

import (
	"context"

	gizmoConfig "github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsSchedule "github.com/lyft/flyteadmin/pkg/async/schedule/aws"
	"github.com/lyft/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/lyft/flyteadmin/pkg/async/schedule/noop"
	"github.com/lyft/flyteadmin/pkg/common"
	managerInterfaces "github.com/lyft/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

type WorkflowSchedulerConfig struct {
	Retries                int
	EventSchedulerConfig   runtimeInterfaces.EventSchedulerConfig
	WorkflowExecutorConfig runtimeInterfaces.WorkflowExecutorConfig
	Scope                  promutils.Scope
}

type WorkflowScheduler interface {
	GetEventScheduler() interfaces.EventScheduler
	GetWorkflowExecutor(executionManager managerInterfaces.ExecutionInterface,
		launchPlanManager managerInterfaces.LaunchPlanInterface) interfaces.WorkflowExecutor
}

type workflowScheduler struct {
	cfg              WorkflowSchedulerConfig
	eventScheduler   interfaces.EventScheduler
	workflowExecutor interfaces.WorkflowExecutor
}

func (w *workflowScheduler) GetEventScheduler() interfaces.EventScheduler {
	return w.eventScheduler
}

func (w *workflowScheduler) GetWorkflowExecutor(
	executionManager managerInterfaces.ExecutionInterface,
	launchPlanManager managerInterfaces.LaunchPlanInterface) interfaces.WorkflowExecutor {
	if w.workflowExecutor == nil {
		sqsConfig := gizmoConfig.SQSConfig{
			QueueName:           w.cfg.WorkflowExecutorConfig.ScheduleQueueName,
			QueueOwnerAccountID: w.cfg.WorkflowExecutorConfig.AccountID,
		}
		sqsConfig.Region = w.cfg.WorkflowExecutorConfig.Region
		w.workflowExecutor = awsSchedule.NewWorkflowExecutor(
			sqsConfig, executionManager, launchPlanManager, w.cfg.Scope.NewSubScope("workflow_executor"))
	}
	return w.workflowExecutor
}

func NewWorkflowScheduler(cfg WorkflowSchedulerConfig) WorkflowScheduler {
	var eventScheduler interfaces.EventScheduler
	var workflowExecutor interfaces.WorkflowExecutor

	switch cfg.EventSchedulerConfig.Scheme {
	case common.AWS:
		awsConfig := aws.NewConfig().WithRegion(cfg.WorkflowExecutorConfig.Region).WithMaxRetries(cfg.Retries)
		sess, err := session.NewSession(awsConfig)
		if err != nil {
			panic(err)
		}
		eventScheduler = awsSchedule.NewCloudWatchScheduler(
			cfg.EventSchedulerConfig.ScheduleRole, cfg.EventSchedulerConfig.TargetName, sess, awsConfig,
			cfg.Scope.NewSubScope("cloudwatch_scheduler"))
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop event scheduler implementation for cloud provider type [%s]",
			cfg.EventSchedulerConfig.Scheme)
		eventScheduler = noop.NewNoopEventScheduler()
	}

	switch cfg.WorkflowExecutorConfig.Scheme {
	case common.AWS:
		// Do nothing, this special case depends on the execution manager and launch plan manager having been
		// initialized and is handled in GetWorkflowExecutor.
		break
	case common.Local:
		fallthrough
	default:
		logger.Infof(context.Background(),
			"Using default noop workflow executor implementation for cloud provider type [%s]",
			cfg.EventSchedulerConfig.Scheme)
		workflowExecutor = noop.NewWorkflowExecutor()
	}
	return &workflowScheduler{
		cfg:              cfg,
		eventScheduler:   eventScheduler,
		workflowExecutor: workflowExecutor,
	}
}
