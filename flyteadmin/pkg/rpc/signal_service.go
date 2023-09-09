package rpc

import (
	"context"
	"fmt"
	"runtime/debug"

	manager "github.com/flyteorg/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	runtimeIfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"

	"github.com/prometheus/client_golang/prometheus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SignalMetrics struct {
	scope        promutils.Scope
	panicCounter prometheus.Counter

	create util.RequestMetrics
	get    util.RequestMetrics
}

func NewSignalMetrics(scope promutils.Scope) SignalMetrics {
	return SignalMetrics{
		scope: scope,
		panicCounter: scope.MustNewCounter("handler_panic",
			"panics encountered while handling requests to the admin service"),
		create: util.NewRequestMetrics(scope, "create_signal"),
		get:    util.NewRequestMetrics(scope, "get_signal"),
	}
}

type SignalService struct {
	service.UnimplementedSignalServiceServer
	signalManager interfaces.SignalInterface
	metrics       SignalMetrics
}

func NewSignalServer(ctx context.Context, configuration runtimeIfaces.Configuration, adminScope promutils.Scope) *SignalService {
	panicCounter := adminScope.MustNewCounter("initialization_panic",
		"panics encountered initializing the signal service")

	defer func() {
		if err := recover(); err != nil {
			panicCounter.Inc()
			logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	databaseConfig := configuration.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := adminScope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(adminScope.NewSubScope("errors")), dbScope)

	signalManager := manager.NewSignalManager(repo, adminScope.NewSubScope("signal_manager"))

	logger.Info(ctx, "Initializing a new SignalService")
	return &SignalService{
		signalManager: signalManager,
		metrics:       NewSignalMetrics(adminScope),
	}
}

// Intercepts all admin requests to handle panics during execution.
func (s *SignalService) interceptPanic(ctx context.Context, request proto.Message) {
	err := recover()
	if err == nil {
		return
	}

	s.metrics.panicCounter.Inc()
	logger.Fatalf(ctx, "panic-ed for request: [%+v] with err: %v with Stack: %v", request, err, string(debug.Stack()))
}

func (s *SignalService) GetOrCreateSignal(
	ctx context.Context, request *admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	defer s.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.Signal
	var err error
	s.metrics.create.Time(func() {
		response, err = s.signalManager.GetOrCreateSignal(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &s.metrics.create)
	}
	s.metrics.create.Success()
	return response, nil
}

func (s *SignalService) ListSignals(
	ctx context.Context, request *admin.SignalListRequest) (*admin.SignalList, error) {
	defer s.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.SignalList
	var err error
	s.metrics.get.Time(func() {
		response, err = s.signalManager.ListSignals(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &s.metrics.get)
	}
	s.metrics.get.Success()
	return response, nil
}

func (s *SignalService) SetSignal(
	ctx context.Context, request *admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	defer s.interceptPanic(ctx, request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}
	var response *admin.SignalSetResponse
	var err error
	s.metrics.get.Time(func() {
		response, err = s.signalManager.SetSignal(ctx, *request)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &s.metrics.get)
	}
	s.metrics.get.Success()
	return response, nil
}
