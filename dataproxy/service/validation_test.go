package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	flyteconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

func TestValidateUploadRequest(t *testing.T) {
	ctx := context.Background()
	cfg := config.DataProxyConfig{
		Upload: config.DataProxyUploadConfig{
			MaxExpiresIn: flyteconfig.Duration{Duration: 1 * time.Hour},
			MaxSize:      resource.MustParse("100Mi"),
		},
	}

	tests := []struct {
		name    string
		req     *dataproxy.CreateUploadLocationRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request with filename_root",
			req: &dataproxy.CreateUploadLocationRequest{
				FilenameRoot: "test-root",
				ExpiresIn:    durationpb.New(30 * time.Minute),
			},
			wantErr: false,
		},
		{
			name: "valid request with content_md5",
			req: &dataproxy.CreateUploadLocationRequest{
				ContentMd5: []byte("test-hash"),
				ExpiresIn:  durationpb.New(30 * time.Minute),
			},
			wantErr: false,
		},
		{
			name: "missing both filename_root and content_md5",
			req: &dataproxy.CreateUploadLocationRequest{
				ExpiresIn: durationpb.New(30 * time.Minute),
			},
			wantErr: true,
			errMsg:  "either filename_root or content_md5 must be provided",
		},
		{
			name: "expires_in exceeds max",
			req: &dataproxy.CreateUploadLocationRequest{
				FilenameRoot: "test-root",
				ExpiresIn:    durationpb.New(2 * time.Hour),
			},
			wantErr: true,
			errMsg:  "exceeds maximum allowed duration",
		},
		{
			name: "content_length exceeds max",
			req: &dataproxy.CreateUploadLocationRequest{
				FilenameRoot:  "test-root",
				ContentLength: 1024 * 1024 * 200,
			},
			wantErr: true,
			errMsg:  "exceeds maximum allowed size",
		},
		{
			name: "valid request without expires_in (will use default)",
			req: &dataproxy.CreateUploadLocationRequest{
				FilenameRoot: "test-root",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUploadRequest(ctx, tt.req, cfg)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
