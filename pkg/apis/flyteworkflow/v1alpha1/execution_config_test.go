package v1alpha1

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestRawOutputConfig(t *testing.T) {
	r := RawOutputDataConfig{&admin.RawOutputDataConfig{
		OutputLocationPrefix: "s3://bucket",
	}}
	assert.Equal(t, "s3://bucket", r.OutputLocationPrefix)
}
