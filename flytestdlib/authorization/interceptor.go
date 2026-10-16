package authorization

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"connectrpc.com/connect"
)

// Config is deliberately file-only: policy is reviewed as a complete document.
type Config struct {
	Enabled bool `json:"enabled"`
	// ClaimsHeader is a JWT header validated and overwritten by a trusted proxy.
	// Authorization explicitly opts into the proxy-validated Bearer JWT path.
	ClaimsHeader string `json:"claimsHeader"`
	Roles        []Role `json:"roles"`
}

type Role struct {
	Name   string   `json:"name"`
	Claim  string   `json:"claim"`
	Values []string `json:"values"`
	// Procedures are exact Connect procedure paths, /package.Service/*, or *.
	Procedures []string `json:"procedures"`
	Scopes     []string `json:"scopes"`
}

type compiledRole struct {
	role   Role
	scopes []Scope
}
type Interceptor struct {
	header string
	roles  []compiledRole
}

var procedurePattern = regexp.MustCompile(`^/[A-Za-z_][A-Za-z0-9_.]*/([A-Za-z_][A-Za-z0-9_]*|\*)$`)

// NewInterceptor relies on upstream authentication, just like forwarded identity
// attribution. It never combines claims from different headers or tokens.
func NewInterceptor(cfg Config, trustForwarded bool) (*Interceptor, error) {
	if !trustForwarded {
		return nil, errors.New("authorization requires trustForwardedIdentityHeaders and a trusted authenticating proxy")
	}
	if cfg.ClaimsHeader == "" || strings.ContainsAny(cfg.ClaimsHeader, " \r\n\t") {
		return nil, errors.New("authorization.claimsHeader is required")
	}
	i := &Interceptor{header: cfg.ClaimsHeader}
	names := map[string]bool{}
	for _, role := range cfg.Roles {
		if role.Name == "" || names[role.Name] || role.Claim == "" || len(role.Values) == 0 || len(role.Procedures) == 0 || len(role.Scopes) == 0 {
			return nil, fmt.Errorf("invalid or duplicate authorization role %q", role.Name)
		}
		names[role.Name] = true
		for _, v := range role.Values {
			if v == "" {
				return nil, fmt.Errorf("empty claim value in role %q", role.Name)
			}
		}
		for _, p := range role.Procedures {
			if p != "*" && !procedurePattern.MatchString(p) {
				return nil, fmt.Errorf("invalid procedure %q", p)
			}
		}
		role.Values = append([]string(nil), role.Values...)
		role.Procedures = append([]string(nil), role.Procedures...)
		c := compiledRole{role: role}
		for _, value := range role.Scopes {
			s, err := ParseScope(value)
			if err != nil {
				return nil, err
			}
			c.scopes = append(c.scopes, s)
		}
		i.roles = append(i.roles, c)
	}
	return i, nil
}

func (i *Interceptor) authorize(ctx context.Context, h http.Header, procedure string) (context.Context, error) {
	if IsSystem(ctx) {
		return ctx, nil
	}
	values := h.Values(i.header)
	if len(values) != 1 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing trusted claims"))
	}
	token := values[0]
	if strings.EqualFold(i.header, "Authorization") {
		parts := strings.Fields(token)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("expected Bearer JWT"))
		}
		token = parts[1]
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid trusted claims JWT"))
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	var claims map[string]json.RawMessage
	if err != nil || json.Unmarshal(payload, &claims) != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid trusted claims"))
	}
	var subject string
	if json.Unmarshal(claims["sub"], &subject) != nil || strings.TrimSpace(subject) == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing trusted subject"))
	}
	var scopes []Scope
	for _, c := range i.roles {
		if !claimMatches(claims[c.role.Claim], c.role.Values) {
			continue
		}
		for _, p := range c.role.Procedures {
			if p == "*" || p == procedure || (strings.HasSuffix(p, "/*") && strings.HasPrefix(procedure, strings.TrimSuffix(p, "*"))) {
				scopes = append(scopes, c.scopes...)
				break
			}
		}
	}
	if len(scopes) == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, ErrDenied)
	}
	ctx = context.WithValue(ctx, forwardedClaimsKey{}, forwardedClaims{header: i.header, value: values[0]})
	return WithScopes(ctx, scopes), nil
}

func claimMatches(raw json.RawMessage, values []string) bool {
	var single string
	var claims []string
	if json.Unmarshal(raw, &single) == nil {
		claims = []string{single}
	} else if json.Unmarshal(raw, &claims) != nil {
		return false
	}
	for _, c := range claims {
		for _, v := range values {
			if c == v {
				return true
			}
		}
	}
	return false
}
func permissionError(err error) error {
	if errors.Is(err, ErrDenied) {
		return connect.NewError(connect.CodePermissionDenied, ErrDenied)
	}
	return err
}
func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		ctx, err := i.authorize(ctx, req.Header(), req.Spec().Procedure)
		if err != nil {
			return nil, err
		}
		resp, err := next(ctx, req)
		return resp, permissionError(err)
	}
}
func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		ctx, err := i.authorize(ctx, conn.RequestHeader(), conn.Spec().Procedure)
		if err != nil {
			return err
		}
		return permissionError(next(ctx, conn))
	}
}
func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

// ForwardClaims forwards only the claims selected by a successful inbound
// authorization decision. Install only on clients to a trusted downstream
// service configured to authenticate the same header. Never forward scopes.
func ForwardClaims() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			claims, ok := ctx.Value(forwardedClaimsKey{}).(forwardedClaims)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing authenticated caller claims"))
			}
			req.Header().Set(claims.header, claims.value)
			return next(ctx, req)
		}
	}
}

type forwardedClaimsKey struct{}
type forwardedClaims struct{ header, value string }
