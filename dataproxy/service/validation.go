package service

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

// validateUploadRequest performs validation on the upload request.
func validateUploadRequest(ctx context.Context, req *dataproxy.CreateUploadLocationRequest, cfg config.DataProxyConfig) error {
	if len(req.FilenameRoot) == 0 && len(req.ContentMd5) == 0 {
		return fmt.Errorf("either filename_root or content_md5 must be provided")
	}

	// Validate expires_in against platform maximum
	if req.ExpiresIn != nil {
		if !req.ExpiresIn.IsValid() {
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
