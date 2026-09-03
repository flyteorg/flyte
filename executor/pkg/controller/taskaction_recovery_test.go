package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func newRecoveredTaskAction(name string, recovered *flyteorgv1.RecoveredFrom) *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: flyteorgv1.TaskActionSpec{
			RunName:       "recovery-run",
			Project:       "test-project",
			Domain:        "test-domain",
			ActionName:    name,
			ActionType:    flyteorgv1.ActionTypeTask,
			TaskType:      "python-task",
			RunOutputBase: "s3://bucket/recovery-run",
			RecoveredFrom: recovered,
		},
	}
}

var _ = Describe("Recovered TaskAction Controller", func() {
	ctx := context.Background()

	newReconciler := func(recorder *recordingEventsClient) *TaskActionReconciler {
		return &TaskActionReconciler{
			Client:       k8sClient,
			Scheme:       k8sClient.Scheme(),
			Recorder:     events.NewFakeRecorder(10),
			eventsClient: recorder,
		}
	}

	getTaskAction := func(nn types.NamespacedName) *flyteorgv1.TaskAction {
		ta := &flyteorgv1.TaskAction{}
		Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
		return ta
	}

	cleanup := func(nn types.NamespacedName) {
		ta := &flyteorgv1.TaskAction{}
		if err := k8sClient.Get(ctx, nn, ta); err == nil {
			Expect(k8sClient.Delete(ctx, ta)).To(Succeed())
		}
	}

	Context("an action recovered from a prior run", func() {
		nn := types.NamespacedName{Name: "recovered-a5", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("settles terminal from the spec without dispatching a plugin", func() {
			Expect(k8sClient.Create(ctx, newRecoveredTaskAction(nn.Name, &flyteorgv1.RecoveredFrom{
				SourceRunName: "source-run",
				OutputUri:     "s3://bucket/source-run/a5/1/outputs.pb",
				Attempts:      2,
				CacheStatus:   int32(core.CatalogCacheStatus_CACHE_HIT),
			}))).To(Succeed())

			recorder := &recordingEventsClient{}
			_, err := newReconciler(recorder).Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			ta := getTaskAction(nn)
			Expect(isTerminal(ta)).To(BeTrue())
			succeeded := meta.FindStatusCondition(ta.Status.Conditions, string(flyteorgv1.ConditionTypeSucceeded))
			Expect(succeeded.Status).To(Equal(metav1.ConditionTrue))
			Expect(succeeded.Reason).To(Equal(string(flyteorgv1.ConditionReasonRecovered)))
			Expect(ta.GetLabels()).To(HaveKeyWithValue(LabelTerminationStatus, LabelValueTerminated))

			By("never adding the finalizer the plugin path would add")
			Expect(ta.Finalizers).To(BeEmpty())

			By("never creating a pod")
			pods := &corev1.PodList{}
			Expect(k8sClient.List(ctx, pods, client.InNamespace("default"))).To(Succeed())
			for _, pod := range pods.Items {
				Expect(pod.Name).NotTo(Equal(nn.Name))
			}

			By("reporting RECOVERED with the source run's outputs, not this run's output base")
			recorded := recorder.RecordedEvents()
			Expect(recorded).To(HaveLen(1))
			Expect(recorded[0].GetPhase()).To(Equal(common.ActionPhase_ACTION_PHASE_RECOVERED))
			Expect(recorded[0].GetOutputs().GetOutputUri()).To(Equal("s3://bucket/source-run/a5/1/outputs.pb"))
			Expect(recorded[0].GetAttempt()).To(Equal(uint32(2)))
			Expect(recorded[0].GetCacheStatus()).To(Equal(core.CatalogCacheStatus_CACHE_HIT))
			Expect(recorded[0].GetId().GetName()).To(Equal(nn.Name))
			Expect(recorded[0].GetId().GetRun().GetName()).To(Equal("recovery-run"))

			By("surfacing as RECOVERED to the actions service, not SUCCEEDED")
			Expect(actionsk8s.GetPhaseFromConditions(ta)).To(Equal(common.ActionPhase_ACTION_PHASE_RECOVERED))
		})

		It("replays a recovered condition's signal instead of pausing for a new one", func() {
			signal, err := proto.Marshal(&core.Literal{
				Value: &core.Literal_Scalar{Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
						Value: &core.Primitive_Boolean{Boolean: true},
					}},
				}},
			})
			Expect(err).NotTo(HaveOccurred())

			ta := newRecoveredTaskAction(nn.Name, &flyteorgv1.RecoveredFrom{
				SourceRunName: "source-run",
				Output:        signal,
			})
			// A condition has no plugin and no outputs file; the recovered branch must win
			// over the per-type dispatch that would otherwise park it in Paused.
			ta.Spec.ActionType = flyteorgv1.ActionTypeCondition
			ta.Spec.TaskType = ""
			Expect(k8sClient.Create(ctx, ta)).To(Succeed())

			recorder := &recordingEventsClient{}
			_, err = newReconciler(recorder).Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			got := getTaskAction(nn)
			Expect(isTerminal(got)).To(BeTrue())
			Expect(meta.FindStatusCondition(got.Status.Conditions,
				string(flyteorgv1.ConditionTypeSucceeded)).Reason).
				To(Equal(string(flyteorgv1.ConditionReasonRecovered)))

			By("carrying the source run's resolved value so the actions watcher can ship it")
			Expect(got.Status.SignalValue).To(Equal(signal))
			Expect(actionsk8s.SignalValueFromStatus(ctx, got)).NotTo(BeNil())
		})

		It("re-reconciles without emitting a second event", func() {
			Expect(k8sClient.Create(ctx, newRecoveredTaskAction(nn.Name, &flyteorgv1.RecoveredFrom{
				SourceRunName: "source-run",
				OutputUri:     "s3://bucket/source-run/a5/1/outputs.pb",
			}))).To(Succeed())

			recorder := &recordingEventsClient{}
			r := newReconciler(recorder)
			_, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())
			_, err = r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			Expect(recorder.RecordedEvents()).To(HaveLen(1))
		})
	})

	Context("a recovered action with no source outputs", func() {
		nn := types.NamespacedName{Name: "recovered-no-outputs", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("still settles terminal, with an attempt the event validation accepts", func() {
			Expect(k8sClient.Create(ctx, newRecoveredTaskAction(nn.Name, &flyteorgv1.RecoveredFrom{
				SourceRunName: "source-run",
			}))).To(Succeed())

			recorder := &recordingEventsClient{}
			_, err := newReconciler(recorder).Reconcile(ctx, reconcile.Request{NamespacedName: nn})
			Expect(err).NotTo(HaveOccurred())

			Expect(isTerminal(getTaskAction(nn))).To(BeTrue())
			recorded := recorder.RecordedEvents()
			Expect(recorded).To(HaveLen(1))
			Expect(recorded[0].GetAttempt()).To(Equal(uint32(1)))
			Expect(recorded[0].GetOutputs()).To(BeNil())
		})
	})
})
