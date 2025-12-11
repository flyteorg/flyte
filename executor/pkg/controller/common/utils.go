package common

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/compiler/validators"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/config"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/executors"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/handler"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller/interfaces"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/pbhash"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	idlcore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	maxUniqueIDLength = 20
	MB                = 1024 * 1024 // 1 MB in bytes (1 MiB)
)

// GenerateUniqueID is the UniqueId of a node is unique within a given workflow execution.
// In order to achieve that we track the lineage of the node.
// To compute the uniqueID of a node, we use the uniqueID and retry attempt of the parent node
// For nodes in level 0, there is no parent, and parentInfo is nil
func GenerateUniqueID(parentInfo executors.ImmutableParentInfo, nodeID string) (string, error) {
	var parentUniqueID v1alpha1.NodeID
	var parentRetryAttempt string

	if parentInfo != nil {
		parentUniqueID = parentInfo.GetUniqueID()
		parentRetryAttempt = strconv.Itoa(int(parentInfo.CurrentAttempt()))
	}

	return encoding.FixedLengthUniqueIDForParts(maxUniqueIDLength, []string{parentUniqueID, parentRetryAttempt, nodeID})
}

// CreateParentInfo creates a unique parent id, the unique id of parent is dependent on the unique id and the current
// attempt of the grandparent to track the lineage.
func CreateParentInfo(grandParentInfo executors.ImmutableParentInfo, nodeID string, parentAttempt uint32, nodeIsDynamic bool) (executors.ImmutableParentInfo, error) {
	uniqueID, err := GenerateUniqueID(grandParentInfo, nodeID)
	if err != nil {
		return nil, err
	}
	if nodeIsDynamic || (grandParentInfo != nil && grandParentInfo.IsInDynamicChain()) {
		return executors.NewParentInfo(uniqueID, parentAttempt, true), nil
	}

	return executors.NewParentInfo(uniqueID, parentAttempt, false), nil
}

// ReadLargeLiteral reads the offloaded large literal needed by array node task
func ReadLargeLiteral(ctx context.Context, datastore *storage.DataStore,
	tobeRead *idlcore.Literal) error {
	if tobeRead.GetOffloadedMetadata() == nil {
		return fmt.Errorf("unsupported type for reading offloaded literal")
	}
	dataReference := tobeRead.GetOffloadedMetadata().GetUri()
	if dataReference == "" {
		return fmt.Errorf("uri is empty for offloaded literal")
	}
	// read the offloaded literal
	size := tobeRead.GetOffloadedMetadata().GetSizeBytes()
	if err := datastore.ReadProtobuf(ctx, storage.DataReference(dataReference), tobeRead); err != nil {
		logger.Errorf(ctx, "Failed to  read the offloaded literal at location [%s] with error [%s]", dataReference, err)
		return err
	}

	logger.Infof(ctx, "read offloaded literal at location [%s] with size [%s]", dataReference, size)
	return nil
}

