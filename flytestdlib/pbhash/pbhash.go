// This is a package that provides hashing utilities for Protobuf objects.
package pbhash

import (
	"context"
	"encoding/base64"

	goObjectHash "github.com/benlaurie/objecthash/go/objecthash"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var marshaller = &jsonpb.Marshaler{}

func fromHashToByteArray(input [32]byte) []byte {
	output := make([]byte, 32)
	copy(output, input[:])
	return output
}

// ComputeHash generate a deterministic hash in bytes for the pb object
func ComputeHash(ctx context.Context, pb proto.Message) ([]byte, error) {
	// We marshal the pb object to JSON first which should provide a consistent mapping of pb to json fields as stated
	// here: https://developers.google.com/protocol-buffers/docs/proto3#json
	// jsonpb marshalling includes:
	// - sorting map values to provide a stable output
	// - omitting empty values which supports backwards compatibility of old protobuf definitions
	// We do not use protobuf marshalling because it does not guarantee stable output because of how it handles
	// unknown fields and ordering of fields. https://github.com/protocolbuffers/protobuf/issues/2830
	pbJSON, err := marshaller.MarshalToString(pb)
	if err != nil {
		logger.Warning(ctx, "failed to marshal pb [%+v] to JSON with err %v", pb, err)
		return nil, err
	}

	// Deterministically hash the JSON object to a byte array. The library will sort the map keys of the JSON object
	// so that we do not run into the issues from pb marshalling.
	hash, err := goObjectHash.CommonJSONHash(pbJSON)
	if err != nil {
		logger.Warning(ctx, "failed to hash JSON for pb [%+v] with err %v", pb, err)
		return nil, err
	}

	return fromHashToByteArray(hash), err
}

// Generate a deterministic hash as a base64 encoded string for the pb object.
func ComputeHashString(ctx context.Context, pb proto.Message) (string, error) {
	hashBytes, err := ComputeHash(ctx, pb)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hashBytes), err
}
