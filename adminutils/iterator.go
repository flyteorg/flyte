package adminutils

import (
	"context"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc"
)

type NamedEntityIDLister func(ctx context.Context, in *admin.NamedEntityIdentifierListRequest, opts ...grpc.CallOption) (*admin.NamedEntityIdentifierList, error)

type NamedEntityVisitor func(entities []*admin.NamedEntityIdentifier) error

type ListRequest struct {
	Project string
	Domain  string
	Filters string
}

func IterateThroughAllNamedEntities(ctx context.Context, lister NamedEntityIDLister, visitor NamedEntityVisitor, req ListRequest, opts ...grpc.CallOption) error {
	adminReq := &admin.NamedEntityIdentifierListRequest{
		Project: req.Project,
		Domain:  req.Domain,
		Filters: req.Filters,
		SortBy: &admin.Sort{
			Key:       "name",
			Direction: admin.Sort_ASCENDING,
		},
		Limit: uint32(GetConfig().BatchSize),
	}

	i := 0
	for i < GetConfig().MaxRecords {
		res, err := lister(ctx, adminReq, opts...)
		if err != nil {
			return err
		}
		if len(res.Entities) != 0 {
			if err := visitor(res.Entities); err != nil {
				return err
			}
		}
		i = i + len(res.Entities)
		if len(res.Entities) == 0 || res.Token == "" {
			break
		}
		adminReq.Token = res.Token
	}
	return nil
}

func GetAllNamedEntities(ctx context.Context, lister NamedEntityIDLister, req ListRequest, opts ...grpc.CallOption) ([]*admin.NamedEntityIdentifier, error) {
	var allEntities []*admin.NamedEntityIdentifier
	err := IterateThroughAllNamedEntities(ctx, lister, func(entities []*admin.NamedEntityIdentifier) error {
		allEntities = append(allEntities, entities...)
		return nil
	}, req)
	if err != nil {
		return nil, err
	}
	return allEntities, nil
}

func NamedEntityToProtoMessage(l []*admin.NamedEntityIdentifier) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}