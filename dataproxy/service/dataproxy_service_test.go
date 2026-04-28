package service

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	flyteconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	projectMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	workflowMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

func TestCreateUploadLocation(t *testing.T) {
	ctx := context.Background()

	cfg := config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{
			MaxExpiresIn:          flyteconfig.Duration{Duration: 1 * time.Hour},
			MaxSize:               resource.MustParse("100Mi"), // 100MB
			StoragePrefix:         "uploads",
			DefaultFileNameLength: 20,
		},
	}

	tests := []struct {
		name           string
		req            *dataproxy.CreateUploadLocationRequest
		wantErr        bool
		errContains    string
		validateResult func(t *testing.T, resp *connect.Response[dataproxy.CreateUploadLocationResponse])
	}{
		{
			name: "success with valid request",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
				ContentMd5:   []byte("test-hash"),
				ExpiresIn:    durationpb.New(30 * time.Minute),
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *connect.Response[dataproxy.CreateUploadLocationResponse]) {
				assert.Contains(t, resp.Msg.SignedUrl, "https://test-bucket")
				assert.Contains(t, resp.Msg.NativeUrl, "uploads/test-project/test-domain/test-root/test-file.txt")
			},
		},
		{
			name: "validation error - missing both filename_root and content_md5",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:  "test-project",
				Domain:   "test-domain",
				Filename: "test-file.txt",
			},
			wantErr:     true,
			errContains: "either filename_root or content_md5 must be provided",
		},
		{
			name: "validation error - expires_in exceeds maximum",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
				ExpiresIn:    durationpb.New(2 * time.Hour),
			},
			wantErr:     true,
			errContains: "exceeds maximum allowed duration",
		},
		{
			name: "validation error - content_length exceeds maximum",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:       "test-project",
				Domain:        "test-domain",
				Filename:      "test-file.txt",
				FilenameRoot:  "test-root",
				ContentLength: 1024 * 1024 * 200, // 200MB
			},
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := setupMockDataStore(t)
			mockProjectClient := projectMocks.NewProjectServiceClient(t)
			if !tt.wantErr {
				mockProjectClient.On("GetProject", mock.Anything, mock.Anything).Return(
					connect.NewResponse(&project.GetProjectResponse{}), nil)
			}
			service := NewService(cfg, mockStore, nil, nil, nil, mockProjectClient, nil)

			req := &connect.Request[dataproxy.CreateUploadLocationRequest]{
				Msg: tt.req,
			}

			resp, err := service.CreateUploadLocation(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateResult != nil {
					tt.validateResult(t, resp)
				}
			}
		})
	}
}

