package testutils

import (
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
)

func GetValidTaskRequest() admin.TaskCreateRequest {
	return admin.TaskCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		Spec: &admin.TaskSpec{
			Template: &core.TaskTemplate{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_TASK,
					Project:      "project",
					Domain:       "domain",
					Name:         "name",
					Version:      "version",
				},
				Type: "type",
				Metadata: &core.TaskMetadata{
					Runtime: &core.RuntimeMetadata{
						Version: "runtime version",
					},
				},
				Interface: &core.TypedInterface{},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: "image",
						Command: []string{
							"command",
						},
					},
				},
			},
			Description: &admin.DescriptionEntity{ShortDescription: "hello"},
		},
	}
}

func GetValidTaskRequestWithOverrides(project string, domain string, name string, version string) admin.TaskCreateRequest {
	return admin.TaskCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      project,
			Domain:       domain,
			Name:         name,
			Version:      version,
		},
		Spec: &admin.TaskSpec{
			Template: &core.TaskTemplate{
				Type: "type",
				Metadata: &core.TaskMetadata{
					Runtime: &core.RuntimeMetadata{
						Version: "runtime version",
					},
				},
				Interface: &core.TypedInterface{},
				Target: &core.TaskTemplate_Container{
					Container: &core.Container{
						Image: "image",
						Command: []string{
							"command",
						},
					},
				},
			},
		},
	}
}

func GetValidTaskSpecBytes() []byte {
	bytes, _ := proto.Marshal(GetValidTaskRequest().Spec)
	return bytes
}

func GetWorkflowRequest() admin.WorkflowCreateRequest {
	identifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}
	return admin.WorkflowCreateRequest{
		Id: &identifier,
		Spec: &admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id: &identifier,
				Interface: &core.TypedInterface{
					Inputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"foo": {
								Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
							},
						},
					},
					Outputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"bar": {
								Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
							},
						},
					},
				},
				Nodes: []*core.Node{
					{
						Id: "node 1",
					},
					{
						Id: "node 2",
					},
				},
			},
			Description: &admin.DescriptionEntity{ShortDescription: "hello"},
		},
	}
}

func GetLaunchPlanRequest() admin.LaunchPlanCreateRequest {
	return admin.LaunchPlanCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				ResourceType: core.ResourceType_WORKFLOW,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"foo": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
						Behavior: &core.Parameter_Default{
							Default: coreutils.MustMakeLiteral("foo-value"),
						},
					},
				},
			},
			FixedInputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"bar": coreutils.MustMakeLiteral("bar-value"),
				},
			},
		},
	}
}

func GetLaunchPlanRequestWithCronSchedule(testCronExpr string) admin.LaunchPlanCreateRequest {
	lpRequest := GetLaunchPlanRequest()
	lpRequest.Spec.EntityMetadata = &admin.LaunchPlanMetadata{
		Schedule: &admin.Schedule{
			ScheduleExpression:  &admin.Schedule_CronExpression{CronExpression: testCronExpr},
			KickoffTimeInputArg: "",
		},
	}
	return lpRequest
}

func GetLaunchPlanRequestWithFixedRateSchedule(testRateValue uint32, testRateUnit admin.FixedRateUnit) admin.LaunchPlanCreateRequest {
	lpRequest := GetLaunchPlanRequest()
	lpRequest.Spec.EntityMetadata = &admin.LaunchPlanMetadata{
		Schedule: &admin.Schedule{
			ScheduleExpression: &admin.Schedule_Rate{
				Rate: &admin.FixedRate{
					Value: testRateValue,
					Unit:  testRateUnit,
				},
			},
		},
	}
	return lpRequest
}

func GetExecutionRequest() admin.ExecutionCreateRequest {
	return admin.ExecutionCreateRequest{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
			NotificationOverrides: &admin.ExecutionSpec_Notifications{
				Notifications: &admin.NotificationList{
					Notifications: []*admin.Notification{
						{
							Phases: []core.WorkflowExecution_Phase{
								core.WorkflowExecution_FAILED,
							},
							Type: &admin.Notification_Slack{
								Slack: &admin.SlackNotification{
									RecipientsEmail: []string{
										"slack@example.com",
									},
								},
							},
						},
					},
				},
			},
			RawOutputDataConfig: &admin.RawOutputDataConfig{OutputLocationPrefix: "default_raw_output"},
			Envs:                &admin.Envs{},
			Tags:                []string{"tag1", "tag2"},
		},
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": coreutils.MustMakeLiteral("foo-value-1"),
			},
		},
	}
}

func GetSampleWorkflowSpecForTest() admin.WorkflowSpec {
	return admin.WorkflowSpec{
		Template: &core.WorkflowTemplate{
			Interface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"foo": {
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
						"bar": {
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
					},
				},
			},
		},
	}
}

func GetSampleLpSpecForTest() admin.LaunchPlanSpec {
	return admin.LaunchPlanSpec{
		WorkflowId: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "project",
			Domain:       "domain",
			Name:         "name",
			Version:      "version",
		},
		DefaultInputs: &core.ParameterMap{
			Parameters: map[string]*core.Parameter{
				"foo": {
					Var: &core.Variable{
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
					Behavior: &core.Parameter_Default{
						Default: coreutils.MustMakeLiteral("foo-value"),
					},
				},
			},
		},
		FixedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"bar": coreutils.MustMakeLiteral("bar-value"),
			},
		},
		EntityMetadata: &admin.LaunchPlanMetadata{
			Notifications: []*admin.Notification{
				{
					Phases: []core.WorkflowExecution_Phase{
						core.WorkflowExecution_SUCCEEDED,
					},
					Type: &admin.Notification_Email{
						Email: &admin.EmailNotification{
							RecipientsEmail: []string{
								"slack@example.com",
							},
						},
					},
				},
			},
		},
	}
}

func GetWorkflowRequestInterfaceBytes() []byte {
	bytes, _ := proto.Marshal(GetWorkflowRequest().Spec.Template.Interface)
	return bytes
}
