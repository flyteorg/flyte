package interfaces

import (
	"gorm.io/gorm"

	schedulerInterfaces "github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/interfaces"
)

// The Repository indicates the methods that each Repository must support.
// A Repository indicates a Database which is collection of Tables/models.
// The goal is allow databases to be Plugged in easily.
type Repository interface {
	TaskRepo() TaskRepoInterface
	WorkflowRepo() WorkflowRepoInterface
	LaunchPlanRepo() LaunchPlanRepoInterface
	ExecutionRepo() ExecutionRepoInterface
	ExecutionEventRepo() ExecutionEventRepoInterface
	ProjectRepo() ProjectRepoInterface
	ResourceRepo() ResourceRepoInterface
	NodeExecutionRepo() NodeExecutionRepoInterface
	NodeExecutionEventRepo() NodeExecutionEventRepoInterface
	TaskExecutionRepo() TaskExecutionRepoInterface
	NamedEntityRepo() NamedEntityRepoInterface
	DescriptionEntityRepo() DescriptionEntityRepoInterface
	ConfigurationRepo() ConfigurationInterface
	SchedulableEntityRepo() schedulerInterfaces.SchedulableEntityRepoInterface
	ScheduleEntitiesSnapshotRepo() schedulerInterfaces.ScheduleEntitiesSnapShotRepoInterface
	SignalRepo() SignalRepoInterface

	GetGormDB() *gorm.DB
}
