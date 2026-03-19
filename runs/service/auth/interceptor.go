package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

// BlanketAuthorization is a gRPC unary interceptor that checks the authenticated identity has the "all" scope.
func BlanketAuthorization(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	resp interface{}, err error) {

	identityContext := IdentityContextFromContext(ctx)
	if identityContext == nil {
		return handler(ctx, req)
	}

	for _, scope := range identityContext.Scopes() {
		if scope == ScopeAll {
			return handler(ctx, req)
		}
	}

	logger.Debugf(ctx, "authenticated user doesn't have required scope")
	return nil, status.Errorf(codes.Unauthenticated, "authenticated user doesn't have required scope")
}

// GetAuthenticationCustomMetadataInterceptor produces a gRPC interceptor that translates a custom authorization
// header name to the standard "authorization" header for downstream interceptors.
func GetAuthenticationCustomMetadataInterceptor(cfg config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cfg.GrpcAuthorizationHeader != DefaultAuthorizationHeader {
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				existingHeader := md.Get(cfg.GrpcAuthorizationHeader)
				if len(existingHeader) > 0 {
					logger.Debugf(ctx, "Found existing metadata header %s", cfg.GrpcAuthorizationHeader)
					newAuthorizationMetadata := metadata.Pairs(DefaultAuthorizationHeader, existingHeader[0])
					joinedMetadata := metadata.Join(md, newAuthorizationMetadata)
					newCtx := metadata.NewIncomingContext(ctx, joinedMetadata)
					return handler(newCtx, req)
				}
			}
		}
		return handler(ctx, req)
	}
}

// GetAuthenticationInterceptor returns a function that validates incoming gRPC requests.
// It attempts to extract and validate an access token or ID token from the request metadata.
func GetAuthenticationInterceptor(cfg config.Config, resourceServer OAuth2ResourceServer, oidcProvider *oidc.Provider) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		logger.Debugf(ctx, "Running authentication gRPC interceptor")

		expectedAudience := GetPublicURL(ctx, nil, cfg).String()

		identityContext, accessTokenErr := GRPCGetIdentityFromAccessToken(ctx, expectedAudience, resourceServer)
		if accessTokenErr == nil {
			return identityContext.WithContext(ctx), nil
		}

		logger.Infof(ctx, "Failed to parse Access Token from context. Will attempt to find IDToken. Error: %v", accessTokenErr)

		identityContext, idTokenErr := GRPCGetIdentityFromIDToken(ctx, cfg.UserAuth.OpenID.ClientID, oidcProvider)
		if idTokenErr == nil {
			return identityContext.WithContext(ctx), nil
		}
		logger.Debugf(ctx, "Failed to parse ID Token from context. Error: %v", idTokenErr)

		if !cfg.DisableForGrpc {
			err := fmt.Errorf("[id token err: %w] | [access token err: %w]", idTokenErr, accessTokenErr)
			return ctx, status.Errorf(codes.Unauthenticated, "token parse error %s", err)
		}

		return ctx, nil
	}
}

// AuthenticationLoggingInterceptor logs information about the authenticated user for each gRPC request.
func AuthenticationLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	identityContext := IdentityContextFromContext(ctx)
	if identityContext != nil {
		var emailPlaceholder string
		if len(identityContext.UserInfo().GetEmail()) > 0 {
			emailPlaceholder = fmt.Sprintf(" (%s) ", identityContext.UserInfo().GetEmail())
		}
		logger.Debugf(ctx, "gRPC server info in logging interceptor [%s]%smethod [%s]\n", identityContext.UserID(), emailPlaceholder, info.FullMethod)
	}
	return handler(ctx, req)
}
