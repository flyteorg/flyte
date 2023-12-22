package util

import "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

type NamedEntityLike interface {
	GetProject() string
	GetDomain() string
	GetName() string
	GetOrg() string
}

func GetIdentifierScope(id NamedEntityLike) *admin.NamedEntityIdentifier {
	return &admin.NamedEntityIdentifier{
		Project: id.GetProject(),
		Domain:  id.GetDomain(),
		Name:    id.GetName(),
		Org:     id.GetOrg(),
	}
}
