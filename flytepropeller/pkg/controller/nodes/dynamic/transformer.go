package dynamic

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const futureDataName = "future"

func FromFutureLiteralMapToDynamicJobSpec(futureLiteralMap *core.LiteralMap) (*core.DynamicJobSpec, error) {
	if futureLiteralMap == nil || len(futureLiteralMap.GetLiterals()) == 0 {
		return nil, nil
	}

	futureLiteral, ok := futureLiteralMap.GetLiterals()[futureDataName]
	if !ok {
		return nil, fmt.Errorf("Failed to retrieve Future literal")
	}

	if futureLiteral.GetScalar() == nil || futureLiteral.GetScalar().GetBinary() == nil {
		return nil, fmt.Errorf("Failed to get Future Binary from literal")
	}

	djSpec := &core.DynamicJobSpec{}
	err := proto.Unmarshal(futureLiteral.GetScalar().GetBinary().GetValue(), djSpec)
	if err != nil {
		return nil, err
	}

	return djSpec, nil
}
