package models

import (
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Artifact is a wrapper object for easier handling of additional fields
type Artifact struct {
	artifact.Artifact
	OffloadedMetadata string
	LiteralTypeBytes  []byte
	LiteralValueBytes []byte
}

// Trigger - A trigger is nothing more than a launch plan, so wrap that.
type Trigger struct {
	// The launch plan below doesn't have an ID field for the trigger directly (it's nested), so add one here.
	Project string
	Domain  string
	Name    string
	Version string

	LaunchPlanID core.Identifier

	// The trigger as defined in the user code becomes a launch plan as it is the most similar.
	*admin.LaunchPlan

	RunsOn []core.ArtifactID

	// Additional meta fields relevant to the trigger.
	Active       bool
	SpecBytes    []byte
	ClosureBytes []byte
}
