package cacheservice

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/catalog/datacatalog"
)

func TestGenerateCachedTaskKey(t *testing.T) {

	sampleInputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"a": 1, "b": 2})
	assert.NoError(t, err)
	mockInputReader := &mocks.InputReader{}
	mockInputReader.On("Get", mock.Anything).Return(sampleInputs, nil, nil)

	sampleInterface := core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"a": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
				"b": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}},
			},
		},
	}

	testCases := []struct {
		name        string
		key         catalog.Key
		expectedErr bool
		expectedKey string
	}{
		{
			name: "without ignore vars",
			key: catalog.Key{
				Identifier: core.Identifier{
					Project: "project_1",
					Domain:  "domain_1",
					Name:    "name_1",
					Version: "0",
					Org:     "org_1",
				},
				CacheVersion:   "1.0.0",
				TypedInterface: sampleInterface,
				InputReader:    mockInputReader,
			},
			expectedErr: false,
			expectedKey: "+UmrGhEwHv3FesdpA4gliBluF3FUXz4tshmuOlw1FSk=-gfjJTWtY-GKw-c0Pw-jwhb7XmCTEQeFc_pixCptPWTWg_hezuWfuZAwHJDwWQ-1.0.0",
		},
		{
			name: "different resource version produces same key",
			key: catalog.Key{
				Identifier: core.Identifier{
					Project: "project_1",
					Domain:  "domain_1",
					Name:    "name_1",
					Version: "1",
					Org:     "org_1",
				},
				CacheVersion:   "1.0.0",
				TypedInterface: sampleInterface,
				InputReader:    mockInputReader,
			},
			expectedErr: false,
			expectedKey: "+UmrGhEwHv3FesdpA4gliBluF3FUXz4tshmuOlw1FSk=-gfjJTWtY-GKw-c0Pw-jwhb7XmCTEQeFc_pixCptPWTWg_hezuWfuZAwHJDwWQ-1.0.0",
		},
		{
			name: "ignore inputs",
			key: catalog.Key{
				Identifier: core.Identifier{
					Project: "project_1",
					Domain:  "domain_1",
					Name:    "name_1",
					Version: "0",
					Org:     "org_1",
				},
				CacheVersion:         "1.0.0",
				TypedInterface:       sampleInterface,
				InputReader:          mockInputReader,
				CacheIgnoreInputVars: []string{"a"},
			},
			expectedErr: false,
			expectedKey: "+UmrGhEwHv3FesdpA4gliBluF3FUXz4tshmuOlw1FSk=-gfjJTWtY-GKw-c0Pw-MNjvcaUwW4DzBce6jJ30Pl-A_9guXnrww5_6mn_PVrA-1.0.0",
		},
		{
			name: "different cache version",
			key: catalog.Key{
				Identifier: core.Identifier{
					Project: "project_1",
					Domain:  "domain_1",
					Name:    "name_1",
					Version: "0",
					Org:     "org_1",
				},
				CacheVersion:   "1.0.1",
				TypedInterface: sampleInterface,
				InputReader:    mockInputReader,
			},
			expectedErr: false,
			expectedKey: "+UmrGhEwHv3FesdpA4gliBluF3FUXz4tshmuOlw1FSk=-gfjJTWtY-GKw-c0Pw-jwhb7XmCTEQeFc_pixCptPWTWg_hezuWfuZAwHJDwWQ-1.0.1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskKey, err := GenerateCacheKey(context.TODO(), tc.key)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedKey, taskKey)
			}
		})
	}
}

func TestGenerateCacheMetadata(t *testing.T) {

	tID := &core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Name:         "x",
			Project:      "project",
			Domain:       "development",
			Version:      "ver",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "wf",
				Project: "p1",
				Domain:  "d1",
			},
			NodeId: "n",
		},
		RetryAttempt: 1,
	}
	nID := &core.NodeExecutionIdentifier{
		NodeId: "n",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Name:    "wf",
			Project: "p1",
			Domain:  "d1",
			Org:     "org",
		},
	}

	testCases := []struct {
		name     string
		key      catalog.Key
		metadata catalog.Metadata
		expected map[string]string
	}{
		{
			name:     "task execution identifier",
			key:      catalog.Key{},
			metadata: catalog.Metadata{TaskExecutionIdentifier: tID},
			expected: map[string]string{
				datacatalog.ExecTaskAttemptKey: strconv.Itoa(int(tID.RetryAttempt)),
				datacatalog.ExecProjectKey:     tID.NodeExecutionId.ExecutionId.Project,
				datacatalog.ExecDomainKey:      tID.NodeExecutionId.ExecutionId.Domain,
				datacatalog.ExecNodeIDKey:      tID.NodeExecutionId.NodeId,
				datacatalog.ExecNameKey:        tID.NodeExecutionId.ExecutionId.Name,
				datacatalog.ExecOrgKey:         tID.NodeExecutionId.ExecutionId.Org,
				datacatalog.TaskVersionKey:     tID.TaskId.Version,
			},
		},
		{
			name:     "node execution identifier",
			key:      catalog.Key{},
			metadata: catalog.Metadata{NodeExecutionIdentifier: nID},
			expected: map[string]string{
				datacatalog.ExecNodeIDKey:  nID.NodeId,
				datacatalog.ExecDomainKey:  nID.ExecutionId.Domain,
				datacatalog.ExecNameKey:    nID.ExecutionId.Name,
				datacatalog.ExecProjectKey: nID.ExecutionId.Project,
				datacatalog.ExecOrgKey:     nID.ExecutionId.Org,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			meta := GenerateCacheMetadata(tc.key, tc.metadata)
			assert.Equal(t, tc.expected, meta.KeyMap.Values)
		})
	}
}
