package dataproxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyteadmin/pkg/config"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/flyteorg/stow"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type Service struct {
	service.DataProxyServiceServer

	cfg           config.DataProxyConfig
	dataStore     *storage.DataStore
	shardSelector ioutils.ShardSelector
}

// CreateUploadLocation creates a temporary signed url to allow callers to upload content.
func (s Service) CreateUploadLocation(ctx context.Context, req *service.CreateUploadLocationRequest) (
	*service.CreateUploadLocationResponse, error) {

	if len(req.Project) == 0 || len(req.Domain) == 0 {
		return nil, fmt.Errorf("prjoect and domain are required parameters")
	}

	if expiresIn := req.ExpiresIn; expiresIn != nil {
		if !expiresIn.IsValid() {
			return nil, fmt.Errorf("expiresIn [%v] is invalid", expiresIn)
		}

		if expiresIn.AsDuration() > s.cfg.Upload.MaxExpiresIn.Duration {
			return nil, fmt.Errorf("expiresIn [%v] cannot exceed max allowed expiration [%v]",
				expiresIn.AsDuration().String(), s.cfg.Upload.MaxExpiresIn.String())
		}
	} else {
		req.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	if len(req.Suffix) == 0 {
		req.Suffix = rand.String(20)
	}

	storagePath, err := createShardedStorageLocation(ctx, req, s.shardSelector, s.dataStore, s.cfg.Upload)
	if err != nil {
		return nil, err
	}

	resp, err := s.dataStore.CreateSignedURL(ctx, storagePath, storage.SignedURLProperties{
		Scope:     stow.ClientMethodPut,
		ExpiresIn: req.ExpiresIn.AsDuration(),
		// TODO: pass max allowed upload size
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create a signed url. Error: %w", err)
	}

	return &service.CreateUploadLocationResponse{
		SignedUrl: resp.URL.String(),
		NativeUrl: storagePath.String(),
		ExpiresAt: timestamppb.New(time.Now().Add(req.ExpiresIn.AsDuration())),
	}, nil
}

// createShardedStorageLocation creates a location in storage destination to maximize read/write performance in most
// block stores. The final location should look something like: s3://<my bucket>/<shard length>/<file name>
func createShardedStorageLocation(ctx context.Context, req *service.CreateUploadLocationRequest,
	shardSelector ioutils.ShardSelector, store *storage.DataStore, cfg config.DataProxyUploadConfig) (storage.DataReference, error) {
	keySuffixArr := make([]string, 0, 4)
	if len(cfg.StoragePrefix) > 0 {
		keySuffixArr = append(keySuffixArr, cfg.StoragePrefix)
	}

	keySuffixArr = append(keySuffixArr, req.Project, req.Domain, req.Suffix)
	prefix, err := shardSelector.GetShardPrefix(ctx, []byte(strings.Join(keySuffixArr, "/")))
	if err != nil {
		return "", err
	}

	storagePath, err := store.ConstructReference(ctx, store.GetBaseContainerFQN(ctx),
		append([]string{prefix}, keySuffixArr...)...)
	if err != nil {
		return "", fmt.Errorf("failed to construct datastore reference. Error: %w", err)
	}

	return storagePath, nil
}

func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore) (Service, error) {
	// Context is not used in the constructor. Should ideally be removed.
	selector, err := ioutils.NewBase36PrefixShardSelector(context.TODO())
	if err != nil {
		return Service{}, err
	}

	return Service{
		cfg:           cfg,
		dataStore:     dataStore,
		shardSelector: selector,
	}, nil
}
