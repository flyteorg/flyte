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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type (
	TaskActionConditionType   string
	TaskActionConditionReason string
)

// Condition type constants
// Following Kubernetes API conventions:
// - Condition types describe the current observed state
// - Use Reason field to track sub-states (like Queued, Initializing, Executing)
const (
	// ConditionTypeProgressing indicates whether the TaskAction is actively progressing.
	// This is True when the TaskAction is queued, initializing, or executing.
	// This is False when the TaskAction has completed or failed.
	ConditionTypeProgressing TaskActionConditionType = "Progressing"

	// ConditionTypeSucceeded indicates whether the TaskAction has completed successfully.
	// This is a terminal condition. Once True, the TaskAction will not be reconciled further.
	ConditionTypeSucceeded TaskActionConditionType = "Succeeded"

	// ConditionTypeFailed indicates whether the TaskAction has failed.
	// This is a terminal condition. Once True, the TaskAction will not be reconciled further.
	ConditionTypeFailed TaskActionConditionType = "Failed"
)

// Condition reason constants
// Reasons explain why a condition has a particular status.
// These are used in the Reason field of conditions to provide detailed sub-state information.
const (
	// ConditionReasonQueued indicates the TaskAction is queued and waiting for resources
	ConditionReasonQueued TaskActionConditionReason = "Queued"

	// ConditionReasonInitializing indicates the TaskAction is being initialized
	ConditionReasonInitializing TaskActionConditionReason = "Initializing"

	// ConditionReasonExecuting indicates the TaskAction is actively executing
	ConditionReasonExecuting TaskActionConditionReason = "Executing"

	// ConditionReasonCompleted indicates the TaskAction has completed successfully
	ConditionReasonCompleted TaskActionConditionReason = "Completed"
)

// TaskActionSpec defines the desired state of TaskAction
type TaskActionSpec struct {
	// RunName is the name of the run this action belongs to
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=30
	RunName string `json:"runName"`

	// Org this action belongs to
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Org string `json:"org"`

	// Project this action belongs to
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Project string `json:"project"`

	// Domain this action belongs to
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	Domain string `json:"domain"`

	// ActionName is the unique name of this action within the run
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=30
	ActionName string `json:"actionName"`

	// ParentActionName is the optional name of the parent action
	// +optional
	ParentActionName *string `json:"parentActionName,omitempty"`

	// InputURI is the path to the input data for this action
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	InputURI string `json:"inputUri"`

	// RunOutputBase is the base path where this action should write its output
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	RunOutputBase string `json:"runOutputBase"`
}

func (in *TaskActionSpec) GetActionSpec() (*workflow.ActionSpec, error) {
	// Build ActionSpec from structured fields
	spec := &workflow.ActionSpec{
		ActionId: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     in.Org,
				Project: in.Project,
				Domain:  in.Domain,
				Name:    in.RunName,
			},
			Name: in.ActionName,
		},
		ParentActionName: in.ParentActionName,
		InputUri:         in.InputURI,
		RunOutputBase:    in.RunOutputBase,
	}

	return spec, nil
}

func (in *TaskActionSpec) SetActionSpec(spec *workflow.ActionSpec) error {
	// Populate structured fields from ActionSpec
	if spec.ActionId != nil {
		if spec.ActionId.Run != nil {
			in.Org = spec.ActionId.Run.Org
			in.Project = spec.ActionId.Run.Project
			in.Domain = spec.ActionId.Run.Domain
			in.RunName = spec.ActionId.Run.Name
		}
		in.ActionName = spec.ActionId.Name
	}
	in.ParentActionName = spec.ParentActionName
	in.InputURI = spec.InputUri
	in.RunOutputBase = spec.RunOutputBase

	return nil
}

// TaskActionStatus defines the observed state of TaskAction.
type TaskActionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// StateJSON is the JSON serialized NodeStatus that was last sent to the State Service
	// +optional
	StateJSON string `json:"stateJson,omitempty"`

	// conditions represent the current state of the TaskAction resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Run",type="string",JSONPath=".spec.runName"
// +kubebuilder:printcolumn:name="Action",type="string",JSONPath=".spec.actionName"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type=='Progressing')].reason"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Progressing",type="string",JSONPath=".status.conditions[?(@.type=='Progressing')].status",priority=1
// +kubebuilder:printcolumn:name="Succeeded",type="string",JSONPath=".status.conditions[?(@.type=='Succeeded')].status",priority=1
// +kubebuilder:printcolumn:name="Failed",type="string",JSONPath=".status.conditions[?(@.type=='Failed')].status",priority=1

// TaskAction is the Schema for the taskactions API
type TaskAction struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of TaskAction
	// +required
	Spec TaskActionSpec `json:"spec"`

	// status defines the observed state of TaskAction
	// +optional
	Status TaskActionStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// TaskActionList contains a list of TaskAction
type TaskActionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskAction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskAction{}, &TaskActionList{})
}
