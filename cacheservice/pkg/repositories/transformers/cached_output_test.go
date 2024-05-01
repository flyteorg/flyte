package transformers

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestCreateCachedOutputModel(t *testing.T) {
	ctx := context.Background()

	sampleKey := "test-key"
	sampleURI := "test-uri"
	sampleOutputLiteral, err := coreutils.MakeLiteralMap(map[string]interface{}{"c": 3})
	assert.NoError(t, err)
	sampleOutputLiteralBytes, err := proto.Marshal(sampleOutputLiteral)
	assert.NoError(t, err)

	sampleKeyMapValues := map[string]string{"a": "b"}
	sampleKeyMap := &cacheservice.KeyMapMetadata{
		Values: sampleKeyMapValues,
	}
	sampleKeyMapBytes, err := proto.Marshal(sampleKeyMap)
	assert.NoError(t, err)

	validOutputLiterals := &cacheservice.CachedOutput{
		Output: &cacheservice.CachedOutput_OutputLiterals{
			OutputLiterals: sampleOutputLiteral,
		},
		Metadata: &cacheservice.Metadata{
			SourceIdentifier: &core.Identifier{},
		},
	}

	validOutputKeyMap := &cacheservice.CachedOutput{
		Output: &cacheservice.CachedOutput_OutputUri{
			OutputUri: sampleURI,
		},
		Metadata: &cacheservice.Metadata{
			SourceIdentifier: &core.Identifier{},
			KeyMap:           sampleKeyMap,
		},
	}

	tests := []struct {
		name           string
		key            string
		cachedOutput   *cacheservice.CachedOutput
		expectError    bool
		expectedOutput *models.CachedOutput
	}{
		{
			name:         "valid output literals",
			key:          sampleKey,
			cachedOutput: validOutputLiterals,
			expectError:  false,
			expectedOutput: &models.CachedOutput{
				BaseModel: models.BaseModel{
					ID: sampleKey,
				},
				OutputURI:          "",
				OutputLiteral:      sampleOutputLiteralBytes,
				Identifier:         models.Identifier{},
				SerializedMetadata: []byte{},
			},
		},
		{
			name:         "valid output key map",
			key:          sampleKey,
			cachedOutput: validOutputKeyMap,
			expectError:  false,
			expectedOutput: &models.CachedOutput{
				BaseModel: models.BaseModel{
					ID: sampleKey,
				},
				OutputURI:          sampleURI,
				OutputLiteral:      nil,
				Identifier:         models.Identifier{},
				SerializedMetadata: sampleKeyMapBytes,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := CreateCachedOutputModel(ctx, tc.key, tc.cachedOutput)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.EqualValues(t, tc.expectedOutput, result)
			}
		})
	}
}

func TestFromCachedOutputModel(t *testing.T) {
	ctx := context.Background()

	createdAt := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		cachedOutputModel *models.CachedOutput
		expectError       bool
		errorCode         codes.Code
	}{
		{
			name: "valid output literal",
			cachedOutputModel: &models.CachedOutput{
				OutputLiteral: []byte{},
				Identifier:    models.Identifier{},
				BaseModel: models.BaseModel{
					CreatedAt: createdAt,
				},
			},
			expectError: false,
		},
		{
			name: "valid output URI",
			cachedOutputModel: &models.CachedOutput{
				OutputURI:  "some-uri",
				Identifier: models.Identifier{},
				BaseModel: models.BaseModel{
					CreatedAt: createdAt,
				},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := FromCachedOutputModel(ctx, tc.cachedOutputModel)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.EqualValues(t, timestamppb.New(createdAt), result.Metadata.CreatedAt)
			}
		})
	}
}
