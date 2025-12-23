package interfaces

type SortOrder int

// Set of sort orders available for database queries.
const (
	SortOrderDescending SortOrder = iota
	SortOrderAscending
)

type SortParameter interface {
	GetGormOrderExpr() string
}
