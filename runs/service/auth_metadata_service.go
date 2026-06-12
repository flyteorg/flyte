package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	"github.com/flyteorg/flyte/v2/runs/config"
)

const defaultOAuth2MetadataPath = ".well-known/oauth-authorization-server"

// AuthMetadataService implements the AuthMetadataServiceHandler interface.
type AuthMetadataService struct {
	authconnect.UnimplementedAuthMetadataServiceHandler
	dataplaneDomain string
	cfg             config.AuthMetadataConfig
}

// NewAuthMetadataService creates a new AuthMetadataService instance. When
// cfg.ExternalAuthServerBaseURL is set, GetOAuth2Metadata proxies that server's
// metadata. cfg.FlyteClient is advertised to SDKs via GetPublicClientConfig.
func NewAuthMetadataService(dataplaneDomain string, cfg config.AuthMetadataConfig) *AuthMetadataService {
	return &AuthMetadataService{
		dataplaneDomain: dataplaneDomain,
		cfg:             cfg,
	}
}

var _ authconnect.AuthMetadataServiceHandler = (*AuthMetadataService)(nil)

// GetPublicClientConfig returns the public (CLI/SDK) OAuth2 client settings.
func (s *AuthMetadataService) GetPublicClientConfig(
	ctx context.Context,
	req *connect.Request[auth.GetPublicClientConfigRequest],
) (*connect.Response[auth.GetPublicClientConfigResponse], error) {
	authMetadataKey := s.cfg.AuthorizationMetadataKey
	if authMetadataKey == "" {
		authMetadataKey = "authorization"
	}
	return connect.NewResponse(&auth.GetPublicClientConfigResponse{
		ClientId:                 s.cfg.FlyteClient.ClientID,
		RedirectUri:              s.cfg.FlyteClient.RedirectURI,
		Scopes:                   s.cfg.FlyteClient.Scopes,
		AuthorizationMetadataKey: authMetadataKey,
		Audience:                 s.cfg.FlyteClient.Audience,
		DataplaneDomain:          s.dataplaneDomain,
	}), nil
}

// GetOAuth2Metadata proxies the configured external authorization server's
// metadata document (RFC 8414 / OAuth2 Authorization Server Metadata). This lets
// flyte clients that discover auth at this deployment obtain tokens from the
// external IdP (e.g. Okta) directly, so a single token satisfies both this
// deployment and any upstream (ALB) JWT validation keyed to the same issuer.
//
// The external-fetch logic is adapted from flyteorg/flyte#6998.
func (s *AuthMetadataService) GetOAuth2Metadata(
	ctx context.Context,
	_ *connect.Request[auth.GetOAuth2MetadataRequest],
) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	if s.cfg.ExternalAuthServerBaseURL == "" {
		return nil, connect.NewError(connect.CodeUnimplemented,
			errors.New("oauth2 metadata is not configured; set runs.authMetadata.externalAuthServerBaseUrl"))
	}

	baseURL, err := url.Parse(s.cfg.ExternalAuthServerBaseURL)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("invalid external auth server base URL %q: %w", s.cfg.ExternalAuthServerBaseURL, err))
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("external auth server base URL must be absolute (include scheme and host): %q", s.cfg.ExternalAuthServerBaseURL))
	}

	// Issuer URLs conventionally do not end with a '/', but metadata URLs are
	// relative to them. Add a trailing '/' so ResolveReference behaves intuitively.
	baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + "/"

	metadataPath := s.cfg.ExternalMetadataURL
	if metadataPath == "" {
		metadataPath = defaultOAuth2MetadataPath
	}
	relURL, err := url.Parse(metadataPath)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("invalid external metadata path %q: %w", metadataPath, err))
	}
	// MetadataURL is expected to be a relative path resolved against BaseURL.
	// Reject absolute, scheme-relative, or root-relative paths so BaseURL cannot be bypassed.
	if relURL.IsAbs() || relURL.Host != "" || strings.HasPrefix(relURL.Path, "/") {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("external metadata path must be relative to externalAuthServerBaseUrl (no leading '/'), got %q", metadataPath))
	}
	externalMetadataURL := baseURL.ResolveReference(relURL)

	retryAttempts := s.cfg.RetryAttempts
	if retryAttempts <= 0 {
		retryAttempts = 5
	}
	retryDelay := s.cfg.RetryDelay
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := sendAndRetryHTTPRequest(ctx, client, externalMetadataURL.String(), retryAttempts, retryDelay)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to fetch OAuth2 metadata: %w", err))
	}
	defer func() { _ = response.Body.Close() }()

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read OAuth2 metadata response: %w", err))
	}

	resp := &auth.GetOAuth2MetadataResponse{}
	if err := unmarshalResp(response, raw, resp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unmarshal OAuth2 metadata: %w", err))
	}

	return connect.NewResponse(resp), nil
}

