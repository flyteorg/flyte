package service

import (
	"context"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"slices"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/flyteorg/stow"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/logs"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	flyteIdlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

type Service struct {
	dataproxyconnect.UnimplementedDataProxyServiceHandler

	cfg           config.DataProxyConfig
	dataStore     *storage.DataStore
	taskClient    taskconnect.TaskServiceClient
	triggerClient triggerconnect.TriggerServiceClient
	runClient     workflowconnect.RunServiceClient
	projectClient projectconnect.ProjectServiceClient
	logStreamer   logs.LogStreamer
}

// NewService creates a new DataProxyService instance.
func NewService(cfg config.DataProxyConfig, dataStore *storage.DataStore, taskClient taskconnect.TaskServiceClient, triggerClient triggerconnect.TriggerServiceClient, runClient workflowconnect.RunServiceClient, projectClient projectconnect.ProjectServiceClient, logStreamer logs.LogStreamer) *Service {
	return &Service{
		cfg:           cfg,
		dataStore:     dataStore,
		taskClient:    taskClient,
		triggerClient: triggerClient,
		runClient:     runClient,
		projectClient: projectClient,
		logStreamer:   logStreamer,
	}
}

// CreateUploadLocation generates a signed URL for uploading data to the configured storage backend.
func (s *Service) CreateUploadLocation(
	ctx context.Context,
	req *connect.Request[dataproxy.CreateUploadLocationRequest],
) (*connect.Response[dataproxy.CreateUploadLocationResponse], error) {
	logger.Infof(ctx, "CreateUploadLocation request for project=%s, domain=%s, filename=%s",
		req.Msg.Project, req.Msg.Domain, req.Msg.Filename)

	// Validation on request
	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateUploadLocation request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := validateUploadRequest(ctx, req.Msg, s.cfg); err != nil {
		logger.Errorf(ctx, "Request validation failed: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := s.validateProjectExists(ctx, req.Msg.GetProject()); err != nil {
		return nil, err
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

// validateProjectExists checks that the given project ID exists by calling the ProjectService.
func (s *Service) validateProjectExists(ctx context.Context, projectID string) error {
	if _, err := s.projectClient.GetProject(ctx, connect.NewRequest(&project.GetProjectRequest{
		Id: projectID,
	})); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("project %q not found", projectID))
		}
		logger.Errorf(ctx, "Failed to validate project %q: %v", projectID, err)
		return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to validate project: %w", err))
	}
	return nil
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
//   - storage_prefix/project/domain/filename_root/filename (if filename_root is provided)
//   - storage_prefix/project/domain/base32_hash/filename (if only content_md5 is provided)
func (s *Service) constructStoragePath(ctx context.Context, req *dataproxy.CreateUploadLocationRequest) (storage.DataReference, error) {
	baseRef := s.dataStore.GetBaseContainerFQN(ctx)

	// Build path components: storage_prefix/project/domain/prefix/filename
	pathComponents := []string{s.cfg.Upload.StoragePrefix, req.GetProject(), req.GetDomain()}

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

// UploadInputs persists the given inputs to storage and returns a URI and hash
// that can be passed to CreateRun via OffloadedInputData.
func (s *Service) UploadInputs(
	ctx context.Context,
	req *connect.Request[dataproxy.UploadInputsRequest],
) (*connect.Response[dataproxy.UploadInputsResponse], error) {
	logger.Infof(ctx, "UploadInputs request received")

	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid UploadInputs request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	// Resolve org/project/domain from the identifier.
	var org, project, domain string
	switch id := req.Msg.Id.(type) {
	case *dataproxy.UploadInputsRequest_RunId:
		org = id.RunId.GetOrg()
		project = id.RunId.GetProject()
		domain = id.RunId.GetDomain()
	case *dataproxy.UploadInputsRequest_ProjectId:
		org = id.ProjectId.GetOrganization()
		project = id.ProjectId.GetName()
		domain = id.ProjectId.GetDomain()
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	if err := s.validateProjectExists(ctx, project); err != nil {
		return nil, err
	}

	// Resolve the task template to get cache_ignore_input_vars.
	taskTemplate, err := s.resolveTaskTemplate(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	// Filter out cache-ignored inputs before hashing.
	filteredInputs := filterInputs(req.Msg.GetInputs(), taskTemplate.GetMetadata().GetCacheIgnoreInputVars())

	// Deterministically hash the filtered inputs for cache key computation.
	inputsHash, err := hashInputsProto(filteredInputs)
	if err != nil {
		logger.Errorf(ctx, "Failed to hash inputs: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to hash inputs: %w", err))
	}

	// Build the storage path: storagePrefix/org/project/domain/offloaded-inputs/<hash>/inputs.pb
	storagePrefix := strings.TrimRight(s.cfg.Upload.StoragePrefix, "/")
	pathComponents := []string{storagePrefix, org, project, domain, "offloaded-inputs", inputsHash}
	pathComponents = lo.Filter(pathComponents, func(key string, _ int) bool {
		return key != ""
	})

	baseRef := s.dataStore.GetBaseContainerFQN(ctx)
	dirRef, err := s.dataStore.ConstructReference(ctx, baseRef, pathComponents...)
	if err != nil {
		logger.Errorf(ctx, "Failed to construct storage path: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to construct storage path: %w", err))
	}

	inputRef, err := s.dataStore.ConstructReference(ctx, dirRef, "inputs.pb")
	if err != nil {
		logger.Errorf(ctx, "Failed to construct input ref: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to construct input ref: %w", err))
	}

	// Store all inputs (unfiltered) — the hash is over the filtered set for caching.
	if err := s.dataStore.WriteProtobuf(ctx, inputRef, storage.Options{}, req.Msg.GetInputs()); err != nil {
		logger.Errorf(ctx, "Failed to write inputs to storage: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to write inputs: %w", err))
	}

	logger.Infof(ctx, "Successfully uploaded inputs to %s (hash=%s)", inputRef, inputsHash)

	return connect.NewResponse(&dataproxy.UploadInputsResponse{
		OffloadedInputData: &common.OffloadedInputData{
			Uri:        string(dirRef),
			InputsHash: inputsHash,
		},
	}), nil
}

// CreateDownloadLink generates signed URL(s) for downloading an artifact associated with a run action.
func (s *Service) CreateDownloadLink(
	ctx context.Context,
	req *connect.Request[dataproxy.CreateDownloadLinkRequest],
) (*connect.Response[dataproxy.CreateDownloadLinkResponse], error) {
	logger.Infof(ctx, "CreateDownloadLink request received")

	if err := req.Msg.Validate(); err != nil {
		logger.Errorf(ctx, "Invalid CreateDownloadLink request: %v", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	if req.Msg.GetArtifactType() != dataproxy.ArtifactType_ARTIFACT_TYPE_REPORT {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("artifact_type is required"))
	}

	// Set expires_in to default if not provided in request
	if req.Msg.GetExpiresIn() == nil {
		req.Msg.ExpiresIn = durationpb.New(s.cfg.Download.MaxExpiresIn.Duration)
	}
	expiresIn := req.Msg.GetExpiresIn().AsDuration()

	nativeURL, err := s.resolveArtifactURL(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	ref := storage.DataReference(nativeURL)
	meta, err := s.dataStore.Head(ctx, ref)
	if err != nil {
		logger.Errorf(ctx, "Failed to head artifact at [%s]: %v", nativeURL, err)
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("failed to check artifact existence: %w", err))
	}
	if !meta.Exists() {
		return nil, connect.NewError(connect.CodeNotFound,
			fmt.Errorf("artifact not found at [%s]", nativeURL))
	}

	signedResp, err := s.dataStore.CreateSignedURL(ctx, ref, storage.SignedURLProperties{
		Scope:     stow.ClientMethodGet,
		ExpiresIn: expiresIn,
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to create signed URL for [%s]: %v", nativeURL, err)
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("failed to create signed URL: %w", err))
	}

	expiresAt := timestamppb.New(time.Now().Add(expiresIn))
	return connect.NewResponse(&dataproxy.CreateDownloadLinkResponse{
		PreSignedUrls: &dataproxy.PreSignedURLs{
			SignedUrl: []string{signedResp.URL.String()},
			ExpiresAt: expiresAt,
		},
	}), nil
}

// resolveArtifactURL resolves the native storage URL for the requested artifact type and source.
func (s *Service) resolveArtifactURL(ctx context.Context, req *dataproxy.CreateDownloadLinkRequest) (string, error) {
	attemptIDEnvelope, ok := req.GetSource().(*dataproxy.CreateDownloadLinkRequest_ActionAttemptId)
	if !ok {
		return "", connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("unsupported source type"))
	}

	attemptID := attemptIDEnvelope.ActionAttemptId
	actionResp, err := s.runClient.GetActionDetails(ctx, connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: attemptID.GetActionId(),
	}))
	if err != nil {
		logger.Errorf(ctx, "Failed to get action details for %v: %v", attemptID.GetActionId(), err)
		return "", connect.NewError(connect.CodeNotFound,
			fmt.Errorf("failed to get action details: %w", err))
	}

	// Find the matching attempt by attempt number.
	var matchedAttempt *workflow.ActionAttempt
	for _, attempt := range actionResp.Msg.GetDetails().GetAttempts() {
		if attempt.GetAttempt() == attemptID.GetAttempt() {
			matchedAttempt = attempt
			break
		}
	}
	if matchedAttempt == nil {
		return "", connect.NewError(connect.CodeNotFound,
			fmt.Errorf("attempt %d not found for action [%v]", attemptID.GetAttempt(), attemptID.GetActionId()))
	}

	switch req.GetArtifactType() {
	case dataproxy.ArtifactType_ARTIFACT_TYPE_REPORT:
		reportURI := matchedAttempt.GetOutputs().GetReportUri()
		if reportURI == "" {
			return "", connect.NewError(connect.CodeNotFound,
				fmt.Errorf("no report URI found for action [%v] attempt %d", attemptID.GetActionId(), attemptID.GetAttempt()))
		}
		return reportURI, nil
	default:
		return "", connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("unsupported artifact type: %v", req.GetArtifactType()))
	}
}

