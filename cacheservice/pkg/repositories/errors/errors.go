package errors

type ErrorTransformer interface {
	ToCacheServiceError(err error) error
}
