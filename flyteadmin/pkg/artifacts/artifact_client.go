package artifacts

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewArtifactConnection(_ context.Context, cfg *Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if opts == nil {
		// Initialize opts list to the potential number of options we will add. Initialization optimizes memory
		// allocation.
		opts = make([]grpc.DialOption, 0, 5)
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithInsecure())
	} else {
		tlsConfig := &tls.Config{} //nolint
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	return grpc.Dial(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), opts...)
}

func InitializeArtifactClient(ctx context.Context, cfg *Config, opts ...grpc.DialOption) artifact.ArtifactRegistryClient {
	if cfg == nil {
		logger.Warningf(ctx, "Artifact config is not set, skipping creation of client...")
		return nil
	}
	conn, err := NewArtifactConnection(ctx, cfg, opts...)
	if err != nil {
		logger.Panicf(ctx, "failed to initialize Artifact connection. Err: %s", err.Error())
		panic(err)
	}
	return artifact.NewArtifactRegistryClient(conn)
}