// resolveTaskTemplate resolves the task template from the request's task oneof.
func (s *Service) resolveTaskTemplate(ctx context.Context, req *dataproxy.UploadInputsRequest) (*flyteIdlCore.TaskTemplate, error) {
	switch t := req.Task.(type) {
	case *dataproxy.UploadInputsRequest_TaskSpec:
		return t.TaskSpec.GetTaskTemplate(), nil
	case *dataproxy.UploadInputsRequest_TaskId:
		resp, err := s.taskClient.GetTaskDetails(ctx, connect.NewRequest(&task.GetTaskDetailsRequest{
			TaskId: t.TaskId,
		}))
		if err != nil {
			logger.Errorf(ctx, "Failed to get task details for %v: %v", t.TaskId, err)
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("failed to get task: %w", err))
		}
		return resp.Msg.GetDetails().GetSpec().GetTaskTemplate(), nil
	case *dataproxy.UploadInputsRequest_TriggerName:
		triggerResp, err := s.triggerClient.GetTriggerDetails(ctx, connect.NewRequest(&trigger.GetTriggerDetailsRequest{
			Name: t.TriggerName,
		}))
		if err != nil {
			logger.Errorf(ctx, "Failed to get trigger details for %v: %v", t.TriggerName, err)
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("failed to get trigger: %w", err))
		}
		triggerDetails := triggerResp.Msg.GetTrigger()
		taskID := &task.TaskIdentifier{
			Org:     t.TriggerName.GetOrg(),
			Project: t.TriggerName.GetProject(),
			Domain:  t.TriggerName.GetDomain(),
			Name:    t.TriggerName.GetTaskName(),
			Version: triggerDetails.GetSpec().GetTaskVersion(),
		}
		taskResp, err := s.taskClient.GetTaskDetails(ctx, connect.NewRequest(&task.GetTaskDetailsRequest{
			TaskId: taskID,
		}))
		if err != nil {
			logger.Errorf(ctx, "Failed to get task details for trigger %v: %v", t.TriggerName, err)
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("failed to get task for trigger: %w", err))
		}
		return taskResp.Msg.GetDetails().GetSpec().GetTaskTemplate(), nil
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("task is required"))
	}
}

