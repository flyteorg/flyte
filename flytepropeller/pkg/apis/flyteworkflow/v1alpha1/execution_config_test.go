package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestRawOutputConfig(t *testing.T) {
	r := RawOutputDataConfig{&admin.RawOutputDataConfig{
		OutputLocationPrefix: "s3://bucket",
	}}
	assert.Equal(t, "s3://bucket", r.OutputLocationPrefix)
}
