/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// TaskActionReconciler reconciles a TaskAction object
type TaskActionReconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	StateServiceURL    string
	stateServiceClient workflowconnect.StateServiceClient
}

// NewTaskActionReconciler creates a new TaskActionReconciler with initialized clients
func NewTaskActionReconciler(c client.Client, scheme *runtime.Scheme, stateServiceURL string) *TaskActionReconciler {
	// Create HTTP/2 cleartext (h2c) client for buf connect
	// This is required because the state service uses h2c (HTTP/2 without TLS)
	httpClient := &http.Client{
		Transport: &http2.Transport{
			// Allow HTTP/2 without TLS (h2c)
			AllowHTTP: true,
			// Use HTTP/2 dialer
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				// Dial without TLS for h2c
				return net.Dial(network, addr)
			},
		},
	}

	return &TaskActionReconciler{
		Client:             c,
		Scheme:             scheme,
		StateServiceURL:    stateServiceURL,
		stateServiceClient: workflowconnect.NewStateServiceClient(httpClient, stateServiceURL),
	}
}

// Phase constants
const (
	PhaseQueued       = "PHASE_QUEUED"
	PhaseInitializing = "PHASE_INITIALIZING"
	PhaseRunning      = "PHASE_RUNNING"
	PhaseSucceeded    = "PHASE_SUCCEEDED"
	PhaseFailed       = "PHASE_FAILED"
	PhaseAborted      = "PHASE_ABORTED"
)

// +kubebuilder:rbac:groups=flyte.org.flyte.org,resources=taskactions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flyte.org.flyte.org,resources=taskactions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flyte.org.flyte.org,resources=taskactions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TaskActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the TaskAction instance
	taskAction := &flyteorgv1.TaskAction{}
	if err := r.Get(ctx, req.NamespacedName, taskAction); err != nil {
		if errors.IsNotFound(err) {
			// TaskAction was deleted
			logger.Info("TaskAction not found, likely deleted", "name", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TaskAction")
		return ctrl.Result{}, err
	}

	// Get the ActionSpec from the TaskAction
	actionSpec, err := taskAction.Spec.GetActionSpec()
	if err != nil {
		logger.Error(err, "Failed to unmarshal ActionSpec")
		return ctrl.Result{}, err
	}

	// Determine current phase (default to empty if new)
	currentPhase := taskAction.Status.Phase

	// State machine logic
	var nextPhase string
	var requeueAfter time.Duration

	//time.Sleep(2 * time.Second)
	switch currentPhase {
	case "":
		// New TaskAction - transition to Queued
		nextPhase = PhaseQueued
		logger.Info("New TaskAction detected, transitioning to Queued",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseQueued:
		// Queued → Initializing
		nextPhase = PhaseInitializing
		logger.Info("Transitioning from Queued to Initializing",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseInitializing:
		// Initializing → Running
		nextPhase = PhaseRunning
		logger.Info("Transitioning from Initializing to Running",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseRunning:
		// Running → Succeeded (simulated execution)
		nextPhase = PhaseSucceeded
		logger.Info("Transitioning from Running to Succeeded",
			"name", taskAction.Name, "action", actionSpec.ActionId.Name)

	case PhaseSucceeded, PhaseFailed, PhaseAborted:
		// Terminal states - no further transitions
		logger.Info("TaskAction in terminal state",
			"name", taskAction.Name, "phase", currentPhase)
		return ctrl.Result{}, nil

	default:
		logger.Info("Unknown phase, resetting to Queued",
			"name", taskAction.Name, "phase", currentPhase)
		nextPhase = PhaseQueued
	}

	// Create state JSON (simplified NodeStatus)
	stateJSON := r.createStateJSON(actionSpec, nextPhase)

	// Send state update to State Service
	if err := r.updateStateService(ctx, actionSpec.ActionId, actionSpec.ParentActionName, stateJSON); err != nil {
		logger.Error(err, "Failed to update state service", "phase", nextPhase)
		// Continue anyway - we'll try again on next reconcile
	}

	// Update TaskAction status
	taskAction.Status.Phase = nextPhase
	taskAction.Status.StateJSON = stateJSON
	taskAction.Status.Message = fmt.Sprintf("Transitioned to %s", nextPhase)

	if err := r.Status().Update(ctx, taskAction); err != nil {
		logger.Error(err, "Failed to update TaskAction status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully updated TaskAction status",
		"name", taskAction.Name, "phase", nextPhase)

	// Requeue after a delay to simulate state transitions
	// For non-terminal states, requeue after 5 seconds
	if nextPhase != PhaseSucceeded && nextPhase != PhaseFailed && nextPhase != PhaseAborted {
		requeueAfter = 5 * time.Second
	}

	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// createStateJSON creates a simplified NodeStatus JSON representation
func (r *TaskActionReconciler) createStateJSON(actionSpec *workflow.ActionSpec, phase string) string {
	// Create a simplified state object
	state := map[string]interface{}{
		"phase":     phase,
		"actionId":  fmt.Sprintf("%s/%s", actionSpec.ActionId.Run.Name, actionSpec.ActionId.Name),
		"timestamp": time.Now().Format(time.RFC3339),
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return "{}"
	}

	return string(stateBytes)
}

// updateStateService sends a state update to the State Service via bidirectional streaming
func (r *TaskActionReconciler) updateStateService(ctx context.Context, actionID *common.ActionIdentifier, parentActionName *string, stateJSON string) error {
	// Create a bidirectional stream
	stream := r.stateServiceClient.Put(ctx)

	// Send PutRequest
	req := &workflow.PutRequest{
		ActionId:         actionID,
		ParentActionName: parentActionName,
		State:            stateJSON,
	}

	if err := stream.Send(req); err != nil {
		return fmt.Errorf("failed to send put request: %w", err)
	}

	// Receive PutResponse
	resp, err := stream.Receive()
	if err != nil {
		return fmt.Errorf("failed to receive put response: %w", err)
	}

	// Check response status
	if resp.Status.Code != 0 {
		return fmt.Errorf("state service returned error: %s", resp.Status.Message)
	}

	// Close the stream
	if err := stream.CloseRequest(); err != nil {
		return fmt.Errorf("failed to close request stream: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskActionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&flyteorgv1.TaskAction{}).
		Named("taskaction").
		Complete(r)
}
