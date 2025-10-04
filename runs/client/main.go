package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

func main() {
	// Create a client
	client := workflowconnect.NewRunServiceClient(
		http.DefaultClient,
		"http://localhost:8090",
	)

	ctx := context.Background()

	// Test 1: Create a run with a task spec
	fmt.Println("Test 1: Creating a run with task spec...")
	createRunReq := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Org:     "my-org",
				Project: "my-project",
				Domain:  "development",
				Name:    "test-run-001",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{
				TaskTemplate: &core.TaskTemplate{
					Target: &core.TaskTemplate_Container{
						Container: &core.Container{
							Image: "alpine:latest",
							Args:  []string{"echo", "Hello from Runs Service!"},
						},
					},
				},
			},
		},
		Inputs: &task.Inputs{},
	}

	createResp, err := client.CreateRun(ctx, connect.NewRequest(createRunReq))
	if err != nil {
		log.Fatalf("Failed to create run: %v", err)
	}
	fmt.Printf("✓ Run created successfully: %+v\n\n", createResp.Msg.Run.Action.Id.Run.Name)

	// Test 2: Create another run with ProjectId (auto-generated name)
	fmt.Println("Test 2: Creating a run with auto-generated name...")
	createRunReq2 := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_ProjectId{
			ProjectId: &common.ProjectIdentifier{
				Organization: "my-org",
				Name:         "my-project",
				Domain:       "development",
			},
		},
		Task: &workflow.CreateRunRequest_TaskSpec{
			TaskSpec: &task.TaskSpec{
				TaskTemplate: &core.TaskTemplate{
					Target: &core.TaskTemplate_Container{
						Container: &core.Container{
							Image: "python:3.9",
							Args:  []string{"python", "-c", "print('Auto-generated run')"},
						},
					},
				},
			},
		},
		Inputs: &task.Inputs{},
	}

	createResp2, err := client.CreateRun(ctx, connect.NewRequest(createRunReq2))
	if err != nil {
		log.Fatalf("Failed to create second run: %v", err)
	}
	runName := createResp2.Msg.Run.Action.Id.Run.Name
	fmt.Printf("✓ Run with auto-generated name created: %s\n\n", runName)

	// Test 3: Get run details
	fmt.Println("Test 3: Getting run details...")
	getRunReq := &workflow.GetRunDetailsRequest{
		RunId: &common.RunIdentifier{
			Org:     "my-org",
			Project: "my-project",
			Domain:  "development",
			Name:    "test-run-001",
		},
	}

	getRunResp, err := client.GetRunDetails(ctx, connect.NewRequest(getRunReq))
	if err != nil {
		log.Fatalf("Failed to get run details: %v", err)
	}
	fmt.Printf("✓ Retrieved run details: %+v\n\n", getRunResp.Msg)

	// Test 4: List runs
	fmt.Println("Test 4: Listing runs...")
	listRunsReq := &workflow.ListRunsRequest{
		ScopeBy: &workflow.ListRunsRequest_ProjectId{
			ProjectId: &common.ProjectIdentifier{
				Organization: "my-org",
				Name:         "my-project",
				Domain:       "development",
			},
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}

	listRunsResp, err := client.ListRuns(ctx, connect.NewRequest(listRunsReq))
	if err != nil {
		log.Fatalf("Failed to list runs: %v", err)
	}
	fmt.Printf("✓ Found %d runs\n\n", len(listRunsResp.Msg.Runs))

	// Test 5: List actions for a run
	fmt.Println("Test 5: Listing actions for the run...")
	listActionsReq := &workflow.ListActionsRequest{
		RunId: &common.RunIdentifier{
			Org:     "my-org",
			Project: "my-project",
			Domain:  "development",
			Name:    "test-run-001",
		},
		Request: &common.ListRequest{
			Limit: 10,
		},
	}

	listActionsResp, err := client.ListActions(ctx, connect.NewRequest(listActionsReq))
	if err != nil {
		log.Fatalf("Failed to list actions: %v", err)
	}
	fmt.Printf("✓ Found %d actions\n\n", len(listActionsResp.Msg.Actions))

	// Test 6: Abort an action
	if len(listActionsResp.Msg.Actions) > 0 {
		fmt.Println("Test 6: Aborting an action...")
		actionToAbort := listActionsResp.Msg.Actions[0]
		abortActionReq := &workflow.AbortActionRequest{
			ActionId: actionToAbort.Id,
			Reason:   "Testing abort functionality",
		}

		_, err = client.AbortAction(ctx, connect.NewRequest(abortActionReq))
		if err != nil {
			log.Fatalf("Failed to abort action: %v", err)
		}
		fmt.Println("✓ Action aborted successfully\n")
	}

	// Test 7: Abort a run
	fmt.Println("Test 7: Aborting a run...")
	abortRunReq := &workflow.AbortRunRequest{
		RunId: &common.RunIdentifier{
			Org:     "my-org",
			Project: "my-project",
			Domain:  "development",
			Name:    runName, // Use the auto-generated run name
		},
		Reason: strPtr("Testing run abort functionality"),
	}

	_, err = client.AbortRun(ctx, connect.NewRequest(abortRunReq))
	if err != nil {
		log.Fatalf("Failed to abort run: %v", err)
	}
	fmt.Println("✓ Run aborted successfully\n")

	fmt.Println("All tests completed successfully! 🎉")
}

func strPtr(s string) *string {
	return &s
}
