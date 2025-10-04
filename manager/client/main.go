package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

const (
	runsServiceURL = "http://localhost:8090"
	org            = "my-org"
	project        = "my-project"
	domain         = "development"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("\n🛑 Shutting down...")
		cancel()
	}()

	// Create HTTP client
	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: true,
		},
	}

	// Create service clients
	runClient := workflowconnect.NewRunServiceClient(httpClient, runsServiceURL)

	log.Println("🚀 Flyte Client Demo")
	log.Println("===================")

	var wg sync.WaitGroup

	// 1. Start watching runs (background goroutine)
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchRuns(ctx, runClient)
	}()

	// Give the watch a moment to start
	time.Sleep(500 * time.Millisecond)

	// 2. Create a run
	runID := createRun(ctx, runClient)
	if runID == nil {
		log.Println("❌ Failed to create run")
		return
	}

	// 3. Watch actions for the run
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchActions(ctx, runClient, runID)
	}()

	// Wait for context cancellation or all watchers to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Println("Context cancelled, waiting for goroutines to finish...")
		wg.Wait()
	case <-done:
		log.Println("All watchers finished")
	}

	log.Println("✅ Client demo complete")
}

// watchRuns watches for run updates and logs them
func watchRuns(ctx context.Context, client workflowconnect.RunServiceClient) {
	log.Println("📡 Starting WatchRuns stream...")

	req := &workflow.WatchRunsRequest{
		Target: &workflow.WatchRunsRequest_ProjectId{
			ProjectId: &common.ProjectIdentifier{
				Organization: org,
				Name:         project,
				Domain:       domain,
			},
		},
	}

	stream, err := client.WatchRuns(ctx, connect.NewRequest(req))
	if err != nil {
		log.Printf("❌ Failed to start WatchRuns: %v", err)
		return
	}
	defer stream.Close()

	log.Println("✅ WatchRuns stream connected")

	for stream.Receive() {
		msg := stream.Msg()
		for _, run := range msg.Runs {
			if run.Action != nil && run.Action.Id != nil && run.Action.Id.Run != nil {
				runID := run.Action.Id.Run
				log.Printf("📥 [WatchRuns] Run Update: %s (org: %s, project: %s, domain: %s)",
					runID.Name, runID.Org, runID.Project, runID.Domain)
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("❌ WatchRuns stream error: %v", err)
	}
}

// watchActions watches for action updates and logs them
func watchActions(ctx context.Context, client workflowconnect.RunServiceClient, runID *common.RunIdentifier) {
	log.Printf("📡 Starting WatchActions stream for run: %s...", runID.Name)

	req := &workflow.WatchActionsRequest{
		RunId: runID,
	}

	stream, err := client.WatchActions(ctx, connect.NewRequest(req))
	if err != nil {
		log.Printf("❌ Failed to start WatchActions: %v", err)
		return
	}
	defer stream.Close()

	log.Println("✅ WatchActions stream connected")

	for stream.Receive() {
		msg := stream.Msg()
		for _, enrichedAction := range msg.EnrichedActions {
			if enrichedAction.Action != nil && enrichedAction.Action.Id != nil {
				var phase workflow.Phase
				if enrichedAction.Action.Status != nil {
					phase = enrichedAction.Action.Status.Phase
				}
				log.Printf("📥 [WatchActions] Action: %s | Phase: %s",
					enrichedAction.Action.Id.Name,
					formatPhase(phase))
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Printf("❌ WatchActions stream error: %v", err)
	}
}

// createRun creates a new workflow run
func createRun(ctx context.Context, client workflowconnect.RunServiceClient) *common.RunIdentifier {
	log.Println("\n📝 Creating a new run...")

	// Generate a unique run name
	runName := fmt.Sprintf("run-%d", time.Now().Unix())

	runID := &common.RunIdentifier{
		Org:     org,
		Project: project,
		Domain:  domain,
		Name:    runName,
	}

	// Create a simple task identifier
	taskID := &task.TaskIdentifier{
		Org:     org,
		Project: project,
		Domain:  domain,
		Name:    "example-task",
		Version: "v1",
	}

	req := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: runID,
		},
		Task: &workflow.CreateRunRequest_TaskId{
			TaskId: taskID,
		},
		Inputs: &task.Inputs{},
	}

	resp, err := client.CreateRun(ctx, connect.NewRequest(req))
	if err != nil {
		log.Printf("❌ Failed to create run: %v", err)
		return nil
	}

	if resp.Msg.Run != nil && resp.Msg.Run.Action != nil && resp.Msg.Run.Action.Id != nil && resp.Msg.Run.Action.Id.Run != nil {
		createdRunID := resp.Msg.Run.Action.Id.Run
		log.Printf("✅ Run created: %s", createdRunID.Name)
		log.Printf("   Organization: %s", createdRunID.Org)
		log.Printf("   Project: %s", createdRunID.Project)
		log.Printf("   Domain: %s", createdRunID.Domain)
		log.Println()
	}

	return runID
}

// formatPhase converts a Phase enum to a readable string with emoji
func formatPhase(phase workflow.Phase) string {
	switch phase {
	case workflow.Phase_PHASE_UNSPECIFIED:
		return "⚪ UNSPECIFIED"
	case workflow.Phase_PHASE_QUEUED:
		return "🔵 QUEUED"
	case workflow.Phase_PHASE_WAITING_FOR_RESOURCES:
		return "⏳ WAITING_FOR_RESOURCES"
	case workflow.Phase_PHASE_INITIALIZING:
		return "🔄 INITIALIZING"
	case workflow.Phase_PHASE_RUNNING:
		return "🏃 RUNNING"
	case workflow.Phase_PHASE_SUCCEEDED:
		return "✅ SUCCEEDED"
	case workflow.Phase_PHASE_FAILED:
		return "❌ FAILED"
	case workflow.Phase_PHASE_ABORTED:
		return "🛑 ABORTED"
	case workflow.Phase_PHASE_TIMED_OUT:
		return "⏰ TIMED_OUT"
	default:
		return fmt.Sprintf("❓ UNKNOWN(%d)", phase)
	}
}
