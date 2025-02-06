// This is its own package to resolve a circular dependency issues between auth -> common -> auth
package isolation

import (
	"context"

	"github.com/flyteorg/flyte/flytestdlib/contextutils"
)

const (
	ContextKeyIsolationContext = contextutils.Key("isolation_context")
)

// An IsolationContext provides context around how to isolate or filter resources for a particular API request.
type IsolationContext struct {
	resourceScopes []ResourceScope
}

func NewIsolationContext(resourceScopes []ResourceScope) IsolationContext {
	return IsolationContext{resourceScopes: resourceScopes}
}

// GetResourceScopes gets the scope of resources that the client can access. These are typically Flyte projects as well
// as domains within projects.
func (c IsolationContext) GetResourceScopes() []ResourceScope {
	return c.resourceScopes
}

// ResourceScope is a hierarchical representation of what scope of resources a user has access to. Project -> Domain.
// Empty strings are considered wildcard access.
type ResourceScope struct {
	Project string
	Domain  string
}

// TargetResourceScopeDepth represents the depth/scope of an individual resource. Sometimes only project level scope is
// applicable to a resource while the user has domain level scope within a project. In such cases, a user's resource
// scope may need to be truncated to match the depth of the target resource scope during filtering operations.
type TargetResourceScopeDepth = int

const (
	ProjectTargetResourceScopeDepth = 0 // The resource can only be filtered at a project level
	DomainTargetResourceScopeDepth  = 1 // THe resource can be filtered at both a project and domain level
)

// WithContext adds the isolation context to the go context
func (c IsolationContext) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextKeyIsolationContext, c)
}

// IsolationContextFromContext extracts the isolation context from a go context
func IsolationContextFromContext(ctx context.Context) IsolationContext {
	existing := ctx.Value(ContextKeyIsolationContext)
	if existing != nil {
		return existing.(IsolationContext)
	}

	return NewIsolationContext([]ResourceScope{})
}
