package transformers

import (
	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
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