func TestCheckFileExists(t *testing.T) {
	ctx := context.Background()

	cfg := config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{
			StoragePrefix: "uploads",
		},
	}

	tests := []struct {
		name            string
		req             *dataproxy.CreateUploadLocationRequest
		existingFileMD5 string // Empty means file doesn't exist
		expectErr       bool
		errContains     string
	}{
		{
			name: "file does not exist",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
			},
			existingFileMD5: "", // File doesn't exist
			expectErr:       false,
		},
		{
			name: "file exists without hash provided",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
				// No ContentMd5 provided
			},
			existingFileMD5: "existing-hash",
			expectErr:       true,
			errContains:     "content_md5 is required to verify safe overwrite",
		},
		{
			name: "file exists with matching hash",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
				ContentMd5:   []byte("test-hash-123"),
			},
			existingFileMD5: base64.StdEncoding.EncodeToString([]byte("test-hash-123")),
			expectErr:       false,
		},
		{
			name: "file exists with different hash",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
				ContentMd5:   []byte("different-hash"),
			},
			existingFileMD5: base64.StdEncoding.EncodeToString([]byte("existing-hash")),
			expectErr:       true,
			errContains:     "with different content (hash mismatch)",
		},
		{
			name: "skip check when filename is empty",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "", // Empty filename
				FilenameRoot: "test-root",
			},
			existingFileMD5: "",
			expectErr:       false,
		},
		{
			name: "skip check when filename_root is empty",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "", // Empty filename_root
			},
			existingFileMD5: "",
			expectErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockStore *storage.DataStore
			if tt.existingFileMD5 == "" {
				mockStore = setupMockDataStore(t)
			} else {
				mockStore = setupMockDataStoreWithExistingFile(t, tt.existingFileMD5)
			}

			service := NewService(cfg, mockStore, nil, nil, nil, nil, nil)
			storagePath := storage.DataReference("s3://test-bucket/uploads/test-project/test-domain/test-root/test-file.txt")

			err := service.checkFileExists(ctx, storagePath, tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConstructStoragePath(t *testing.T) {
	ctx := context.Background()

	cfg := config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{
			StoragePrefix: "uploads",
		},
	}

	tests := []struct {
		name         string
		req          *dataproxy.CreateUploadLocationRequest
		expectedPath string
	}{
		{
			name: "with filename_root",
			req: &dataproxy.CreateUploadLocationRequest{
				Org:          "test-org",
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
			},
			expectedPath: "s3://test-bucket/uploads/test-project/test-domain/test-root/test-file.txt",
		},
		{
			name: "with content_md5 uses base32 encoding",
			req: &dataproxy.CreateUploadLocationRequest{
				Project:    "test-project",
				Domain:     "test-domain",
				Filename:   "test-file.txt",
				ContentMd5: []byte("test-hash"),
			},
			// base32 encoding for "test-hash" is ORSXG5BNNBQXG2A=
			expectedPath: "s3://test-bucket/uploads/test-project/test-domain/ORSXG5BNNBQXG2A=/test-file.txt",
		},
		{
			name: "filters empty org component",
			req: &dataproxy.CreateUploadLocationRequest{
				Org:          "", // Empty org
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
			},
			expectedPath: "s3://test-bucket/uploads/test-project/test-domain/test-root/test-file.txt",
		},
		{
			name: "with all components including org",
			req: &dataproxy.CreateUploadLocationRequest{
				Org:          "test-org",
				Project:      "test-project",
				Domain:       "test-domain",
				Filename:     "test-file.txt",
				FilenameRoot: "test-root",
			},
			expectedPath: "s3://test-bucket/uploads/test-project/test-domain/test-root/test-file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := setupMockDataStore(t)
			service := NewService(cfg, mockStore, nil, nil, nil, nil, nil)

			path, err := service.constructStoragePath(ctx, tt.req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPath, path.String())
		})
	}
}

// Helper functions to setup mocks

// simpleRefConstructor is a simple implementation of ReferenceConstructor for testing
type simpleRefConstructor struct{}

func (s *simpleRefConstructor) ConstructReference(ctx context.Context, base storage.DataReference, keys ...string) (storage.DataReference, error) {
	path := string(base)
	for _, key := range keys {
		if key != "" {
			path += "/" + key
		}
	}
	return storage.DataReference(path), nil
}

func setupMockDataStore(t *testing.T) *storage.DataStore {
	mockComposedStore := storageMocks.NewComposedProtobufStore(t)

	// Setup base container
	mockComposedStore.On("GetBaseContainerFQN", mock.Anything).Return(storage.DataReference("s3://test-bucket")).Maybe()

	// Setup Head to return file does not exist
	mockMetadata := storageMocks.NewMetadata(t)
	mockMetadata.On("Exists").Return(false).Maybe()
	mockComposedStore.On("Head", mock.Anything, mock.Anything).Return(mockMetadata, nil).Maybe()

	// Setup CreateSignedURL
	testURL, _ := url.Parse("https://test-bucket.s3.amazonaws.com/signed-url")
	mockComposedStore.On("CreateSignedURL", mock.Anything, mock.Anything, mock.Anything).Return(
		storage.SignedURLResponse{
			URL:                    *testURL,
			RequiredRequestHeaders: map[string]string{"Content-Type": "application/octet-stream"},
		}, nil).Maybe()

	return &storage.DataStore{
		ComposedProtobufStore: mockComposedStore,
		ReferenceConstructor:  &simpleRefConstructor{},
	}
}