// OAuth2MetadataHTTPHandler serves the OAuth2 authorization-server metadata
// document at the RFC 8414 well-known path. OAuth2/OIDC discovery clients
// (flytectl, pyflyte) fetch this path directly rather than the Connect RPC.
func OAuth2MetadataHTTPHandler(svc *AuthMetadataService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		resp, err := svc.GetOAuth2Metadata(r.Context(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
		if err != nil {
			http.Error(w, err.Error(), connectCodeToHTTPStatus(err))
			return
		}
		// Marshal with proto3 JSON (camelCase) to match what flyteadmin and
		// other flyte clients expect from this endpoint.
		body, marshalErr := protojson.Marshal(resp.Msg)
		if marshalErr != nil {
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, writeErr := w.Write(body); writeErr != nil {
			logger.Warnf(r.Context(), "failed to write oauth2 metadata response: %v", writeErr)
		}
	})
}

func connectCodeToHTTPStatus(err error) int {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return http.StatusInternalServerError
	}
	switch connectErr.Code() {
	case connect.CodeUnimplemented:
		return http.StatusNotImplemented
	case connect.CodeInvalidArgument:
		return http.StatusBadRequest
	case connect.CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// unmarshalResp unmarshals a JSON response body into a protobuf message. It uses
// protojson.Unmarshal which accepts both the camelCase form used by proto3 JSON
// serialization and the snake_case form matching proto field names. This matters
// because external authorization servers (including flyteadmin) emit camelCase
// keys while the Go proto struct tags are snake_case.
//
// Adapted from flyteorg/flyte#6998.
func unmarshalResp(r *http.Response, body []byte, v proto.Message) error {
	// DiscardUnknown: real authorization servers (e.g. Okta) return many metadata
	// fields beyond those modelled here (introspection_endpoint, claims_supported,
	// …); without this, unmarshal fails on the first unknown field.
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(body, v); err == nil {
		return nil
	} else {
		ct := r.Header.Get("Content-Type")
		mediaType, _, parseErr := mime.ParseMediaType(ct)
		if parseErr == nil && mediaType == "application/json" {
			return fmt.Errorf("got Content-Type = application/json, but could not unmarshal as JSON: %w", err)
		}
		return fmt.Errorf("expected Content-Type = application/json, got %q: %w", ct, err)
	}
}

// sendAndRetryHTTPRequest fetches the given URL with retry logic.
//
// Adapted from flyteorg/flyte#6998.
func sendAndRetryHTTPRequest(ctx context.Context, client *http.Client, targetURL string, retryAttempts int, retryDelay time.Duration) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response
	for i := 0; i < retryAttempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warnf(ctx, "Failed to fetch %s (attempt %d/%d): %v", targetURL, i+1, retryAttempts, err)
			continue
		}

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp, nil
		}

		_ = resp.Body.Close()
		lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, targetURL)
		lastResp = resp
		logger.Warnf(ctx, "Unexpected status code from %s (attempt %d/%d): %d", targetURL, i+1, retryAttempts, resp.StatusCode)
	}

	if lastResp != nil {
		return nil, fmt.Errorf("failed to get oauth metadata with status code %v: %w", lastResp.StatusCode, lastErr)
	}
	return nil, fmt.Errorf("all %d attempts failed for %s: %w", retryAttempts, targetURL, lastErr)
}
