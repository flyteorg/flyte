package rpc

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetOrCreateSignal(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(&admin.Signal{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, &admin.SignalGetOrCreateRequest{})
		assert.NoError(t, err)
	})

	t.Run("NilRequestError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.GetOrCreateSignal(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.OnGetOrCreateSignalMatch(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

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
		signalManager.OnListSignalsMatch(mock.Anything, mock.Anything).Return(&admin.SignalList{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.ListSignals(ctx, &admin.SignalListRequest{})
		assert.NoError(t, err)
	})

	t.Run("NilRequestError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.ListSignals(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.OnListSignalsMatch(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

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
		signalManager.OnSetSignalMatch(mock.Anything, mock.Anything).Return(&admin.SignalSetResponse{}, nil)

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.NoError(t, err)
	})

	t.Run("NilRequestError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("ManagerError", func(t *testing.T) {
		signalManager := mocks.SignalInterface{}
		signalManager.OnSetSignalMatch(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

		testScope := mockScope.NewTestScope()
		mockServer := &SignalService{
			signalManager: &signalManager,
			metrics:       NewSignalMetrics(testScope),
		}

		_, err := mockServer.SetSignal(ctx, &admin.SignalSetRequest{})
		assert.Error(t, err)
	})
}
