package authorization

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
)

func claimsHeader(claims string) http.Header {
	h := http.Header{}
	h.Set("X-Verified-Claims", "header."+base64.RawURLEncoding.EncodeToString([]byte(claims))+".signature")
	return h
}

func TestScopeIsolation(t *testing.T) {
	for _, invalid := range []string{"", "prod", "org:prod:p", ":p", "prod:", "pro*:p", "prod: p"} {
		_, err := ParseScope(invalid)
		require.Error(t, err)
	}
	scope, err := ParseScope("production:payments")
	require.NoError(t, err)
	grants := []Scope{scope, {Domain: "development", Project: "*"}}
	ctx := WithScopes(context.Background(), grants)
	grants[0] = Scope{Domain: "*", Project: "*"}
	require.NoError(t, Check(ctx, "production", "payments"))
	require.NoError(t, Check(ctx, "development", "anything"))
	require.ErrorIs(t, Check(ctx, "production", "anything"), ErrDenied)
	require.ErrorIs(t, Check(ctx, "staging", "payments"), ErrDenied)
	got, _ := Scopes(ctx)
	got[0] = Scope{Domain: "*", Project: "*"}
	require.ErrorIs(t, Check(ctx, "staging", "payments"), ErrDenied)
	require.ErrorIs(t, Check(WithScopes(context.Background(), nil), "production", "payments"), ErrDenied)
	var guard Guard
	guard.EnableAuthorization()
	require.ErrorIs(t, guard.RequireAuthorizationContext(context.Background()), ErrDenied)
	require.NoError(t, guard.RequireAuthorizationContext(WithSystem(context.Background())))
	require.NoError(t, Check(WithSystem(ctx), "anything", "anything"))
}

func TestClaimsAndProcedureScopesStayCoupled(t *testing.T) {
	cfg := Config{ClaimsHeader: "X-Verified-Claims", Roles: []Role{
		{Name: "reader", Claim: "groups", Values: []string{"readers"}, Procedures: []string{"/test.Service/Get"}, Scopes: []string{"*:*"}},
		{Name: "writer", Claim: "groups", Values: []string{"writers"}, Procedures: []string{"/test.Service/Update"}, Scopes: []string{"development:payments"}},
	}}
	_, err := NewInterceptor(cfg, false)
	require.Error(t, err)
	i, err := NewInterceptor(cfg, true)
	require.NoError(t, err)
	h := claimsHeader(`{"sub":"alice","groups":["readers","writers"]}`)
	ctx, err := i.authorize(context.Background(), h, "/test.Service/Get")
	require.NoError(t, err)
	require.NoError(t, Check(ctx, "production", "anything"))
	ctx, err = i.authorize(context.Background(), h, "/test.Service/Update")
	require.NoError(t, err)
	require.NoError(t, Check(ctx, "development", "payments"))
	require.ErrorIs(t, Check(ctx, "production", "payments"), ErrDenied)
	for _, procedure := range []string{"/test.Service/Delete", "/test.Other/Get"} {
		_, err = i.authorize(context.Background(), h, procedure)
		require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	}
	for _, claims := range []string{`{"sub":"alice","groups":"unknown"}`, `{"sub":"alice","groups":123}`, `{"sub":"alice","groups":["writers",123]}`} {
		_, err = i.authorize(context.Background(), claimsHeader(claims), "/test.Service/Update")
		require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	}
	for _, h := range []http.Header{nil, claimsHeader(`{`), claimsHeader(`{"groups":"writers"}`), {"X-Verified-Claims": []string{"one", "two"}}} {
		_, err = i.authorize(context.Background(), h, "/test.Service/Update")
		require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	}
	_, err = i.authorize(context.Background(), claimsHeader(`{"sub":"alice","groups":"writers"}`), "/test.Service/Update")
	require.NoError(t, err)
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(permissionError(connect.NewError(connect.CodeInternal, fmt.Errorf("wrapped: %w", ErrDenied)))))
}

