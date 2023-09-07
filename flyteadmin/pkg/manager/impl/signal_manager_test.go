package impl

import (
	"context"
	"errors"
	"testing"

	repositoryMocks "github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	mockScope "github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	signalID = &core.SignalIdentifier{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		SignalId: "signal",
	}

	signalType = &core.LiteralType{
		Type: &core.LiteralType_Simple{
			Simple: core.SimpleType_BOOLEAN,
		},
	}

	signalValue = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Boolean{
							Boolean: false,
						},
					},
				},
			},
		},
	}
)

func TestGetOrCreateSignal(t *testing.T) {
	t.Run("Happy", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).OnGetOrCreateMatch(mock.Anything, mock.Anything).Return(nil)

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Id:   signalID,
			Type: signalType,
		}

		response, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(&admin.Signal{
			Id:   signalID,
			Type: signalType,
		}, response))
	})

	t.Run("ValidationError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Type: signalType,
		}

		_, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).OnGetOrCreateMatch(mock.Anything, mock.Anything).Return(errors.New("foo"))

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalGetOrCreateRequest{
			Id:   signalID,
			Type: signalType,
		}

		_, err := signalManager.GetOrCreateSignal(context.Background(), request)
		assert.Error(t, err)
	})
}

func TestListSignals(t *testing.T) {
	signalModel, err := transformers.CreateSignalModel(signalID, signalType, nil)
	assert.NoError(t, err)

	t.Run("Happy", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnListMatch(mock.Anything, mock.Anything).Return(
			[]models.Signal{signalModel},
			nil,
		)

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalListRequest{
			WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Limit: 20,
		}

		response, err := signalManager.ListSignals(context.Background(), request)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(
			&admin.SignalList{
				Signals: []*admin.Signal{
					&admin.Signal{
						Id:   signalID,
						Type: signalType,
					},
				},
			},
			response,
		))
	})

	t.Run("ValidationError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalListRequest{
			WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
		}

		_, err := signalManager.ListSignals(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnListMatch(mock.Anything, mock.Anything).Return(nil, errors.New("foo"))

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalListRequest{
			WorkflowExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Limit: 20,
		}

		_, err := signalManager.ListSignals(context.Background(), request)
		assert.Error(t, err)
	})
}

func TestSetSignal(t *testing.T) {
	signalModel, err := transformers.CreateSignalModel(signalID, signalType, nil)
	assert.NoError(t, err)

	t.Run("Happy", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnGetMatch(mock.Anything, mock.Anything, mock.Anything).Return(signalModel, nil)
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnUpdateMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil)

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalID,
			Value: signalValue,
		}

		response, err := signalManager.SetSignal(context.Background(), request)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(&admin.SignalSetResponse{}, response))
	})

	t.Run("ValidationError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBGetError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnGetMatch(mock.Anything, mock.Anything).Return(
			models.Signal{},
			errors.New("foo"),
		)

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalID,
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})

	t.Run("DBUpdateError", func(t *testing.T) {
		mockRepository := repositoryMocks.NewMockRepository()
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnGetMatch(mock.Anything, mock.Anything).Return(signalModel, nil)
		mockRepository.SignalRepo().(*repositoryMocks.SignalRepoInterface).
			OnUpdateMatch(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("foo"))

		signalManager := NewSignalManager(mockRepository, mockScope.NewTestScope())
		request := admin.SignalSetRequest{
			Id:    signalID,
			Value: signalValue,
		}

		_, err := signalManager.SetSignal(context.Background(), request)
		assert.Error(t, err)
	})
}
