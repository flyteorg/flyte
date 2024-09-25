package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestGetOrCreateSignal(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().GetOrCreateSignal(mock.Anything, mock.Anything).Return(&admin.Signal{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, &admin.SignalGetOrCreateRequest{})
		assert.NoError(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().GetOrCreateSignal(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, &admin.SignalGetOrCreateRequest{})
		assert.Error(t, err)
	})
}

func TestListSignals(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().ListSignals(mock.Anything, mock.Anything).Return(&admin.SignalList{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.ListSignals(ctx, &admin.SignalListRequest{})
		assert.NoError(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().ListSignals(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.ListSignals(ctx, &admin.SignalListRequest{})
		assert.Error(t, err)
	})
}

func TestSetSignal(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().SetSignal(mock.Anything, mock.Anything).Return(&admin.SignalSetResponse{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.NoError(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.EXPECT().SetSignal(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.Error(t, err)
	})
}
