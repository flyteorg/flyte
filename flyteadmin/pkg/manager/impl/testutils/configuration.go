package testutils

import (
	"google.golang.org/grpc/codes"

	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

// Redefine EncodeDocumentKey here to capture that if we change the way to encode document keys
func EncodeDocumentKey(id *admin.ConfigurationID) (string, error) {
	key, err := utils.MarshalPbToString(id)
	if err != nil {
		return "", flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to marshal document key")
	}
	return key, nil
}

var WorkflowExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"1", "2", "3",
			},
		},
	},
}

var WorkflowTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:    "1",
				Memory: "1Gi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:    "2",
				Memory: "2Gi",
			},
		},
	},
}

var ProjectDomainExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"4", "5", "6",
			},
		},
	},
}

var ProjectDomainTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:    "3",
				Memory: "3Gi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:    "4",
				Memory: "4Gi",
			},
		},
	},
}

var ProjectExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"7", "8", "9",
			},
		},
	},
}

var ProjectTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:    "5",
				Memory: "5Gi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:    "6",
				Memory: "6Gi",
			},
		},
	},
}

var UpdateTaskResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_TaskResourceAttributes{
		TaskResourceAttributes: &admin.TaskResourceAttributes{
			Defaults: &admin.TaskResourceSpec{
				Cpu:    "7",
				Memory: "7Gi",
			},
			Limits: &admin.TaskResourceSpec{
				Cpu:    "8",
				Memory: "8Gi",
			},
		},
	},
}

func MockConfigurations(org, project, domain, workflow string) (map[string]*admin.Configuration, error) {
	mockConfigurations := make(map[string]*admin.Configuration)
	key, err := EncodeDocumentKey(&admin.ConfigurationID{
		Org:      org,
		Project:  project,
		Domain:   domain,
		Workflow: workflow,
	})
	if err != nil {
		return nil, err
	}
	mockConfigurations[key] = &admin.Configuration{
		ExecutionQueueAttributes: WorkflowExecutionQueueAttributes.GetExecutionQueueAttributes(),
		TaskResourceAttributes:   WorkflowTaskResourceAttributes.GetTaskResourceAttributes(),
	}

	key, err = EncodeDocumentKey(&admin.ConfigurationID{
		Org:     org,
		Project: project,
		Domain:  domain,
	})
	if err != nil {
		return nil, err
	}
	mockConfigurations[key] = &admin.Configuration{
		ExecutionQueueAttributes: ProjectDomainExecutionQueueAttributes.GetExecutionQueueAttributes(),
		TaskResourceAttributes:   ProjectDomainTaskResourceAttributes.GetTaskResourceAttributes(),
	}

	key, err = EncodeDocumentKey(&admin.ConfigurationID{
		Org:     org,
		Project: project,
	})
	if err != nil {
		return nil, err
	}
	mockConfigurations[key] = &admin.Configuration{
		ExecutionQueueAttributes: ProjectExecutionQueueAttributes.GetExecutionQueueAttributes(),
		TaskResourceAttributes:   ProjectTaskResourceAttributes.GetTaskResourceAttributes(),
	}

	key, err = EncodeDocumentKey(&admin.ConfigurationID{})
	if err != nil {
		return nil, err
	}
	mockConfigurations[key] = &admin.Configuration{
		WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{},
	}

	return mockConfigurations, nil
}
