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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// buildConditionSpecBytes marshals a minimal workflow.ConditionAction with a
// bool declared type and the given timeout (nil = no timeout).
func buildConditionSpecBytes(timeout *time.Duration) []byte {
	spec := &workflow.ConditionAction{
		Name: "approve",
		Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BOOLEAN}},
	}
	if timeout != nil {
		spec.Timeout = durationpb.New(*timeout)
	}
	data, err := proto.Marshal(spec)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func buildBoolLiteralBytes(v bool) []byte {
	lit := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: v},
			}},
		}},
	}
	data, err := proto.Marshal(lit)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func newConditionTaskAction(name string, conditionSpec []byte) *flyteorgv1.TaskAction {
	return &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: flyteorgv1.TaskActionSpec{
			RunName:       "test-run",
			Project:       "test-project",
			Domain:        "test-domain",
			ActionName:    name,
			ActionType:    flyteorgv1.ActionTypeCondition,
			ConditionSpec: conditionSpec,
		},
	}
}

func newConditionReconciler() *TaskActionReconciler {
	return &TaskActionReconciler{
		Client:       k8sClient,
		Scheme:       k8sClient.Scheme(),
		Recorder:     events.NewFakeRecorder(10),
		eventsClient: &fakeEventsClient{},
	}
}

