package catalog

import (
	"context"
	"encoding/base64"

	"k8s.io/utils/strings/slices"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
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
		literals := literal.GetCollection().GetLiterals()
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
		for key, lit := range literal.GetMap().GetLiterals() {
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

func HashLiteralMap(ctx context.Context, literalMap *core.LiteralMap, cacheIgnoreInputVars []string) (string, error) {
	if literalMap == nil || len(literalMap.GetLiterals()) == 0 {
		literalMap = &emptyLiteralMap
	}

	// Hashify, i.e. generate a copy of the literal map where each literal value is removed
	// in case the corresponding hash is set.
	hashifiedLiteralMap := make(map[string]*core.Literal, len(literalMap.GetLiterals()))
	for name, literal := range literalMap.GetLiterals() {
		if !slices.Contains(cacheIgnoreInputVars, name) {
			hashifiedLiteralMap[name] = hashify(literal)
		}
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
