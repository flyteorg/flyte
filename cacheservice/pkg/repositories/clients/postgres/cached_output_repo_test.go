package postgres

import (
	"context"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

var sampleKey = "sampleKey"

func TestGetOutputNotFound(t *testing.T) {
	ctx := context.Background()

	cachedOutputRepo := NewPostgresOutputRepo(GetDbForTest(t), promutils.NewTestScope())
	result, err := cachedOutputRepo.Get(ctx, sampleKey)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualValues(t, codes.NotFound, status.Code(err))

}

func TestGetOutputError(t *testing.T) {
	ctx := context.Background()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(`SELECT * FROM "cached_outputs" WHERE "cached_outputs"."id" = $1 LIMIT 1`).WithReply([]map[string]interface{}{
		{
			"id": "sampleKey",
		},
	}).WithError(gorm.ErrInvalidDB)

	cachedOutputRepo := NewPostgresOutputRepo(GetDbForTest(t), promutils.NewTestScope())
	result, err := cachedOutputRepo.Get(ctx, sampleKey)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NotEqualf(t, codes.NotFound, status.Code(err), "Expected error other than NotFound")

}

func TestGetOutput(t *testing.T) {
	ctx := context.Background()

	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	GlobalMock.NewMock().WithQuery(`SELECT * FROM "cached_outputs" WHERE "cached_outputs"."id" = $1 LIMIT 1`).WithReply([]map[string]interface{}{
		{
			"id": "sampleKey",
		},
	})
	cachedOutputRepo := NewPostgresOutputRepo(GetDbForTest(t), promutils.NewTestScope())
	result, err := cachedOutputRepo.Get(ctx, sampleKey)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}
