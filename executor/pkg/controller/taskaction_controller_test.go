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
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	testingclock "k8s.io/utils/clock/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorplugin "github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	pluginserrors "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	k8sPlugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

var errFakeWebhookDenied = stderrors.New(`admission webhook denied: secret "test1" not found`)

// fakePlugin is a minimal Plugin implementation for unit tests.
type fakePlugin struct {
	id             string
	handleFunc     func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error)
	transitions    []pluginsCore.Transition
	handleErrors   []error
	abortErrors    []error
	finalizeErrors []error
	handleCalls    int
	abortCalls     int
	finalizeCalls  int
}

func (f *fakePlugin) GetID() string { return f.id }
func (f *fakePlugin) GetProperties() pluginsCore.PluginProperties {
	return pluginsCore.PluginProperties{}
}
func (f *fakePlugin) Handle(ctx context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	call := f.handleCalls
	f.handleCalls++
	if f.handleFunc != nil {
		return f.handleFunc(ctx, tCtx)
	}
	if call < len(f.handleErrors) && f.handleErrors[call] != nil {
		return pluginsCore.UnknownTransition, f.handleErrors[call]
	}
	if len(f.transitions) == 0 {
		return pluginsCore.UnknownTransition, nil
	}
	if call >= len(f.transitions) {
		call = len(f.transitions) - 1
	}
	return f.transitions[call], nil
}
func (f *fakePlugin) Abort(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	call := f.abortCalls
	f.abortCalls++
	if call < len(f.abortErrors) {
		return f.abortErrors[call]
	}
	return nil
}
func (f *fakePlugin) Finalize(_ context.Context, _ pluginsCore.TaskExecutionContext) error {
	call := f.finalizeCalls
	f.finalizeCalls++
	if call < len(f.finalizeErrors) {
		return f.finalizeErrors[call]
	}
	return nil
}

// fakeLaunchState is the plugin state of a fake that records whether it waits to launch or launched its resource.
type fakeLaunchState struct{ Waiting, Launched bool }

// fakeEventsClient is a no-op implementation of EventsProxyServiceClient for tests.
type fakeEventsClient struct{}

func (f *fakeEventsClient) Record(_ context.Context, _ *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	return connect.NewResponse(&workflow.RecordResponse{}), nil
}

type failingEventsClient struct {
	recordingEventsClient
	failures int
}

func (f *failingEventsClient) Record(ctx context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	if f.failures > 0 {
		f.failures--
		return nil, stderrors.New("events unavailable")
	}
	return f.recordingEventsClient.Record(ctx, req)
}

// retryEventRejectingClient records every event but the ones a system retry publishes,
// standing in for a store that cannot accept those (an unencodable version, say).
type retryEventRejectingClient struct {
	recordingEventsClient
	rejected int
}

func (c *retryEventRejectingClient) Record(ctx context.Context, req *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	for _, e := range req.Msg.GetEvents() {
		if e.GetVersion() >= systemRetryEventVersionBase {
			c.rejected++
			return nil, stderrors.New("unable to encode version")
		}
	}
	return c.recordingEventsClient.Record(ctx, req)
}

// systemRetryEvents picks out the events a system retry published, which are the only
// ones versioned from systemRetryEventVersionBase.
func systemRetryEvents(events []*workflow.ActionEvent) []*workflow.ActionEvent {
	var out []*workflow.ActionEvent
	for _, e := range events {
		if e.GetVersion() >= systemRetryEventVersionBase {
			out = append(out, e)
		}
	}
	return out
}

type failingStatusClient struct {
	client.Client
	failUpdates bool
}

func (f *failingStatusClient) Status() client.SubResourceWriter {
	return &failingStatusWriter{
		SubResourceWriter: f.Client.Status(),
		client:            f,
	}
}

type failingStatusWriter struct {
	client.SubResourceWriter
	client *failingStatusClient
}

func (f *failingStatusWriter) Update(
	ctx context.Context,
	obj client.Object,
	opts ...client.SubResourceUpdateOption,
) error {
	if f.client.failUpdates {
		return stderrors.New("status unavailable")
	}
	return f.SubResourceWriter.Update(ctx, obj, opts...)
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

func buildTaskTemplateBytesWithTimeoutAndRetries(
	taskType, image string,
	timeout time.Duration,
	retries uint32,
) []byte {
	tmpl := &core.TaskTemplate{}
	Expect(proto.Unmarshal(buildTaskTemplateBytes(taskType, image), tmpl)).To(Succeed())
	if timeout > 0 {
		tmpl.Metadata.Timeout = durationpb.New(timeout)
	}
	tmpl.Metadata.Retries = &core.RetryStrategy{Retries: retries}
	data, err := proto.Marshal(tmpl)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func buildTaskTemplateBytesWithQueuedTimeoutAndRetries(retries uint32) []byte {
	tmpl := &core.TaskTemplate{}
	Expect(proto.Unmarshal(buildTaskTemplateBytes("timeout-test", "busybox"), tmpl)).To(Succeed())
	tmpl.Metadata.Timeouts = &core.TimeoutStrategy{
		QueuedTimeout: durationpb.New(2 * time.Second),
	}
	tmpl.Metadata.Retries = &core.RetryStrategy{Retries: retries}
	data, err := proto.Marshal(tmpl)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func buildSerializedCacheTemplateBytesWithQueuedTimeout() []byte {
	tmpl := &core.TaskTemplate{}
	Expect(proto.Unmarshal(buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0), tmpl)).To(Succeed())
	tmpl.Id = &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Name:         "queued-timeout-test",
		Project:      "timeout-project",
		Domain:       "timeout-domain",
		Version:      "v1",
	}
	tmpl.Metadata.Discoverable, tmpl.Metadata.CacheSerializable = true, true
	data, err := proto.Marshal(tmpl)
	Expect(err).NotTo(HaveOccurred())
	return data
}

// launchThenPoll fakes a plugin that manages a resource the way the K8s plugin does: a call with no saved
// state launches or adopts the resource and can only report it Queued, and later calls read the resource.
func launchThenPoll(
	now func() time.Time,
	poll func() pluginsCore.Transition,
) func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
	return func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
		state := fakeLaunchState{}
		if _, err := tCtx.PluginStateReader().Get(&state); err != nil {
			return pluginsCore.UnknownTransition, err
		}
		if state.Launched {
			return poll(), nil
		}
		if err := tCtx.PluginStateWriter().Put(1, &fakeLaunchState{Launched: true}); err != nil {
			return pluginsCore.UnknownTransition, err
		}
		return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
			now(), pluginsCore.DefaultPhaseVersion, "task submitted")), nil
	}
}

func recordedPhases(recorded []*workflow.ActionEvent) []common.ActionPhase {
	phases := make([]common.ActionPhase, 0, len(recorded))
	for _, event := range recorded {
		phases = append(phases, event.GetPhase())
	}
	return phases
}

func withRetryBackoff(template []byte, backoff *core.Backoff) []byte {
	tmpl := &core.TaskTemplate{}
	Expect(proto.Unmarshal(template, tmpl)).To(Succeed())
	tmpl.Metadata.Retries.Backoff = backoff
	data, err := proto.Marshal(tmpl)
	Expect(err).NotTo(HaveOccurred())
	return data
}

type fakeCorePluginRegistry struct {
	plugin *fakePlugin
}

func (f fakeCorePluginRegistry) GetCorePlugins() []pluginsCore.PluginEntry {
	return []pluginsCore.PluginEntry{{
		ID:                  f.plugin.id,
		RegisteredTaskTypes: []pluginsCore.TaskType{"timeout-test"},
		LoadPlugin: func(context.Context, pluginsCore.SetupContext) (pluginsCore.Plugin, error) {
			return f.plugin, nil
		},
	}}
}

func (fakeCorePluginRegistry) GetK8sPlugins() []k8sPlugin.PluginEntry {
	return nil
}

