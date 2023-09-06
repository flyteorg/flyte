package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// Defines the interface for interacting with launch plan models.
type LaunchPlanRepoInterface interface {
	// Inserts a launch plan model into the database store.
	Create(ctx context.Context, input models.LaunchPlan) error
	// Updates an existing launch plan in the database store.
	Update(ctx context.Context, input models.LaunchPlan) error
	// Sets the state to active for an existing launch plan in the database store
	// (and deactivates the formerly active version if the toDisable model exists).
	SetActive(ctx context.Context, toEnable models.LaunchPlan, toDisable *models.LaunchPlan) error
	// Returns a matching launch plan if it exists.
	Get(ctx context.Context, input Identifier) (models.LaunchPlan, error)
	// Returns launch plan revisions matching query parameters. A limit must be provided for the results page size.
	List(ctx context.Context, input ListResourceInput) (LaunchPlanCollectionOutput, error)
	// Returns a list of identifiers for launch plans.  A limit must be provided for the results page size.
	ListLaunchPlanIdentifiers(ctx context.Context, input ListResourceInput) (LaunchPlanCollectionOutput, error)
}

type SetStateInput struct {
	Identifier core.Identifier
	Version    string
	Closure    []byte
}

// Response format for a query on workflows.
type LaunchPlanCollectionOutput struct {
	LaunchPlans []models.LaunchPlan
}
