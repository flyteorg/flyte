package cacheservice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
)

// GenerateCacheKey creates a cache key utilizing hashes of the resource's identifier, signature, and inputs.
func GenerateCacheKey(ctx context.Context, key catalog.Key) (string, error) {
	inputs := &core.LiteralMap{}
	if key.TypedInterface.Inputs != nil {
		retInputs, err := key.InputReader.Get(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to read inputs when trying to query cache service")
		}
		inputs = retInputs
	}

	identifierHash, err := catalog.HashIdentifierExceptVersion(ctx, key.Identifier)
	if err != nil {
		return "", err
	}

	signatureHash, err := datacatalog.GenerateInterfaceSignatureHash(ctx, key.TypedInterface)
	if err != nil {
		return "", err
	}

	inputsHash, err := catalog.HashLiteralMap(ctx, inputs, key.CacheIgnoreInputVars)
	if err != nil {
		return "", err
	}

	cacheKey := fmt.Sprintf("%s-%s-%s-%s", identifierHash, signatureHash, inputsHash, key.CacheVersion)
	return cacheKey, nil
}

// GenerateCacheMetadata creates cache metadata to identify the source of a cache entry
func GenerateCacheMetadata(key catalog.Key, metadata catalog.Metadata) *cacheservice.Metadata {
	if metadata.TaskExecutionIdentifier != nil {
		return &cacheservice.Metadata{
			SourceIdentifier: &key.Identifier,
			KeyMap: &cacheservice.KeyMapMetadata{
				Values: map[string]string{
					datacatalog.TaskVersionKey:     metadata.TaskExecutionIdentifier.TaskId.GetVersion(),
					datacatalog.ExecProjectKey:     metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetProject(),
					datacatalog.ExecDomainKey:      metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetDomain(),
					datacatalog.ExecNameKey:        metadata.TaskExecutionIdentifier.NodeExecutionId.GetExecutionId().GetName(),
					datacatalog.ExecNodeIDKey:      metadata.TaskExecutionIdentifier.NodeExecutionId.GetNodeId(),
					datacatalog.ExecTaskAttemptKey: strconv.Itoa(int(metadata.TaskExecutionIdentifier.GetRetryAttempt())),
					datacatalog.ExecOrgKey:         metadata.TaskExecutionIdentifier.GetNodeExecutionId().GetExecutionId().GetOrg(),
				},
			},
		}
	} else if metadata.NodeExecutionIdentifier != nil {
		return &cacheservice.Metadata{
			SourceIdentifier: &key.Identifier,
			KeyMap: &cacheservice.KeyMapMetadata{
				Values: map[string]string{
					datacatalog.ExecProjectKey: metadata.NodeExecutionIdentifier.GetExecutionId().GetProject(),
					datacatalog.ExecDomainKey:  metadata.NodeExecutionIdentifier.GetExecutionId().GetDomain(),
					datacatalog.ExecNameKey:    metadata.NodeExecutionIdentifier.GetExecutionId().GetName(),
					datacatalog.ExecNodeIDKey:  metadata.NodeExecutionIdentifier.GetNodeId(),
					datacatalog.ExecOrgKey:     metadata.NodeExecutionIdentifier.GetExecutionId().GetOrg(),
				},
			},
		}
	}

	return &cacheservice.Metadata{
		SourceIdentifier: &key.Identifier,
	}
}

func GetSourceFromMetadata(metadata *cacheservice.Metadata) (*core.TaskExecutionIdentifier, error) {
	if metadata == nil {
		return nil, errors.Errorf("Output does not have metadata")
	}
	if metadata.SourceIdentifier == nil {
		return nil, errors.Errorf("Output does not have source identifier")
	}

	val := datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecTaskAttemptKey, "0")
	attempt, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return nil, err
	}

	return &core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: metadata.SourceIdentifier.ResourceType,
			Project:      metadata.SourceIdentifier.Project,
			Domain:       metadata.SourceIdentifier.Domain,
			Name:         metadata.SourceIdentifier.Name,
			Version:      datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.TaskVersionKey, "unknown"),
			Org:          metadata.SourceIdentifier.Org,
		},
		RetryAttempt: uint32(attempt),
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecNodeIDKey, "unknown"),
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecProjectKey, metadata.SourceIdentifier.GetProject()),
				Domain:  datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecDomainKey, metadata.SourceIdentifier.GetDomain()),
				Name:    datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecNameKey, "unknown"),
				Org:     datacatalog.GetOrDefault(metadata.GetKeyMap().GetValues(), datacatalog.ExecOrgKey, metadata.SourceIdentifier.GetOrg()),
			},
		},
	}, nil
}

func GenerateCatalogMetadata(source *core.TaskExecutionIdentifier, metadata *cacheservice.Metadata) *core.CatalogMetadata {
	if metadata == nil {
		return nil
	}

	return &core.CatalogMetadata{
		DatasetId: metadata.SourceIdentifier,
		SourceExecution: &core.CatalogMetadata_SourceTaskExecution{
			SourceTaskExecution: source,
		},
	}
}