func TestUploadInputs(t *testing.T) {
	ctx := context.Background()

	cfg := config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{
			StoragePrefix: "uploads",
		},
	}

	testTaskSpec := &task.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Id:       &core.Identifier{Name: "test-task"},
			Metadata: &core.TaskMetadata{},
		},
	}

	testTaskSpecWithIgnoredVars := &task.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Id: &core.Identifier{Name: "test-task"},
			Metadata: &core.TaskMetadata{
				CacheIgnoreInputVars: []string{"y"},
			},
		},
	}

	tests := []struct {
		name           string
		req            *dataproxy.UploadInputsRequest
		wantErr        bool
		errCode        connect.Code
		validateResult func(t *testing.T, resp *connect.Response[dataproxy.UploadInputsResponse])
	}{
		{
			name: "success with run_id and task_spec",
			req: &dataproxy.UploadInputsRequest{
				Id: &dataproxy.UploadInputsRequest_RunId{
					RunId: &common.RunIdentifier{
						Org:     "test-org",
						Project: "test-project",
						Domain:  "test-domain",
						Name:    "test-run",
					},
				},
				Task: &dataproxy.UploadInputsRequest_TaskSpec{TaskSpec: testTaskSpec},
				Inputs: &task.Inputs{
					Literals: []*task.NamedLiteral{
						{Name: "x", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 42}}}}}}},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *connect.Response[dataproxy.UploadInputsResponse]) {
				assert.NotNil(t, resp.Msg.OffloadedInputData)
				assert.NotEmpty(t, resp.Msg.OffloadedInputData.Uri)
				assert.NotEmpty(t, resp.Msg.OffloadedInputData.InputsHash)
				assert.Contains(t, resp.Msg.OffloadedInputData.Uri, "uploads/test-org/test-project/test-domain/offloaded-inputs/")
			},
		},
		{
			name: "success with project_id",
			req: &dataproxy.UploadInputsRequest{
				Id: &dataproxy.UploadInputsRequest_ProjectId{
					ProjectId: &common.ProjectIdentifier{
						Organization: "test-org",
						Name:         "test-project",
						Domain:       "test-domain",
					},
				},
				Task: &dataproxy.UploadInputsRequest_TaskSpec{TaskSpec: testTaskSpec},
				Inputs: &task.Inputs{
					Literals: []*task.NamedLiteral{
						{Name: "y", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "hello"}}}}}}},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *connect.Response[dataproxy.UploadInputsResponse]) {
				assert.NotNil(t, resp.Msg.OffloadedInputData)
				assert.Contains(t, resp.Msg.OffloadedInputData.Uri, "uploads/test-org/test-project/test-domain/offloaded-inputs/")
			},
		},
		{
			name: "cache_ignore_input_vars excludes inputs from hash",
			req: &dataproxy.UploadInputsRequest{
				Id: &dataproxy.UploadInputsRequest_RunId{
					RunId: &common.RunIdentifier{
						Org: "org", Project: "proj", Domain: "dom", Name: "run1",
					},
				},
				Task: &dataproxy.UploadInputsRequest_TaskSpec{TaskSpec: testTaskSpecWithIgnoredVars},
				Inputs: &task.Inputs{
					Literals: []*task.NamedLiteral{
						{Name: "x", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}}}}}}},
						{Name: "y", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 2}}}}}}},
					},
				},
			},
			wantErr: false,
			validateResult: func(t *testing.T, resp *connect.Response[dataproxy.UploadInputsResponse]) {
				assert.NotEmpty(t, resp.Msg.OffloadedInputData.InputsHash)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := setupMockDataStoreWithWriteProtobuf(t)
			mockProjectClient := projectMocks.NewProjectServiceClient(t)
			if !tt.wantErr {
				mockProjectClient.On("GetProject", mock.Anything, mock.Anything).Return(
					connect.NewResponse(&project.GetProjectResponse{}), nil)
			}
			svc := NewService(cfg, mockStore, nil, nil, nil, mockProjectClient, nil)

			req := &connect.Request[dataproxy.UploadInputsRequest]{
				Msg: tt.req,
			}

			resp, err := svc.UploadInputs(ctx, req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.errCode != 0 {
					assert.Equal(t, tt.errCode, connect.CodeOf(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateResult != nil {
					tt.validateResult(t, resp)
				}
			}
		})
	}
}