var _ = Describe("Condition TaskAction Controller", func() {
	ctx := context.Background()

	reconcileOnce := func(r *TaskActionReconciler, nn types.NamespacedName) reconcile.Result {
		result, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: nn})
		Expect(err).NotTo(HaveOccurred())
		return result
	}

	getTaskAction := func(nn types.NamespacedName) *flyteorgv1.TaskAction {
		ta := &flyteorgv1.TaskAction{}
		Expect(k8sClient.Get(ctx, nn, ta)).To(Succeed())
		return ta
	}

	conditionReason := func(ta *flyteorgv1.TaskAction, condType flyteorgv1.TaskActionConditionType) string {
		cond := meta.FindStatusCondition(ta.Status.Conditions, string(condType))
		if cond == nil || cond.Status != metav1.ConditionTrue {
			return ""
		}
		return cond.Reason
	}

	cleanup := func(nn types.NamespacedName) {
		ta := &flyteorgv1.TaskAction{}
		if err := k8sClient.Get(ctx, nn, ta); err == nil {
			Expect(k8sClient.Delete(ctx, ta)).To(Succeed())
		}
	}

	Context("paused → signaled", func() {
		nn := types.NamespacedName{Name: "cond-signaled", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("pauses on create, then succeeds when a signal value lands", func() {
			Expect(k8sClient.Create(ctx, newConditionTaskAction(nn.Name, buildConditionSpecBytes(nil)))).To(Succeed())
			r := newConditionReconciler()

			By("first reconcile sets Progressing=True/Paused with no requeue deadline")
			result := reconcileOnce(r, nn)
			Expect(result.RequeueAfter).To(BeZero())
			ta := getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeProgressing)).To(Equal(string(flyteorgv1.ConditionReasonPaused)))
			Expect(isTerminal(ta)).To(BeFalse())
			Expect(ta.Finalizers).To(BeEmpty())

			By("writing a signal value to status, as the Signal RPC would")
			ta.Status.SignalValue = buildBoolLiteralBytes(true)
			ta.Status.SignalledBy = "user@example.com"
			now := metav1.Now()
			ta.Status.SignalledAt = &now
			Expect(k8sClient.Status().Update(ctx, ta)).To(Succeed())

			By("next reconcile flips to Succeeded=True/Signaled and stamps GC labels")
			reconcileOnce(r, nn)
			ta = getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeSucceeded)).To(Equal(string(flyteorgv1.ConditionReasonSignaled)))
			Expect(meta.IsStatusConditionTrue(ta.Status.Conditions, string(flyteorgv1.ConditionTypeProgressing))).To(BeFalse())
			Expect(ta.GetLabels()).To(HaveKeyWithValue(LabelTerminationStatus, LabelValueTerminated))
			// Signal fields must survive the reconciler's status write.
			Expect(ta.Status.SignalValue).NotTo(BeEmpty())
			Expect(ta.Status.SignalledBy).To(Equal("user@example.com"))
			// PhaseHistory records the Paused → Signaled timeline.
			Expect(ta.Status.PhaseHistory).To(HaveLen(2))
		})
	})

	Context("paused → timed out", func() {
		nn := types.NamespacedName{Name: "cond-timeout", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("requeues until the deadline, then fails with TimedOut", func() {
			// CreationTimestamp has second precision, so the deadline can land up to
			// 1s earlier than wall-clock create time; keep the timeout comfortably above that.
			timeout := 2 * time.Second
			Expect(k8sClient.Create(ctx, newConditionTaskAction(nn.Name, buildConditionSpecBytes(&timeout)))).To(Succeed())
			r := newConditionReconciler()

			By("first reconcile pauses and requeues for the remaining deadline")
			result := reconcileOnce(r, nn)
			Expect(result.RequeueAfter).To(BeNumerically(">", 0))
			ta := getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeProgressing)).To(Equal(string(flyteorgv1.ConditionReasonPaused)))

			By("reconciling after the deadline marks Failed=True/TimedOut")
			time.Sleep(timeout + 200*time.Millisecond)
			reconcileOnce(r, nn)
			ta = getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeFailed)).To(Equal(string(flyteorgv1.ConditionReasonTimedOut)))
			Expect(ta.Status.ErrorState).NotTo(BeNil())
			Expect(ta.Status.ErrorState.Code).To(Equal(ConditionTimedOutCode))
			Expect(ta.Status.ErrorState.Kind).To(Equal("USER"))
			Expect(ta.GetLabels()).To(HaveKeyWithValue(LabelTerminationStatus, LabelValueTerminated))
		})
	})

	Context("signal after terminal", func() {
		nn := types.NamespacedName{Name: "cond-late-signal", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("ignores a signal that lands after the condition timed out", func() {
			timeout := time.Millisecond
			Expect(k8sClient.Create(ctx, newConditionTaskAction(nn.Name, buildConditionSpecBytes(&timeout)))).To(Succeed())
			r := newConditionReconciler()

			time.Sleep(50 * time.Millisecond)
			reconcileOnce(r, nn)
			ta := getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeFailed)).To(Equal(string(flyteorgv1.ConditionReasonTimedOut)))

			By("a late signal write does not flip the terminal state")
			ta.Status.SignalValue = buildBoolLiteralBytes(true)
			Expect(k8sClient.Status().Update(ctx, ta)).To(Succeed())
			reconcileOnce(r, nn)
			ta = getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeFailed)).To(Equal(string(flyteorgv1.ConditionReasonTimedOut)))
			Expect(meta.IsStatusConditionTrue(ta.Status.Conditions, string(flyteorgv1.ConditionTypeSucceeded))).To(BeFalse())
		})
	})

	Context("abort by delete", func() {
		nn := types.NamespacedName{Name: "cond-abort", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("deletes immediately — no finalizer is ever added to a condition", func() {
			Expect(k8sClient.Create(ctx, newConditionTaskAction(nn.Name, buildConditionSpecBytes(nil)))).To(Succeed())
			r := newConditionReconciler()
			reconcileOnce(r, nn)

			ta := getTaskAction(nn)
			Expect(ta.Finalizers).To(BeEmpty())
			Expect(k8sClient.Delete(ctx, ta)).To(Succeed())

			err := k8sClient.Get(ctx, nn, &flyteorgv1.TaskAction{})
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("invalid conditionSpec", func() {
		nn := types.NamespacedName{Name: "cond-invalid", Namespace: "default"}
		AfterEach(func() { cleanup(nn) })

		It("marks a condition with an empty spec as terminally failed", func() {
			Expect(k8sClient.Create(ctx, newConditionTaskAction(nn.Name, nil))).To(Succeed())
			r := newConditionReconciler()
			reconcileOnce(r, nn)

			ta := getTaskAction(nn)
			Expect(conditionReason(ta, flyteorgv1.ConditionTypeFailed)).To(Equal(string(flyteorgv1.ConditionReasonInvalidSpec)))
			Expect(ta.GetLabels()).To(HaveKeyWithValue(LabelTerminationStatus, LabelValueTerminated))
		})
	})
})