// OffloadLargeLiteral offloads the large literal if meets the threshold conditions
func OffloadLargeLiteral(ctx context.Context, datastore *storage.DataStore, dataReference storage.DataReference,
	toBeOffloaded *idlcore.Literal, literalOffloadingConfig config.LiteralOffloadingConfig) error {
	literalSizeBytes := int64(proto.Size(toBeOffloaded))
	literalSizeMB := literalSizeBytes / MB
	// check if the literal is large
	if literalSizeMB >= literalOffloadingConfig.MaxSizeInMBForOffloading {
		return fmt.Errorf("literal size [%d] MB is larger than the max size [%d] MB for offloading", literalSizeMB, literalOffloadingConfig.MaxSizeInMBForOffloading)
	}
	if literalSizeMB < literalOffloadingConfig.MinSizeInMBForOffloading {
		logger.Debugf(ctx, "Literal size [%d] MB is smaller than the min size [%d] MB for offloading", literalSizeMB, literalOffloadingConfig.MinSizeInMBForOffloading)
		return nil
	}

	inferredType := validators.LiteralTypeForLiteral(toBeOffloaded)
	if inferredType == nil {
		return fmt.Errorf("failed to determine literal type for offloaded literal")
	}

	// offload the literal
	if err := datastore.WriteProtobuf(ctx, dataReference, storage.Options{}, toBeOffloaded); err != nil {
		logger.Errorf(ctx, "Failed to offload literal at location [%s] with error [%s]", dataReference, err)
		return err
	}

	if toBeOffloaded.GetHash() == "" {
		// compute the hash of the literal
		literalDigest, err := pbhash.ComputeHash(ctx, toBeOffloaded)
		if err != nil {
			logger.Errorf(ctx, "Failed to compute hash for offloaded literal with error [%s]", err)
			return err
		}
		// Set the hash or else respect what the user set in the literal
		toBeOffloaded.Hash = base64.RawURLEncoding.EncodeToString(literalDigest)
	}
	// update the literal with the offloaded URI, size and inferred type
	toBeOffloaded.Value = &idlcore.Literal_OffloadedMetadata{
		OffloadedMetadata: &idlcore.LiteralOffloadedMetadata{
			Uri:          dataReference.String(),
			SizeBytes:    uint64(literalSizeBytes), // #nosec G115
			InferredType: inferredType,
		},
	}
	logger.Infof(ctx, "Offloaded literal at location [%s] with size [%d] MB and inferred type [%s]", dataReference, literalSizeMB, inferredType)
	return nil
}

// CheckOffloadingCompat checks if the upstream and downstream nodes are compatible with the literal offloading feature and returns an error if not contained in phase info object
func CheckOffloadingCompat(ctx context.Context, nCtx interfaces.NodeExecutionContext, inputLiterals map[string]*idlcore.Literal, node v1.ExecutableNode, literalOffloadingConfig config.LiteralOffloadingConfig) *handler.PhaseInfo {
	consumesOffloadLiteral := false
	for _, val := range inputLiterals {
		if val != nil && val.GetOffloadedMetadata() != nil {
			consumesOffloadLiteral = true
			break
		}
	}
	if !consumesOffloadLiteral {
		return nil
	}
	var phaseInfo handler.PhaseInfo

	// Return early if the node is not of type NodeKindTask
	if node.GetKind() != v1.NodeKindTask {
		return nil
	}

	// Process NodeKindTask
	taskID := *node.GetTaskID()
	taskNode, err := nCtx.ExecutionContext().GetTask(taskID)
	if err != nil {
		phaseInfo = handler.PhaseInfoFailure(idlcore.ExecutionError_SYSTEM, "GetTaskIDFailure", err.Error(), nil)
		return &phaseInfo
	}
	runtimeData := taskNode.CoreTask().GetMetadata().GetRuntime()
	if !literalOffloadingConfig.IsSupportedSDKVersion(ctx, runtimeData.GetType().String(), runtimeData.GetVersion()) {
		if !literalOffloadingConfig.Enabled {
			errMsg := fmt.Sprintf("task [%s] is trying to consume offloaded literals but feature is not enabled", taskID)
			logger.Errorf(ctx, errMsg)
			phaseInfo = handler.PhaseInfoFailure(idlcore.ExecutionError_USER, "LiteralOffloadingDisabled", errMsg, nil)
			return &phaseInfo
		}
		leastSupportedVersion := literalOffloadingConfig.GetSupportedSDKVersion(runtimeData.GetType().String())
		errMsg := fmt.Sprintf("Literal offloading is not supported for this task as its registered with SDK version [%s] which is less than the least supported version [%s] for this feature", runtimeData.GetVersion(), leastSupportedVersion)
		logger.Errorf(ctx, errMsg)
		phaseInfo = handler.PhaseInfoFailure(idlcore.ExecutionError_USER, "LiteralOffloadingNotSupported", errMsg, nil)
		return &phaseInfo
	}

	return nil
}
