package transformers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
)

func marshalMetadata(metadata *datacatalog.Metadata) ([]byte, error) {
	// if it is nil, marshal empty protobuf
	if metadata == nil {
		metadata = &datacatalog.Metadata{}
	}
	return proto.Marshal(metadata)
}

func unmarshalMetadata(serializedMetadata []byte) (*datacatalog.Metadata, error) {
	if serializedMetadata == nil {
		return nil, errors.NewDataCatalogErrorf(codes.Unknown, "Serialized metadata should never be nil")
	}
	var metadata datacatalog.Metadata
	err := proto.Unmarshal(serializedMetadata, &metadata)
	return &metadata, err
}
