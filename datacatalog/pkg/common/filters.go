package common

// Common constants and types for Filtering
const (
	DefaultPageOffset = 0
	MaxPageLimit      = 50
)

// Common Entity types that can be used on any filters
type Entity string

const (
	Artifact  Entity = "Artifact"
	Dataset   Entity = "Dataset"
	Partition Entity = "Partition"
	Tag       Entity = "Tag"
)

// Supported operators that can be used on filters
type ComparisonOperator int

const (
	Equal ComparisonOperator = iota
	// Add more operators as needed, ie., gte, lte
)
