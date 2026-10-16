// Package authorization provides opt-in Cedar authorization for Connect services.
package authorization

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	cedar "github.com/cedar-policy/cedar-go"
	"golang.org/x/net/http/httpguts"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Config is file-only. Claims must be validated and client-supplied copies stripped
// by a trusted authenticating proxy. Never expose this listener directly to clients.
type Config struct {
	Enabled              bool   `json:"enabled"`
	TrustForwardedClaims bool   `json:"trustForwardedClaims"`
	ClaimsHeader         string `json:"claimsHeader"`
	Issuer               string `json:"issuer"`
	PolicyFile           string `json:"policyFile"`
}

// Requirement describes one permission. All requirements for a request must pass.
type Requirement struct {
	Action   string
	Resource cedar.EntityUID
	Entities cedar.EntityMap
}

// Resolver must explicitly handle supported procedures and reject unknown ones.
type Resolver func(procedure string, message any) ([]Requirement, error)

// Interceptor evaluates an immutable policy snapshot for each inbound request.
type Interceptor struct {
	cfg      Config
	policies *cedar.PolicySet
	version  string
	resolve  Resolver
}

// New loads an immutable policy snapshot. Policy changes require a restart.
func New(cfg Config, resolve Resolver) (*Interceptor, error) {
	if !cfg.TrustForwardedClaims || cfg.Issuer == "" || cfg.PolicyFile == "" || resolve == nil {
		return nil, errors.New("cedar requires trusted forwarded claims, issuer, policyFile, and a resolver")
	}
	if !httpguts.ValidHeaderFieldName(cfg.ClaimsHeader) {
		return nil, errors.New("cedar requires a valid claimsHeader")
	}
	raw, err := os.ReadFile(cfg.PolicyFile) // #nosec G304 -- operator-configured policy file
	if err != nil {
		return nil, fmt.Errorf("read Cedar policies: %w", err)
	}
	policies, err := cedar.NewPolicySetFromBytes(cfg.PolicyFile, raw)
	if err != nil {
		return nil, fmt.Errorf("parse Cedar policies: %w", err)
	}
	return &Interceptor{cfg: cfg, policies: policies, version: fmt.Sprintf("%x", sha256.Sum256(raw)), resolve: resolve}, nil
}

type caller struct {
	principal     cedar.EntityUID
	entities      cedar.EntityMap
	header, token string
}
type callerKey struct{}

func unauthenticated() error {
	return connect.NewError(connect.CodeUnauthenticated, errors.New("missing or invalid trusted claims"))
}

func (i *Interceptor) authenticate(h http.Header) (caller, error) {
	values := h.Values(i.cfg.ClaimsHeader)
	if len(values) != 1 {
		return caller{}, unauthenticated()
	}
	token := values[0]
	if strings.EqualFold(i.cfg.ClaimsHeader, "Authorization") {
		parts := strings.Fields(token)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return caller{}, unauthenticated()
		}
		token = parts[1]
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return caller{}, unauthenticated()
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	var claims struct {
		Subject string   `json:"sub"`
		Issuer  string   `json:"iss"`
		Groups  []string `json:"groups"`
	}
	if err != nil || json.Unmarshal(payload, &claims) != nil || strings.TrimSpace(claims.Subject) == "" || claims.Issuer != i.cfg.Issuer {
		return caller{}, unauthenticated()
	}
	uid := cedar.NewEntityUID("Flyte::Principal", cedar.String(claims.Subject))
	entities := cedar.EntityMap{}
	parents := make([]cedar.EntityUID, 0, len(claims.Groups))
	for _, group := range claims.Groups {
		if group == "" {
			return caller{}, unauthenticated()
		}
		id := cedar.NewEntityUID("Flyte::Group", cedar.String(group))
		parents = append(parents, id)
		entities[id] = cedar.Entity{UID: id}
	}
	entities[uid] = cedar.Entity{UID: uid, Parents: cedar.NewEntityUIDSet(parents...), Attributes: cedar.NewRecord(cedar.RecordMap{"issuer": cedar.String(claims.Issuer)})}
	return caller{principal: uid, entities: entities, header: i.cfg.ClaimsHeader, token: values[0]}, nil
}

func (i *Interceptor) authorize(ctx context.Context, c caller, procedure string, message any) error {
	checks, err := i.resolve(procedure, message)
	if err != nil {
		return err
	}
	if len(checks) == 0 {
		return Denied()
	}
	for _, check := range checks {
		entities := cedar.EntityMap{}
		for id, entity := range c.entities {
			entities[id] = entity
		}
		for id, entity := range check.Entities {
			entities[id] = entity
		}
		decision, diagnostic := cedar.Authorize(i.policies, entities, cedar.Request{
			Principal: c.principal, Action: cedar.NewEntityUID("Flyte::Action", cedar.String(check.Action)),
			Resource: check.Resource, Context: cedar.NewRecord(cedar.RecordMap{}),
		})
		allowed := decision == cedar.Allow && len(diagnostic.Errors) == 0
		logger.Infof(ctx, "Cedar decision principal=%q action=%q resource=%q policy_version=%s allowed=%t evaluation_errors=%d", c.principal.String(), check.Action, check.Resource.String(), i.version, allowed, len(diagnostic.Errors))
		if !allowed {
			return Denied()
		}
	}
	return nil
}

// Denied deliberately exposes neither policy text nor resource existence.
func Denied() error {
	return connect.NewError(connect.CodePermissionDenied, errors.New("authorization denied"))
}

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		c, err := i.authenticate(req.Header())
		if err != nil {
			return nil, err
		}
		if err := i.authorize(ctx, c, req.Spec().Procedure, req.Any()); err != nil {
			return nil, err
		}
		return next(context.WithValue(ctx, callerKey{}, c), req)
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		c, err := i.authenticate(conn.RequestHeader())
		if err != nil {
			return err
		}
		ctx = context.WithValue(ctx, callerKey{}, c)
		return next(ctx, &authorizedConn{StreamingHandlerConn: conn, check: func(message any) error {
			return i.authorize(ctx, c, conn.Spec().Procedure, message)
		}})
	}
}

type authorizedConn struct {
	connect.StreamingHandlerConn
	check func(any) error
}

// Check every received message, before the handler can use it (including bidi).
func (c *authorizedConn) Receive(message any) error {
	if err := c.StreamingHandlerConn.Receive(message); err != nil {
		return err
	}
	return c.check(message)
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// ForwardClaims is only for clients to trusted services in the same authentication
// boundary. The receiving service re-evaluates its own operation. Do not combine
// with a transport that replaces Authorization with broad service credentials.
func ForwardClaims() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			c, ok := ctx.Value(callerKey{}).(caller)
			if !ok {
				return nil, unauthenticated()
			}
			req.Header().Set(c.header, c.token)
			return next(ctx, req)
		}
	}
}
