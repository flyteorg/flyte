package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=SignalInterface --output=../mocks --case=underscore --with-expecter

// Interface for managing Flyte Signals
type SignalInterface interface {
	GetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error)
	ListSignals(ctx context.Context, request admin.SignalListRequest) (*admin.SignalList, error)
	SetSignal(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error)
}