// filterInputs returns a new Inputs with cache-ignored variables removed.
func filterInputs(inputs *task.Inputs, ignoreVars []string) *task.Inputs {
	if len(ignoreVars) == 0 {
		return inputs
	}
	var filtered []*task.NamedLiteral
	for _, nl := range inputs.GetLiterals() {
		if !slices.Contains(ignoreVars, nl.GetName()) {
			filtered = append(filtered, nl)
		}
	}
	return &task.Inputs{Literals: filtered}
}

// GetActionData gets input and output data for an action by calling RunService for URIs
// and reading the data from storage.
func (s *Service) GetActionData(
	ctx context.Context,
	req *connect.Request[dataproxy.GetActionDataRequest],
) (*connect.Response[dataproxy.GetActionDataResponse], error) {
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	actionId := req.Msg.GetActionId()

	urisResp, err := s.runClient.GetActionDataURIs(ctx, connect.NewRequest(&workflow.GetActionDataURIsRequest{
		ActionId: actionId,
	}))
	if err != nil {
		return nil, err
	}

	resp := &dataproxy.GetActionDataResponse{
		Inputs:  &task.Inputs{},
		Outputs: &task.Outputs{},
	}

	group, groupCtx := errgroup.WithContext(ctx)

	if urisResp.Msg.GetInputsUri() != "" {
		group.Go(func() error {
			baseRef := storage.DataReference(urisResp.Msg.GetInputsUri())
			inputRef, err := s.dataStore.ConstructReference(groupCtx, baseRef, "inputs.pb")
			if err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to construct input ref: %w", err))
			}
			logger.Infof(groupCtx, "GetActionData: reading inputs from %s", inputRef)
			if err := s.dataStore.ReadProtobuf(groupCtx, inputRef, resp.Inputs); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "GetActionData: failed to read inputs from %s: %v", inputRef, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read inputs from %s: %w", inputRef, err))
				}
			} else {
				logger.Debugf(groupCtx, "Read %d input literals and %d action contexts", len(resp.Inputs.Literals), len(resp.Inputs.Context))
			}
			return nil
		})
	} else {
		logger.Warnf(ctx, "Action %s has empty InputURI", req.Msg.ActionId.Name)
	}

	if urisResp.Msg.GetOutputsUri() != "" {
		group.Go(func() error {
			outputRef := storage.DataReference(urisResp.Msg.GetOutputsUri())
			logger.Infof(groupCtx, "GetActionData: reading outputs from %s", outputRef)
			var inputsOrOutputs task.Inputs
			if err := s.dataStore.ReadProtobuf(groupCtx, outputRef, &inputsOrOutputs); err != nil {
				if !storage.IsNotFound(err) {
					logger.Errorf(groupCtx, "GetActionData: failed to read outputs from %s: %v", outputRef, err)
					return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read outputs from %s: %w", outputRef, err))
				}
				logger.Debugf(groupCtx, "Outputs not found at %s (action may not have finished)", urisResp.Msg.GetOutputsUri())
			} else {
				resp.Outputs = &task.Outputs{
					Literals: inputsOrOutputs.GetLiterals(),
				}
				logger.Debugf(groupCtx, "Read %d output literals", len(resp.Outputs.Literals))
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

// TailLogs streams logs for an action attempt.
func (s *Service) TailLogs(ctx context.Context, req *connect.Request[dataproxy.TailLogsRequest], stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	// Get log context from RunService
	logCtxResp, err := s.runClient.GetActionLogContext(ctx, connect.NewRequest(&workflow.GetActionLogContextRequest{
		ActionId: req.Msg.GetActionId(),
		Attempt:  req.Msg.GetAttempt(),
	}))
	if err != nil {
		return err
	}

	logContext := logCtxResp.Msg.GetLogContext()
	if logContext == nil {
		return connect.NewError(connect.CodeNotFound, fmt.Errorf("no log context found"))
	}

	return s.logStreamer.TailLogs(ctx, logContext, stream)
}

// hashInputsProto computes a deterministic FNV-64a hash of the serialized inputs.
func hashInputsProto(inputs proto.Message) (string, error) {
	marshaller := proto.MarshalOptions{Deterministic: true}
	data, err := marshaller.Marshal(inputs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal inputs: %w", err)
	}
	h := fnv.New64a()
	_, _ = h.Write(data)
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}
