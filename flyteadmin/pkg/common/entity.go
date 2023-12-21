package common

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type Entity string

const (
	Execution           Entity = "e"
	LaunchPlan          Entity = "l"
	NodeExecution       Entity = "ne"
	NodeExecutionEvent  Entity = "nee"
	Task                Entity = "t"
	TaskExecution       Entity = "te"
	Workflow            Entity = "w"
	NamedEntity         Entity = "nen"
	NamedEntityMetadata Entity = "nem"
	Project             Entity = "p"
	Signal              Entity = "s"
	AdminTag            Entity = "at"
	ExecutionAdminTag   Entity = "eat"
)

// ResourceTypeToEntity maps a resource type to an entity suitable for use with Database filters
var ResourceTypeToEntity = map[core.ResourceType]Entity{
	core.ResourceType_LAUNCH_PLAN: LaunchPlan,
	core.ResourceType_TASK:        Task,
	core.ResourceType_WORKFLOW:    Workflow,
}
