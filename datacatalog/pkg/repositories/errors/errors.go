// Generic errors used in the repos layer
package errors

import (
	"github.com/golang/protobuf/proto"
	"github.com/lyft/datacatalog/pkg/errors"
	"google.golang.org/grpc/codes"
)

const (
	notFound = "missing entity of type %s with identifier %v"
)

func GetMissingEntityError(entityType string, identifier proto.Message) error {
	return errors.NewDataCatalogErrorf(codes.NotFound, notFound, entityType, identifier)
}
