package service

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"sort"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/stow"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"golang.org/x/sync/errgroup"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

type Service struct {
	dataproxy.UnimplementedDataProxyServiceServer

	cfg       config.DataProxyConfig
	dataStore *storage.DataStore
	repo      interfaces.Repository
}

// NewService creates a new DataProxyService instance.
func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore, repo interfaces.Repository) *Service {
	return &Service{
		cfg:       cfg,
		dataStore: dataStore,
		repo:      repo,
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
	expiresIn := req.Msg.GetExpiresIn().AsDuration()
	props := storage.SignedURLProperties{
		Scope:                 stow.ClientMethodPut,
		ExpiresIn:             expiresIn,
		ContentMD5:            base64.StdEncoding.EncodeToString(req.Msg.GetContentMd5()),
		AddContentMD5Metadata: req.Msg.GetAddContentMd5Metadata(),
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

// GetActionData gets input and output data for an action by reading from storage.
func (s *Service) GetActionData(
	ctx context.Context,
	req *connect.Request[workflow.GetActionDataRequest],
) (*connect.Response[workflow.GetActionDataResponse], error) {
	logger.Infof(ctx, "Received GetActionData request for: %s/%s",
		req.Msg.ActionId.Run.Name, req.Msg.ActionId.Name)

	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	action, err := s.repo.ActionRepo().GetAction(ctx, req.Msg.ActionId)
	if err != nil {
		logger.Errorf(ctx, "Failed to get action: %v", err)
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found: %w", err))
	}

	inputURI, _ := extractStorageURIs(action.ActionSpec)

	info := &workflow.RunInfo{}
	if err := proto.Unmarshal(action.DetailedInfo, info); err != nil {
		return nil, err
	}

	resp := &workflow.GetActionDataResponse{
		Inputs:  &task.Inputs{},
		Outputs: &task.Outputs{},
	}

	group, groupCtx := errgroup.WithContext(ctx)
	if inputURI != "" {
		group.Go(func() error {
			inputRef := storage.DataReference(inputURI)
			logger.Debugf(groupCtx, "Reading inputs from: %s", inputRef)
			if err := s.dataStore.ReadProtobuf(groupCtx, inputRef, resp.Inputs); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "Failed to read inputs from %s: %v", inputRef, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read inputs: %w", err))
				}
				logger.Debugf(groupCtx, "Inputs not found at %s", inputRef)
			} else {
				logger.Debugf(groupCtx, "Read %d input literals and %d action contexts", len(resp.Inputs.Literals), len(resp.Inputs.Context))
			}
			return nil
		})
	} else {
		logger.Warnf(ctx, "Action %s has empty InputURI", req.Msg.ActionId.Name)
	}

	if action.Phase == int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED) {
		group.Go(func() error {
			if workflow.ActionType(action.ActionType) == workflow.ActionType_ACTION_TYPE_TRACE {
				if info.GetOutputsUri() == "" {
					return nil
				}
				logger.Debugf(groupCtx, "Reading outputs from: %s", info.GetOutputsUri())

				outputMap := &core.LiteralMap{}
				if err := s.dataStore.ReadProtobuf(groupCtx, storage.DataReference(info.GetOutputsUri()), outputMap); err != nil {
					if !storage.IsNotFound(err) {
						logger.Errorf(groupCtx, "Failed to read outputs from %s: %v", info.GetOutputsUri(), err)
						return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read outputs: %w", err))
					}
					logger.Debugf(groupCtx, "Outputs not found at %s (action may not have finished)", info.GetOutputsUri())
				} else {
					resp.Outputs = literalMapToOutputs(outputMap)
					logger.Debugf(groupCtx, "Read %d output literals", len(resp.Outputs.Literals))
				}
				return nil
			}

			attempts, err := s.getAttempts(groupCtx, req.Msg.GetActionId())
			if err != nil {
				return err
			}

			if len(attempts) == 0 {
				return app.NewServerError(codes.NotFound, "outputs not available, no attempts for action")
			}

			outputUri := attempts[len(attempts)-1].GetOutputs().GetOutputUri()
			if outputUri == "" {
				return app.NewServerError(codes.NotFound, "outputs not available")
			}

			logger.Debugf(groupCtx, "Reading outputs from: %s", outputUri)
			outputMap := &core.LiteralMap{}
			if err := s.dataStore.ReadProtobuf(groupCtx, storage.DataReference(outputUri), outputMap); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "Failed to read outputs from %s: %v", outputUri, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read outputs: %w", err))
				}
				logger.Debugf(groupCtx, "Outputs not found at %s (action may not have finished)", outputUri)
			} else {
				resp.Outputs = literalMapToOutputs(outputMap)
				logger.Debugf(groupCtx, "Read %d output literals", len(resp.Outputs.Literals))
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	logger.Infof(ctx, "Retrieved action data for: %s (inputs=%d, outputs=%d)",
		req.Msg.ActionId.Name, len(resp.Inputs.Literals), len(resp.Outputs.Literals))
	return connect.NewResponse(resp), nil
}

// getAttempts looks up all events for an action and builds a list of attempts.
func (s *Service) getAttempts(ctx context.Context, actionId *common.ActionIdentifier) ([]*workflow.ActionAttempt, error) {
	eventModels, err := s.repo.ActionRepo().ListEvents(ctx, actionId, 500)
	if err != nil {
		return nil, err
	}

	events := make([]*workflow.ActionEvent, 0, len(eventModels))
	for _, m := range eventModels {
		event, err := m.ToActionEvent()
		if err != nil {
			logger.Warnf(ctx, "failed to convert action event model for action %v: %v", actionId, err)
			return nil, err
		}
		events = append(events, event)
	}

	attemptToEvents := map[uint32][]*workflow.ActionEvent{}
	for _, event := range events {
		attemptToEvents[event.GetAttempt()] = append(attemptToEvents[event.GetAttempt()], event)
	}

	attempts := make([]*workflow.ActionAttempt, 0, len(attemptToEvents))
	for attempt, evts := range attemptToEvents {
		merged := &workflow.ActionAttempt{Attempt: attempt}
		if len(evts) > 0 {
			lastEvent := evts[len(evts)-1]
			if lastEvent.GetOutputs() != nil {
				merged.Outputs = lastEvent.GetOutputs()
			}
		}
		attempts = append(attempts, merged)
	}
	sort.SliceStable(attempts, func(i, j int) bool {
		return attempts[i].GetAttempt() < attempts[j].GetAttempt()
	})

	return attempts, nil
}

// extractStorageURIs parses ActionSpec protobuf to extract InputUri and RunOutputBase.
func extractStorageURIs(specBytes []byte) (inputURI, runOutputBase string) {
	if len(specBytes) == 0 {
		return
	}
	var spec workflow.ActionSpec
	if err := proto.Unmarshal(specBytes, &spec); err != nil {
		return
	}
	return spec.GetInputUri(), spec.GetRunOutputBase()
}

// literalMapToOutputs converts a LiteralMap to task.Outputs.
func literalMapToOutputs(m *core.LiteralMap) *task.Outputs {
	if m == nil || len(m.Literals) == 0 {
		return &task.Outputs{}
	}
	literals := make([]*task.NamedLiteral, 0, len(m.Literals))
	for name, val := range m.Literals {
		literals = append(literals, &task.NamedLiteral{Name: name, Value: val})
	}
	sort.Slice(literals, func(i, j int) bool { return literals[i].Name < literals[j].Name })
	return &task.Outputs{Literals: literals}
}
