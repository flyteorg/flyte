package catalog

import (
	"context"
	"encoding/base64"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/pbhash"
)

var emptyLiteralMap = core.LiteralMap{Literals: map[string]*core.Literal{}}

// Hashify a literal, in other words, produce a new literal where the corresponding value is removed in case
// the literal hash is set.
func hashify(literal *core.Literal) *core.Literal {
	// If the hash is set, return an empty literal with the same hash,
	// regardless of type (scalar/collection/map).
	if literal.GetHash() != "" {
		return &core.Literal{
			Hash: literal.GetHash(),
		}
	}

	// Two recursive cases:
	//   1. A collection of literals or
	//   2. A map of literals
	if literal.GetCollection() != nil {
		literals := literal.GetCollection().Literals
		literalsHash := make([]*core.Literal, 0)
		for _, lit := range literals {
			literalsHash = append(literalsHash, hashify(lit))
		}
		return &core.Literal{
			Value: &core.Literal_Collection{
				Collection: &core.LiteralCollection{
					Literals: literalsHash,
				},
			},
		}
	}
	if literal.GetMap() != nil {
		literalsMap := make(map[string]*core.Literal)
		for key, lit := range literal.GetMap().Literals {
			literalsMap[key] = hashify(lit)
		}
		return &core.Literal{
			Value: &core.Literal_Map{
				Map: &core.LiteralMap{
					Literals: literalsMap,
				},
			},
		}
	}

	return literal
}

func HashLiteralMap(ctx context.Context, literalMap *core.LiteralMap) (string, error) {
	if literalMap == nil || len(literalMap.Literals) == 0 {
		literalMap = &emptyLiteralMap
	}

	// Hashify, i.e. generate a copy of the literal map where each literal value is removed
	// in case the corresponding hash is set.
	hashifiedLiteralMap := make(map[string]*core.Literal, len(literalMap.Literals))
	for name, literal := range literalMap.Literals {
		hashifiedLiteralMap[name] = hashify(literal)
	}
	hashifiedInputs := &core.LiteralMap{
		Literals: hashifiedLiteralMap,
	}

	inputsHash, err := pbhash.ComputeHash(ctx, hashifiedInputs)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(inputsHash), nil
}
