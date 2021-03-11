package common

import (
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"
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

func NewSortParameter(sort admin.Sort) (SortParameter, error) {
	var gormOrderExpression string
	switch sort.Direction {
	case admin.Sort_DESCENDING:
		gormOrderExpression = fmt.Sprintf(gormDescending, sort.Key)
	case admin.Sort_ASCENDING:
		gormOrderExpression = fmt.Sprintf(gormAscending, sort.Key)
	default:
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid sort order specified: %v", sort)
	}
	return &sortParamImpl{
		gormOrderExpression: gormOrderExpression,
	}, nil
}
