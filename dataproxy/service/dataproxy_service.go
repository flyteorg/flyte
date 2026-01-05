package service

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/stow"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

type Service struct {
	dataproxy.UnimplementedDataProxyServiceServer

	cfg       config.DataProxyConfig
	dataStore *storage.DataStore
}

// NewService creates a new DataProxyService instance.
func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore) *Service {
	return &Service{
		cfg:       cfg,
		dataStore: dataStore,
	}
}

// CreateUploadLocation generates a signed URL for uploading data to the configured storage backend.
func (s *Service) CreateUploadLocation(
	ctx context.Context,
	req *connect.Request[dataproxy.CreateUploadLocationRequest],
) (*connect.Response[dataproxy.CreateUploadLocationResponse], error) {
	logger.Infof(ctx, "CreateUploadLocation request for project=%s, domain=%s, org=%s, filename=%s",
		req.Msg.Project, req.Msg.Domain, req.Msg.Org, req.Msg.Filename)

	// Validation on request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateUploadLocation request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := validateUploadRequest(ctx, req.Msg, s.cfg); err != nil {
		logger.Errorf(ctx, "Request validation failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Build the storage path
	storagePath, err := s.constructStoragePath(ctx, req.Msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to construct storage path: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to construct storage path: %w", err))
	}

	// Check if file already exists and validate for safe upload
	if err := s.checkFileExists(ctx, storagePath, req.Msg); err != nil {
		return nil, err
	}

	// Set expires_in to default if not provided in request
	if req.Msg.GetExpiresIn() == nil {
		req.Msg.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	// Create signed URL properties
	expiresIn := req.Msg.ExpiresIn.AsDuration()
	props := storage.SignedURLProperties{
		Scope:      stow.ClientMethodPut,
		ExpiresIn:  expiresIn,
		ContentMD5: string(req.Msg.ContentMd5),
	}

	// Generate signed URL
	signedResp, err := s.dataStore.CreateSignedURL(ctx, storagePath, props)
	if err != nil {
		logger.Errorf(ctx, "Failed to create signed URL: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create signed URL: %w", err))
	}

	// Build response
	expiresAt := time.Now().Add(expiresIn)
	resp := &dataproxy.CreateUploadLocationResponse{
		SignedUrl: signedResp.URL.String(),
		NativeUrl: storagePath.String(),
		ExpiresAt: timestamppb.New(expiresAt),
		Headers:   signedResp.RequiredRequestHeaders,
	}

	logger.Infof(ctx, "Successfully created upload location: native_url=%s, expires_at=%s",
		resp.NativeUrl, resp.ExpiresAt.AsTime().Format(time.RFC3339))

	return connect.NewResponse(resp), nil
}

// checkFileExists validates whether a file upload is safe by checking existing files.
// Returns an error if:
//   - File exists without content_md5 provided (cannot verify safe overwrite)
//   - File exists with different content_md5 (prevents accidental overwrite)
//
// Returns nil if:
//   - File does not exist (safe to upload)
//   - File exists with matching content_md5 (safe to re-upload same content)
func (s *Service) checkFileExists(ctx context.Context, storagePath storage.DataReference, req *dataproxy.CreateUploadLocationRequest) error {
	// Only check if both filename and filename_root are provided
	if len(req.GetFilename()) == 0 || len(req.GetFilenameRoot()) == 0 {
		return nil
	}

	metadata, err := s.dataStore.Head(ctx, storagePath)
	if err != nil {
		logger.Errorf(ctx, "Failed to check if file exists at location [%s]: %v", storagePath.String(), err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to check if file exists at location [%s]: %w", storagePath.String(), err))
	}

	if !metadata.Exists() {
		return nil
	}

	// Validate based on content hash if file exists
	// NOTE: This is a best-effort check. Race conditions may occur when multiple clients
	// upload to the same location simultaneously.
	if len(req.GetContentMd5()) == 0 {
		// Cannot verify content, reject to prevent accidental overwrites
		return connect.NewError(connect.CodeAlreadyExists,
			fmt.Errorf("file already exists at [%v]; content_md5 is required to verify safe overwrite", storagePath))
	}

	// Validate hash matches
	base64Digest := base64.StdEncoding.EncodeToString(req.GetContentMd5())
	if base64Digest != metadata.ContentMD5() {
		// Hash mismatch, reject to prevent overwriting different content
		logger.Errorf(ctx, "File exists at [%v] with different content hash", storagePath)
		return connect.NewError(connect.CodeAlreadyExists,
			fmt.Errorf("file already exists at [%v] with different content (hash mismatch)", storagePath))
	}

	// File exists with matching hash, allow upload to proceed
	logger.Debugf(ctx, "File already exists at [%v] with matching hash, allowing upload", storagePath)
	return nil
}

// constructStoragePath builds the storage path based on the request parameters.
// Path patterns:
//   - storage_prefix/org/project/domain/filename_root/filename (if filename_root is provided)
//   - storage_prefix/org/project/domain/base32_hash/filename (if only content_md5 is provided)
func (s *Service) constructStoragePath(ctx context.Context, req *dataproxy.CreateUploadLocationRequest) (storage.DataReference, error) {
	baseRef := s.dataStore.GetBaseContainerFQN(ctx)

	// Build path components: storage_prefix/org/project/domain/prefix/filename
	pathComponents := []string{s.cfg.Upload.StoragePrefix, req.GetOrg(), req.GetProject(), req.GetDomain()}

	// Set filename_root or base32-encoded content hash as prefix
	if len(req.GetFilenameRoot()) > 0 {
		pathComponents = append(pathComponents, req.GetFilenameRoot())
	} else {
		// URL-safe base32 encoding of content hash
		pathComponents = append(pathComponents, base32.StdEncoding.EncodeToString(req.GetContentMd5()))
	}

	pathComponents = append(pathComponents, req.GetFilename())

	// Filter out empty components to avoid double slashes in path
	pathComponents = lo.Filter(pathComponents, func(key string, _ int) bool {
		return key != ""
	})

	return s.dataStore.ConstructReference(ctx, baseRef, pathComponents...)
}
