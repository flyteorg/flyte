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
	"github.com/flyteorg/flyte/v2/executor/pkg/fakeplugins"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/stretchr/testify/mock"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/net/http2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
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

// +kubebuilder:rbac:groups=flyte.org,resources=taskactions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=flyte.org,resources=taskactions/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *TaskActionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Init plugin
	p, err := r.invokePlugin()
	if err != nil {
		logger.Error(err, "Failed to initialize plugin")
		return ctrl.Result{}, err
	}

	// Fetch the TaskAction instance
	taskAction := &flyteorgv1.TaskAction{}
	if err := r.Get(ctx, req.NamespacedName, taskAction); err != nil {
		if errors.IsNotFound(err) {
			// Call handler to abort
			logger.Info("TaskAction not found, likely deleted", "name", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get TaskAction")
		return ctrl.Result{}, err
	}

	actionSpec, err := taskAction.Spec.GetActionSpec()
	if err != nil {
		logger.Error(err, "Failed to unmarshal ActionSpec")
		return ctrl.Result{}, err
	}

	// Determine current phase (default to empty if new)
	currentPhase := taskAction.Status.Phase
	var requeueAfter time.Duration

	time.Sleep(2 * time.Second)

	// Create mock TaskExecutionContext with PluginStateReader
	// TODO(alex): We might need to remove NodeExecutionContext from TaskExecutionContext interface as we don't have NodeExecutor now
	mockPluginStateReader := &mocks.PluginStateReader{}
	mockPluginStateReader.OnGet(mock.Anything).Return(uint8(r.getPhaseID(currentPhase)), nil)
	mockTaskExecutionContext := &mocks.TaskExecutionContext{}
	mockTaskExecutionContext.OnPluginStateReader().Return(mockPluginStateReader)

	pluginTrns, err := p.Handle(ctx, mockTaskExecutionContext)
	if err != nil {
		logger.Error(err, "Failed to handle TaskAction")
		return ctrl.Result{}, err
	}

	nextPhase := r.getPhaseString(pluginTrns.Info().Phase())

	if nextPhase != "" && currentPhase != nextPhase {
		// Create state JSON (simplified NodeStatus)
		stateJSON := r.createStateJSON(actionSpec, nextPhase)

		// Send state update to State Service
		if err := r.updateStateService(ctx, actionSpec.ActionId, actionSpec.ParentActionName, stateJSON); err != nil {
			logger.Error(err, "Failed to update state service", "phase", nextPhase)
			return ctrl.Result{}, err
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
	} else {
		// If taskAction CR was in terminal state, we should stop reconciling
		return ctrl.Result{}, nil
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

// updateStateService sends a state update to the State Service via unary RPC
func (r *TaskActionReconciler) updateStateService(ctx context.Context, actionID *common.ActionIdentifier, parentActionName *string, stateJSON string) error {
	// Create PutRequest
	reqMsg := &workflow.PutRequest{
		ActionId:         actionID,
		ParentActionName: parentActionName,
		State:            stateJSON,
	}

	// Make unary Put call
	resp, err := r.stateServiceClient.Put(ctx, connect.NewRequest(reqMsg))
	if err != nil {
		return fmt.Errorf("failed to call put: %w", err)
	}

	// Check response status
	if resp.Msg.Status.Code != 0 {
		return fmt.Errorf("state service returned error: %s", resp.Msg.Status.Message)
	}

	return nil
}

// Invoke plugin based on task type
func (r *TaskActionReconciler) invokePlugin() (core.Plugin, error) {
	// This is a sample implementation, we should invoke the plugin by task type in the future
	return fakeplugins.NewPhaseBasedPlugin(), nil
}

// Get phase ID from phase string
// This function should be removed in the future
func (r *TaskActionReconciler) getPhaseID(phase string) core.Phase {
	switch phase {
	case "":
		return core.PhaseNotReady
	case PhaseQueued:
		return core.PhaseQueued
	case PhaseInitializing:
		return core.PhaseInitializing
	case PhaseRunning:
		return core.PhaseRunning
	case PhaseSucceeded:
		return core.PhaseSuccess
	case PhaseFailed:
		return core.PhaseRetryableFailure
	case PhaseAborted:
		return core.PhaseAborted
	default:
		return core.PhaseUndefined
	}
}

// get phase string from phase ID
// This function should be removed in the future
func (r *TaskActionReconciler) getPhaseString(phase core.Phase) string {
	switch phase {
	case core.PhaseQueued:
		return PhaseQueued
	case core.PhaseInitializing:
		return PhaseInitializing
	case core.PhaseRunning:
		return PhaseRunning
	case core.PhaseSuccess:
		return PhaseSucceeded
	case core.PhaseRetryableFailure:
		return PhaseFailed
	case core.PhaseAborted:
		return PhaseAborted
	default:
		return ""
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskActionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&flyteorgv1.TaskAction{}).
		Named("taskaction").
		Complete(r)
}