func TestSettingsAndSharedProjectScopes(t *testing.T) {
	ctx := WithScopes(context.Background(), []Scope{{Domain: "production", Project: "payments"}})
	for _, key := range []string{"v1:::", "v1::production:", "v1::production:payments"} {
		require.NoError(t, CheckSettings(ctx, key, false))
	}
	for _, key := range []string{"v1:::", "v1::production:", "v1::production:other", "v1::staging:payments"} {
		require.ErrorIs(t, CheckSettings(ctx, key, true), ErrDenied)
	}
	require.NoError(t, CheckSettings(ctx, "v1::production:payments", true))
	require.ErrorIs(t, CheckSettings(ctx, "v1::staging:payments", false), ErrDenied)
	require.NoError(t, CheckProjectRead(ctx, "payments"))
	require.ErrorIs(t, Check(ctx, "*", "payments"), ErrDenied)
	require.ErrorIs(t, CheckTaskSpec(ctx, "digest"), ErrDenied)
	require.NoError(t, CheckTaskSpec(WithTaskSpec(ctx, "digest"), "digest"))
	require.ErrorIs(t, CheckTaskSpec(WithTaskSpec(ctx, "digest"), "other"), ErrDenied)
}

func TestInvalidPolicyAndServicePatterns(t *testing.T) {
	cfg := Config{ClaimsHeader: "X-Verified-Claims", Roles: []Role{{Name: "reader", Claim: "groups", Values: []string{"reader"}, Procedures: []string{"/test.Service/*"}, Scopes: []string{"development:*"}}}}
	i, err := NewInterceptor(cfg, true)
	require.NoError(t, err)
	_, err = i.authorize(context.Background(), claimsHeader(`{"sub":"a","groups":"reader"}`), "/test.Service/Get")
	require.NoError(t, err)
	_, err = i.authorize(context.Background(), claimsHeader(`{"sub":"a","groups":"reader"}`), "/test.ServiceOther/Get")
	require.Error(t, err)
	for _, pattern := range []string{"Get", "/test.*/Get", "/test.Service/", "/test.Service/Get*"} {
		cfg.Roles[0].Procedures = []string{pattern}
		_, err = NewInterceptor(cfg, true)
		require.Error(t, err)
	}
}

// Exercise the real Connect transport to verify unary and streaming requests both
// receive method-specific context, including error translation from repositories.
type scopedRunHandler struct {
	workflowconnect.UnimplementedRunServiceHandler
}

