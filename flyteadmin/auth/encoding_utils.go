package auth

import (
	"context"
	"encoding/base64"

	"github.com/flyteorg/flytestdlib/logger"
)

// EncodeBase64 returns the base64 encoded version of the data
func EncodeBase64(raw []byte) string {
	return base64.RawStdEncoding.EncodeToString(raw)
}

// DecodeFromBase64 returns the original encoded bytes and logs warning in case of error
func DecodeFromBase64(encodedData string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		logger.Warnf(context.TODO(), "Unable to decode %v due to %v", encodedData, err)
	}
	return decodedData, err
}
