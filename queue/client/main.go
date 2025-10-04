package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

func main() {
	// Create a client
	client := workflowconnect.NewQueueServiceClient(
		http.DefaultClient,
		"http://localhost:8089",
	)

	ctx := context.Background()

	// Test 1: Enqueue an action
	fmt.Println("Test 1: Enqueuing an action...")
	enqueueReq := &workflow.EnqueueActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     "my-org",
				Project: "my-project",
				Domain:  "development",
				Name:    "run-001",
			},
			Name: "task-001",
		},
		InputUri:      "s3://my-bucket/inputs/run-001/task-001",
		RunOutputBase: "s3://my-bucket/outputs/run-001",
		Spec: &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Target: &core.TaskTemplate_Container{
							Container: &core.Container{
								Image: "alpine:latest",
								Args:  []string{"echo", "Hello from Queue Service!"},
							},
						},
					},
				},
			},
		},
	}

	enqueueResp, err := client.EnqueueAction(ctx, connect.NewRequest(enqueueReq))
	if err != nil {
		log.Fatalf("Failed to enqueue action: %v", err)
	}
	fmt.Printf("✓ Action enqueued successfully: %+v\n\n", enqueueResp.Msg)

	// Test 2: Enqueue another action in the same run
	fmt.Println("Test 2: Enqueuing another action in the same run...")
	enqueueReq2 := &workflow.EnqueueActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     "my-org",
				Project: "my-project",
				Domain:  "development",
				Name:    "run-001",
			},
			Name: "task-002",
		},
		ParentActionName: strPtr("task-001"),
		InputUri:         "s3://my-bucket/inputs/run-001/task-002",
		RunOutputBase:    "s3://my-bucket/outputs/run-001",
		Spec: &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Target: &core.TaskTemplate_Container{
							Container: &core.Container{
								Image: "alpine:latest",
								Args:  []string{"echo", "Second task!"},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.EnqueueAction(ctx, connect.NewRequest(enqueueReq2))
	if err != nil {
		log.Fatalf("Failed to enqueue second action: %v", err)
	}
	fmt.Println("✓ Second action enqueued successfully\n")

	// Test 3: Abort a specific action
	fmt.Println("Test 3: Aborting a specific action...")
	abortActionReq := &workflow.AbortQueuedActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     "my-org",
				Project: "my-project",
				Domain:  "development",
				Name:    "run-001",
			},
			Name: "task-002",
		},
		Reason: strPtr("Testing abort functionality"),
	}

	_, err = client.AbortQueuedAction(ctx, connect.NewRequest(abortActionReq))
	if err != nil {
		log.Fatalf("Failed to abort action: %v", err)
	}
	fmt.Println("✓ Action aborted successfully\n")

	// Test 4: Enqueue action for a different run
	fmt.Println("Test 4: Enqueuing action for a different run...")
	enqueueReq3 := &workflow.EnqueueActionRequest{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     "my-org",
				Project: "my-project",
				Domain:  "development",
				Name:    "run-002",
			},
			Name: "task-001",
		},
		InputUri:      "s3://my-bucket/inputs/run-002/task-001",
		RunOutputBase: "s3://my-bucket/outputs/run-002",
		Spec: &workflow.EnqueueActionRequest_Task{
			Task: &workflow.TaskAction{
				Spec: &task.TaskSpec{
					TaskTemplate: &core.TaskTemplate{
						Target: &core.TaskTemplate_Container{
							Container: &core.Container{
								Image: "alpine:latest",
								Args:  []string{"echo", "Different run!"},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.EnqueueAction(ctx, connect.NewRequest(enqueueReq3))
	if err != nil {
		log.Fatalf("Failed to enqueue third action: %v", err)
	}
	fmt.Println("✓ Action for run-002 enqueued successfully\n")

	// Test 5: Abort entire run
	fmt.Println("Test 5: Aborting entire run-001...")
	abortRunReq := &workflow.AbortQueuedRunRequest{
		RunId: &common.RunIdentifier{
			Org:     "my-org",
			Project: "my-project",
			Domain:  "development",
			Name:    "run-001",
		},
		Reason: strPtr("Testing run abort functionality"),
	}

	_, err = client.AbortQueuedRun(ctx, connect.NewRequest(abortRunReq))
	if err != nil {
		log.Fatalf("Failed to abort run: %v", err)
	}
	fmt.Println("✓ Run aborted successfully\n")

	fmt.Println("All tests completed successfully! 🎉")
}

func strPtr(s string) *string {
	return &s
}
