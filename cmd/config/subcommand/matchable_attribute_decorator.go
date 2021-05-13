package subcommand

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

// MatchableAttributeDecorator defines a decorator for any matchable attribute target.
type MatchableAttributeDecorator interface {
	Decorate() *admin.MatchingAttributes
}

// MatchableAttributeUnDecorator defines a undecorator to get the target.
type MatchableAttributeUnDecorator interface {
	UnDecorate(*admin.MatchingAttributes)
}
