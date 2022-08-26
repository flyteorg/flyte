package interfaces

import (
	"github.com/flyteorg/flyteadmin/pkg/common"
)

// Parameters for getting an individual resource.
type Identifier struct {
	Project string
	Domain  string
	Name    string
	Version string
}

// Parameters for querying multiple resources.
type ListResourceInput struct {
	Limit         int
	Offset        int
	InlineFilters []common.InlineFilter
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
	InlineFilters []common.InlineFilter
	// MapFilters refers to primary entity filters defined as map values rather than inline sql queries.
	// These exist to permit filtering on "IS NULL" which isn't permitted with inline filter queries and
	// pq driver value substitution.
	MapFilters []common.MapFilter
	// A set of the entities (besides the primary table being queried) that should be joined with when performing
	// the count query. This enables filtering on non-primary entity attributes.
	JoinTableEntities map[common.Entity]bool
}
