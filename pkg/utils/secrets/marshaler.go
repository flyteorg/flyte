package secrets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	annotationPrefix = "flyte.secrets/s"
	PodLabel         = "inject-flyte-secrets"
	PodLabelValue    = "true"
)

// Copied from:
// https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/apimachinery/pkg/api/validation/objectmeta.go#L36
const totalAnnotationSizeLimitB int = 256 * (1 << 10) // 256 kB

func encodeSecret(secretAsString string) string {
	res := encoding.Base32Encoder.EncodeToString([]byte(secretAsString))
	return strings.TrimSuffix(res, "=")
}

func decodeSecret(encoded string) (string, error) {
	decodedRaw, err := encoding.Base32Encoder.DecodeString(encoded)
	if err != nil {
		return encoded, err
	}

	return string(decodedRaw), nil
}

func marshalSecret(s *core.Secret) string {
	return encodeSecret(proto.MarshalTextString(s))
}

func unmarshalSecret(encoded string) (*core.Secret, error) {
	decoded, err := decodeSecret(encoded)
	if err != nil {
		return nil, err
	}

	s := &core.Secret{}
	err = proto.UnmarshalText(decoded, s)
	return s, err
}

func MarshalSecretsToMapStrings(secrets []*core.Secret) (map[string]string, error) {
	res := make(map[string]string, len(secrets))
	for index, s := range secrets {
		if _, found := core.Secret_MountType_name[int32(s.MountRequirement)]; !found {
			return nil, fmt.Errorf("invalid mount requirement [%v]", s.MountRequirement)
		}

		encodedSecret := marshalSecret(s)
		res[annotationPrefix+strconv.Itoa(index)] = encodedSecret

		if len(encodedSecret) > totalAnnotationSizeLimitB {
			return nil, fmt.Errorf("secret descriptor cannot exceed [%v]", totalAnnotationSizeLimitB)
		}
	}

	return res, nil
}

func UnmarshalStringMapToSecrets(m map[string]string) ([]*core.Secret, error) {
	res := make([]*core.Secret, 0, len(m))
	for key, val := range m {
		if strings.HasPrefix(key, annotationPrefix) {
			s, err := unmarshalSecret(val)
			if err != nil {
				return nil, fmt.Errorf("error unmarshaling secret [%v]. Error: %w", key, err)
			}

			res = append(res, s)
		}
	}

	return res, nil
}
