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
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	k8sPlugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
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

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied, 0)
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

			res, err := r.recordSystemError(ctx, ta, original, "pod", errFakeWebhookDenied, 0)
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

			res, err := r.recordSystemError(ctx, ta, original, "ray", handleErr, 0)
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
			Expect(getTaskAction(nn).Status.AttemptStartedAt).To(BeNil())

			fakeClock.Step(24 * time.Hour)
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(fake.handleCalls).To(Equal(2))
			Expect(fake.abortCalls).To(BeZero())
			Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
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
			Expect(getTaskAction(nn).Status.AttemptStartedAt).To(BeNil())

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

		It("bootstraps an existing Running action from its earliest current-attempt history", func() {
			const timeout = 2 * time.Hour
			now := time.Date(2026, time.August, 25, 4, 0, 0, 0, time.UTC)
			queuedAt := now.Add(-2 * time.Hour)
			firstExecutingAt := now.Add(-90 * time.Minute)
			laterExecutingAt := now.Add(-time.Hour)
			fakeClock := testingclock.NewFakeClock(now)
			fake := &fakePlugin{
				id: "timeout-plugin",
				transitions: []pluginsCore.Transition{
					pluginsCore.DoTransition(pluginsCore.PhaseInfoRunning(
						pluginsCore.DefaultPhaseVersion,
						&pluginsCore.TaskInfo{},
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
				{Phase: string(flyteorgv1.ConditionReasonExecuting), OccurredAt: metav1.NewTime(firstExecutingAt)},
				{Phase: string(flyteorgv1.ConditionReasonExecuting), OccurredAt: metav1.NewTime(laterExecutingAt)},
			}
			Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

			result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", firstExecutingAt))
			Expect(isTerminal(persisted)).To(BeFalse())
		})

		It("does not bootstrap a retried attempt from history before its latest Queued boundary", func() {
			const timeout = time.Hour
			now := time.Date(2026, time.August, 25, 4, 0, 0, 0, time.UTC)
			previousQueuedAt := now.Add(-4 * time.Hour)
			previousExecutingAt := now.Add(-3 * time.Hour)
			retryQueuedAt := now.Add(-10 * time.Minute)
			currentExecutingAt := now.Add(-5 * time.Minute)
			cases := []struct {
				name          string
				currentStart  *time.Time
				expectedStart time.Time
			}{
				{name: "plugin-stale", expectedStart: now},
				{name: "current-history", currentStart: &currentExecutingAt, expectedStart: currentExecutingAt},
			}

			for _, testCase := range cases {
				fakeClock := testingclock.NewFakeClock(now)
				fake := &fakePlugin{
					id:          "timeout-plugin",
					transitions: []pluginsCore.Transition{runningTransition(previousExecutingAt)},
				}
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				name := "timeout-rb-stale"
				if testCase.name == "current-history" {
					name = "timeout-rb-history"
				}
				nn := createTaskAction(
					name,
					buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 1),
				)
				resource := getTaskAction(nn)
				resource.Status.Attempts = 2
				resource.Status.PluginPhase = pluginsCore.PhaseRunning.String()
				resource.Status.PhaseHistory = []flyteorgv1.PhaseTransition{
					{Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(previousQueuedAt)},
					{Phase: string(flyteorgv1.ConditionReasonExecuting), OccurredAt: metav1.NewTime(previousExecutingAt)},
					{Phase: string(flyteorgv1.ConditionReasonQueued), OccurredAt: metav1.NewTime(retryQueuedAt)},
				}
				if testCase.currentStart != nil {
					resource.Status.PhaseHistory = append(
						resource.Status.PhaseHistory,
						flyteorgv1.PhaseTransition{
							Phase:      string(flyteorgv1.ConditionReasonExecuting),
							OccurredAt: metav1.NewTime(*testCase.currentStart),
						},
					)
				}
				Expect(k8sClient.Status().Update(ctx, resource)).To(Succeed())

				_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
				Expect(err).NotTo(HaveOccurred())
				Expect(getTaskAction(nn).Status.AttemptStartedAt.Time).To(
					BeTemporally("==", testCase.expectedStart),
				)
			}
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

		It("waits for the local deadline before classifying future terminal reports", func() {
			const timeout = time.Minute
			base := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
			cases := []string{"late", "missing", "zero"}

			for i, testCase := range cases {
				startedAt := base.Add(time.Duration(i) * time.Hour)
				lateCompletion := startedAt.Add(timeout + time.Second)
				zeroCompletion := time.Time{}
				info := &pluginsCore.TaskInfo{OccurredAt: &lateCompletion}
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
				r := newReconciler(fake, fakeClock, &recordingEventsClient{}, nil)
				nn := createTaskAction(
					"timeout-future-"+testCase,
					buildTaskTemplateBytesWithTimeoutAndRetries("timeout-test", "busybox", timeout, 0),
				)
				request := reconcile.Request{NamespacedName: nn}

				_, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				fakeClock.Step(timeout / 2)
				result, err := r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.RequeueAfter).To(Equal(TaskActionDefaultRequeueDuration))
				Expect(isTerminal(getTaskAction(nn))).To(BeFalse())
				Expect(fake.abortCalls).To(BeZero())

				fakeClock.Step(timeout / 2)
				_, err = r.Reconcile(ctx, request)
				Expect(err).NotTo(HaveOccurred())
				Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
				Expect(fake.abortCalls).To(Equal(1))
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

		It("retries status persistence and retains the authoritative start time", func() {
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

			fakeClock.Step(timeout)
			failingClient.failUpdates = false
			_, err = r.Reconcile(ctx, request)
			Expect(err).NotTo(HaveOccurred())
			persisted := getTaskAction(nn)
			Expect(persisted.Status.AttemptStartedAt.Time).To(BeTemporally("==", base))
			Expect(persisted.Status.TimeoutAt.Time).To(BeTemporally("==", base.Add(timeout)))
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
