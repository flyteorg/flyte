package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAscii(t *testing.T) {
	assert.Equal(t, "bmls", EncodeBase64([]byte("nil")))
	assert.Equal(t, "w4RwZmVs", EncodeBase64([]byte("Äpfel")))
}

func TestDecodeFromAscii(t *testing.T) {
	type data struct {
		decoded     string
		encoded     string
		expectedErr error
	}
	tt := []data{
		{decoded: "nil", encoded: "bmls", expectedErr: nil},
		{decoded: "Äpfel", encoded: "w4RwZmVs", expectedErr: nil},
		{decoded: "", encoded: "Äpfel", expectedErr: base64.CorruptInputError(0)},
	}
	for _, testdata := range tt {
		actualDecoded, actualErr := DecodeFromBase64(testdata.encoded)
		assert.Equal(t, []byte(testdata.decoded), actualDecoded)
		assert.Equal(t, testdata.expectedErr, actualErr)
	}
}
