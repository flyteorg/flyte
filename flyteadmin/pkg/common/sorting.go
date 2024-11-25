package common

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

const gormDescending = "%s desc"
const gormAscending = "%s asc"

type SortParameter interface {
	GetGormOrderExpr() string
}

type sortParamImpl struct {
	gormOrderExpression string
}

func (s *sortParamImpl) GetGormOrderExpr() string {
	return s.gormOrderExpression
}

func NewSortParameter(sort *admin.Sort, allowed sets.String) (SortParameter, error) {
	if sort == nil {
		return nil, nil
	}

	key := sort.GetKey()
	if !allowed.Has(key) {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort key '%s'", key)
	}

	var gormOrderExpression string
	switch sort.GetDirection() {
	case admin.Sort_DESCENDING:
		gormOrderExpression = fmt.Sprintf(gormDescending, key)
	case admin.Sort_ASCENDING:
		gormOrderExpression = fmt.Sprintf(gormAscending, key)
	default:
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort order specified: %v", sort)
	}
	return &sortParamImpl{
		gormOrderExpression: gormOrderExpression,
	}, nil
}