func newFakePluginRegistry(fake *fakePlugin) *executorplugin.Registry {
	registry := executorplugin.NewRegistry(nil, fakeCorePluginRegistry{plugin: fake})
	Expect(registry.Initialize(context.Background())).To(Succeed())
	return registry
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
						TaskType:      "python",
						TaskTemplate:  buildTaskTemplateBytes("python", "python:3.11"),
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
				Recorder:       events.NewFakeRecorder(10),
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
					TaskType:      "python",
					TaskTemplate:  buildTaskTemplateBytes("python", "python:3.11"),
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
					TaskType:      "python",
					TaskTemplate:  buildTaskTemplateBytes("python", "python:3.11"),
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
				Recorder:          events.NewFakeRecorder(10),
				MaxSystemFailures: 3,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()
			startingAttempts := ta.Status.Attempts

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied, taskTimeouts{})
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
				Recorder:          events.NewFakeRecorder(10),
				eventsClient:      &fakeEventsClient{},
				MaxSystemFailures: 2,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			ta.Status.SystemFailures = 2
			Expect(k8sClient.Status().Update(ctx, ta)).To(Succeed())
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied, taskTimeouts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeZero(), "terminal — should not requeue")
			Expect(ta.Status.SystemFailures).To(Equal(uint32(3)))
			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal(MaxSystemFailuresExceededCode))
			Expect(ta.Status.ErrorState.Kind).To(Equal("SYSTEM"))
			Expect(ta.Status.ErrorState.Message).To(ContainSubstring("admission webhook"))
			Expect(isTerminal(ta)).To(BeTrue())
		})

		It("fails immediately on a non-retryable error without consuming any attempts", func() {
			r := &TaskActionReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				Recorder:          events.NewFakeRecorder(10),
				eventsClient:      &fakeEventsClient{},
				MaxSystemFailures: 3,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()
			startingAttempts := ta.Status.Attempts

			handleErr := pluginserrors.Errorf(pluginserrors.BadTaskSpecification, "invalid ray submission mode %q", "HttpMode")

			res, err := r.recordSystemError(ctx, ta, original, "ray", handleErr, taskTimeouts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeZero(), "terminal - should not requeue")
			Expect(ta.Status.SystemFailures).To(BeZero(), "a deterministic failure must not consume system attempts")
			Expect(ta.Status.Attempts).To(Equal(startingAttempts), "user retry budget must not be consumed")
			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal(pluginserrors.BadTaskSpecification))
			Expect(ta.Status.ErrorState.Kind).To(Equal("USER"))
			Expect(ta.Status.ErrorState.Message).To(ContainSubstring("invalid ray submission mode"))
			Expect(isTerminal(ta)).To(BeTrue())
		})

		It("fails an Apple Silicon task immediately when the backend has no bridge", func() {
			r := &TaskActionReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				Recorder:          events.NewFakeRecorder(10),
				eventsClient:      &fakeEventsClient{},
				MaxSystemFailures: 3,
			}
			ta := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
			original := ta.DeepCopy()
			startingAttempts := ta.Status.Attempts

			handleErr := pluginserrors.Errorf(pluginserrors.AppleSiliconNotAvailable,
				"gpu='AppleSilicon' requires an Apple Silicon devbox; this backend has no Apple Silicon bridge.")

			res, err := r.recordSystemError(ctx, ta, original, "container", handleErr, taskTimeouts{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.RequeueAfter).To(BeZero(), "terminal - should not requeue")
			Expect(ta.Status.SystemFailures).To(BeZero(), "retrying cannot conjure a bridge")
			Expect(ta.Status.Attempts).To(Equal(startingAttempts))
			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal(pluginserrors.AppleSiliconNotAvailable))
			Expect(ta.Status.ErrorState.Kind).To(Equal("USER"))
			Expect(isTerminal(ta)).To(BeTrue())
		})
	})

	Context("task max runtime", func() {
		ctx := context.Background()
		var created []types.NamespacedName

		createTaskAction := func(name string, template []byte) types.NamespacedName {
			nn := types.NamespacedName{Name: name, Namespace: "default"}
			resource := &flyteorgv1.TaskAction{
				ObjectMeta: metav1.ObjectMeta{
					Name:       nn.Name,
					Namespace:  nn.Namespace,
					Finalizers: []string{taskActionFinalizer},
				},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "timeout-run",
					Project:       "timeout-project",
					Domain:        "timeout-domain",
					ActionName:    name,
					InputURI:      "file:///tmp/input",
					RunOutputBase: "file:///tmp/output",
					TaskType:      "timeout-test",
					TaskTemplate:  template,
				},
			}
			Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			created = append(created, nn)
			return nn
		}

		getTaskAction := func(nn types.NamespacedName) *flyteorgv1.TaskAction {
			resource := &flyteorgv1.TaskAction{}
			Expect(k8sClient.Get(ctx, nn, resource)).To(Succeed())
			return resource
		}

		newReconciler := func(
			fake *fakePlugin,
			fakeClock *testingclock.FakeClock,
			eventsClient workflowconnect.EventsProxyServiceClient,
			kubeClient client.Client,
		) *TaskActionReconciler {
			if kubeClient == nil {
				kubeClient = k8sClient
			}
			return &TaskActionReconciler{
				Client:         kubeClient,
				Scheme:         k8sClient.Scheme(),
				Recorder:       events.NewFakeRecorder(20),
				PluginRegistry: newFakePluginRegistry(fake),
				DataStore:      dataStore,
				eventsClient:   eventsClient,
				Clock:          fakeClock,
			}
		}

		runningTransition := func(startedAt time.Time) pluginsCore.Transition {
			return pluginsCore.DoTransition(pluginsCore.PhaseInfoRunning(
				pluginsCore.DefaultPhaseVersion,
				&pluginsCore.TaskInfo{OccurredAt: &startedAt},
			))
		}

		AfterEach(func() {
			for _, nn := range created {
				resource := &flyteorgv1.TaskAction{}
				if err := k8sClient.Get(ctx, nn, resource); err != nil {
					Expect(errors.IsNotFound(err)).To(BeTrue())
					continue
				}
				resource.Finalizers = nil
				Expect(k8sClient.Update(ctx, resource)).To(Succeed())
				Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
			}
			created = nil
		})

		It("leaves tasks with no max runtime unlimited", func() {
			startedAt := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(startedAt)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(startedAt)},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-unset",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
			// The clock is anchored even without a bound; nothing reads it.
			Expect(getTaskAction(nn).Status.AttemptStartedAt.Time).To(BeTemporally("==", startedAt))

			fakeClock.Step(24 * time.Hour)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(BeZero())
			Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
		})

		It("times out a queued task after queued_timeout elapses", func() {
			const queuedTimeout = 2 * time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
						base,
						pluginsCore.DefaultPhaseVersion,
						"queued",
					)),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"queued-timeout-missing",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0),
			)
			fakeClock.SetTime(getTaskAction(nn).CreationTimestamp.Time)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.Conditions).To(ContainElement(And(
				HaveField("Type", string(flyteorgv1.ConditionTypeProgressing)),
				HaveField("Reason", string(flyteorgv1.ConditionReasonQueued)),
			)))

			fakeClock.Step(queuedTimeout)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState).To(Equal(&flyteorgv1.ErrorState{
				Code:    TaskExecutionTimedOutCode,
				Kind:    "USER",
				Message: "task attempt 1 exceeded its queued timeout of 2s",
			}))
			Expect(persisted.Status.Conditions).To(ContainElement(And(
				HaveField("Type", string(flyteorgv1.ConditionTypeFailed)),
				HaveField("Status", metav1.ConditionTrue),
				HaveField("Reason", string(flyteorgv1.ConditionReasonTimedOut)),
			)))
			Expect(fake.abortCalls).To(Equal(1))
			Expect(fake.finalizeCalls).To(Equal(1))
		})

		It("preserves a running attempt after its queued event fails to advance", func() {
			nn := createTaskAction("queued-timeout-event-outage",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{
				pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(base, pluginsCore.DefaultPhaseVersion, "queued")),
				runningTransition(base.Add(time.Second)),
			}}
			recorded := &failingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			req := reconcile.Request{NamespacedName: nn}
			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Second)
			recorded.failures = 1
			_, err = r.Reconcile(ctx, req)
			Expect(err).To(MatchError("events unavailable"))
			persisted := getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))
			Expect(persisted.Status.AttemptStartedAt).NotTo(BeNil())
			fakeClock.Step(2 * time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseRunning.String()))
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("gives retries a fresh queued timeout after their backoff", func() {
			const budget = 2 * time.Second
			const lookupFailure = "lookup"
			const runtimeFailure = "runtime"
			const userFailure = "user"
			for _, failure := range []string{"queued", lookupFailure, runtimeFailure, userFailure} {
				for _, backoff := range []time.Duration{0, 10 * time.Second} {
					template := &core.TaskTemplate{}
					Expect(proto.Unmarshal(buildTaskTemplateBytesWithQueuedTimeoutAndRetries(1), template)).To(Succeed())
					template.Metadata.Retries.Backoff = &core.Backoff{Base: durationpb.New(backoff)}
					if failure == runtimeFailure {
						template.Metadata.Timeout = durationpb.New(time.Second)
					}
					data, err := proto.Marshal(template)
					Expect(err).NotTo(HaveOccurred())
					nn := createTaskAction("queued-retry-"+failure+"-"+backoff.String(), data)
					base := getTaskAction(nn).CreationTimestamp.Time
					fakeClock := testingclock.NewFakeClock(base)
					queued := pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(base, pluginsCore.DefaultPhaseVersion, "queued"))
					fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{queued, queued, queued}}
					step := budget
					switch failure {
					case runtimeFailure:
						fake.transitions[0], fake.transitions[1] = runningTransition(base), runningTransition(base)
						step = time.Second
					case userFailure:
						fake.transitions[1] = pluginsCore.DoTransition(
							pluginsCore.PhaseInfoRetryableFailure("UserError", "task failed", nil))
						step = time.Second
					case lookupFailure:
						fake.handleErrors = []error{nil, stderrors.New("temporary lookup failure")}
					}
					recorded := &recordingEventsClient{}
					r := newReconciler(fake, fakeClock, recorded, nil)
					req := reconcile.Request{NamespacedName: nn}
					_, err = r.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					fakeClock.Step(step)
					_, err = r.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					if failure == lookupFailure {
						_, err = r.Reconcile(ctx, req)
						Expect(err).NotTo(HaveOccurred())
					}
					retryAt := fakeClock.Now().Add(backoff)
					persisted := getTaskAction(nn)
					Expect(persisted.Status.Attempts).To(Equal(uint32(2)))
					Expect(persisted.Status.AttemptQueuedAt).NotTo(BeNil())
					Expect(persisted.Status.AttemptQueuedAt.Time).To(BeTemporally("==", retryAt))
					if backoff > 0 {
						handled := fake.handleCalls
						fakeClock.Step(backoff - time.Second)
						result, reconcileErr := r.Reconcile(ctx, req)
						Expect(reconcileErr).NotTo(HaveOccurred())
						Expect(result.RequeueAfter).To(Equal(time.Second))
						Expect(fake.handleCalls).To(Equal(handled))
					}
					fakeClock.SetTime(retryAt)
					result, err := r.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RequeueAfter).To(Equal(budget))
					Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
					fakeClock.Step(budget - time.Nanosecond)
					result, err = r.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					Expect(result.RequeueAfter).To(Equal(time.Nanosecond))
					fakeClock.Step(time.Nanosecond)
					_, err = r.Reconcile(ctx, req)
					Expect(err).NotTo(HaveOccurred())
					persisted = getTaskAction(nn)
					Expect(isTerminal(persisted)).To(BeTrue())
					Expect(persisted.Status.ErrorState.Message).To(ContainSubstring("queued timeout"))
					var timedOut []uint32
					for _, event := range recorded.RecordedEvents() {
						if event.GetPhase() == common.ActionPhase_ACTION_PHASE_TIMED_OUT {
							timedOut = append(timedOut, event.GetAttempt())
						}
					}
					if failure == userFailure {
						Expect(timedOut).To(Equal([]uint32{2}))
					} else {
						Expect(timedOut).To(Equal([]uint32{1, 2}))
					}
				}
			}
		})

		DescribeTable("waits for current plugin state when queued timeout recovery encounters an outage",
			func(outage string) {
				nn := createTaskAction("queued-outage-"+outage, buildSerializedCacheTemplateBytesWithQueuedTimeout())
				base := getTaskAction(nn).CreationTimestamp.Time
				startedAt := base.Add(time.Second)
				fakeClock := testingclock.NewFakeClock(base)
				fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(base, pluginsCore.DefaultPhaseVersion, "queued")),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
						startedAt, pluginsCore.DefaultPhaseVersion, "initializing", &pluginsCore.TaskInfo{OccurredAt: &startedAt})),
				}}
				unavailable := false
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				r.Catalog = &stubCatalogClient{
					getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
						return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
					},
					getOrExtendReservationFunc: func(
						_ context.Context, _ catalog.Key, owner string, _ time.Duration,
					) (*cacheservice.Reservation, error) {
						if unavailable {
							return nil, stderrors.New("catalog unavailable")
						}
						return &cacheservice.Reservation{OwnerId: owner}, nil
					},
					releaseReservationFunc: func(context.Context, catalog.Key, string) error { return nil },
				}
				req := reconcile.Request{NamespacedName: nn}
				_, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				if outage == "plugin" {
					fake.handleErrors = []error{nil, stderrors.New("plugin unavailable")}
				} else {
					unavailable = true
				}
				fakeClock.Step(3 * time.Second)
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
				Expect(fake.abortCalls).To(BeZero())
				Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
				unavailable = false
				_, err = r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseInitializing.String()))
				Expect(getTaskAction(nn).Status.TimeoutAt).To(BeNil())
			},
			Entry("plugin", "plugin"),
			Entry("catalog", "catalog"),
		)

		It("keeps the system failure count when the first plugin call fails", func() {
			nn := createTaskAction("qt-first-call-fails",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			fakeClock := testingclock.NewFakeClock(getTaskAction(nn).CreationTimestamp.Time)
			fake := &fakePlugin{id: "timeout-plugin", handleErrors: []error{stderrors.New("temporary lookup failure")}}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptQueuedAt).NotTo(BeNil())
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(1)))
		})

		It("starts the queue budget of a retry scheduled before upgrade when its backoff ends", func() {
			const budget = 2 * time.Second
			nn := createTaskAction("qt-upgraded-retry",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(1))
			base := getTaskAction(nn).CreationTimestamp.Time
			launchableAt := metav1.NewTime(base.Add(10 * time.Second))
			// What a controller without queue budgets wrote: the retry was queued at base
			// and may launch once its backoff ends.
			upgraded := getTaskAction(nn)
			upgraded.Status.Attempts = 2
			upgraded.Status.PluginPhase = pluginsCore.PhaseQueued.String()
			upgraded.Status.NextAttemptAt = &launchableAt
			upgraded.Status.PhaseHistory = []flyteorgv1.PhaseTransition{{
				Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(base),
			}}
			Expect(k8sClient.Status().Update(ctx, upgraded)).To(Succeed())
			fakeClock := testingclock.NewFakeClock(launchableAt.Add(time.Second))
			queued := pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(base, pluginsCore.DefaultPhaseVersion, "queued"))
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{queued}}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse(), "one second into the budget")
			Expect(persisted.Status.AttemptQueuedAt.Time).To(BeTemporally("==", launchableAt.Time))
			Expect(fake.abortCalls).To(BeZero())

			fakeClock.Step(budget)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
		})

		It("does not end a retried attempt on the Queued report that launched or adopted its resource", func() {
			nn := createTaskAction("qt-first-report",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(1))
			base := getTaskAction(nn).CreationTimestamp.Time
			// The retry was queued before the controller went away, so it is not fresh work: its
			// first call may find the resource of a launch whose state was never saved.
			retried := getTaskAction(nn)
			retried.Status.Attempts = 2
			retried.Status.PluginPhase = pluginsCore.PhaseQueued.String()
			retried.Status.AttemptQueuedAt = &metav1.Time{Time: base}
			Expect(k8sClient.Status().Update(ctx, retried)).To(Succeed())
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
				state := fakeLaunchState{}
				if _, err := tCtx.PluginStateReader().Get(&state); err != nil {
					return pluginsCore.UnknownTransition, err
				}
				if !state.Launched {
					if err := tCtx.PluginStateWriter().Put(1, &fakeLaunchState{Launched: true}); err != nil {
						return pluginsCore.UnknownTransition, err
					}
					// Adopting a resource only says that it exists, whatever it has done since.
					return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
						fakeClock.Now(), pluginsCore.DefaultPhaseVersion, "task submitted")), nil
				}
				return runningTransition(base.Add(time.Second)), nil
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			req := reconcile.Request{NamespacedName: nn}
			fakeClock.Step(10 * time.Second) // the controller was away for several budgets

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))
			Expect(fake.abortCalls).To(BeZero())

			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseRunning.String()))
			Expect(persisted.Status.AttemptStartedAt).NotTo(BeNil())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("does not end an attempt on the launch report that follows a saved wait to launch", func() {
			nn := createTaskAction("qt-saved-wait",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
				state := fakeLaunchState{}
				if _, err := tCtx.PluginStateReader().Get(&state); err != nil {
					return pluginsCore.UnknownTransition, err
				}
				if state.Launched {
					return runningTransition(base.Add(time.Second)), nil
				}
				if state.Waiting {
					if err := tCtx.PluginStateWriter().Put(1, &fakeLaunchState{Launched: true}); err != nil {
						return pluginsCore.UnknownTransition, err
					}
					// Adopting a resource only says that it exists, whatever it has done since.
					return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
						fakeClock.Now(), pluginsCore.DefaultPhaseVersion, "task submitted")), nil
				}
				if err := tCtx.PluginStateWriter().Put(1, &fakeLaunchState{Waiting: true}); err != nil {
					return pluginsCore.UnknownTransition, err
				}
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResourcesInfo(
					fakeClock.Now(), pluginsCore.DefaultPhaseVersion, "waiting for the previous resource", nil)), nil
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseWaitingForResources.String()))
			Expect(persisted.Status.PluginState).NotTo(BeEmpty())

			fakeClock.Step(10 * time.Second) // the wait outlasted the budget
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))
			Expect(fake.abortCalls).To(BeZero())

			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseRunning.String()))
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("ends an attempt still waiting for resources past its budget when the plugin saves state each poll", func() {
			nn := createTaskAction("qt-wait-every-poll",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
				if err := tCtx.PluginStateWriter().Put(1, &fakeLaunchState{Waiting: true}); err != nil {
					return pluginsCore.UnknownTransition, err
				}
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoWaitingForResourcesInfo(
					fakeClock.Now(), pluginsCore.DefaultPhaseVersion, "waiting for resources", nil)), nil
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseWaitingForResources.String()))

			fakeClock.Step(10 * time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.TimeoutAt).NotTo(BeNil())
			Expect(fake.abortCalls).To(Equal(1))
		})

		It("keeps an attempt seen initializing in time when that event never publishes", func() {
			nn := createTaskAction("qt-initialized-in-time",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			calls := 0
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
				calls++
				observedAt := fakeClock.Now()
				if calls == 1 {
					return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
						observedAt, pluginsCore.DefaultPhaseVersion, "queued")), nil
				}
				// Dask, Ray and Spark report the time of the poll as the time they initialized.
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
					observedAt, pluginsCore.DefaultPhaseVersion, "initializing",
					&pluginsCore.TaskInfo{OccurredAt: &observedAt})), nil
			}
			recorded := &failingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Second) // inside the 2s budget
			recorded.failures = 1
			_, err = r.Reconcile(ctx, req)
			Expect(err).To(MatchError("events unavailable"))
			Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))

			fakeClock.Step(2 * time.Second) // the events came back after the budget ran out
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(persisted.Status.AttemptInitializedAt).NotTo(BeNil())
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseInitializing.String()))
			Expect(fake.abortCalls).To(BeZero())
		})

		DescribeTable("keeps an attempt that initialized while the controller was away unless the plugin proves it was late",
			func(name string, initializing func(polledAt, queuedAt time.Time) pluginsCore.PhaseInfo, wantTimedOut bool) {
				nn := createTaskAction(name, buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
				// A controller clock an hour behind the wall clock: a plugin that leaves the stamp to the
				// framework gets the wall time, which is after the poll just as it is in production.
				queuedAt := time.Now().Add(-time.Hour).Truncate(time.Second)
				polledAt := queuedAt.Add(10 * time.Second) // five budgets after the attempt was queued
				seeded := getTaskAction(nn)
				seeded.Status.PluginPhase = pluginsCore.PhaseQueued.String()
				seeded.Status.AttemptQueuedAt = &metav1.Time{Time: queuedAt}
				Expect(k8sClient.Status().Update(ctx, seeded)).To(Succeed())
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					return pluginsCore.DoTransition(initializing(polledAt, queuedAt)), nil
				}
				r := newReconciler(fake, testingclock.NewFakeClock(polledAt), &recordingEventsClient{}, nil)

				_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				if wantTimedOut {
					Expect(isTerminal(persisted)).To(BeTrue())
					Expect(persisted.Status.ErrorState.Message).To(ContainSubstring("queued timeout"))
					return
				}
				Expect(isTerminal(persisted)).To(BeFalse())
				Expect(persisted.Status.TimeoutAt).To(BeNil())
				Expect(persisted.Status.AttemptInitializedAt).NotTo(BeNil())
				Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseInitializing.String()))
				Expect(fake.abortCalls).To(BeZero())
			},
			// Dask, Ray, Spark and the clustered plugin stamp the poll, not the resource's own transition:
			// dask.go and clustered/phase.go set time.Now() themselves, ray.go does the same in its task
			// info and passes the cluster's creation time to a constructor that ignores it, and spark.go
			// sets nothing, which leaves phaseInfo in pluginmachinery/core/phase.go to stamp time.Now().
			Entry("dask stamps the poll", "qt-recovery-dask",
				func(polledAt, _ time.Time) pluginsCore.PhaseInfo {
					return pluginsCore.PhaseInfoInitializing(polledAt, pluginsCore.DefaultPhaseVersion, "job created",
						&pluginsCore.TaskInfo{OccurredAt: &polledAt})
				}, false),
			Entry("ray stamps the poll and ignores the creation time", "qt-recovery-ray",
				func(polledAt, queuedAt time.Time) pluginsCore.PhaseInfo {
					return pluginsCore.PhaseInfoInitializing(queuedAt, pluginsCore.DefaultPhaseVersion, "cluster is creating",
						&pluginsCore.TaskInfo{OccurredAt: &polledAt})
				}, false),
			Entry("spark leaves the stamp to the framework", "qt-recovery-spark",
				func(polledAt, _ time.Time) pluginsCore.PhaseInfo {
					return pluginsCore.PhaseInfoInitializing(polledAt, pluginsCore.DefaultPhaseVersion, "job submitted",
						&pluginsCore.TaskInfo{})
				}, false),
			// The pod plugin reports when the pod itself changed, which can prove the attempt late or in time.
			Entry("a pod that changed inside the budget", "qt-recovery-pod-early",
				func(_, queuedAt time.Time) pluginsCore.PhaseInfo {
					at := queuedAt.Add(time.Second)
					return pluginsCore.PhaseInfoInitializing(at, pluginsCore.DefaultPhaseVersion, "pod initializing",
						&pluginsCore.TaskInfo{OccurredAt: &at})
				}, false),
			Entry("a pod that changed after the budget", "qt-recovery-pod-late",
				func(_, queuedAt time.Time) pluginsCore.PhaseInfo {
					at := queuedAt.Add(5 * time.Second)
					return pluginsCore.PhaseInfoInitializing(at, pluginsCore.DefaultPhaseVersion, "pod initializing",
						&pluginsCore.TaskInfo{OccurredAt: &at})
				}, true),
		)

		It("ends an attempt whose queue budget ran out before its first reconcile without launching it", func() {
			nn := createTaskAction("qt-late-first-reconcile",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			// The controller was away for longer than the 2s budget.
			fakeClock := testingclock.NewFakeClock(base.Add(3 * time.Second))
			// Like core-sleep, the plugin reports Running from its first call.
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{
				runningTransition(base.Add(3 * time.Second)),
			}}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Message).To(Equal("task attempt 1 exceeded its queued timeout of 2s"))
			Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", base.Add(2*time.Second)))
			Expect(fake.handleCalls).To(BeZero(), "nothing is launched for an attempt that has no budget left")
		})

		It("keeps an attempt that started inside its queue budget when a later reconcile is late", func() {
			nn := createTaskAction("qt-started-then-late",
				buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base.Add(time.Second))
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{
				runningTransition(base.Add(time.Second)),
			}}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.AttemptStartedAt).NotTo(BeNil())

			fakeClock.Step(10 * time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(BeZero())
		})

		It("keeps a launched attempt whose serialized cache reservation moved to another owner", func() {
			nn := createTaskAction("qt-cache-moved-reservation", buildSerializedCacheTemplateBytesWithQueuedTimeout())
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			owner := nn.String()
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = launchThenPoll(fakeClock.Now, func() pluginsCore.Transition {
				return runningTransition(base.Add(time.Second))
			})
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			r.Catalog = &stubCatalogClient{
				getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
					return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
				},
				getOrExtendReservationFunc: func(
					context.Context, catalog.Key, string, time.Duration,
				) (*cacheservice.Reservation, error) {
					return &cacheservice.Reservation{OwnerId: owner}, nil
				},
				releaseReservationFunc: func(context.Context, catalog.Key, string) error { return nil },
			}
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req) // launches the resource inside the budget
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.PluginState).NotTo(BeEmpty())

			// The controller was away long enough for the reservation to lapse. A cache wait says
			// nothing about a resource that already exists, so it cannot show that one is still queued.
			owner = "default/another-action"
			fakeClock.Step(10 * time.Second)
			result, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseWaitingForCache.String()))
			Expect(fake.abortCalls).To(BeZero())
			Expect(result).To(Equal(reconcile.Result{RequeueAfter: TaskActionDefaultRequeueDuration}),
				"an expired deadline that cannot be enforced must not spin the queue")

			owner = nn.String()
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(persisted.Status.PluginPhase).To(Equal(pluginsCore.PhaseRunning.String()))
			Expect(persisted.Status.AttemptStartedAt).NotTo(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("ends a cache waiter that never launched when its queue budget runs out", func() {
			nn := createTaskAction("qt-cache-waiter", buildSerializedCacheTemplateBytesWithQueuedTimeout())
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			r.Catalog = &stubCatalogClient{
				getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
					return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
				},
				getOrExtendReservationFunc: func(
					context.Context, catalog.Key, string, time.Duration,
				) (*cacheservice.Reservation, error) {
					return &cacheservice.Reservation{OwnerId: "default/another-action"}, nil
				},
				releaseReservationFunc: func(context.Context, catalog.Key, string) error { return nil },
			}
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseWaitingForCache.String()))

			fakeClock.Step(10 * time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Message).To(ContainSubstring("queued timeout"))
			Expect(fake.handleCalls).To(BeZero(), "a task blocked on the cache is never launched")
		})

		It("ends an attempt whose cache reservation keeps failing before it launched when its queue budget runs out", func() {
			nn := createTaskAction("qt-cache-error-unlaunched", buildSerializedCacheTemplateBytesWithQueuedTimeout())
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			r.Catalog = &stubCatalogClient{
				getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
					return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
				},
				getOrExtendReservationFunc: func(
					context.Context, catalog.Key, string, time.Duration,
				) (*cacheservice.Reservation, error) {
					return nil, stderrors.New("cache service unavailable")
				},
				releaseReservationFunc: func(context.Context, catalog.Key, string) error { return nil },
			}
			req := reconcile.Request{NamespacedName: nn}

			result, err := r.Reconcile(ctx, req) // the reservation fails before the first Handle call
			Expect(err).NotTo(HaveOccurred())
			Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
			Expect(result).To(Equal(reconcile.Result{RequeueAfter: 2 * time.Second}),
				"the requeue still follows the queue budget")

			fakeClock.Step(10 * time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Message).To(ContainSubstring("queued timeout"))
			Expect(fake.handleCalls).To(BeZero(), "a task blocked on a failing cache is never launched")
		})

		It("keeps a launched attempt whose cache reservation fails after its queue budget ran out", func() {
			nn := createTaskAction("qt-cache-error-launched", buildSerializedCacheTemplateBytesWithQueuedTimeout())
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			failing := false
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = launchThenPoll(fakeClock.Now, func() pluginsCore.Transition {
				return runningTransition(base.Add(time.Second))
			})
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			r.Catalog = &stubCatalogClient{
				getFunc: func(context.Context, catalog.Key) (catalog.Entry, error) {
					return catalog.Entry{}, grpcstatus.Error(codes.NotFound, "miss")
				},
				getOrExtendReservationFunc: func(
					context.Context, catalog.Key, string, time.Duration,
				) (*cacheservice.Reservation, error) {
					if failing {
						return nil, stderrors.New("cache service unavailable")
					}
					return &cacheservice.Reservation{OwnerId: nn.String()}, nil
				},
				releaseReservationFunc: func(context.Context, catalog.Key, string) error { return nil },
			}
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req) // launches the resource inside the budget
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.PluginState).NotTo(BeEmpty())

			// A failing cache says nothing about a resource that already exists.
			failing = true
			fakeClock.Step(10 * time.Second)
			result, err := r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
			Expect(result).To(Equal(reconcile.Result{RequeueAfter: TaskActionDefaultRequeueDuration}),
				"an expired deadline that cannot be enforced must not spin the queue")

			failing = false
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseRunning.String()))
			Expect(fake.abortCalls).To(BeZero())
		})

		DescribeTable("judges a launched attempt on its next poll while its events keep failing",
			func(name string, poll func(time.Time) pluginsCore.Transition, wantTimedOut bool) {
				nn := createTaskAction(name, buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
				base := getTaskAction(nn).CreationTimestamp.Time
				fakeClock := testingclock.NewFakeClock(base)
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = launchThenPoll(fakeClock.Now, func() pluginsCore.Transition { return poll(base) })
				recorded := &failingEventsClient{failures: 1000}
				r := newReconciler(fake, fakeClock, recorded, nil)
				req := reconcile.Request{NamespacedName: nn}

				_, err := r.Reconcile(ctx, req) // launches inside the budget, but its Queued event cannot be published
				Expect(err).To(MatchError("events unavailable"))

				fakeClock.Step(10 * time.Second) // the budget runs out while the events service is still away
				_, err = r.Reconcile(ctx, req)
				Expect(err).To(MatchError("events unavailable"))
				persisted := getTaskAction(nn)
				if wantTimedOut {
					Expect(persisted.Status.TimeoutAt).NotTo(BeNil())
					Expect(fake.abortCalls).To(Equal(1))
				} else {
					Expect(persisted.Status.TimeoutAt).To(BeNil())
					Expect(persisted.Status.AttemptStartedAt).NotTo(BeNil())
					Expect(fake.abortCalls).To(BeZero())
				}
				Expect(persisted.Status.PluginState).NotTo(BeEmpty(), "the launch must not be forgotten")

				recorded.failures = 0
				_, err = r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				persisted = getTaskAction(nn)
				Expect(isTerminal(persisted)).To(Equal(wantTimedOut))
				if wantTimedOut {
					Expect(recordedPhases(recorded.RecordedEvents())).To(ContainElement(common.ActionPhase_ACTION_PHASE_TIMED_OUT))
				}
			},
			Entry("still queued", "qt-events-down-queued",
				func(at time.Time) pluginsCore.Transition {
					return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(at, pluginsCore.DefaultPhaseVersion, "scheduling"))
				}, true),
			Entry("already running", "qt-events-down-running",
				func(at time.Time) pluginsCore.Transition { return runningTransition(at.Add(time.Second)) }, false),
		)

		It("still publishes the Queued event of a launch whose first publication failed", func() {
			nn := createTaskAction("qt-launch-event-retried", buildTaskTemplateBytesWithQueuedTimeoutAndRetries(0))
			base := getTaskAction(nn).CreationTimestamp.Time
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = launchThenPoll(fakeClock.Now, func() pluginsCore.Transition {
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(base, pluginsCore.DefaultPhaseVersion, "scheduling"))
			})
			recorded := &failingEventsClient{failures: 1}
			r := newReconciler(fake, fakeClock, recorded, nil)
			req := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, req)
			Expect(err).To(MatchError("events unavailable"))
			Expect(recorded.RecordedEvents()).To(BeEmpty())

			fakeClock.Step(time.Second)
			_, err = r.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(recordedPhases(recorded.RecordedEvents())).To(ContainElement(common.ActionPhase_ACTION_PHASE_QUEUED))
			Expect(getTaskAction(nn).Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))
		})

		It("leaves a valid zero max runtime unlimited", func() {
			startedAt := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(startedAt)
			template := &core.TaskTemplate{}
			Expect(proto.Unmarshal(buildTaskTemplateBytes("timeout-test", "busybox"), template)).To(Succeed())
			template.Metadata.Timeout = durationpb.New(0)
			templateBytes, err := proto.Marshal(template)
			Expect(err).NotTo(HaveOccurred())
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(startedAt)},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction("timeout-zero", templateBytes)
			request := reconcile.Request{NamespacedName: nn}

			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
			// The clock is anchored even without a bound; nothing reads it.
			Expect(getTaskAction(nn).Status.AttemptStartedAt.Time).To(BeTemporally("==", startedAt))

			fakeClock.Step(24 * time.Hour)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.abortCalls).To(BeZero())
			Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
		})

		It("fails malformed task templates and timeout durations as InvalidSpec", func() {
			template := &core.TaskTemplate{}
			Expect(proto.Unmarshal(buildTaskTemplateBytes("timeout-test", "busybox"), template)).To(Succeed())
			template.Metadata.Timeout = &durationpb.Duration{Seconds: 1, Nanos: 1_000_000_000}
			invalidDuration, err := proto.Marshal(template)
			Expect(err).NotTo(HaveOccurred())
			template.Metadata.Timeout = durationpb.New(-time.Second)
			negativeDuration, err := proto.Marshal(template)
			Expect(err).NotTo(HaveOccurred())
			cases := []struct {
				name     string
				template []byte
			}{
				{name: "template", template: []byte{0xff}},
				{name: "duration", template: invalidDuration},
				{name: "negative", template: negativeDuration},
			}

			for _, testCase := range cases {
				fakeClock := testingclock.NewFakeClock(time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC))
				fake := &fakePlugin{id: "timeout-plugin"}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				nn := createTaskAction("timeout-invalid-"+testCase.name, testCase.template)

				_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				Expect(isTerminal(persisted)).To(BeTrue())
				Expect(persisted.Status.Conditions).To(ContainElement(And(
					HaveField("Type", string(flyteorgv1.ConditionTypeFailed)),
					HaveField("Status", metav1.ConditionTrue),
					HaveField("Reason", string(flyteorgv1.ConditionReasonInvalidSpec)),
				)))
				Expect(fake.handleCalls).To(BeZero())
			}
		})

		It("excludes queueing and initialization and requeues at the exact deadline", func() {
			const timeout = 2 * time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			startedAt := base.Add(2 * time.Hour)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(
						base,
						pluginsCore.DefaultPhaseVersion,
						"queued",
					)),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoInitializing(
						base.Add(time.Hour),
						pluginsCore.DefaultPhaseVersion,
						"initializing",
						nil,
					)),
					runningTransition(startedAt),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-deadline",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.AttemptStartedAt).To(BeNil())

			fakeClock.Step(time.Hour)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.AttemptStartedAt).To(BeNil())

			fakeClock.Step(time.Hour)
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(timeout))
			Expect(getTaskAction(nn).Status.AttemptStartedAt.Time).To(BeTemporally("==", startedAt))

			fakeClock.Step(timeout - time.Nanosecond)
			result, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(time.Nanosecond))
			Expect(fake.abortCalls).To(BeZero())

			fakeClock.Step(time.Nanosecond)
			result, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			persisted := getTaskAction(nn)
			Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", startedAt.Add(timeout)))
			Expect(persisted.Status.ErrorState).To(Equal(&flyteorgv1.ErrorState{
				Code:    TaskExecutionTimedOutCode,
				Kind:    "USER",
				Message: "task attempt 1 exceeded its max runtime of 2s",
			}))
			Expect(persisted.Status.Conditions).To(ContainElement(And(
				HaveField("Type", string(flyteorgv1.ConditionTypeFailed)),
				HaveField("Status", metav1.ConditionTrue),
				HaveField("Reason", string(flyteorgv1.ConditionReasonTimedOut)),
			)))
			Expect(fake.abortCalls).To(Equal(1))
			Expect(fake.finalizeCalls).To(Equal(1))
		})

		It("anchors the attempt clock on now when the attempt has no recorded Queued", func() {
			const timeout = time.Hour
			now := time.Date(2026, time.August, 25, 4, 0, 0, 0, time.UTC)
			startedAt := now.Add(-10 * time.Minute)
			fakeClock := testingclock.NewFakeClock(now)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(startedAt)},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-anchor-plugin",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)

			result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
			// PhaseHistory is written after recordAttemptStart, so the first
			// reconcile has no Queued entry to validate the reported time
			// against and falls back to the controller clock.
			Expect(getTaskAction(nn).Status.AttemptStartedAt.Time).To(BeTemporally("==", now))
		})

		It("bootstraps an action already Running from its recorded Executing transition", func() {
			const timeout = 2 * time.Hour
			now := time.Date(2026, time.August, 25, 4, 0, 0, 0, time.UTC)
			queuedAt := now.Add(-2 * time.Hour)
			executingAt := now.Add(-90 * time.Minute)
			fakeClock := testingclock.NewFakeClock(now)
			// A terminal report, so there is no live Running phase to anchor on —
			// the state an upgrade mid-attempt leaves behind.
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(
						&pluginsCore.TaskInfo{OccurredAt: &now},
					)),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-existing-running",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			resource := getTaskAction(nn)
			resource.Status.Attempts = 1
			resource.Status.PluginPhase = pluginsCore.PhaseRunning.String()
			resource.Status.PhaseHistory = []flyteorgv1.PhaseTransition{
				{Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(queuedAt)},
				{Phase: string(flyteorgv1.ConditionReasonExecuting), OccurredAt: metav1.NewTime(executingAt)},
			}
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", executingAt))
			// 90 minutes into a 2 hour budget: the run completed inside it.
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("ignores a reported start time from before the attempt was queued", func() {
			const timeout = time.Hour
			now := time.Date(2026, time.August, 25, 4, 0, 0, 0, time.UTC)
			previousQueuedAt := now.Add(-4 * time.Hour)
			previousExecutingAt := now.Add(-3 * time.Hour)
			retryQueuedAt := now.Add(-10 * time.Minute)
			fakeClock := testingclock.NewFakeClock(now)
			// The plugin reports the *previous* attempt's start time. Taking it at
			// face value would expire attempt 2 the moment it starts.
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(previousExecutingAt)},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-stale-start",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 1),
			)
			resource := getTaskAction(nn)
			resource.Status.Attempts = 2
			resource.Status.PhaseHistory = []flyteorgv1.PhaseTransition{
				{Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(previousQueuedAt)},
				{Phase: string(flyteorgv1.ConditionReasonExecuting), OccurredAt: metav1.NewTime(previousExecutingAt)},
				{Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(retryQueuedAt)},
			}
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", now))
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(BeZero())
		})

		It("uses authoritative completion time after controller downtime", func() {
			const timeout = time.Minute
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			cases := []struct {
				name             string
				completedAt      time.Time
				expectTimedOut   bool
				expectedTerminal common.ActionPhase
			}{
				{
					name:             "before",
					completedAt:      base.Add(timeout - time.Nanosecond),
					expectedTerminal: common.ActionPhase_ACTION_PHASE_SUCCEEDED,
				},
				{
					name:             "at",
					completedAt:      base.Add(timeout),
					expectedTerminal: common.ActionPhase_ACTION_PHASE_SUCCEEDED,
				},
				{
					name:             "after",
					completedAt:      base.Add(timeout + time.Nanosecond),
					expectTimedOut:   true,
					expectedTerminal: common.ActionPhase_ACTION_PHASE_TIMED_OUT,
				},
			}

			for _, testCase := range cases {
				fakeClock := testingclock.NewFakeClock(base)
				fake := &fakePlugin{
					id: "timeout-plugin",
					transitions: []pluginsCore.Transition{
						runningTransition(base),
						pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(
							&pluginsCore.TaskInfo{OccurredAt: &testCase.completedAt},
						)),
					},
				}
				recorded := &recordingEventsClient{}
				r := newReconciler(fake, fakeClock, recorded, nil)
				nn := createTaskAction(
					"timeout-completion-"+testCase.name,
					buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
				)
				request := reconcile.Request{NamespacedName: nn}

				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(timeout + time.Hour)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				persisted := getTaskAction(nn)
				Expect(isTerminal(persisted)).To(BeTrue())
				Expect(fake.handleCalls).To(Equal(2))
				if testCase.expectTimedOut {
					Expect(fake.abortCalls).To(Equal(1))
				} else {
					Expect(fake.abortCalls).To(BeZero())
				}
				allEvents := recorded.RecordedEvents()
				Expect(allEvents[len(allEvents)-1].GetPhase()).To(Equal(testCase.expectedTerminal))
			}
		})

		It("honours a terminal result observed inside the deadline whatever it reports", func() {
			// The bound is enforced against the controller clock. A task seen
			// finishing with half its budget left has not overrun, however
			// unusable the timestamp it came with — a skewed or absent
			// completion time must not manufacture a timeout.
			const timeout = time.Minute
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			cases := []string{"future", "missing", "zero"}

			for i, testCase := range cases {
				startedAt := base.Add(time.Duration(i) * time.Hour)
				futureCompletion := startedAt.Add(timeout + time.Second)
				zeroCompletion := time.Time{}
				info := &pluginsCore.TaskInfo{OccurredAt: &futureCompletion}
				if testCase == "missing" {
					info = &pluginsCore.TaskInfo{}
				} else if testCase == "zero" {
					info = &pluginsCore.TaskInfo{OccurredAt: &zeroCompletion}
				}
				fakeClock := testingclock.NewFakeClock(startedAt)
				fake := &fakePlugin{
					id: "timeout-plugin",
					transitions: []pluginsCore.Transition{
						runningTransition(startedAt),
						pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(info)),
					},
				}
				recorded := &recordingEventsClient{}
				r := newReconciler(fake, fakeClock, recorded, nil)
				nn := createTaskAction(
					"timeout-inside-"+testCase,
					buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
				)
				request := reconcile.Request{NamespacedName: nn}

				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(timeout / 2)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				persisted := getTaskAction(nn)
				Expect(isTerminal(persisted)).To(BeTrue())
				Expect(persisted.Status.TimeoutAt).To(BeNil())
				Expect(fake.abortCalls).To(BeZero())
				allEvents := recorded.RecordedEvents()
				Expect(allEvents[len(allEvents)-1].GetPhase()).To(
					Equal(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
				)
			}
		})

		It("times out a terminal result that cannot prove it finished in time", func() {
			// Same unusable timestamps, but now observed after the deadline: the
			// task cannot show it finished inside its budget, so the bound wins.
			const timeout = time.Minute
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			cases := []string{"future", "missing", "zero"}

			for i, testCase := range cases {
				startedAt := base.Add(time.Duration(i) * time.Hour)
				futureCompletion := startedAt.Add(timeout + time.Second)
				zeroCompletion := time.Time{}
				info := &pluginsCore.TaskInfo{OccurredAt: &futureCompletion}
				if testCase == "missing" {
					info = &pluginsCore.TaskInfo{}
				} else if testCase == "zero" {
					info = &pluginsCore.TaskInfo{OccurredAt: &zeroCompletion}
				}
				fakeClock := testingclock.NewFakeClock(startedAt)
				fake := &fakePlugin{
					id: "timeout-plugin",
					transitions: []pluginsCore.Transition{
						runningTransition(startedAt),
						pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(info)),
					},
				}
				recorded := &recordingEventsClient{}
				r := newReconciler(fake, fakeClock, recorded, nil)
				nn := createTaskAction(
					"timeout-outside-"+testCase,
					buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
				)
				request := reconcile.Request{NamespacedName: nn}

				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(timeout)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())

				persisted := getTaskAction(nn)
				Expect(isTerminal(persisted)).To(BeTrue())
				Expect(persisted.Status.ErrorState.Code).To(Equal(TaskExecutionTimedOutCode))
				Expect(fake.abortCalls).To(Equal(1))
				allEvents := recorded.RecordedEvents()
				Expect(allEvents[len(allEvents)-1].GetPhase()).To(
					Equal(common.ActionPhase_ACTION_PHASE_TIMED_OUT),
				)
			}
		})

		It("resets the runtime per user retry and consumes the configured retry count", func() {
			const timeout = time.Second
			const retries = uint32(2)
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			attemptStarts := map[uint32]time.Time{}
			fake := &fakePlugin{id: "timeout-plugin"}
			fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
				attempt := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetRetryAttempt() + 1
				startedAt, ok := attemptStarts[attempt]
				if !ok {
					startedAt = fakeClock.Now()
					attemptStarts[attempt] = startedAt
				}
				return runningTransition(startedAt), nil
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"timeout-retries",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, retries),
			)
			request := reconcile.Request{NamespacedName: nn}

			for attempt := uint32(1); attempt <= retries+1; attempt++ {
				if attempt > 1 {
					fakeClock.Step(time.Hour)
				}

				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				running := getTaskAction(nn)
				Expect(running.Status.Attempts).To(Equal(attempt))
				Expect(running.Status.AttemptStartedAt.Time).To(BeTemporally("==", fakeClock.Now()))

				fakeClock.Step(timeout)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				afterTimeout := getTaskAction(nn)

				if attempt <= retries {
					Expect(isTerminal(afterTimeout)).To(BeFalse())
					Expect(afterTimeout.Status.Attempts).To(Equal(attempt + 1))
					Expect(afterTimeout.Status.AttemptStartedAt).To(BeNil())
					Expect(afterTimeout.Status.TimeoutAt).To(BeNil())
					Expect(afterTimeout.Status.PluginPhase).To(Equal(pluginsCore.PhaseQueued.String()))
				} else {
					Expect(isTerminal(afterTimeout)).To(BeTrue())
					Expect(afterTimeout.Status.Attempts).To(Equal(retries + 1))
					Expect(afterTimeout.Status.ErrorState.Code).To(Equal(TaskExecutionTimedOutCode))
				}
			}

			events := recorded.RecordedEvents()
			expectedHistory := []struct {
				phase   common.ActionPhase
				attempt uint32
			}{
				{common.ActionPhase_ACTION_PHASE_RUNNING, 1},
				{common.ActionPhase_ACTION_PHASE_TIMED_OUT, 1},
				{common.ActionPhase_ACTION_PHASE_QUEUED, 2},
				{common.ActionPhase_ACTION_PHASE_RUNNING, 2},
				{common.ActionPhase_ACTION_PHASE_TIMED_OUT, 2},
				{common.ActionPhase_ACTION_PHASE_QUEUED, 3},
				{common.ActionPhase_ACTION_PHASE_RUNNING, 3},
				{common.ActionPhase_ACTION_PHASE_TIMED_OUT, 3},
			}
			Expect(events).To(HaveLen(len(expectedHistory)))
			for i, expected := range expectedHistory {
				Expect(events[i].GetPhase()).To(Equal(expected.phase))
				Expect(events[i].GetAttempt()).To(Equal(expected.attempt))
				if expected.phase == common.ActionPhase_ACTION_PHASE_TIMED_OUT {
					Expect(events[i].GetUpdatedTime().AsTime()).To(
						BeTemporally("==", attemptStarts[expected.attempt].Add(timeout)),
					)
				}
			}
			Expect(fake.abortCalls).To(Equal(int(retries + 1)))
			Expect(fake.finalizeCalls).To(Equal(int(retries + 1)))

			handleCalls := fake.handleCalls
			abortCalls := fake.abortCalls
			finalizeCalls := fake.finalizeCalls
			restarted := newReconciler(fake, fakeClock, recorded, nil)
			_, err := restarted.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(handleCalls))
			Expect(fake.abortCalls).To(Equal(abortCalls))
			Expect(fake.finalizeCalls).To(Equal(finalizeCalls))
		})

		It("retries Abort failures without publishing terminal timeout", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(base)},
				abortErrors: []error{stderrors.New("abort failed"), nil},
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"timeout-abort-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(timeout)
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))

			pending := getTaskAction(nn)
			Expect(pending.Status.TimeoutAt).NotTo(BeNil())
			Expect(isTerminal(pending)).To(BeFalse())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(Equal(1))
			Expect(fake.finalizeCalls).To(BeZero())
			for _, event := range recorded.RecordedEvents() {
				Expect(event.GetPhase()).NotTo(Equal(common.ActionPhase_ACTION_PHASE_TIMED_OUT))
			}

			restarted := newReconciler(fake, fakeClock, recorded, nil)
			_, err = restarted.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(Equal(2))
			Expect(fake.finalizeCalls).To(Equal(1))
		})

		It("retries Finalize failures without publishing terminal timeout", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id:             "timeout-plugin",
				transitions:    []pluginsCore.Transition{runningTransition(base)},
				finalizeErrors: []error{stderrors.New("finalize failed"), nil},
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"timeout-finalize-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(timeout)
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))

			pending := getTaskAction(nn)
			Expect(pending.Status.TimeoutAt).NotTo(BeNil())
			Expect(isTerminal(pending)).To(BeFalse())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(Equal(1))
			Expect(fake.finalizeCalls).To(Equal(1))
			for _, event := range recorded.RecordedEvents() {
				Expect(event.GetPhase()).NotTo(Equal(common.ActionPhase_ACTION_PHASE_TIMED_OUT))
			}

			restarted := newReconciler(fake, fakeClock, recorded, nil)
			_, err = restarted.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(Equal(2))
			Expect(fake.finalizeCalls).To(Equal(2))
		})

		It("persists attempt time independently from ActionEvent publication", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(base)},
			}
			failingEvents := &failingEventsClient{failures: 1}
			r := newReconciler(fake, fakeClock, failingEvents, nil)
			nn := createTaskAction(
				"timeout-event-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).To(MatchError("events unavailable"))
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", base))
			Expect(persisted.Status.StateJSON).To(BeEmpty())

			fakeClock.Step(timeout)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", base.Add(timeout)))
			Expect(fake.abortCalls).To(Equal(1))
		})

		It("waits out the retry backoff before relaunching a failed attempt", func() {
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			failed := pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("UserError", "task raised", nil))
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{failed}}
			r := newReconciler(fake, fakeClock, &fakeEventsClient{}, nil)
			backoff := &core.Backoff{
				Base:   durationpb.New(10 * time.Second),
				Factor: proto.Float64(2),
				Cap:    durationpb.New(10 * time.Minute),
			}
			nn := createTaskAction(
				"retry-backoff",
				withRetryBackoff(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 2), backoff),
			)
			request := reconcile.Request{NamespacedName: nn}

			// The first attempt fails: its retry is due in 10s and the plugin stays untouched until then.
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(10 * time.Second))
			persisted := getTaskAction(nn)
			Expect(persisted.Status.Attempts).To(Equal(uint32(2)))
			Expect(persisted.Status.NextAttemptAt.Time).To(BeTemporally("==", base.Add(10*time.Second)))
			Expect(fake.handleCalls).To(Equal(1))

			fakeClock.Step(4 * time.Second)
			result, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(6 * time.Second))
			Expect(fake.handleCalls).To(Equal(1))

			// Once due, the retry launches; its own failure waits twice as long.
			fakeClock.Step(6 * time.Second)
			result, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(result.RequeueAfter).To(Equal(20 * time.Second))
			persisted = getTaskAction(nn)
			Expect(persisted.Status.Attempts).To(Equal(uint32(3)))
			Expect(persisted.Status.NextAttemptAt.Time).To(BeTemporally("==", base.Add(30*time.Second)))

			// The last attempt exhausts the retries: terminal, with nothing left to schedule.
			fakeClock.Step(20 * time.Second)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(3))
			persisted = getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.NextAttemptAt).To(BeNil())
		})

		It("waits out the retry backoff after a max-runtime timeout", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{runningTransition(base)}}
			r := newReconciler(fake, fakeClock, &fakeEventsClient{}, nil)
			backoff := &core.Backoff{
				Base:   durationpb.New(10 * time.Second),
				Factor: proto.Float64(2),
				Cap:    durationpb.New(10 * time.Minute),
			}
			nn := createTaskAction(
				"timeout-retry-backoff",
				withRetryBackoff(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 1), backoff),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(1))
			running := getTaskAction(nn)
			Expect(running.Status.AttemptStartedAt.Time).To(BeTemporally("==", base))
			Expect(running.Status.NextAttemptAt).To(BeNil())

			// The attempt times out and is retried: the retry is due 10s after the timeout reconcile.
			fakeClock.Step(timeout)
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(10 * time.Second))
			retrying := getTaskAction(nn)
			Expect(retrying.Status.Attempts).To(Equal(uint32(2)))
			Expect(retrying.Status.TimeoutAt).To(BeNil())
			Expect(retrying.Status.NextAttemptAt.Time).To(BeTemporally("==", fakeClock.Now().Add(10*time.Second)))
			Expect(fake.abortCalls).To(Equal(1))
			// The timeout reconcile handled the plugin once more before enforcing the deadline.
			handled := fake.handleCalls

			fakeClock.Step(4 * time.Second)
			result, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(6 * time.Second))
			Expect(fake.handleCalls).To(Equal(handled))

			fakeClock.Step(6 * time.Second)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(handled + 1))
			Expect(getTaskAction(nn).Status.NextAttemptAt).To(BeNil())
		})

		It("retains the timeout deadline when retry event publication fails", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(base)},
			}
			failingEvents := &failingEventsClient{}
			r := newReconciler(fake, fakeClock, failingEvents, nil)
			nn := createTaskAction(
				"timeout-retry-event-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 1),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			failingEvents.failures = 1
			fakeClock.Step(timeout)
			result, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))

			pending := getTaskAction(nn)
			Expect(pending.Status.Attempts).To(Equal(uint32(1)))
			Expect(pending.Status.TimeoutAt.Time).To(BeTemporally("==", base.Add(timeout)))
			Expect(isTerminal(pending)).To(BeFalse())
			Expect(fake.abortCalls).To(Equal(1))
			Expect(fake.finalizeCalls).To(Equal(1))

			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			retrying := getTaskAction(nn)
			Expect(retrying.Status.Attempts).To(Equal(uint32(2)))
			Expect(retrying.Status.AttemptStartedAt).To(BeNil())
			Expect(retrying.Status.TimeoutAt).To(BeNil())
			Expect(fake.abortCalls).To(Equal(2))
			Expect(fake.finalizeCalls).To(Equal(2))
		})

		It("retries status persistence and anchors the clock on the retrying reconcile", func() {
			const timeout = time.Second
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id:          "timeout-plugin",
				transitions: []pluginsCore.Transition{runningTransition(base)},
			}
			failingClient := &failingStatusClient{Client: k8sClient, failUpdates: true}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, failingClient)
			nn := createTaskAction(
				"timeout-status-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).To(MatchError("status unavailable"))
			Expect(getTaskAction(nn).Status.AttemptStartedAt).To(BeNil())

			// The retry anchors on the controller clock: the failed reconcile left
			// no Queued entry to validate the plugin's reported time against.
			fakeClock.Step(timeout)
			failingClient.failUpdates = false
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", base.Add(timeout)))
			// The deadline is read before recordAttemptStart, so enforcement starts
			// on the following reconcile.
			Expect(persisted.Status.TimeoutAt).To(BeNil())

			fakeClock.Step(timeout)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			persisted = getTaskAction(nn)
			Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", base.Add(2*timeout)))
			Expect(isTerminal(persisted)).To(BeTrue())
		})

		It("does not misclassify ordinary failure as timeout", func() {
			const timeout = time.Hour
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure(
						"OrdinaryFailure",
						"ordinary failure",
						&pluginsCore.TaskInfo{OccurredAt: &base},
					)),
				},
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"timeout-ordinary-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
			)

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Code).To(Equal("OrdinaryFailure"))
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(recorded.RecordedEvents()).To(HaveLen(1))
			Expect(recorded.RecordedEvents()[0].GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_FAILED))
			Expect(fake.abortCalls).To(BeZero())
		})

		It("does not consume a user attempt for a system retry", func() {
			const timeout = time.Hour
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			systemFailureAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
						"TransientSystemFailure",
						"transient failure",
						&pluginsCore.TaskInfo{OccurredAt: &systemFailureAt},
					)),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"timeout-system-retry",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Minute)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			persisted := getTaskAction(nn)
			Expect(persisted.Status.Attempts).To(Equal(uint32(1)))
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(1)))
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", base))
			Expect(persisted.Status.TimeoutAt).To(BeNil())
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(fake.abortCalls).To(Equal(1))
		})

		It("reports the system failure before relaunching the attempt in place", func() {
			base := time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			preemptedAt := base.Add(time.Minute)
			const code = "WorkloadEvictedDueToPreempted"
			const message = "Pod was terminated by an external controller: Preempted to accommodate a workload"
			fake := &fakePlugin{
				id: "kueue-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
						code,
						message,
						&pluginsCore.TaskInfo{OccurredAt: &preemptedAt},
					)),
				},
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"system-retry-reported",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Minute)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			events := recorded.RecordedEvents()
			Expect(len(events)).To(BeNumerically(">=", 2))
			Expect(events[len(events)-2].GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_RUNNING))
			retry := events[len(events)-1]
			Expect(retry.GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_QUEUED))
			Expect(retry.GetAttempt()).To(Equal(uint32(1)))
			// Events are stored under (action, attempt, phase, version) and a duplicate key
			// is dropped, so this event must not reuse a version the launch events of the
			// same attempt already spent in the same phase. Plugins count up from zero.
			Expect(retry.GetVersion()).To(Equal(systemRetryEventVersionBase))
			Expect(retry.GetVersion()).To(BeNumerically(">", 1000))
			// The store keeps this column in a signed 32-bit integer.
			Expect(retry.GetVersion()).To(BeNumerically("<", uint32(1)<<31))
			Expect(retry.GetErrorInfo()).To(BeNil())
			Expect(retry.GetClusterEvents()).To(HaveLen(2))
			Expect(retry.GetClusterEvents()[0].GetMessage()).To(Equal(systemRetryReason))
			warning := retry.GetClusterEvents()[1]
			Expect(warning.GetType()).To(Equal(workflow.ClusterEvent_TYPE_WARNING))
			Expect(warning.GetReason()).To(Equal(code))
			Expect(warning.GetMessage()).To(Equal("[" + code + "] " + message))
			Expect(warning.GetOccurredAt().AsTime()).To(BeTemporally("==", preemptedAt))

			persisted := getTaskAction(nn)
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(1)))
			Expect(persisted.Status.SystemRetries).To(Equal(uint32(1)))
			Expect(persisted.Status.PluginState).To(BeNil())
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(fake.abortCalls).To(Equal(1))
		})

		It("gives every system retry of an attempt its own event version", func() {
			base := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			preempted := func(at time.Time) pluginsCore.Transition {
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
					"WorkloadEvictedDueToPreempted",
					"preempted",
					&pluginsCore.TaskInfo{OccurredAt: &at},
				))
			}
			// Preempted, relaunched and running again, then preempted a second time. Both
			// failures belong to the same attempt, so the progress in between does not
			// reset the system-failure count.
			fake := &fakePlugin{
				id: "kueue-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					preempted(base.Add(time.Minute)),
					runningTransition(base.Add(2 * time.Minute)),
					preempted(base.Add(3 * time.Minute)),
				},
			}
			recorded := &recordingEventsClient{}
			r := newReconciler(fake, fakeClock, recorded, nil)
			nn := createTaskAction(
				"system-retry-twice",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			for i := 0; i < len(fake.transitions); i++ {
				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(time.Minute)
			}

			// Both retries belong to attempt 1 and report Queued, so the version is all
			// that keeps the store from dropping the second as a duplicate of the first.
			retries := systemRetryEvents(recorded.RecordedEvents())
			Expect(retries).To(HaveLen(2))
			for _, retry := range retries {
				Expect(retry.GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_QUEUED))
				Expect(retry.GetAttempt()).To(Equal(uint32(1)))
			}
			Expect(retries[0].GetVersion()).To(Equal(systemRetryEventVersionBase))
			Expect(retries[1].GetVersion()).To(Equal(systemRetryEventVersionBase + 1))

			persisted := getTaskAction(nn)
			Expect(persisted.Status.Attempts).To(Equal(uint32(1)))
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(2)), "both failures count against the attempt")
			Expect(persisted.Status.SystemRetries).To(Equal(uint32(2)))
			Expect(isTerminal(persisted)).To(BeFalse())
		})

		It("fails the TaskAction once an attempt is relaunched past the system-failure limit", func() {
			base := time.Date(2026, time.October, 24, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			deleted := func(at time.Time) pluginsCore.Transition {
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
					"ResourceDeletedExternally",
					"pod deleted",
					&pluginsCore.TaskInfo{OccurredAt: &at},
				))
			}
			// The pod is deleted every time it starts running. Each relaunch reports
			// Running again, which must not reset the count, or the pod would be
			// relaunched forever.
			var transitions []pluginsCore.Transition
			for i := 0; i <= int(DefaultMaxSystemFailures); i++ {
				transitions = append(transitions,
					runningTransition(base.Add(time.Duration(2*i)*time.Minute)),
					deleted(base.Add(time.Duration(2*i+1)*time.Minute)),
				)
			}
			fake := &fakePlugin{id: "k8s-plugin", transitions: transitions}
			r := newReconciler(fake, fakeClock, &fakeEventsClient{}, nil)
			nn := createTaskAction(
				"system-retry-limit",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			for i := 0; i < len(transitions); i++ {
				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(time.Minute)
				if i < len(transitions)-1 {
					Expect(isTerminal(getTaskAction(nn))).To(BeFalse(), "reconcile %d", i)
				}
			}

			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState).NotTo(BeNil())
			Expect(persisted.Status.ErrorState.Code).To(Equal(MaxSystemFailuresExceededCode))
			Expect(persisted.Status.SystemFailures).To(Equal(DefaultMaxSystemFailures + 1))
			Expect(persisted.Status.Attempts).To(Equal(uint32(1)), "system failures do not consume user retries")
		})

		It("resets the system-failure count when a user failure starts a new attempt", func() {
			base := time.Date(2026, time.October, 24, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			preemptedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: "k8s-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
						"WorkloadEvictedDueToPreempted",
						"preempted",
						&pluginsCore.TaskInfo{OccurredAt: &preemptedAt},
					)),
					runningTransition(base.Add(2 * time.Minute)),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("UserError", "task raised", nil)),
				},
			}
			r := newReconciler(fake, fakeClock, &fakeEventsClient{}, nil)
			nn := createTaskAction(
				"system-failures-new-attempt",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			for i := 0; i < 3; i++ {
				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(time.Minute)
			}
			Expect(getTaskAction(nn).Status.SystemFailures).To(Equal(uint32(1)), "running again does not reset the count")

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.Attempts).To(Equal(uint32(2)))
			Expect(persisted.Status.SystemFailures).To(BeZero(), "the new attempt starts with a clean count")
			Expect(persisted.Status.SystemRetries).To(Equal(uint32(1)))
			Expect(isTerminal(persisted)).To(BeFalse())
		})

		It("republishes a system retry under the same version until it is persisted", func() {
			base := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			preemptedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: "kueue-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
						"WorkloadEvictedDueToPreempted",
						"preempted",
						&pluginsCore.TaskInfo{OccurredAt: &preemptedAt},
					)),
				},
			}
			recorded := &recordingEventsClient{}
			failingClient := &failingStatusClient{Client: k8sClient}
			r := newReconciler(fake, fakeClock, recorded, failingClient)
			nn := createTaskAction(
				"system-retry-republished",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			// The event goes out but the status does not, so the next reconcile meets the
			// same failure with the same counters and publishes it again.
			fakeClock.Step(time.Minute)
			failingClient.failUpdates = true
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(getTaskAction(nn).Status.SystemRetries).To(BeZero())

			fakeClock.Step(time.Minute)
			failingClient.failUpdates = false
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			retries := systemRetryEvents(recorded.RecordedEvents())
			Expect(retries).To(HaveLen(2))
			Expect(retries[0].GetVersion()).To(Equal(systemRetryEventVersionBase))
			Expect(retries[1].GetVersion()).To(Equal(systemRetryEventVersionBase), "the store drops it as a duplicate")
			Expect(getTaskAction(nn).Status.SystemRetries).To(Equal(uint32(1)))
		})

		It("retries even when the system retry event cannot be published", func() {
			base := time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			failedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: "kueue-plugin",
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoSystemRetryableFailure(
						"WorkloadEvictedDueToPreempted",
						"preempted",
						&pluginsCore.TaskInfo{OccurredAt: &failedAt},
					)),
				},
			}
			// The event store rejects the retry event, as it does when its version cannot
			// be encoded. The attempt must still be retried rather than held back.
			rejecting := &retryEventRejectingClient{}
			r := newReconciler(fake, fakeClock, rejecting, nil)
			nn := createTaskAction(
				"system-retry-event-unavailable",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Minute)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			Expect(rejecting.rejected).To(BeNumerically(">=", 1), "the retry event was rejected")
			persisted := getTaskAction(nn)
			Expect(persisted.Status.SystemFailures).To(Equal(uint32(1)), "the retry proceeded")
			Expect(persisted.Status.PluginState).To(BeNil(), "the plugin resource was reset")
			Expect(isTerminal(persisted)).To(BeFalse())
			Expect(fake.abortCalls).To(Equal(1))
		})

		const cleanupPluginID = "cleanup-plugin"

		It("cleans up the resource when a retryable failure with cleanup has no retries left", func() {
			base := time.Date(2026, time.October, 7, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			evictedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: cleanupPluginID,
				transitions: []pluginsCore.Transition{
					runningTransition(base),
					pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailureWithCleanup(
						"GangEvicted", "gang evicted by kueue (Preempted)", &pluginsCore.TaskInfo{OccurredAt: &evictedAt},
					)),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"cleanup-no-retries-left",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 0),
			)
			request := reconcile.Request{NamespacedName: nn}

			_, err := r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			fakeClock.Step(time.Minute)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())

			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Code).To(Equal("GangEvicted"))
			Expect(fake.abortCalls).To(Equal(1), "the resource is deleted so it cannot run again")
		})

		It("cleans up the resource when the plugin reports a permanent failure with cleanup", func() {
			base := time.Date(2026, time.October, 7, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			failedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: cleanupPluginID,
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoFailed(
						pluginsCore.PhasePermanentFailure,
						&core.ExecutionError{Kind: core.ExecutionError_USER, Code: "KueueLocalQueueNotFound", Message: "no such queue"},
						&pluginsCore.TaskInfo{OccurredAt: &failedAt},
					).WithCleanupOnFailure()),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"cleanup-permanent-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			persisted := getTaskAction(nn)
			Expect(isTerminal(persisted)).To(BeTrue())
			Expect(persisted.Status.ErrorState.Code).To(Equal("KueueLocalQueueNotFound"))
			Expect(persisted.Status.Attempts).To(Equal(uint32(1)), "a permanent failure is not retried")
			Expect(fake.abortCalls).To(Equal(1))
		})

		It("leaves the resource alone for a permanent failure without cleanup", func() {
			base := time.Date(2026, time.October, 7, 0, 0, 0, 0, time.UTC)
			fakeClock := testingclock.NewFakeClock(base)
			failedAt := base.Add(time.Minute)
			fake := &fakePlugin{
				id: cleanupPluginID,
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoFailure(
						"OrdinaryFailure", "ordinary failure", &pluginsCore.TaskInfo{OccurredAt: &failedAt},
					)),
				},
			}
			r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
			nn := createTaskAction(
				"no-cleanup-permanent-failure",
				buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", time.Hour, 2),
			)

			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
			Expect(fake.abortCalls).To(BeZero())
		})

		Context("action deadline", func() {
			withDeadline := func(template []byte, deadline time.Duration) []byte {
				tmpl := &core.TaskTemplate{}
				Expect(proto.Unmarshal(template, tmpl)).To(Succeed())
				tmpl.Metadata.Timeouts = &core.TimeoutStrategy{Deadline: durationpb.New(deadline)}
				data, err := proto.Marshal(tmpl)
				Expect(err).NotTo(HaveOccurred())
				return data
			}

			// The deadline is anchored on CreationTimestamp, which the API server sets from
			// the real clock, so each test starts its fake clock there.
			createdAt := func(nn types.NamespacedName) time.Time {
				return getTaskAction(nn).CreationTimestamp.Time
			}

			queuedTransition := func(at time.Time) pluginsCore.Transition {
				return pluginsCore.DoTransition(pluginsCore.PhaseInfoQueued(at, pluginsCore.DefaultPhaseVersion, "waiting"))
			}

			expectDeadlineExceeded := func(ta *flyteorgv1.TaskAction) {
				Expect(isTerminal(ta)).To(BeTrue())
				Expect(ta.Status.ErrorState).NotTo(BeNil())
				Expect(ta.Status.ErrorState.Code).To(Equal(TaskDeadlineExceededCode))
				Expect(ta.Status.ErrorState.Kind).To(Equal("USER"))
				Expect(ta.Status.Conditions).To(ContainElement(And(
					HaveField("Type", string(flyteorgv1.ConditionTypeFailed)),
					HaveField("Status", metav1.ConditionTrue),
					HaveField("Reason", string(flyteorgv1.ConditionReasonTimedOut)),
				)))
			}

			It("counts queue time and times out an action that never started", func() {
				const deadline = 30 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					return queuedTransition(fakeClock.Now()), nil
				}
				recorded := &recordingEventsClient{}
				r := newReconciler(fake, fakeClock, recorded, nil)
				nn := createTaskAction(
					"deadline-queued",
					withDeadline(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 2), deadline),
				)
				created := createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				// Close to the deadline, the requeue is cut short so the deadline is not overshot.
				fakeClock.SetTime(created.Add(25 * time.Second))
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(5 * time.Second))
				Expect(fake.handleCalls).To(Equal(1))

				fakeClock.SetTime(created.Add(deadline))
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				expectDeadlineExceeded(persisted)
				Expect(persisted.Status.Attempts).To(Equal(uint32(1)))
				Expect(persisted.Status.AttemptStartedAt).To(BeNil())
				Expect(fake.handleCalls).To(Equal(1))
				Expect(fake.abortCalls).To(Equal(1))
				Expect(fake.finalizeCalls).To(Equal(1))

				events := recorded.RecordedEvents()
				last := events[len(events)-1]
				Expect(last.GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_TIMED_OUT))
				Expect(last.GetUpdatedTime().AsTime()).To(BeTemporally("==", created.Add(deadline)))
				Expect(last.GetErrorInfo().GetCode()).To(Equal(TaskDeadlineExceededCode))
			})

			It("spans attempts and is not reset by a user retry", func() {
				const deadline = 30 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					if tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetRetryAttempt() == 0 {
						return pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("UserError", "task raised", nil)), nil
					}
					return runningTransition(fakeClock.Now()), nil
				}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				nn := createTaskAction(
					"deadline-spans-attempts",
					withDeadline(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 3), deadline),
				)
				created := createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				// Attempt 1 fails 10s in and is retried at once.
				fakeClock.SetTime(created.Add(10 * time.Second))
				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(getTaskAction(nn).Status.Attempts).To(Equal(uint32(2)))

				// Attempt 2 starts running; its requeue is bounded by what is left of the deadline.
				fakeClock.SetTime(created.Add(20 * time.Second))
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(10 * time.Second))
				Expect(isTerminal(getTaskAction(nn))).To(BeFalse())

				// Retries remain, but the deadline ends the action anyway.
				fakeClock.SetTime(created.Add(deadline))
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				expectDeadlineExceeded(persisted)
				Expect(persisted.Status.Attempts).To(Equal(uint32(2)))
				Expect(persisted.Status.ErrorState.Message).To(ContainSubstring("deadline of 30s across all attempts (attempt 2)"))
			})

			It("cuts a retry backoff short and times out without relaunching", func() {
				const deadline = 30 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				failed := pluginsCore.DoTransition(pluginsCore.PhaseInfoRetryableFailure("UserError", "task raised", nil))
				fake := &fakePlugin{id: "timeout-plugin", transitions: []pluginsCore.Transition{failed}}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				backoff := &core.Backoff{Base: durationpb.New(time.Minute)}
				nn := createTaskAction(
					"deadline-backoff",
					withDeadline(
						withRetryBackoff(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 2), backoff),
						deadline,
					),
				)
				created := createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				fakeClock.SetTime(created)
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(getTaskAction(nn).Status.NextAttemptAt.Time).To(BeTemporally("==", created.Add(time.Minute)))
				Expect(result.RequeueAfter).To(Equal(deadline))

				fakeClock.SetTime(created.Add(10 * time.Second))
				result, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(20 * time.Second))
				Expect(fake.handleCalls).To(Equal(1))

				fakeClock.SetTime(created.Add(deadline))
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				expectDeadlineExceeded(persisted)
				Expect(persisted.Status.Attempts).To(Equal(uint32(2)))
				Expect(persisted.Status.NextAttemptAt).To(BeNil())
				Expect(fake.handleCalls).To(Equal(1))
			})

			It("keeps max-runtime retries until the deadline ends the action", func() {
				const timeout = 5 * time.Second
				const deadline = 12 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				attemptStarts := map[uint32]time.Time{}
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					attempt := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().GetRetryAttempt() + 1
					startedAt, ok := attemptStarts[attempt]
					if !ok {
						startedAt = fakeClock.Now()
						attemptStarts[attempt] = startedAt
					}
					return runningTransition(startedAt), nil
				}
				recorded := &recordingEventsClient{}
				r := newReconciler(fake, fakeClock, recorded, nil)
				nn := createTaskAction(
					"deadline-after-max-runtime",
					withDeadline(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 5), deadline),
				)
				created := createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				fakeClock.SetTime(created)
				for attempt := uint32(1); attempt <= 2; attempt++ {
					_, err := r.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
					fakeClock.Step(timeout)
					_, err = r.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
					persisted := getTaskAction(nn)
					Expect(isTerminal(persisted)).To(BeFalse())
					Expect(persisted.Status.Attempts).To(Equal(attempt + 1))
				}

				// Attempt 3 would run until created+15s, but the deadline at created+12s comes first.
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(2 * time.Second))

				fakeClock.Step(2 * time.Second)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				expectDeadlineExceeded(persisted)
				Expect(persisted.Status.Attempts).To(Equal(uint32(3)))

				var timeoutCodes []string
				for _, event := range recorded.RecordedEvents() {
					if event.GetPhase() == common.ActionPhase_ACTION_PHASE_TIMED_OUT {
						timeoutCodes = append(timeoutCodes, event.GetErrorInfo().GetCode())
					}
				}
				Expect(timeoutCodes).To(Equal([]string{
					TaskExecutionTimedOutCode,
					TaskExecutionTimedOutCode,
					TaskDeadlineExceededCode,
				}))
			})

			It("lets a result that finished before the deadline win over a late reconcile", func() {
				const deadline = 30 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				var created time.Time
				fake := &fakePlugin{id: "timeout-plugin"}
				fake.handleFunc = func(_ context.Context, tCtx pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					Expect(tCtx.PluginStateWriter().Put(1, "launched")).To(Succeed())
					if fake.handleCalls == 1 {
						return runningTransition(fakeClock.Now()), nil
					}
					finishedAt := created.Add(20 * time.Second)
					return pluginsCore.DoTransition(pluginsCore.PhaseInfoSuccess(&pluginsCore.TaskInfo{OccurredAt: &finishedAt})), nil
				}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				nn := createTaskAction(
					"deadline-success-before",
					withDeadline(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 0), deadline),
				)
				created = createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				fakeClock.SetTime(created)
				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(getTaskAction(nn).Status.PluginState).NotTo(BeEmpty())

				// The controller only gets back to the action after the deadline.
				fakeClock.SetTime(created.Add(time.Minute))
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				persisted := getTaskAction(nn)
				Expect(persisted.Status.ErrorState).To(BeNil())
				Expect(persisted.Status.Conditions).To(ContainElement(And(
					HaveField("Type", string(flyteorgv1.ConditionTypeSucceeded)),
					HaveField("Status", metav1.ConditionTrue),
				)))
				Expect(fake.abortCalls).To(BeZero())
			})

			It("resumes deadline cleanup after an Abort failure", func() {
				const deadline = 30 * time.Second
				fakeClock := testingclock.NewFakeClock(time.Now())
				fake := &fakePlugin{
					id:          "timeout-plugin",
					abortErrors: []error{stderrors.New("abort failed")},
				}
				fake.handleFunc = func(context.Context, pluginsCore.TaskExecutionContext) (pluginsCore.Transition, error) {
					return queuedTransition(fakeClock.Now()), nil
				}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				nn := createTaskAction(
					"deadline-abort-failure",
					withDeadline(buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, 2), deadline),
				)
				created := createdAt(nn)
				request := reconcile.Request{NamespacedName: nn}

				fakeClock.SetTime(created.Add(deadline))
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
				pending := getTaskAction(nn)
				Expect(isTerminal(pending)).To(BeFalse())
				Expect(pending.Status.TimeoutAt.Time).To(BeTemporally("==", created.Add(deadline)))

				fakeClock.Step(TaskActionDefaultRequeueDuration)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				expectDeadlineExceeded(getTaskAction(nn))
				Expect(fake.abortCalls).To(Equal(2))
				Expect(fake.handleCalls).To(BeZero())
			})

			Context("with a queued timeout", func() {
				withQueuedTimeout := func(template []byte, queued time.Duration) []byte {
					tmpl := &core.TaskTemplate{}
					Expect(proto.Unmarshal(template, tmpl)).To(Succeed())
					tmpl.Metadata.Timeouts.QueuedTimeout = durationpb.New(queued)
					data, err := proto.Marshal(tmpl)
					Expect(err).NotTo(HaveOccurred())
					return data
				}

				newQueuedAction := func(
					name string, queued, deadline time.Duration, retries uint32,
				) (types.NamespacedName, *fakePlugin, *testingclock.FakeClock, *recordingEventsClient, *TaskActionReconciler) {
					fakeClock := testingclock.NewFakeClock(time.Now())
					fake := &fakePlugin{id: "timeout-plugin"}
					fake.handleFunc = launchThenPoll(fakeClock.Now, func() pluginsCore.Transition {
						return queuedTransition(fakeClock.Now())
					})
					recorded := &recordingEventsClient{}
					r := newReconciler(fake, fakeClock, recorded, nil)
					template := withDeadline(
						buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", 0, retries), deadline)
					nn := createTaskAction(name, withQueuedTimeout(template, queued))
					return nn, fake, fakeClock, recorded, r
				}

				DescribeTable("ends the action with its deadline when the queue budget also ran out",
					func(name string, launchFirst bool) {
						const queued = 2 * time.Second
						const deadline = 10 * time.Second
						nn, fake, fakeClock, recorded, r := newQueuedAction(name, queued, deadline, 2)
						created := createdAt(nn)
						request := reconcile.Request{NamespacedName: nn}

						if launchFirst {
							fakeClock.SetTime(created)
							_, err := r.Reconcile(ctx, request)
							Expect(err).NotTo(HaveOccurred())
							Expect(getTaskAction(nn).Status.PluginState).NotTo(BeEmpty())
						}

						// Past both the queue budget and the deadline, retries remain, and the deadline still ends the action.
						fakeClock.SetTime(created.Add(time.Minute))
						_, err := r.Reconcile(ctx, request)
						Expect(err).NotTo(HaveOccurred())
						persisted := getTaskAction(nn)
						expectDeadlineExceeded(persisted)
						Expect(observedAttempts(persisted)).To(Equal(uint32(1)), "the deadline ends the action without a retry")
						Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", created.Add(deadline)))
						events := recorded.RecordedEvents()
						Expect(events[len(events)-1].GetErrorInfo().GetCode()).To(Equal(TaskDeadlineExceededCode))
						if launchFirst {
							Expect(fake.handleCalls).To(Equal(2))
						} else {
							Expect(fake.handleCalls).To(BeZero(), "an attempt past its deadline is not launched")
						}
					},
					Entry("an attempt that launched inside the budget", "qt-deadline-launched", true),
					Entry("an attempt that never launched", "qt-deadline-unlaunched", false),
				)

				It("ends a queued attempt with its queued timeout while the deadline is still ahead", func() {
					const queued = 2 * time.Second
					nn, _, fakeClock, _, r := newQueuedAction("qt-queued-first", queued, 30*time.Second, 0)
					created := createdAt(nn)
					request := reconcile.Request{NamespacedName: nn}

					fakeClock.SetTime(created)
					_, err := r.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())

					fakeClock.SetTime(created.Add(queued))
					_, err = r.Reconcile(ctx, request)
					Expect(err).NotTo(HaveOccurred())
					persisted := getTaskAction(nn)
					Expect(isTerminal(persisted)).To(BeTrue())
					Expect(persisted.Status.ErrorState.Code).To(Equal(TaskExecutionTimedOutCode))
					Expect(persisted.Status.ErrorState.Message).To(Equal("task attempt 1 exceeded its queued timeout of 2s"))
					Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", created.Add(queued)))
				})

				DescribeTable("requeues a queued attempt for whichever of its queue budget and deadline ends first",
					func(name string, queued, deadline, wantRequeue time.Duration) {
						nn, _, fakeClock, _, r := newQueuedAction(name, queued, deadline, 0)
						fakeClock.SetTime(createdAt(nn).Add(time.Second))
						result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
						Expect(err).NotTo(HaveOccurred())
						Expect(result.RequeueAfter).To(Equal(wantRequeue))
					},
					Entry("the queue budget", "qt-requeue-queued", 2*time.Second, 30*time.Second, time.Second),
					Entry("the deadline", "qt-requeue-deadline", 20*time.Second, 5*time.Second, 4*time.Second),
				)
			})
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
					Name:       abortResourceName,
					Namespace:  "default",
					Finalizers: []string{taskActionFinalizer},
				},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "abort-run",
					Project:       "abort-project",
					Domain:        "abort-domain",
					ActionName:    "abort-action",
					InputURI:      "/tmp/input",
					RunOutputBase: "/tmp/output",
					TaskType:      "python",
					TaskTemplate:  buildTaskTemplateBytes("python", "python:3.11"),
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
				Recorder:       events.NewFakeRecorder(10),
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

		It("should finalize a terminal TaskAction without aborting it or emitting ACTION_PHASE_ABORTED", func() {
			fake := &fakePlugin{id: "timeout-plugin"}
			terminalResource := &flyteorgv1.TaskAction{
				ObjectMeta: metav1.ObjectMeta{
					Name:       "abort-terminal-resource",
					Namespace:  "default",
					Finalizers: []string{taskActionFinalizer},
				},
				Spec: flyteorgv1.TaskActionSpec{
					RunName:       "terminal-run",
					Project:       "terminal-project",
					Domain:        "terminal-domain",
					ActionName:    "terminal-action",
					InputURI:      "/tmp/input",
					RunOutputBase: "/tmp/output",
					TaskType:      "timeout-test",
					TaskTemplate:  buildTaskTemplateBytes("timeout-test", "busybox"),
				},
			}
			Expect(k8sClient.Create(ctx, terminalResource)).To(Succeed())
			terminalResource.Status.Conditions = []metav1.Condition{{
				Type:               string(flyteorgv1.ConditionTypeSucceeded),
				Status:             metav1.ConditionTrue,
				Reason:             string(flyteorgv1.ConditionReasonCompleted),
				Message:            "TaskAction completed successfully",
				LastTransitionTime: metav1.Now(),
			}}
			Expect(k8sClient.Status().Update(ctx, terminalResource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, terminalResource)).To(Succeed())

			recorder := &recordingEventsClient{}
			reconciler := &TaskActionReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				Recorder:       events.NewFakeRecorder(10),
				PluginRegistry: newFakePluginRegistry(fake),
				DataStore:      dataStore,
				eventsClient:   recorder,
			}

			result, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: "abort-terminal-resource", Namespace: "default"},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
			Expect(fake.abortCalls).To(BeZero())
			Expect(fake.finalizeCalls).To(Equal(1))

			deleted := &flyteorgv1.TaskAction{}
			Expect(errors.IsNotFound(k8sClient.Get(ctx, types.NamespacedName{Name: "abort-terminal-resource", Namespace: "default"}, deleted))).To(BeTrue())

			recorded := recorder.RecordedEvents()
			phases := make([]interface{}, 0, len(recorded))
			for _, e := range recorded {
				phases = append(phases, e.GetPhase())
			}
			Expect(phases).NotTo(ContainElement(common.ActionPhase_ACTION_PHASE_ABORTED))
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

		It("should carry a reason's gpu fault over as data and never read one from the text", func() {
			occurredAt := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
			fault := &core.GpuFault{Kind: core.GpuFault_KIND_XID, Code: 79, Node: "ip-10-0-0-1"}
			faultMessage := "[gpu-health] [critical] Xid 79 (GPU has fallen off the bus) on GPU 0. xid=79 severity=critical"

			phaseInfo := pluginsCore.PhaseInfoRunning(pluginsCore.DefaultPhaseVersion, &pluginsCore.TaskInfo{
				OccurredAt: &occurredAt,
				AdditionalReasons: []pluginsCore.ReasonInfo{
					{Reason: faultMessage, OccurredAt: &occurredAt, GpuFault: fault},
					// The plugin did not trust this one, so the message alone must not
					// turn into a fault.
					{Reason: faultMessage, OccurredAt: &occurredAt},
				},
			})

			events := toClusterEvents(phaseInfo, timestamppb.New(occurredAt))
			Expect(events).To(HaveLen(2))
			Expect(proto.Equal(events[0].GetGpuFault(), fault)).To(BeTrue())
			Expect(events[1].GetGpuFault()).To(BeNil())
		})
	})
})
