package impl

import (
	"context"

	"github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// EnableAuthorization makes missing request context fail closed. Call only during
// construction; trusted background workers must explicitly use WithSystem.
func EnableAuthorization(repo any) { repo.(interface{ EnableAuthorization() }).EnableAuthorization() }

// authorizationGuard combines context validation, resource checks, and SQL filtering.
type authorizationGuard struct{ authorization.Guard }

// authorizationFilter validates the context before constructing a predicate.
// It is independent of client-supplied filters. Always combine
// it with AND; each complete grant is ORed, preserving domain/project pairing.
func (g *authorizationGuard) authorizationFilter(ctx context.Context, projects bool) (interfaces.Filter, error) {
	if err := g.RequireAuthorizationContext(ctx); err != nil {
		return nil, err
	}
	scopes, restricted := authorization.Scopes(ctx)
	if !restricted {
		return nil, nil
	}
	var result interfaces.Filter = NewRawFilter("FALSE")
	for _, scope := range scopes {
		var grant interfaces.Filter = NewRawFilter("TRUE")
		if !projects && scope.Domain != "*" {
			grant = grant.And(NewEqualFilter("domain", scope.Domain))
		}
		column := "project"
		if projects {
			column = "identifier"
		}
		if scope.Project != "*" {
			grant = grant.And(NewEqualFilter(column, scope.Project))
		}
		result = result.Or(grant)
	}
	return result, nil
}
func (g *authorizationGuard) andAuthorization(filter interfaces.Filter, ctx context.Context, projects bool) (interfaces.Filter, error) {
	auth, err := g.authorizationFilter(ctx, projects)
	if err != nil {
		return nil, err
	}
	if auth == nil {
		return filter, nil
	}
	if filter == nil {
		return auth, nil
	}
	return filter.And(auth), nil
}
