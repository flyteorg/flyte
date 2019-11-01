package adminservice

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func (m *AdminService) ListNamedEntities(
	ctx context.Context, request *admin.NamedEntityListRequest) (*admin.NamedEntityList, error) {
	// TODO(schottra): Implement this method.
	return nil, nil
}

func (m *AdminService) GetNamedEntity(
	ctx context.Context, request *admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	// TODO(schottra): Implement this method.
	return nil, nil
}

func (m *AdminService) UpdateNamedEntity(
	ctx context.Context, request *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
	// TODO(schottra): Implement this method.
	return nil, nil
}
