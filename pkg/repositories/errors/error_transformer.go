package errors

// Defines the basic error transformer interface that all database types must implement.
type ErrorTransformer interface {
	ToDataCatalogError(err error) error
}