func setupMockDataStoreWithWriteProtobuf(t *testing.T) *storage.DataStore {
	mockComposedStore := storageMocks.NewComposedProtobufStore(t)

	mockComposedStore.On("GetBaseContainerFQN", mock.Anything).Return(storage.DataReference("s3://test-bucket")).Maybe()
	mockComposedStore.On("WriteProtobuf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	return &storage.DataStore{
		ComposedProtobufStore: mockComposedStore,
		ReferenceConstructor:  &simpleRefConstructor{},
	}
}

func TestGetActionData(t *testing.T) {
	ctx := context.Background()
	cfg := config.DataProxyConfig{}

	actionID := &common.ActionIdentifier{
		Name: "a0",
		Run: &common.RunIdentifier{
			Org: "org", Project: "proj", Domain: "dom", Name: "run1",
		},
	}

	storedInputs := &task.Inputs{
		Literals: []*task.NamedLiteral{
			{Name: "x", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_Integer{Integer: 1}}}}}}},
		},
	}
	storedOutputs := &task.Inputs{
		Literals: []*task.NamedLiteral{
			{Name: "o", Value: &core.Literal{Value: &core.Literal_Scalar{Scalar: &core.Scalar{Value: &core.Scalar_Primitive{Primitive: &core.Primitive{Value: &core.Primitive_StringValue{StringValue: "result"}}}}}}},
		},
	}

	tests := []struct {
		name             string
		inputsURI        string
		outputsURI       string
		runClientErr     error
		readInputsErr    error
		readOutputsErr   error
		wantErr          bool
		expectInputsLen  int
		expectOutputsLen int
	}{
		{
			name:             "success with both inputs and outputs",
			inputsURI:        "s3://test-bucket/inputs-dir",
			outputsURI:       "s3://test-bucket/outputs/outputs.pb",
			expectInputsLen:  1,
			expectOutputsLen: 1,
		},
		{
			name:             "success with only inputs",
			inputsURI:        "s3://test-bucket/inputs-dir",
			outputsURI:       "",
			expectInputsLen:  1,
			expectOutputsLen: 0,
		},
		{
			name:             "success with only outputs",
			inputsURI:        "",
			outputsURI:       "s3://test-bucket/outputs/outputs.pb",
			expectInputsLen:  0,
			expectOutputsLen: 1,
		},
		{
			name:             "success with neither inputs nor outputs",
			inputsURI:        "",
			outputsURI:       "",
			expectInputsLen:  0,
			expectOutputsLen: 0,
		},
		{
			name:         "RunService error propagates",
			runClientErr: connect.NewError(connect.CodeNotFound, assertErr("not found")),
			wantErr:      true,
		},
		{
			name:          "read inputs error propagates",
			inputsURI:     "s3://test-bucket/inputs-dir",
			readInputsErr: assertErr("read failed"),
			wantErr:       true,
		},
		{
			name:           "read outputs error propagates",
			outputsURI:     "s3://test-bucket/outputs/outputs.pb",
			readOutputsErr: assertErr("read failed"),
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runClient := workflowMocks.NewRunServiceClient(t)
			if tt.runClientErr != nil {
				runClient.EXPECT().GetActionDataURIs(mock.Anything, mock.Anything).Return(nil, tt.runClientErr)
			} else {
				runClient.EXPECT().GetActionDataURIs(mock.Anything, mock.Anything).Return(
					connect.NewResponse(&workflow.GetActionDataURIsResponse{
						InputsUri:  tt.inputsURI,
						OutputsUri: tt.outputsURI,
					}), nil)
			}

			mockComposedStore := storageMocks.NewComposedProtobufStore(t)

			if tt.inputsURI != "" {
				expectedInputRef := storage.DataReference(tt.inputsURI + "/inputs.pb")
				call := mockComposedStore.On("ReadProtobuf", mock.Anything, expectedInputRef, mock.Anything)
				if tt.readInputsErr != nil {
					call.Return(tt.readInputsErr).Maybe()
				} else {
					call.Run(func(args mock.Arguments) {
						msg := args.Get(2).(proto.Message)
						proto.Reset(msg)
						proto.Merge(msg, storedInputs)
					}).Return(nil).Maybe()
				}
			}

			if tt.outputsURI != "" {
				expectedOutputRef := storage.DataReference(tt.outputsURI)
				call := mockComposedStore.On("ReadProtobuf", mock.Anything, expectedOutputRef, mock.Anything)
				if tt.readOutputsErr != nil {
					call.Return(tt.readOutputsErr).Maybe()
				} else {
					call.Run(func(args mock.Arguments) {
						msg := args.Get(2).(proto.Message)
						proto.Reset(msg)
						proto.Merge(msg, storedOutputs)
					}).Return(nil).Maybe()
				}
			}

			ds := &storage.DataStore{
				ComposedProtobufStore: mockComposedStore,
				ReferenceConstructor:  &simpleRefConstructor{},
			}
			svc := NewService(cfg, ds, nil, nil, runClient, nil, nil)

			resp, err := svc.GetActionData(ctx, connect.NewRequest(&dataproxy.GetActionDataRequest{
				ActionId: actionID,
			}))

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Len(t, resp.Msg.GetInputs().GetLiterals(), tt.expectInputsLen)
			assert.Len(t, resp.Msg.GetOutputs().GetLiterals(), tt.expectOutputsLen)
			if tt.expectInputsLen > 0 {
				assert.Equal(t, "x", resp.Msg.GetInputs().GetLiterals()[0].GetName())
			}
			if tt.expectOutputsLen > 0 {
				assert.Equal(t, "o", resp.Msg.GetOutputs().GetLiterals()[0].GetName())
			}
		})
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

