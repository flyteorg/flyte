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
	stderrors "errors"
	"sync"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	k8sPlugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

var errFakeWebhookDenied = stderrors.New(`admission webhook denied: secret "test1" not found`)

// fakePlugin is a minimal Plugin implementation for unit tests.
type fakePlugin struct {
	id         string
	abortCalls int
}

func (f *fakePlugin) GetID() string                              { return f.id }
func (f *fakePlugin) GetProperties() pluginsCore.PluginProperties { return pluginsCore.PluginProperties{} }
func (f *fakePlugin) Handle(_ context.Context, _ pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	return pluginsCore.UnknownTransition, nil
}
func (f *fakePlugin) Abort(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	f.abortCalls++
	return nil
}
func (f *fakePlugin) Finalize(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	return nil
}

// fakeEventsClient is a no-op implementation of EventsProxyServiceClient for tests.
type fakeEventsClient struct{}

func (f *fakeEventsClient) Record(_ context.Context, _ *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	return connect.NewResponse(&workflow.RecordResponse{}), nil
}

// recordingEventsClient captures all recorded ActionEvents for assertion in tests.
type recordingEventsClient struct {
	mu     sync.Mutex
	events []*workflow.ActionEvent
}

func (r *recordingEventsClient) Record(_ context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, req.Msg.GetEvents()...)
	return connect.NewResponse(&workflow.RecordResponse{}), nil
}

func (r *recordingEventsClient) RecordedEvents() []*workflow.ActionEvent {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*workflow.ActionEvent, len(r.events))
	copy(out, r.events)
	return out
}

// buildTaskTemplateBytes creates a minimal protobuf-serialized TaskTemplate
// with a container spec that the pod plugin can use to build a Pod.
func buildTaskTemplateBytes(taskType, image string) []byte {
	tmpl := &core.TaskTemplate{
		Type: taskType,
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image:   image,
				Command: []string{"echo"},
				Args:    []string{"hello"},
			},
		},
		Metadata: &core.TaskMetadata{
			Runtime: &core.RuntimeMetadata{
				Type: core.RuntimeMetadata_FLYTE_SDK,
			},
		},
		Interface: &core.TypedInterface{},
	}
	data, err := proto.Marshal(tmpl)
	Expect(err).NotTo(HaveOccurred())
	return data
}

// emptyPluginRegistry satisfies plugin.PluginRegistryIface with no registered plugins.
type emptyPluginRegistry struct{}

func (emptyPluginRegistry) GetCorePlugins() []pluginsCore.PluginEntry { return nil }
func (emptyPluginRegistry) GetK8sPlugins() []k8sPlugin.PluginEntry    { return nil }

