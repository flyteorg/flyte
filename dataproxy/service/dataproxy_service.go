package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/stow"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/durationpb"

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

	// - TODO: start from here: filename and filename root provided, need to check if file already exists -> check what is the best practice to handle it
	// - check expires in set (already done in validation)
	// - generate filename if not provided
	// - create storage location (prefix is filename root or md5)
	// - created signed url
	// - create response

	if len(req.Msg.GetFilename()) > 0 && len(req.Msg.GetFilenameRoot()) > 0 {
		// check if file already exists
	}

	// Build the storage path
	storagePath, err := s.constructStoragePath(ctx, req.Msg)
	if err != nil {
		logger.Errorf(ctx, "Failed to construct storage path: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to construct storage path: %w", err))
	}

	// Set expires in to default if not provided in request
	if expiresIn := req.Msg.GetExpiresIn(); expiresIn == nil {
		req.Msg.ExpiresIn = durationpb.New(s.cfg.Upload.MaxExpiresIn.Duration)
	}

	// 3. Create signed URL properties
	props := storage.SignedURLProperties{
		Scope:      stow.ClientMethodPut,
		ExpiresIn:  req.Msg.ExpiresIn.AsDuration(),
		ContentMD5: string(req.Msg.ContentMd5),
	}

	// 4. Generate signed URL
	signedResp, err := s.dataStore.CreateSignedURL(ctx, storagePath, props)
	if err != nil {
		logger.Errorf(ctx, "Failed to create signed URL: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create signed URL: %w", err))
	}

	// 5. Build response
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

// constructStoragePath builds the storage path based on the request parameters.
// Path patterns:
//   - project/domain/(md5_hash)/filename (if filename is present)
//   - project/domain/filename_root/filename (if filename_root and filename are present)
func (s *Service) constructStoragePath(ctx context.Context, req *dataproxy.CreateUploadLocationRequest) (storage.DataReference, error) {
	// TODO: rewrite this
	// Get base container
	baseRef := s.dataStore.GetBaseContainerFQN(ctx)

	// Build path components
	pathComponents := []string{}

	// Add org if present
	if req.Org != "" {
		pathComponents = append(pathComponents, req.Org)
	}

	// Add project and domain
	pathComponents = append(pathComponents, req.Project, req.Domain)

	// Add filename_root or md5 hash
	if req.FilenameRoot != "" {
		// Use filename_root if provided
		pathComponents = append(pathComponents, req.FilenameRoot)
	} else if len(req.ContentMd5) > 0 {
		// Use MD5 hash as directory name
		md5Hash := hex.EncodeToString(req.ContentMd5)
		pathComponents = append(pathComponents, md5Hash)
	} else {
		// Generate a hash from content_md5 for consistency
		hasher := md5.New()
		hasher.Write([]byte(fmt.Sprintf("%s/%s/%s", req.Project, req.Domain, req.Filename)))
		md5Hash := hex.EncodeToString(hasher.Sum(nil))
		pathComponents = append(pathComponents, md5Hash)
	}

	// Add filename if present
	if req.Filename != "" {
		pathComponents = append(pathComponents, req.Filename)
	}

	// Construct the full reference
	return s.dataStore.ConstructReference(ctx, baseRef, pathComponents...)
}

// validateUploadRequest performs validation on the upload request
func validateUploadRequest(ctx context.Context, req *dataproxy.CreateUploadLocationRequest, cfg config.DataProxyConfig) error {

	if len(req.FilenameRoot) == 0 && len(req.ContentMd5) == 0 {
		return fmt.Errorf("Either filename_root or content_md5 must be provided")
	}

	// Validate expires_in against platform maximum
	if req.ExpiresIn != nil {
		if req.ExpiresIn.IsValid() {
			return fmt.Errorf("expires_in (%v) is invalid", req.ExpiresIn)
		}
		maxExpiration := cfg.Upload.MaxExpiresIn.Duration
		if maxExpiration > 0 && req.ExpiresIn.AsDuration() > maxExpiration {
			return fmt.Errorf("expires_in (%v) exceeds maximum allowed duration of %v",
				req.ExpiresIn.AsDuration(), maxExpiration)
		}
	}

	// Validate content_length against platform maximum
	if req.ContentLength > 0 {
		maxSize := cfg.Upload.MaxSize.Value()
		if maxSize > 0 && req.ContentLength > maxSize {
			return fmt.Errorf("content_length (%d bytes) exceeds maximum allowed size of %d bytes",
				req.ContentLength, maxSize)
		}
	}

	return nil
}
