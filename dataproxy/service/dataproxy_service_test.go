package service

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	flyteconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
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
			service := NewService(cfg, mockStore)

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
		name             string
		req              *dataproxy.CreateUploadLocationRequest
		existingFileMD5  string // Empty means file doesn't exist
		expectErr          bool
		errContains      string
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
			expectErr:         false,
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
			expectErr:         true,
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
			expectErr:         false,
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
			expectErr:         true,
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
			expectErr:         false,
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
			expectErr:         false,
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

			service := NewService(cfg, mockStore)
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
			expectedPath: "s3://test-bucket/uploads/test-org/test-project/test-domain/test-root/test-file.txt",
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
			expectedPath: "s3://test-bucket/uploads/test-org/test-project/test-domain/test-root/test-file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := setupMockDataStore(t)
			service := NewService(cfg, mockStore)

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
