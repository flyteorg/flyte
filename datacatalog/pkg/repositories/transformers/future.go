package transformers

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func FromFutureLiteralToDynamicJobSpec(ctx context.Context, literal *core.Literal) (*core.DynamicJobSpec, error) {
	binary := literal.GetScalar().GetBinary().GetValue()
	var djSpec core.DynamicJobSpec
	err := proto.Unmarshal(binary, &djSpec)
	if err != nil {
		return nil, err
	}
	return &djSpec, nil
}

func FromDynamicJobSpecToLiteral(djSpec *core.DynamicJobSpec) (*core.Literal, error) {
	binary, err := proto.Marshal(djSpec)
	if err != nil {
		return nil, err
	}
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Binary{
					Binary: &core.Binary{
						Value: binary,
					},
				},
			},
		},
	}, nil
}
