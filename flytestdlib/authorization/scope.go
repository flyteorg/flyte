// Package authorization carries request authorization into the repository layer.
package authorization

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var ErrDenied = errors.New("authorization denied")

// Scope is a domain:project grant. A whole segment may be *, never a partial glob.
// Project metadata is global across domains; its mutations require domain=*.
type Scope struct{ Domain, Project string }

func ParseScope(value string) (Scope, error) {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return Scope{}, fmt.Errorf("invalid scope %q: expected domain:project", value)
	}
	for _, p := range parts {
		if p == "" || strings.TrimSpace(p) != p || (p != "*" && strings.ContainsAny(p, "*? \t\r\n")) {
			return Scope{}, fmt.Errorf("invalid scope %q", value)
		}
	}
	return Scope{Domain: parts[0], Project: parts[1]}, nil
}

func matches(grant, resource string) bool { return grant == "*" || grant == resource }
func (s Scope) Covers(domain, project string) bool {
	return matches(s.Domain, domain) && matches(s.Project, project)
}

type contextKey struct{}
type access struct {
	scopes []Scope
	system bool
}

// WithScopes copies the grants so callers cannot mutate a request's authority.
// An explicitly empty scope set denies all resources.
func WithScopes(ctx context.Context, scopes []Scope) context.Context {
	return context.WithValue(ctx, contextKey{}, access{scopes: append([]Scope(nil), scopes...)})
}

// WithSystem is only for trusted in-process workers. Never derive it from a header.
func WithSystem(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, access{system: true})
}
func IsSystem(ctx context.Context) bool { a, _ := ctx.Value(contextKey{}).(access); return a.system }

// Scopes returns a copy of the effective RPC grants. false means legacy/unrestricted
// or a trusted system context. Guard.RequireAuthorizationContext still rejects missing request context
// when authorization is enabled on the repository.
func Scopes(ctx context.Context) ([]Scope, bool) {
	a, ok := ctx.Value(contextKey{}).(access)
	if !ok || a.system {
		return nil, false
	}
	return append([]Scope(nil), a.scopes...), true
}

func Check(ctx context.Context, domain, project string) error {
	scopes, restricted := Scopes(ctx)
	if !restricted {
		return nil
	}
	for _, s := range scopes {
		if s.Covers(domain, project) {
			return nil
		}
	}
	return ErrDenied
}

// Guard is configured once, before publishing a repository to other goroutines.
type Guard struct{ required bool }

func (g *Guard) EnableAuthorization() { g.required = true }
func (g *Guard) RequireAuthorizationContext(ctx context.Context) error {
	if _, ok := ctx.Value(contextKey{}).(access); !ok && g.required {
		return ErrDenied
	}
	return nil
}

// Check validates both the required authorization context and resource scope.
func (g *Guard) Check(ctx context.Context, domain, project string) error {
	if err := g.RequireAuthorizationContext(ctx); err != nil {
		return err
	}
	return Check(ctx, domain, project)
}

// CheckProjectRead validates context and access to shared project metadata.
func (g *Guard) CheckProjectRead(ctx context.Context, project string) error {
	if err := g.RequireAuthorizationContext(ctx); err != nil {
		return err
	}
	return CheckProjectRead(ctx, project)
}

// CheckSettings validates context and access to the settings scope.
func (g *Guard) CheckSettings(ctx context.Context, key string, write bool) error {
	if err := g.RequireAuthorizationContext(ctx); err != nil {
		return err
	}
	return CheckSettings(ctx, key, write)
}

// CheckTaskSpec validates context and access to an authorized action's digest.
func (g *Guard) CheckTaskSpec(ctx context.Context, digest string) error {
	if err := g.RequireAuthorizationContext(ctx); err != nil {
		return err
	}
	return CheckTaskSpec(ctx, digest)
}

// CheckProjectRead projects domain-specific access onto the shared project row.
func CheckProjectRead(ctx context.Context, project string) error {
	scopes, restricted := Scopes(ctx)
	if !restricted {
		return nil
	}
	for _, s := range scopes {
		if matches(s.Project, project) {
			return nil
		}
	}
	return ErrDenied
}

// CheckSettings permits reading ancestors used to resolve inherited settings.
// Mutations must cover the entire target scope.
func CheckSettings(ctx context.Context, key string, write bool) error {
	parts := strings.Split(key, ":")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "" {
		return ErrDenied
	}
	domain, project := parts[2], parts[3]
	if domain == "" {
		domain = "*"
	}
	if project == "" {
		project = "*"
	}
	if write {
		return Check(ctx, domain, project)
	}
	scopes, restricted := Scopes(ctx)
	if !restricted {
		return nil
	}
	for _, s := range scopes {
		if (domain == "*" || matches(s.Domain, domain)) && (project == "*" || matches(s.Project, project)) {
			return nil
		}
	}
	return ErrDenied
}

// A digest is readable only after resolving it from an authorized action. The
// content-addressed task_specs table itself has no domain/project columns.
type digestKey struct{}

func WithTaskSpec(ctx context.Context, digest string) context.Context {
	return context.WithValue(ctx, digestKey{}, digest)
}
func CheckTaskSpec(ctx context.Context, digest string) error {
	if _, restricted := Scopes(ctx); !restricted {
		return nil
	}
	if allowed, ok := ctx.Value(digestKey{}).(string); ok && allowed == digest {
		return nil
	}
	return ErrDenied
}