func setupMockDataStoreWithExistingFile(t *testing.T, contentMD5 string) *storage.DataStore {
	mockComposedStore := storageMocks.NewComposedProtobufStore(t)

	// Setup base container
	mockComposedStore.On("GetBaseContainerFQN", mock.Anything).Return(storage.DataReference("s3://test-bucket")).Maybe()

	// Setup Head to return file exists with given hash
	mockMetadata := storageMocks.NewMetadata(t)
	mockMetadata.On("Exists").Return(true)
	mockMetadata.On("ContentMD5").Return(contentMD5).Maybe()
	mockComposedStore.On("Head", mock.Anything, mock.Anything).Return(mockMetadata, nil)

	return &storage.DataStore{
		ComposedProtobufStore: mockComposedStore,
		ReferenceConstructor:  &simpleRefConstructor{},
	}
}

// mockLogStreamer implements logs.LogStreamer for tests.
type mockLogStreamer struct {
	mock.Mock
}

func (m *mockLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	args := m.Called(ctx, logContext, stream)
	return args.Error(0)
}

func newTailLogsTestClient(t *testing.T, svc *Service) dataproxyconnect.DataProxyServiceClient {
	path, handler := dataproxyconnect.NewDataProxyServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return dataproxyconnect.NewDataProxyServiceClient(http.DefaultClient, server.URL)
}