var _ = Describe("TaskAction Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		taskaction := &flyteorgv1.TaskAction{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind TaskAction")
			err := k8sClient.Get(ctx, typeNamespacedName, taskaction)
			if err != nil && errors.IsNotFound(err) {
				resource := &flyteorgv1.TaskAction{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: flyteorgv1.TaskActionSpec{
						RunName:       "test-run",
						Project:       "test-project",
						Domain:        "test-domain",
						ActionName:    "test-action",
						InputURI:      "/tmp/input",
						RunOutputBase: "/tmp/output",
						TaskType:      "python-task",
						TaskTemplate:  buildTaskTemplateBytes("python-task", "python:3.11"),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &flyteorgv1.TaskAction{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance TaskAction")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")

			controllerReconciler := &TaskActionReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				Recorder:       record.NewFakeRecorder(10),
				PluginRegistry: pluginRegistry,
				DataStore:      dataStore,
				eventsClient:   &fakeEventsClient{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// After the first reconciliation the controller should have added
			// a finalizer and/or set conditions on the TaskAction status.
			updatedTaskAction := &flyteorgv1.TaskAction{}
			err = k8sClient.Get(ctx, typeNamespacedName, updatedTaskAction)
			Expect(err).NotTo(HaveOccurred())

			// The first reconcile adds the finalizer; a second reconcile
			// drives the plugin Handle path which sets conditions.
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, typeNamespacedName, updatedTaskAction)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedTaskAction.Status.Conditions).NotTo(BeEmpty())
		})
	})

	Context("When reconciling a terminal TaskAction", func() {
		const terminalResourceName = "terminal-test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      terminalResourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			By("creating a terminal TaskAction")
			resource := &flyteorgv1.TaskAction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      terminalResourceName,
					Namespace: "default",
				},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "test-run",
					Project:       "test-project",
					Domain:        "test-domain",
					ActionName:    "test-action",
					InputURI:      "/tmp/input",
					RunOutputBase: "/tmp/output",
					TaskType:      "python-task",
					TaskTemplate:  buildTaskTemplateBytes("python-task", "python:3.11"),
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())

			// Set terminal condition on status
			resource.Status.Conditions = []metav1.Condition{
				{
					Type:               string(flyteorgv1.ConditionTypeSucceeded),
					Status:             metav1.ConditionTrue,
					Reason:             string(flyteorgv1.ConditionReasonCompleted),
					Message:            "TaskAction completed successfully",
					LastTransitionTime: metav1.Now(),
				},
			}
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &flyteorgv1.TaskAction{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
		})

		It("should set GC labels on terminal TaskAction", func() {
			By("Reconciling the terminal resource")

			controllerReconciler := &TaskActionReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify GC labels are set
			updatedTaskAction := &flyteorgv1.TaskAction{}
			err = k8sClient.Get(ctx, typeNamespacedName, updatedTaskAction)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedTaskAction.GetLabels()).To(HaveKeyWithValue(LabelTerminationStatus, LabelValueTerminated))
			Expect(updatedTaskAction.GetLabels()).To(HaveKey(LabelCompletedTime))
		})
	})

	Context("mapPhaseToConditions", func() {
		It("should keep PhaseHistory using controller time, not pod time", func() {
			ta := &flyteorgv1.TaskAction{}
			podTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) // far in the past
			info := pluginsCore.PhaseInfoRunning(0, &pluginsCore.TaskInfo{
				OccurredAt: &podTime,
			})
			before := time.Now().Add(-time.Second)
			mapPhaseToConditions(ta, info)
			after := time.Now().Add(time.Second)

			Expect(ta.Status.PhaseHistory).To(HaveLen(1))
			phTime := ta.Status.PhaseHistory[0].OccurredAt.Time
			Expect(phTime.After(before)).To(BeTrue(), "PhaseHistory should use controller time, not pod time")
			Expect(phTime.Before(after)).To(BeTrue(), "PhaseHistory should use controller time, not pod time")
		})

		It("should persist ErrorState (Code/Kind/Message) on permanent failure", func() {
			ta := &flyteorgv1.TaskAction{}
			info := pluginsCore.PhaseInfoFailure("OOMKilled", "Pod OOMKilled", nil)
			mapPhaseToConditions(ta, info)

			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal("OOMKilled"))
			Expect(ta.Status.ErrorState.Kind).To(Equal("USER"))
			Expect(ta.Status.ErrorState.Message).To(Equal("Pod OOMKilled"))
		})

		It("should persist ErrorState on retryable failure with system kind", func() {
			ta := &flyteorgv1.TaskAction{}
			info := pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", "node lost", nil)
			mapPhaseToConditions(ta, info)

			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal("ResourceDeletedExternally"))
			Expect(ta.Status.ErrorState.Kind).To(Equal("SYSTEM"))
		})
	})

	Context("maxSystemFailures", func() {
		It("returns the default when MaxSystemFailures is zero", func() {
			r := &TaskActionReconciler{}
			Expect(r.maxSystemFailures()).To(Equal(DefaultMaxSystemFailures))
		})

		It("returns the configured value when set", func() {
			r := &TaskActionReconciler{MaxSystemFailures: 7}
			Expect(r.maxSystemFailures()).To(Equal(uint32(7)))
		})
	})

	Context("recordSystemError", func() {
		const handleErrResource = "handle-err-resource"
		ctx := context.Background()
		nn := types.NamespacedName{Name: handleErrResource, Namespace: "default"}

		BeforeEach(func() {
			resource := &flyteorgv1.TaskAction{
				ObjectMeta: metav1.ObjectMeta{Name: handleErrResource, Namespace: "default"},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "test-run",
					Project:       "test-project",
					Domain:        "test-domain",
					ActionName:    "test-action",
					InputURI:      "/tmp/input",
					RunOutputBase: "/tmp/output",
					TaskType:      "python-task",
					TaskTemplate:  buildTaskTemplateBytes("python-task", "python:3.11"),
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &flyteorgv1.TaskAction{}
			if err := k8sClient.Get(ctx, nn, resource); err == nil {
				resource.Finalizers = nil
				_ = k8sClient.Update(ctx, resource)
				_ = k8sClient.Delete(ctx, resource)
			}
		})

		It("increments SystemFailures and requeues without consuming user retries", func() {
			r := &TaskActionReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				Recorder:          record.NewFakeRecorder(10),
				MaxSystemFailures: 3,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()
			startingAttempts := ta.Status.Attempts

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
			Expect(ta.Status.SystemFailures).To(Equal(uint32(1)))
			Expect(ta.Status.Attempts).To(Equal(startingAttempts), "user retry budget must not be consumed by system errors")
			Expect(isTerminal(ta)).To(BeFalse())

			persisted := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, persisted)).To(Succeed())
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(1)))
		})

		It("converts to PermanentFailure once the threshold is exceeded", func() {
			r := &TaskActionReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				Recorder:          record.NewFakeRecorder(10),
				eventsClient:      &fakeEventsClient{},
				MaxSystemFailures: 2,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			ta.Status.SystemFailures = 2
			Expect(k8sClient.Status().Update(ctx, ta)).To(Succeed())
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeZero(), "terminal — should not requeue")
			Expect(ta.Status.SystemFailures).To(Equal(uint32(3)))
			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal(MaxSystemFailuresExceededCode))
			Expect(ta.Status.ErrorState.Kind).To(Equal("SYSTEM"))
			Expect(ta.Status.ErrorState.Message).To(ContainSubstring("admission webhook"))
			Expect(isTerminal(ta)).To(BeTrue())
		})
	})

	Context("resetPluginResource", func() {
		It("aborts the plugin and clears persisted plugin state", func() {
			r := &TaskActionReconciler{}
			ta := &flyteorgv1.TaskAction{}
			ta.Status.PluginState = []byte("stale")
			ta.Status.PluginStateVersion = 1

			fp := &fakePlugin{id: "pod"}
			r.resetPluginResource(context.Background(), ta, fp, nil)

			Expect(fp.abortCalls).To(Equal(1))
			Expect(ta.Status.PluginState).To(BeNil())
			Expect(ta.Status.PluginStateVersion).To(Equal(uint8(0)))
		})
	})

	Context("isSystemRetryableFailure", func() {
		It("is true for PhaseInfoSystemRetryableFailure (PhaseRetryableFailure + kind SYSTEM)", func() {
			info := pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", "node lost", nil)
			Expect(isSystemRetryableFailure(info)).To(BeTrue())
		})

		It("is false for a user-kind retryable failure", func() {
			info := pluginsCore.PhaseInfoRetryableFailure("OOMKilled", "container OOMKilled", nil)
			Expect(isSystemRetryableFailure(info)).To(BeFalse())
		})

		It("is false for a permanent failure", func() {
			info := pluginsCore.PhaseInfoFailure("BadInput", "invalid spec", nil)
			Expect(isSystemRetryableFailure(info)).To(BeFalse())
		})

		It("is false for a running phase", func() {
			info := pluginsCore.PhaseInfoRunning(0, nil)
			Expect(isSystemRetryableFailure(info)).To(BeFalse())
		})
	})

	Context("systemErrorFromPhaseInfo", func() {
		It("formats code and message from the ExecutionError", func() {
			info := pluginsCore.PhaseInfoSystemRetryableFailure("ResourceDeletedExternally", "node lost", nil)
			err := systemErrorFromPhaseInfo(info)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ResourceDeletedExternally"))
			Expect(err.Error()).To(ContainSubstring("node lost"))
		})
	})

	Context("errorStateFromExecError", func() {
		It("returns nil for nil input", func() {
			Expect(errorStateFromExecError(nil)).To(BeNil())
		})

		It("preserves Code, Kind=USER, and Message", func() {
			es := errorStateFromExecError(&core.ExecutionError{
				Code:    "OOMKilled",
				Kind:    core.ExecutionError_USER,
				Message: "container OOMKilled",
			})
			Expect(es).NotTo(BeNil())
			Expect(es.Code).To(Equal("OOMKilled"))
			Expect(es.Kind).To(Equal("USER"))
			Expect(es.Message).To(Equal("container OOMKilled"))
		})

		It("maps ExecutionError_SYSTEM kind to \"SYSTEM\"", func() {
			es := errorStateFromExecError(&core.ExecutionError{
				Code: "InternalError",
				Kind: core.ExecutionError_SYSTEM,
			})
			Expect(es.Kind).To(Equal("SYSTEM"))
		})

		It("leaves Kind empty when ExecutionError kind is UNKNOWN", func() {
			es := errorStateFromExecError(&core.ExecutionError{
				Code: "Unknown",
				Kind: core.ExecutionError_UNKNOWN,
			})
			Expect(es.Kind).To(Equal(""))
		})
	})

	Context("taskActionStatusChanged", func() {
		It("should detect PhaseHistory changes", func() {
			oldStatus := flyteorgv1.TaskActionStatus{
				PhaseHistory: []flyteorgv1.PhaseTransition{
					{Phase: "Queued", OccurredAt: metav1.Now()},
				},
			}
			newStatus := flyteorgv1.TaskActionStatus{
				PhaseHistory: []flyteorgv1.PhaseTransition{
					{Phase: "Queued", OccurredAt: metav1.Now()},
					{Phase: "Executing", OccurredAt: metav1.Now()},
				},
			}
			Expect(taskActionStatusChanged(oldStatus, newStatus)).To(BeTrue())
		})

		It("should return false when nothing changed", func() {
			now := metav1.Now()
			status := flyteorgv1.TaskActionStatus{
				PhaseHistory: []flyteorgv1.PhaseTransition{
					{Phase: "Queued", OccurredAt: now},
				},
			}
			Expect(taskActionStatusChanged(status, status)).To(BeFalse())
		})

		It("should detect PhaseHistory addition from empty", func() {
			oldStatus := flyteorgv1.TaskActionStatus{}
			newStatus := flyteorgv1.TaskActionStatus{
				PhaseHistory: []flyteorgv1.PhaseTransition{
					{Phase: "Queued", OccurredAt: metav1.Now()},
				},
			}
			Expect(taskActionStatusChanged(oldStatus, newStatus)).To(BeTrue())
		})
	})

	Context("When a TaskAction is deleted (abort flow)", func() {
		const abortResourceName = "abort-test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      abortResourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			resource := &flyteorgv1.TaskAction{
				ObjectMeta: metav1.ObjectMeta{
					Name:      abortResourceName,
					Namespace: "default",
					Finalizers: []string{taskActionFinalizer},
				},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "abort-run",
					Project:       "abort-project",
					Domain:        "abort-domain",
					ActionName:    "abort-action",
					InputURI:      "/tmp/input",
					RunOutputBase: "/tmp/output",
					TaskType:      "python-task",
					TaskTemplate:  buildTaskTemplateBytes("python-task", "python:3.11"),
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		AfterEach(func() {
			resource := &flyteorgv1.TaskAction{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			if err == nil {
				resource.Finalizers = nil
				Expect(k8sClient.Update(ctx, resource)).To(Succeed())
			}
		})

		It("should emit an ACTION_PHASE_ABORTED event before removing the finalizer", func() {
			recorder := &recordingEventsClient{}
			reconciler := &TaskActionReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				Recorder:       record.NewFakeRecorder(10),
				PluginRegistry: pluginRegistry,
				DataStore:      dataStore,
				eventsClient:   recorder,
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())

			// Finalizer should have been removed — object is gone.
			deleted := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, deleted)).NotTo(Succeed())

			// An ABORTED event must have been emitted.
			recorded := recorder.RecordedEvents()
			Expect(recorded).NotTo(BeEmpty())
			phases := make([]interface{}, len(recorded))
			for i, e := range recorded {
				phases[i] = e.GetPhase()
			}
			Expect(phases).To(ContainElement(common.ActionPhase_ACTION_PHASE_ABORTED))
		})
	})

	Context("toClusterEvents", func() {
		It("should include both phase reason and additional reasons", func() {
			phaseOccurredAt := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
			eventOccurredAt := phaseOccurredAt.Add(2 * time.Minute)
			fallbackTime := metav1.NewTime(phaseOccurredAt.Add(5 * time.Minute))

			phaseInfo := pluginsCore.PhaseInfoQueuedWithTaskInfo(
				phaseOccurredAt,
				pluginsCore.DefaultPhaseVersion,
				"cluster is creating",
				&pluginsCore.TaskInfo{
					OccurredAt: &phaseOccurredAt,
					AdditionalReasons: []pluginsCore.ReasonInfo{
						{
							Reason:     "Head pod pending",
							OccurredAt: &eventOccurredAt,
						},
					},
				},
			)

			events := toClusterEvents(phaseInfo, timestamppb.New(fallbackTime.Time))
			Expect(events).To(HaveLen(2))
			Expect(events[0].GetMessage()).To(Equal("cluster is creating"))
			Expect(events[0].GetOccurredAt().AsTime()).To(Equal(phaseOccurredAt))
			Expect(events[1].GetMessage()).To(Equal("Head pod pending"))
			Expect(events[1].GetOccurredAt().AsTime()).To(Equal(eventOccurredAt))
		})
	})
})
