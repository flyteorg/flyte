package transformers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
)

func TestFromReservationModel(t *testing.T) {
	t.Run("TestFromReservationModel", func(t *testing.T) {

		testKey := "test-key"
		testOwnerID := "test-owner-id"
		testTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

		mockReservationModel := &models.CacheReservation{
			Key:       testKey,
			OwnerID:   testOwnerID,
			ExpiresAt: testTime,
		}

		reservation := FromReservationModel(context.TODO(), mockReservationModel)
		assert.Equal(t, testKey, reservation.GetKey())
		assert.Equal(t, testOwnerID, reservation.GetOwnerId())
		assert.Equal(t, testTime, reservation.GetExpiresAt().AsTime())
	})
}
