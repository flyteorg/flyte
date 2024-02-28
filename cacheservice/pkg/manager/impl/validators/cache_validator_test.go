package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var validKey = "key"

func TestValidateCacheKey(t *testing.T) {
	testCases := []struct {
		name        string
		key         string
		expectError bool
	}{
		{
			name:        "empty key",
			key:         "",
			expectError: true,
		},
		{
			name:        "valid key",
			key:         "key",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCacheKey(tc.key)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputUri(t *testing.T) {
	testCases := []struct {
		name        string
		outputURI   string
		expectError bool
	}{
		{
			name:        "empty output uri",
			outputURI:   "",
			expectError: true,
		},
		{
			name:        "valid output uri",
			outputURI:   "uri",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOutputURI(tc.outputURI)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputLiteral(t *testing.T) {
	testCases := []struct {
		name        string
		literal     *core.LiteralMap
		expectError bool
	}{
		{
			name:        "nil literal",
			literal:     nil,
			expectError: true,
		},
		{
			name:        "valid literal",
			literal:     &core.LiteralMap{},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOutputLiteral(tc.literal)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutputMeta(t *testing.T) {
	testCases := []struct {
		name        string
		metadata    *cacheservice.Metadata
		expectError bool
	}{
		{
			name:        "nil metadata",
			metadata:    nil,
			expectError: true,
		},
		{
			name:        "nil source identifier",
			metadata:    &cacheservice.Metadata{},
			expectError: true,
		},
		{
			name:        "valid metadata",
			metadata:    &cacheservice.Metadata{SourceIdentifier: &core.Identifier{}},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOutputMetadata(tc.metadata)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOutput(t *testing.T) {
	validSourceIdentifier := &core.Identifier{}
	validMetadata := &cacheservice.Metadata{SourceIdentifier: validSourceIdentifier}
	validOutput := &cacheservice.CachedOutput_OutputUri{OutputUri: "uri"}

	testCases := []struct {
		name        string
		output      *cacheservice.CachedOutput
		expectError bool
	}{
		{
			name:        "nil output",
			output:      nil,
			expectError: true,
		},
		{
			name:        "nil output field",
			output:      &cacheservice.CachedOutput{},
			expectError: true,
		},
		{
			name:        "invalid metadata",
			output:      &cacheservice.CachedOutput{Output: &cacheservice.CachedOutput_OutputUri{OutputUri: "uri"}, Metadata: &cacheservice.Metadata{}},
			expectError: true,
		},
		{
			name:        "invalid output uri",
			output:      &cacheservice.CachedOutput{Output: &cacheservice.CachedOutput_OutputUri{OutputUri: ""}, Metadata: validMetadata},
			expectError: true,
		},
		{
			name:        "invalid output literal",
			output:      &cacheservice.CachedOutput{Output: &cacheservice.CachedOutput_OutputLiterals{OutputLiterals: nil}, Metadata: validMetadata},
			expectError: true,
		},
		{
			name:        "valid output",
			output:      &cacheservice.CachedOutput{Output: validOutput, Metadata: validMetadata},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOutput(tc.output)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateOwnerID(t *testing.T) {
	testCases := []struct {
		name        string
		OwnerID     string
		expectError bool
	}{
		{
			name:        "empty owner id",
			OwnerID:     "",
			expectError: true,
		},
		{
			name:        "valid owner id",
			OwnerID:     "owner",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateOwnerID(tc.OwnerID)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePutCacheRequest(t *testing.T) {
	validSourceIdentifier := &core.Identifier{}
	validMetadata := &cacheservice.Metadata{SourceIdentifier: validSourceIdentifier}
	validOutput := &cacheservice.CachedOutput_OutputUri{OutputUri: "uri"}

	invalidKey := ""

	testCases := []struct {
		name        string
		req         *cacheservice.PutCacheRequest
		expectError bool
	}{
		{
			name:        "nil request",
			req:         nil,
			expectError: true,
		},
		{
			name:        "nil output",
			req:         &cacheservice.PutCacheRequest{},
			expectError: true,
		},
		{
			name:        "invalid key",
			req:         &cacheservice.PutCacheRequest{Key: invalidKey, Output: &cacheservice.CachedOutput{}},
			expectError: true,
		},
		{
			name:        "invalid output",
			req:         &cacheservice.PutCacheRequest{Key: validKey, Output: &cacheservice.CachedOutput{}},
			expectError: true,
		},
		{
			name:        "valid request",
			req:         &cacheservice.PutCacheRequest{Key: validKey, Output: &cacheservice.CachedOutput{Output: validOutput, Metadata: validMetadata}},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePutCacheRequest(tc.req)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGetCacheRequest(t *testing.T) {
	invalidKey := ""

	testCases := []struct {
		name        string
		req         *cacheservice.GetCacheRequest
		expectError bool
	}{
		{
			name:        "nil request",
			req:         nil,
			expectError: true,
		},
		{
			name:        "invalid key",
			req:         &cacheservice.GetCacheRequest{Key: invalidKey},
			expectError: true,
		},
		{
			name:        "valid request",
			req:         &cacheservice.GetCacheRequest{Key: validKey},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGetCacheRequest(tc.req)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetOrExtendReservationRequest(t *testing.T) {
	invalidKey := ""

	validOwnerID := "owner"
	invalidOwnerID := ""

	testCases := []struct {
		name        string
		req         *cacheservice.GetOrExtendReservationRequest
		expectError bool
	}{
		{
			name:        "nil request",
			req:         nil,
			expectError: true,
		},
		{
			name:        "invalid key",
			req:         &cacheservice.GetOrExtendReservationRequest{Key: invalidKey},
			expectError: true,
		},
		{
			name:        "invalid owner id",
			req:         &cacheservice.GetOrExtendReservationRequest{Key: validKey, OwnerId: invalidOwnerID},
			expectError: true,
		},
		{
			name:        "valid request",
			req:         &cacheservice.GetOrExtendReservationRequest{Key: validKey, OwnerId: validOwnerID},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGetOrExtendReservationRequest(tc.req)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReleaseReservationRequest(t *testing.T) {
	invalidKey := ""

	validOwnerID := "owner"
	invalidOwnerID := ""

	testCases := []struct {
		name        string
		req         *cacheservice.ReleaseReservationRequest
		expectError bool
	}{
		{
			name:        "nil request",
			req:         nil,
			expectError: true,
		},
		{
			name:        "invalid key",
			req:         &cacheservice.ReleaseReservationRequest{Key: invalidKey},
			expectError: true,
		},
		{
			name:        "invalid owner id",
			req:         &cacheservice.ReleaseReservationRequest{Key: validKey, OwnerId: invalidOwnerID},
			expectError: true,
		},
		{
			name:        "valid request",
			req:         &cacheservice.ReleaseReservationRequest{Key: validKey, OwnerId: validOwnerID},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateReleaseReservationRequest(tc.req)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