func (scopedRunHandler) ListRuns(ctx context.Context, _ *connect.Request[workflow.ListRunsRequest]) (*connect.Response[workflow.ListRunsResponse], error) {
	if err := Check(ctx, "prod", "payments"); err != nil {
		return nil, fmt.Errorf("repository: %w", err)
	}
	return connect.NewResponse(&workflow.ListRunsResponse{}), nil
}
func (scopedRunHandler) WatchRuns(ctx context.Context, _ *connect.Request[workflow.WatchRunsRequest], stream *connect.ServerStream[workflow.WatchRunsResponse]) error {
	if err := Check(ctx, "prod", "payments"); err != nil {
		return err
	}
	return stream.Send(&workflow.WatchRunsResponse{})
}
func TestConnectUnaryAndStreaming(t *testing.T) {
	cfg := Config{ClaimsHeader: "X-Verified-Claims", Roles: []Role{{Name: "reader", Claim: "groups", Values: []string{"reader"}, Procedures: []string{"/flyteidl2.workflow.RunService/*"}, Scopes: []string{"prod:payments"}}}}
	i, err := NewInterceptor(cfg, true)
	require.NoError(t, err)
	path, handler := workflowconnect.NewRunServiceHandler(scopedRunHandler{}, connect.WithInterceptors(i))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	defer server.Close()
	client := workflowconnect.NewRunServiceClient(server.Client(), server.URL)
	_, err = client.ListRuns(context.Background(), connect.NewRequest(&workflow.ListRunsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	request := connect.NewRequest(&workflow.WatchRunsRequest{})
	request.Header().Set("X-Verified-Claims", claimsHeader(`{"sub":"alice","groups":"reader"}`).Get("X-Verified-Claims"))
	stream, err := client.WatchRuns(context.Background(), request)
	require.NoError(t, err)
	require.True(t, stream.Receive())
	require.NotNil(t, stream.Msg())
	require.False(t, stream.Receive())
	require.NoError(t, stream.Err())
	require.NoError(t, stream.Close())
	cfg.Roles[0].Scopes = []string{"dev:payments"}
	denied, err := NewInterceptor(cfg, true)
	require.NoError(t, err)
	path, handler = workflowconnect.NewRunServiceHandler(scopedRunHandler{}, connect.WithInterceptors(denied))
	mux2 := http.NewServeMux()
	mux2.Handle(path, handler)
	server2 := httptest.NewServer(mux2)
	defer server2.Close()
	client2 := workflowconnect.NewRunServiceClient(server2.Client(), server2.URL)
	unary := connect.NewRequest(&workflow.ListRunsRequest{})
	unary.Header().Set("X-Verified-Claims", request.Header().Get("X-Verified-Claims"))
	_, err = client2.ListRuns(context.Background(), unary)
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	stream, err = client2.WatchRuns(context.Background(), request)
	if err == nil {
		require.False(t, stream.Receive())
		err = stream.Err()
		_ = stream.Close()
	}
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

func TestGuardResourceChecksRequireContext(t *testing.T) {
	var guard Guard
	checks := map[string]func(context.Context) error{
		"resource":  func(ctx context.Context) error { return guard.Check(ctx, "production", "payments") },
		"project":   func(ctx context.Context) error { return guard.CheckProjectRead(ctx, "payments") },
		"settings":  func(ctx context.Context) error { return guard.CheckSettings(ctx, "v1::production:payments", true) },
		"task spec": func(ctx context.Context) error { return guard.CheckTaskSpec(ctx, "digest") },
	}
	for name, check := range checks {
		t.Run(name+" disabled", func(t *testing.T) {
			require.NoError(t, check(context.Background()))
		})
	}
	guard.EnableAuthorization()
	allowed := WithScopes(context.Background(), []Scope{{Domain: "production", Project: "payments"}})
	allowed = WithTaskSpec(allowed, "digest")
	denied := WithScopes(context.Background(), []Scope{{Domain: "development", Project: "other"}})
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			require.ErrorIs(t, check(context.Background()), ErrDenied)
			require.NoError(t, check(allowed))
			require.ErrorIs(t, check(denied), ErrDenied)
			require.NoError(t, check(WithSystem(context.Background())))
		})
	}
}

func TestForwardClaimsRequiresAuthenticatedContext(t *testing.T) {
	cfg := Config{ClaimsHeader: "X-Verified-Claims", Roles: []Role{{Name: "reader", Claim: "groups", Values: []string{"reader"}, Procedures: []string{"*"}, Scopes: []string{"prod:payments"}}}}
	i, err := NewInterceptor(cfg, true)
	require.NoError(t, err)
	headers := claimsHeader(`{"sub":"alice","groups":"reader"}`)
	headers.Set("Authorization", "Bearer unselected-token")
	ctx, err := i.authorize(context.Background(), headers, "/test.Service/Get")
	require.NoError(t, err)
	calls := 0
	forward := ForwardClaims().WrapUnary(func(_ context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		calls++
		require.Equal(t, headers.Get("X-Verified-Claims"), req.Header().Get("X-Verified-Claims"))
		require.Empty(t, req.Header().Get("Authorization"))
		return nil, nil
	})
	_, err = forward(ctx, connect.NewRequest(&workflow.ListRunsRequest{}))
	require.NoError(t, err)
	require.Equal(t, 1, calls)
	_, err = forward(WithScopes(context.Background(), []Scope{{Domain: "*", Project: "*"}}), connect.NewRequest(&workflow.ListRunsRequest{}))
	require.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	require.Equal(t, 1, calls)
}
