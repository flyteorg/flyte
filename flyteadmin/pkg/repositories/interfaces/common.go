package interfaces

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// Parameters for querying multiple resources.
type ListResourceInput struct {
	Limit           int
	Offset          int
	IdentifierScope *admin.NamedEntityIdentifier
	InlineFilters   []common.InlineFilter
	// MapFilters refers to primary entity filters defined as map values rather than inline sql queries.
	// These exist to permit filtering on "IS NULL" which isn't permitted with inline filter queries and
	// pq driver value substitution.
	MapFilters    []common.MapFilter
	SortParameter common.SortParameter
	// A set of the entities (besides the primary table being queried) that should be joined with when performing
	// the list query. This enables filtering on non-primary entity attributes.
	JoinTableEntities map[common.Entity]bool
}

// Describes a set of resources for which to apply attribute updates.
type UpdateResourceInput struct {
	Filters    []common.InlineFilter
	Attributes map[string]interface{}
}

// Parameters for counting multiple resources.
type CountResourceInput struct {
	IdentifierScope *admin.NamedEntityIdentifier
	InlineFilters   []common.InlineFilter
	// MapFilters refers to primary entity filters defined as map values rather than inline sql queries.
	// These exist to permit filtering on "IS NULL" which isn't permitted with inline filter queries and
	// pq driver value substitution.
	MapFilters []common.MapFilter
	// A set of the entities (besides the primary table being queried) that should be joined with when performing
	// the count query. This enables filtering on non-primary entity attributes.
	JoinTableEntities map[common.Entity]bool
}
