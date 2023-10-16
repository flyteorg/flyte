// Implementation of an error transformer for test.
package errors

import (
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
)

type transformFunc func(err error) errors.FlyteAdminError

type testErrorTransformer struct {
	transformers []transformFunc
}

// Special method for test.
func (t *testErrorTransformer) AddTransform(transformer transformFunc) {
	t.transformers = append(t.transformers, transformer)
}

func (t *testErrorTransformer) ToFlyteAdminError(err error) errors.FlyteAdminError {
	for _, transformer := range t.transformers {
		if adminErr := transformer(err); adminErr != nil {
			return adminErr
		}
	}
	return errors.NewFlyteAdminError(codes.Unknown, "Test transformer failed to find transformation to apply")
}

func NewTestErrorTransformer() ErrorTransformer {
	return &testErrorTransformer{}
}
