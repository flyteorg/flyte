package qubole

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	quboleMocks "github.com/lyft/flyteplugins/go/tasks/v1/qubole/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
	"github.com/lyft/flytestdlib/storage"
	libUtils "github.com/lyft/flytestdlib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MockSecretsManager struct {
}

func (m MockSecretsManager) GetToken() (string, error) {
	return "sample-token", nil
}

func CreateMockTaskContextWithRealTaskExecId() *mocks.TaskContext {
	taskExId := core.TaskExecutionIdentifier{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "flyteplugins",
			Domain:       "testing",
			Name:         "send_event_test",
		},
		NodeExecutionId: &core.NodeExecutionIdentifier{
			NodeId: "nodeId",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "flyteplugins",
				Domain:  "testing",
				Name:    "node_1",
			},
		},
	}
	mockTaskCtx := &mocks.TaskContext{}
	id := &mocks.TaskExecutionID{}
	id.On("GetGeneratedName").Return("test-hive-job")
	id.On("GetID").Return(taskExId)

	mockTaskCtx.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	mockTaskCtx.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	mockTaskCtx.On("GetTaskExecutionID").Return(id)
	mockTaskCtx.On("GetNamespace").Return("test-namespace")
	mockTaskCtx.On("GetOwnerReference").Return(metav1.OwnerReference{})
	mockTaskCtx.On("GetPhaseVersion").Return(uint32(0))
	dummyStorageReference := storage.DataReference("s3://fake/file/in")
	mockTaskCtx.On("GetInputsFile").Return(dummyStorageReference)
	dummyStorageReferenceOut := storage.DataReference("s3://fake/file/out")
	mockTaskCtx.On("GetOutputsFile").Return(dummyStorageReferenceOut)

	return mockTaskCtx
}

func createDummyHiveCustomObj() *plugins.QuboleHiveJob {
	hiveJob := plugins.QuboleHiveJob{}

	hiveJob.ClusterLabel = "cluster-label"
	hiveJob.Tags = []string{"tag1", "tag2"}
	hiveJob.QueryCollection = &plugins.HiveQueryCollection{Queries: []*plugins.HiveQuery{{Query: "Select 5"}}}
	return &hiveJob
}

func createDummyHiveTaskTemplate(id string) *core.TaskTemplate {
	hiveJob := createDummyHiveCustomObj()
	hiveJobJSON, err := utils.MarshalToString(hiveJob)
	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(hiveJobJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &core.TaskTemplate{
		Id:     &core.Identifier{Name: id},
		Type:   "hive",
		Custom: &structObj,
	}
}

type MockAutoRefreshCache struct {
	*quboleMocks.AutoRefreshCache
	values map[string]libUtils.CacheItem
}

func (m MockAutoRefreshCache) GetOrCreate(item libUtils.CacheItem) (libUtils.CacheItem, error) {
	if cachedItem, ok := m.values[item.ID()]; ok {
		return cachedItem, nil
	} else {
		m.values[item.ID()] = item
	}
	return item, nil
}

func NewMockAutoRefreshCache() MockAutoRefreshCache {
	return MockAutoRefreshCache{
		AutoRefreshCache: &quboleMocks.AutoRefreshCache{},
		values:           make(map[string]libUtils.CacheItem),
	}
}

func (m MockAutoRefreshCache) Get(key string) libUtils.CacheItem {
	if cachedItem, ok := m.values[key]; ok {
		return cachedItem
	} else {
		return nil
	}
}
