package auth

import (
	"context"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
)

const (
	ContextKeyIsolationContext = contextutils.Key("isolation_context")
)

type IsolationContext struct {
	resourceScopes []ResourceScope
}

func NewIsolationContext(resourceScopes []ResourceScope) IsolationContext {
	return IsolationContext{resourceScopes: resourceScopes}
}

func (c IsolationContext) GetResourceScopes() []ResourceScope {
	return c.resourceScopes
}

// ResourceScope is a hierarchical representation of what scope of resources a user has access to. Project -> Domain.
// Empty strings are considered wildcard access.
type ResourceScope struct {
	Project string
	Domain  string
}

// TargetResourceScopeDepth represents the depth/scope of an individual resource. Sometimes only project level or domain level
// scope is applicable to a resource. In such cases, a user's resource scope may need to be truncated to match the depth
// of the target resource scope.
type TargetResourceScopeDepth = int

const (
	ProjectTargetResourceScopeDepth = 0
	DomainTargetResourceScopeDepth  = 1
)

func (c IsolationContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyIsolationContext, c)
}

func IsolationContextFromContext(ctx context.Context) IsolationContext {
	existing := ctx.Value(ContextKeyIsolationContext)
	if existing != nil {
		return existing.(IsolationContext)
	}

	return NewIsolationContext([]ResourceScope{})
}
