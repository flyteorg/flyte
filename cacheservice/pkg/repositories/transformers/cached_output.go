package transformers

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func CreateCachedOutputModel(ctx context.Context, key string, cachedOutput *cacheservice.CachedOutput) (*models.CachedOutput, error) {
	var outputLiteralBytes []byte
	var err error
	if cachedOutput.GetOutputLiterals() != nil {
		outputLiteralBytes, err = proto.Marshal(cachedOutput.GetOutputLiterals())
		if err != nil {
			logger.Debugf(ctx, "Failed to marshal output literal with error %v", err)
			return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal output literal with error %v", err)
		}
	}

	var keyMap *cacheservice.KeyMapMetadata
	if cachedOutput.GetMetadata() == nil || cachedOutput.GetMetadata().GetKeyMap() == nil {
		keyMap = &cacheservice.KeyMapMetadata{}
	} else {
		keyMap = cachedOutput.GetMetadata().GetKeyMap()
	}

	serializedMetadata, err := proto.Marshal(keyMap)
	if err != nil {
		logger.Debugf(ctx, "Failed to marshal output metadata with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to marshal output metadata with error %v", err)
	}

	return &models.CachedOutput{
		BaseModel: models.BaseModel{
			ID: key,
		},
		OutputURI:     cachedOutput.GetOutputUri(),
		OutputLiteral: outputLiteralBytes,
		Identifier: models.Identifier{
			ResourceType: cachedOutput.GetMetadata().GetSourceIdentifier().GetResourceType(),
			Project:      cachedOutput.GetMetadata().GetSourceIdentifier().GetProject(),
			Domain:       cachedOutput.GetMetadata().GetSourceIdentifier().GetDomain(),
			Name:         cachedOutput.GetMetadata().GetSourceIdentifier().GetName(),
			Version:      cachedOutput.GetMetadata().GetSourceIdentifier().GetVersion(),
			Org:          cachedOutput.GetMetadata().GetSourceIdentifier().GetOrg(),
		},
		SerializedMetadata: serializedMetadata,
	}, nil
}

func FromCachedOutputModel(ctx context.Context, cachedOutputModel *models.CachedOutput) (*cacheservice.CachedOutput, error) {
	var keyMapMetadata cacheservice.KeyMapMetadata
	err := proto.Unmarshal(cachedOutputModel.SerializedMetadata, &keyMapMetadata)
	if err != nil {
		logger.Debugf(ctx, "Failed to unmarshal output metadata with error %v", err)
		return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to unmarshal output metadata with error %v", err)
	}

	if cachedOutputModel.OutputLiteral != nil {
		outputLiteral := &core.LiteralMap{}
		err := proto.Unmarshal(cachedOutputModel.OutputLiteral, outputLiteral)
		if err != nil {
			logger.Debugf(ctx, "Failed to unmarshal output literal with error %v", err)
			return nil, errors.NewCacheServiceErrorf(codes.Internal, "Failed to unmarshal output literal with error %v", err)
		}

		return &cacheservice.CachedOutput{
			Output: &cacheservice.CachedOutput_OutputLiterals{OutputLiterals: outputLiteral},
			Metadata: &cacheservice.Metadata{
				SourceIdentifier: &core.Identifier{
					ResourceType: cachedOutputModel.Identifier.ResourceType,
					Project:      cachedOutputModel.Identifier.Project,
					Domain:       cachedOutputModel.Identifier.Domain,
					Name:         cachedOutputModel.Identifier.Name,
					Version:      cachedOutputModel.Identifier.Version,
					Org:          cachedOutputModel.Identifier.Org,
				},
				KeyMap: &keyMapMetadata,
			},
		}, nil
	}

	return &cacheservice.CachedOutput{
		Output: &cacheservice.CachedOutput_OutputUri{OutputUri: cachedOutputModel.OutputURI},
		Metadata: &cacheservice.Metadata{
			SourceIdentifier: &core.Identifier{
				ResourceType: cachedOutputModel.Identifier.ResourceType,
				Project:      cachedOutputModel.Identifier.Project,
				Domain:       cachedOutputModel.Identifier.Domain,
				Name:         cachedOutputModel.Identifier.Name,
				Version:      cachedOutputModel.Identifier.Version,
				Org:          cachedOutputModel.Identifier.Org,
			},
			KeyMap: &keyMapMetadata,
		},
	}, nil
}
