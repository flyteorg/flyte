package mocks

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetNamedEntityFunc func(ctx context.Context, request admin.NamedEntityGetRequest) (*admin.NamedEntity, error)
type UpdateNamedEntityFunc func(ctx context.Context, request admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error)
type ListNamedEntitiesFunc func(ctx context.Context, request admin.NamedEntityListRequest) (*admin.NamedEntityList, error)

type NamedEntityManager struct {
	GetNamedEntityFunc    GetNamedEntityFunc
	UpdateNamedEntityFunc UpdateNamedEntityFunc
	ListNamedEntitiesFunc ListNamedEntitiesFunc
}

func (m *NamedEntityManager) GetNamedEntity(ctx context.Context, request admin.NamedEntityGetRequest) (*admin.NamedEntity, error) {
	if m.GetNamedEntityFunc != nil {
		return m.GetNamedEntityFunc(ctx, request)
	}
	return nil, nil
}

func (m *NamedEntityManager) UpdateNamedEntity(ctx context.Context, request admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
	if m.UpdateNamedEntityFunc != nil {
		return m.UpdateNamedEntityFunc(ctx, request)
	}
	return nil, nil
}

func (m *NamedEntityManager) ListNamedEntities(ctx context.Context, request admin.NamedEntityListRequest) (*admin.NamedEntityList, error) {
	if m.ListNamedEntitiesFunc != nil {
		return m.ListNamedEntitiesFunc(ctx, request)
	}
	return nil, nil
}
