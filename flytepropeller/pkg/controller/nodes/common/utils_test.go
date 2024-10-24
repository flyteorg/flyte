package common

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"

	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	executorMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/pbhash"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type ParentInfo struct {
	uniqueID         string
	attempt          uint32
	isInDynamicChain bool
}

func (p ParentInfo) GetUniqueID() v1alpha1.NodeID {
	return p.uniqueID
}

func (p ParentInfo) CurrentAttempt() uint32 {
	return p.attempt
}

func (p ParentInfo) IsInDynamicChain() bool {
	return p.isInDynamicChain
}

func TestGenerateUniqueID(t *testing.T) {
	p := ParentInfo{
		uniqueID: "u1",
		attempt:  uint32(2),
	}
	uniqueID, err := GenerateUniqueID(p, "n1")
	assert.Nil(t, err)
	assert.Equal(t, "u1-2-n1", uniqueID)
}

func TestGenerateUniqueIDLong(t *testing.T) {
	p := ParentInfo{
		uniqueID: "u1111111111323131231asdadasd",
		attempt:  uint32(2),
	}
	uniqueID, err := GenerateUniqueID(p, "n1")
	assert.Nil(t, err)
	assert.Equal(t, "fraibbty", uniqueID)
}

func TestCreateParentInfo(t *testing.T) {
	gp := ParentInfo{
		uniqueID:         "u1",
		attempt:          uint32(2),
		isInDynamicChain: true,
	}
	parent, err := CreateParentInfo(gp, "n1", uint32(1), false)
	assert.Nil(t, err)
	assert.Equal(t, "u1-2-n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
	assert.True(t, parent.IsInDynamicChain())
}

func TestCreateParentInfoNil(t *testing.T) {
	parent, err := CreateParentInfo(nil, "n1", uint32(1), true)
	assert.Nil(t, err)
	assert.Equal(t, "n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
	assert.True(t, parent.IsInDynamicChain())
}

func init() {
	labeled.SetMetricKeys(contextutils.AppNameKey)
}

func TestReadLargeLiteral(t *testing.T) {
	t.Run("read successful", func(t *testing.T) {
		ctx := context.Background()
		datastore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		dataReference := storage.DataReference("foo/bar")
		toBeRead := &idlCore.Literal{
			Value: &idlCore.Literal_Scalar{
				Scalar: &idlCore.Scalar{
					Value: &idlCore.Scalar_Primitive{
						Primitive: &idlCore.Primitive{
							Value: &idlCore.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
		}
		err := datastore.WriteProtobuf(ctx, dataReference, storage.Options{}, toBeRead)
		assert.Nil(t, err)

		offloadedLiteral := &idlCore.Literal{
			Value: &idlCore.Literal_OffloadedMetadata{
				OffloadedMetadata: &idlCore.LiteralOffloadedMetadata{
					Uri: dataReference.String(),
				},
			},
		}
		err = ReadLargeLiteral(ctx, datastore, offloadedLiteral)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), offloadedLiteral.GetScalar().GetPrimitive().GetInteger())
	})
}

func TestOffloadLargeLiteral(t *testing.T) {
	t.Run("offload successful with valid size", func(t *testing.T) {
		ctx := context.Background()
		datastore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		dataReference := storage.DataReference("foo/bar")
		toBeOffloaded := &idlCore.Literal{
			Value: &idlCore.Literal_Scalar{
				Scalar: &idlCore.Scalar{
					Value: &idlCore.Scalar_Primitive{
						Primitive: &idlCore.Primitive{
							Value: &idlCore.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
		}
		expectedLiteralDigest, err := pbhash.ComputeHash(ctx, toBeOffloaded)
		assert.Nil(t, err)
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			MinSizeInMBForOffloading: 0,
			MaxSizeInMBForOffloading: 1,
		}
		inferredType := &idlCore.LiteralType{
			Type: &idlCore.LiteralType_CollectionType{
				CollectionType: &idlCore.LiteralType{
					Type: &idlCore.LiteralType_Simple{
						Simple: idlCore.SimpleType_INTEGER,
					},
				},
			},
		}
		err = OffloadLargeLiteral(ctx, datastore, dataReference, toBeOffloaded, inferredType, literalOffloadingConfig)
		assert.NoError(t, err)
		assert.Equal(t, "foo/bar", toBeOffloaded.GetOffloadedMetadata().GetUri())
		assert.Equal(t, uint64(6), toBeOffloaded.GetOffloadedMetadata().GetSizeBytes())
		assert.Equal(t, inferredType.GetSimple(), toBeOffloaded.GetOffloadedMetadata().InferredType.GetSimple())
		assert.Equal(t, base64.RawURLEncoding.EncodeToString(expectedLiteralDigest), toBeOffloaded.Hash)
	})

	t.Run("offload successful with valid size and hash passed in", func(t *testing.T) {
		ctx := context.Background()
		datastore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		dataReference := storage.DataReference("foo/bar")
		toBeOffloaded := &idlCore.Literal{
			Value: &idlCore.Literal_Scalar{
				Scalar: &idlCore.Scalar{
					Value: &idlCore.Scalar_Primitive{
						Primitive: &idlCore.Primitive{
							Value: &idlCore.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
			Hash: "hash",
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			MinSizeInMBForOffloading: 0,
			MaxSizeInMBForOffloading: 1,
		}
		inferredType := &idlCore.LiteralType{
			Type: &idlCore.LiteralType_CollectionType{
				CollectionType: &idlCore.LiteralType{
					Type: &idlCore.LiteralType_Simple{
						Simple: idlCore.SimpleType_INTEGER,
					},
				},
			},
		}
		err := OffloadLargeLiteral(ctx, datastore, dataReference, toBeOffloaded, inferredType, literalOffloadingConfig)
		assert.NoError(t, err)
		assert.Equal(t, "hash", toBeOffloaded.Hash)
	})

	t.Run("offload fails with size larger than max", func(t *testing.T) {
		ctx := context.Background()
		datastore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		dataReference := storage.DataReference("foo/bar")
		toBeOffloaded := &idlCore.Literal{
			Value: &idlCore.Literal_Scalar{
				Scalar: &idlCore.Scalar{
					Value: &idlCore.Scalar_Primitive{
						Primitive: &idlCore.Primitive{
							Value: &idlCore.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			MinSizeInMBForOffloading: 0,
			MaxSizeInMBForOffloading: 0,
		}
		inferredType := &idlCore.LiteralType{
			Type: &idlCore.LiteralType_CollectionType{
				CollectionType: &idlCore.LiteralType{
					Type: &idlCore.LiteralType_Simple{
						Simple: idlCore.SimpleType_INTEGER,
					},
				},
			},
		}
		err := OffloadLargeLiteral(ctx, datastore, dataReference, toBeOffloaded, inferredType, literalOffloadingConfig)
		assert.Error(t, err)
	})

	t.Run("offload not attempted with size smaller than min", func(t *testing.T) {
		ctx := context.Background()
		datastore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		dataReference := storage.DataReference("foo/bar")
		toBeOffloaded := &idlCore.Literal{
			Value: &idlCore.Literal_Scalar{
				Scalar: &idlCore.Scalar{
					Value: &idlCore.Scalar_Primitive{
						Primitive: &idlCore.Primitive{
							Value: &idlCore.Primitive_Integer{
								Integer: 1,
							},
						},
					},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			MinSizeInMBForOffloading: 2,
			MaxSizeInMBForOffloading: 3,
		}
		inferredType := &idlCore.LiteralType{
			Type: &idlCore.LiteralType_CollectionType{
				CollectionType: &idlCore.LiteralType{
					Type: &idlCore.LiteralType_Simple{
						Simple: idlCore.SimpleType_INTEGER,
					},
				},
			},
		}
		err := OffloadLargeLiteral(ctx, datastore, dataReference, toBeOffloaded, inferredType, literalOffloadingConfig)
		assert.NoError(t, err)
		assert.Nil(t, toBeOffloaded.GetOffloadedMetadata())
	})
}

func TestCheckOffloadingCompat(t *testing.T) {
	ctx := context.Background()
	nCtx := &nodeMocks.NodeExecutionContext{}
	executionContext := &executorMocks.ExecutionContext{}
	executableTask := &mocks.ExecutableTask{}
	node := &mocks.ExecutableNode{}
	node.OnGetKind().Return(v1alpha1.NodeKindTask)
	nCtx.OnExecutionContext().Return(executionContext)
	executionContext.OnGetTask("task1").Return(executableTask, nil)
	executableTask.OnCoreTask().Return(&idlCore.TaskTemplate{
		Metadata: &idlCore.TaskMetadata{
			Runtime: &idlCore.RuntimeMetadata{
				Type:    idlCore.RuntimeMetadata_FLYTE_SDK,
				Version: "0.16.0",
			},
		},
	})
	taskID := "task1"
	node.OnGetTaskID().Return(&taskID)
	t.Run("supported version success", func(t *testing.T) {
		inputLiterals := map[string]*idlCore.Literal{
			"foo": {
				Value: &idlCore.Literal_OffloadedMetadata{
					OffloadedMetadata: &idlCore.LiteralOffloadedMetadata{},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				idlCore.RuntimeMetadata_FLYTE_SDK.String(): "0.16.0",
			},
			Enabled: true,
		}
		phaseInfo := CheckOffloadingCompat(ctx, nCtx, inputLiterals, node, literalOffloadingConfig)
		assert.Nil(t, phaseInfo)
	})
	t.Run("unsupported version", func(t *testing.T) {
		inputLiterals := map[string]*idlCore.Literal{
			"foo": {
				Value: &idlCore.Literal_OffloadedMetadata{
					OffloadedMetadata: &idlCore.LiteralOffloadedMetadata{},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			SupportedSDKVersions: map[string]string{
				idlCore.RuntimeMetadata_FLYTE_SDK.String(): "0.17.0",
			},
			Enabled: true,
		}
		phaseInfo := CheckOffloadingCompat(ctx, nCtx, inputLiterals, node, literalOffloadingConfig)
		assert.NotNil(t, phaseInfo)
		assert.Equal(t, idlCore.ExecutionError_USER, phaseInfo.GetErr().GetKind())
		assert.Equal(t, "LiteralOffloadingNotSupported", phaseInfo.GetErr().GetCode())
	})
	t.Run("offloading config disabled with offloaded data", func(t *testing.T) {
		inputLiterals := map[string]*idlCore.Literal{
			"foo": {
				Value: &idlCore.Literal_OffloadedMetadata{
					OffloadedMetadata: &idlCore.LiteralOffloadedMetadata{},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			Enabled: false,
		}
		phaseInfo := CheckOffloadingCompat(ctx, nCtx, inputLiterals, node, literalOffloadingConfig)
		assert.NotNil(t, phaseInfo)
		assert.Equal(t, idlCore.ExecutionError_USER, phaseInfo.GetErr().GetKind())
		assert.Equal(t, "LiteralOffloadingDisabled", phaseInfo.GetErr().GetCode())
	})
	t.Run("offloading config enabled with no offloaded data", func(t *testing.T) {
		inputLiterals := map[string]*idlCore.Literal{
			"foo": {
				Value: &idlCore.Literal_Scalar{
					Scalar: &idlCore.Scalar{},
				},
			},
		}
		literalOffloadingConfig := config.LiteralOffloadingConfig{
			Enabled: true,
		}
		phaseInfo := CheckOffloadingCompat(ctx, nCtx, inputLiterals, node, literalOffloadingConfig)
		assert.Nil(t, phaseInfo)
	})
}
