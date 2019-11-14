package interfaces

import (
	"github.com/lyft/flyteadmin/pkg/common"
)

// Parameters for getting an individual resource.
type GetResourceInput struct {
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
}

// Describes a set of resources for which to apply attribute updates.
type UpdateResourceInput struct {
	Filters    []common.InlineFilter
	Attributes map[string]interface{}
}
