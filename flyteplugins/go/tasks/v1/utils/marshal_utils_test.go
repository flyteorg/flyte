package utils

import (
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBackwardsCompatibility(t *testing.T) {
	sidecarProtoMessage := plugins.SidecarJob{
		PrimaryContainerName: "primary",
	}

	sidecarStruct, err := MarshalObjToStruct(sidecarProtoMessage)
	assert.NoError(t, err)

	// Set a new field in the struct to mimic what happens when we add new fields to protobuf messages
	sidecarStruct.Fields["hello"] = &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: "world",
		},
	}

	newSidecarJob := plugins.SidecarJob{}
	err = UnmarshalStruct(sidecarStruct, &newSidecarJob)
	assert.NoError(t, err)
}