func TestTailLogs(t *testing.T) {
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org: "test-org", Project: "test-project", Domain: "test-domain", Name: "rtest12345",
		},
		Name: "a0",
	}

	logContext := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{PodName: "my-pod", Namespace: "ns"},
		},
	}

	t.Run("happy path streams a message", func(t *testing.T) {
		runClient := workflowMocks.NewRunServiceClient(t)
		runClient.EXPECT().GetActionLogContext(mock.Anything, mock.Anything).Return(
			connect.NewResponse(&workflow.GetActionLogContextResponse{
				LogContext: logContext,
				Cluster:    "c1",
			}), nil)

		streamer := &mockLogStreamer{}
		streamer.On("TailLogs", mock.Anything, logContext, mock.Anything).Run(func(args mock.Arguments) {
			stream := args.Get(2).(*connect.ServerStream[dataproxy.TailLogsResponse])
			_ = stream.Send(&dataproxy.TailLogsResponse{})
		}).Return(nil)

		svc := NewService(config.DataProxyConfig{}, nil, nil, nil, runClient, nil, streamer)
		client := newTailLogsTestClient(t, svc)

		stream, err := client.TailLogs(context.Background(), connect.NewRequest(&dataproxy.TailLogsRequest{
			ActionId: actionID,
			Attempt:  1,
		}))
		assert.NoError(t, err)

		assert.True(t, stream.Receive())
		assert.NotNil(t, stream.Msg())
		assert.False(t, stream.Receive())
		assert.NoError(t, stream.Err())

		streamer.AssertExpectations(t)
	})

	t.Run("GetActionLogContext error propagates", func(t *testing.T) {
		runClient := workflowMocks.NewRunServiceClient(t)
		runClient.EXPECT().GetActionLogContext(mock.Anything, mock.Anything).Return(
			nil, connect.NewError(connect.CodeNotFound, assertErr("action missing")))

		streamer := &mockLogStreamer{}
		svc := NewService(config.DataProxyConfig{}, nil, nil, nil, runClient, nil, streamer)
		client := newTailLogsTestClient(t, svc)

		stream, err := client.TailLogs(context.Background(), connect.NewRequest(&dataproxy.TailLogsRequest{
			ActionId: actionID,
			Attempt:  1,
		}))
		assert.NoError(t, err)
		assert.False(t, stream.Receive())
		assert.Error(t, stream.Err())
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(stream.Err()))

		streamer.AssertNotCalled(t, "TailLogs", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("nil log context returns NotFound", func(t *testing.T) {
		runClient := workflowMocks.NewRunServiceClient(t)
		runClient.EXPECT().GetActionLogContext(mock.Anything, mock.Anything).Return(
			connect.NewResponse(&workflow.GetActionLogContextResponse{
				LogContext: nil,
			}), nil)

		streamer := &mockLogStreamer{}
		svc := NewService(config.DataProxyConfig{}, nil, nil, nil, runClient, nil, streamer)
		client := newTailLogsTestClient(t, svc)

		stream, err := client.TailLogs(context.Background(), connect.NewRequest(&dataproxy.TailLogsRequest{
			ActionId: actionID,
			Attempt:  1,
		}))
		assert.NoError(t, err)
		assert.False(t, stream.Receive())
		assert.Error(t, stream.Err())
		assert.Equal(t, connect.CodeNotFound, connect.CodeOf(stream.Err()))

		streamer.AssertNotCalled(t, "TailLogs", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("streamer error propagates", func(t *testing.T) {
		runClient := workflowMocks.NewRunServiceClient(t)
		runClient.EXPECT().GetActionLogContext(mock.Anything, mock.Anything).Return(
			connect.NewResponse(&workflow.GetActionLogContextResponse{LogContext: logContext}), nil)

		streamer := &mockLogStreamer{}
		streamer.On("TailLogs", mock.Anything, logContext, mock.Anything).Return(
			connect.NewError(connect.CodeInternal, assertErr("streamer boom")))

		svc := NewService(config.DataProxyConfig{}, nil, nil, nil, runClient, nil, streamer)
		client := newTailLogsTestClient(t, svc)

		stream, err := client.TailLogs(context.Background(), connect.NewRequest(&dataproxy.TailLogsRequest{
			ActionId: actionID,
			Attempt:  1,
		}))
		assert.NoError(t, err)
		assert.False(t, stream.Receive())
		assert.Error(t, stream.Err())
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(stream.Err()))

		streamer.AssertExpectations(t)
	})

	t.Run("passes action_id and attempt to RunService", func(t *testing.T) {
		runClient := workflowMocks.NewRunServiceClient(t)
		runClient.EXPECT().GetActionLogContext(mock.Anything, mock.MatchedBy(func(r *connect.Request[workflow.GetActionLogContextRequest]) bool {
			return proto.Equal(r.Msg.GetActionId(), actionID) && r.Msg.GetAttempt() == 3
		})).Return(connect.NewResponse(&workflow.GetActionLogContextResponse{LogContext: logContext}), nil)

		streamer := &mockLogStreamer{}
		streamer.On("TailLogs", mock.Anything, logContext, mock.Anything).Return(nil)

		svc := NewService(config.DataProxyConfig{}, nil, nil, nil, runClient, nil, streamer)
		client := newTailLogsTestClient(t, svc)

		stream, err := client.TailLogs(context.Background(), connect.NewRequest(&dataproxy.TailLogsRequest{
			ActionId: actionID,
			Attempt:  3,
		}))
		assert.NoError(t, err)
		assert.False(t, stream.Receive())
		assert.NoError(t, stream.Err())
	})
}
