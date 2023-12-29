package lib

import (
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestRenderDate(t *testing.T) {
	dt := time.Date(2020, 12, 8, 0, 0, 0, 0, time.UTC)
	pt := timestamp.Timestamp{
		Seconds: dt.Unix(),
		Nanos:   0,
	}
	lit := core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Datetime{Datetime: &pt}}}}},
	}

	txt, err := RenderLiteral(&lit)
	assert.NoError(t, err)
	assert.Equal(t, "2020-12-08", txt)
}
