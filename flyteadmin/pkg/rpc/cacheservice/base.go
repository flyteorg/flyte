package cacheservice

import (
	"context"
	"fmt"

	"runtime/debug"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	manager "github.com/flyteorg/flyteadmin/pkg/manager/impl"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/catalog"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
)

type CacheService struct {
	service.UnimplementedCacheServiceServer

	CacheManager interfaces.CacheInterface
	Metrics      CacheMetrics
}

// Intercepts all cache requests to handle panics during execution.
func (s *CacheService) interceptPanic(ctx context.Context, request proto.Message) {
	err := recover()
	if err == nil {
		return
	}

	s.Metrics.PanicCounter.Inc()
	logger.Fatalf(ctx, "panic-ed for request: [%+v] with err: %v with Stack: %v", request, err, string(debug.Stack()))
}

func NewCacheServer(ctx context.Context, config runtimeInterfaces.Configuration, storageClient *storage.DataStore,
	cacheScope promutils.Scope) *CacheService {
	panicCounter := cacheScope.MustNewCounter("initialization_panic",
		"panics encountered initializing the cache service")

	defer func() {
		if err := recover(); err != nil {
			panicCounter.Inc()
			logger.Fatalf(ctx, fmt.Sprintf("caught panic: %v [%+v]", err, string(debug.Stack())))
		}
	}()

	databaseConfig := config.ApplicationConfiguration().GetDbConfig()
	logConfig := logger.GetConfig()

	db, err := repositories.GetDB(ctx, databaseConfig, logConfig)
	if err != nil {
		logger.Fatal(ctx, err)
	}
	dbScope := cacheScope.NewSubScope("database")
	repo := repositories.NewGormRepo(
		db, errors.NewPostgresErrorTransformer(cacheScope.NewSubScope("errors")), dbScope)

	// flyteadmin config/auth optionally used by datacatalog client (if enabled)
	adminCfg := admin.GetConfig(ctx)
	adminCredentialsFuture := admin.NewPerRPCCredentialsFuture()
	adminAuthOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(admin.NewAuthInterceptor(adminCfg, nil, adminCredentialsFuture)),
		grpc.WithPerRPCCredentials(adminCredentialsFuture),
	}

	catalogClient, err := catalog.NewClient(ctx, adminAuthOpts...)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	cacheManager := manager.NewCacheManager(repo, config, catalogClient, cacheScope.NewSubScope("cache_manager"))

	logger.Info(ctx, "Initializing a new CacheService")
	return &CacheService{
		CacheManager: cacheManager,
		Metrics:      InitMetrics(cacheScope),
	}
}
