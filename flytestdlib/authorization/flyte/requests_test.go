package flyte

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	// Register the cluster service descriptor for the coverage assertion.
	_ "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cluster"
)

func TestRequestScopes(t *testing.T) {
	pid := &common.ProjectIdentifier{Name: "payments", Domain: "prod"}
	rid := &common.RunIdentifier{Project: "payments", Domain: "prod", Name: "run"}
	for _, tc := range []struct {
		procedure string
		message   proto.Message
		resources []string
	}{
		{"workflow.RunService/GetRunDetails", &workflow.GetRunDetailsRequest{RunId: rid}, []string{"payments/prod"}},
		{"workflow.RunService/ListRuns", &workflow.ListRunsRequest{ScopeBy: &workflow.ListRunsRequest_ProjectId{ProjectId: pid}}, []string{"payments/prod"}},
		{"workflow.RunService/ListRuns", &workflow.ListRunsRequest{ScopeBy: &workflow.ListRunsRequest_TaskName{TaskName: &task.TaskName{Project: "payments", Domain: "prod"}}}, []string{"default"}},
		{"workflow.RunService/WatchRuns", &workflow.WatchRunsRequest{Target: &workflow.WatchRunsRequest_TaskId{TaskId: &task.TaskIdentifier{Project: "payments", Domain: "prod"}}}, []string{"default"}},
		{"workflow.RunService/WatchRuns", &workflow.WatchRunsRequest{Target: &workflow.WatchRunsRequest_ProjectId{ProjectId: pid}}, []string{"payments/prod"}},
		{"workflow.RunService/CreateRun", &workflow.CreateRunRequest{Id: &workflow.CreateRunRequest_RunId{RunId: rid}, Task: &workflow.CreateRunRequest_TaskId{TaskId: &task.TaskIdentifier{Project: "other", Domain: "dev"}}}, []string{"payments/prod", "other/dev"}},
		{"trigger.TriggerService/UpdateTriggers", &trigger.UpdateTriggersRequest{Names: []*common.TriggerName{{Project: "payments", Domain: "prod"}, {Project: "other", Domain: "dev"}}}, []string{"payments/prod", "other/dev"}},
		{"task.TaskService/DeployTask", &task.DeployTaskRequest{TaskId: &task.TaskIdentifier{Project: "payments", Domain: "prod"}, Triggers: []*task.TaskTrigger{{}}}, []string{"payments/prod", "payments/prod"}},
		{"project.ProjectService/GetProject", &project.GetProjectRequest{Id: "payments"}, []string{"payments"}},
		{"settings.SettingsService/GetSettingsForEdit", &settings.GetSettingsForEditRequest{Key: &settings.SettingsKey{Project: "payments", Domain: "prod"}}, []string{"default"}},
		{"workflow.InternalRunService/RecordAction", &workflow.RecordActionRequest{}, []string{"default"}},
		{"dataproxy.DataProxyService/UploadInputs", &dataproxy.UploadInputsRequest{Id: &dataproxy.UploadInputsRequest_RunId{RunId: rid}, BaseDir: "s3://elsewhere"}, []string{"payments/prod", "default"}},
		{"workflow.RunService/CreateRun", &workflow.CreateRunRequest{Id: &workflow.CreateRunRequest_RunId{RunId: rid}, InputWrapper: &workflow.CreateRunRequest_OffloadedInputData{OffloadedInputData: &common.OffloadedInputData{Uri: "s3://other"}}}, []string{"payments/prod", "default"}},
	} {
		t.Run(tc.procedure, func(t *testing.T) {
			checks, err := Resolve("/flyteidl2."+tc.procedure, tc.message)
			require.NoError(t, err)
			var got []string
			for _, c := range checks {
				got = append(got, string(c.Resource.ID))
			}
			require.Equal(t, tc.resources, got)
		})
	}
}

func TestRejectUnresolvedAndUnsafeRequests(t *testing.T) {
	for _, tc := range []struct {
		procedure string
		message   proto.Message
	}{
		{"workflow.RunService/NewMethod", &workflow.GetRunDetailsRequest{}},
		{"workflow.RunService/GetRunDetails", &workflow.AbortRunRequest{}},
		{"workflow.RunService/GetRunDetails", &workflow.GetRunDetailsRequest{}},
		{"workflow.RunService/GetRunDetails", &workflow.GetRunDetailsRequest{RunId: &common.RunIdentifier{Project: "payments"}}},
		{"trigger.TriggerService/DeleteTriggers", &trigger.DeleteTriggersRequest{}},
		{"dataproxy.DataProxyService/CreateUploadLocation", &dataproxy.CreateUploadLocationRequest{Project: "payments", Domain: "prod", Filename: "../../other"}},
		{"dataproxy.DataProxyService/CreateUploadLocation", &dataproxy.CreateUploadLocationRequest{Project: "payments", Domain: "prod", FilenameRoot: "/other"}},
		{"dataproxy.DataProxyService/CreateUploadLocation", &dataproxy.CreateUploadLocationRequest{Project: "payments", Domain: "prod", Filename: "%2e%2e/%2e%2e/other"}},
		{"dataproxy.DataProxyService/CreateUploadLocation", &dataproxy.CreateUploadLocationRequest{Project: "../other", Domain: "prod"}},
	} {
		_, err := Resolve("/flyteidl2."+tc.procedure, tc.message)
		require.Error(t, err, tc.procedure)
	}
}

// Any API addition must have its authorization semantics explicitly reviewed.
func TestRegisteredServiceCoverage(t *testing.T) {
	for _, name := range []string{"workflow.RunService", "workflow.InternalRunService", "task.TaskService", "trigger.TriggerService", "project.ProjectService", "settings.SettingsService", "dataproxy.DataProxyService", "workflow.TranslatorService", "cluster.ClusterService"} {
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName("flyteidl2." + name))
		require.NoError(t, err)
		methods := desc.(protoreflect.ServiceDescriptor).Methods()
		for idx := 0; idx < methods.Len(); idx++ {
			method := methods.Get(idx)
			r, ok := rules["/flyteidl2."+name+"/"+string(method.Name())]
			require.True(t, ok, "%s.%s needs an authorization rule", name, method.Name())
			for _, path := range r.paths {
				if path == "." {
					continue
				}
				input := method.Input()
				parts := strings.Split(path, ".")
				for idx, part := range parts {
					require.NotNil(t, input, "%s.%s: %s", name, method.Name(), path)
					field := input.Fields().ByName(protoreflect.Name(part))
					require.NotNil(t, field, "%s.%s: %s", name, method.Name(), path)
					if idx < len(parts)-1 {
						input = field.Message()
					}
				}
			}
		}
	}
}
