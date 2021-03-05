package transformers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshaling(t *testing.T) {
	marshaledMetadata, err := marshalMetadata(metadata)
	assert.NoError(t, err)

	unmarshaledMetadata, err := unmarshalMetadata(marshaledMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, unmarshaledMetadata.KeyMap, metadata.KeyMap)
}

func TestMarshalingWithNil(t *testing.T) {
	marshaledMetadata, err := marshalMetadata(nil)
	assert.NoError(t, err)
	var expectedKeymap map[string]string
	unmarshaledMetadata, err := unmarshalMetadata(marshaledMetadata)
	assert.NoError(t, err)
	assert.EqualValues(t, expectedKeymap, unmarshaledMetadata.KeyMap)
}
