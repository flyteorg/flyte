package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

func createTaskAction(ctx context.Context, name string, labels map[string]string) *flyteorgv1.TaskAction {
	ta := &flyteorgv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			Labels:    labels,
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
	ExpectWithOffset(1, k8sClient.Create(ctx, ta)).To(Succeed())
	return ta
}

func deleteTaskAction(ctx context.Context, name string) {
	ta := &flyteorgv1.TaskAction{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: "default"}, ta)
	if err == nil {
		_ = k8sClient.Delete(ctx, ta)
	}
}

var _ = Describe("GarbageCollector", func() {
	ctx := context.Background()

	AfterEach(func() {
		// Clean up all TaskActions in default namespace
		var list flyteorgv1.TaskActionList
		Expect(k8sClient.List(ctx, &list, client.InNamespace("default"))).To(Succeed())
		for i := range list.Items {
			_ = k8sClient.Delete(ctx, &list.Items[i])
		}
	})

	It("should delete TaskActions with expired completed-time label", func() {
		expiredTime := time.Now().UTC().Add(-2 * time.Hour).Format(labelTimeFormat)
		createTaskAction(ctx, "gc-expired", map[string]string{
			LabelTerminationStatus: LabelValueTerminated,
			LabelCompletedTime:     expiredTime,
		})

		gc := NewGarbageCollector(k8sClient, 1*time.Minute, 1*time.Hour)
		Expect(gc.collect(ctx)).To(Succeed())

		ta := &flyteorgv1.TaskAction{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: "gc-expired", Namespace: "default"}, ta)
		Expect(err).To(HaveOccurred())
		Expect(client.IgnoreNotFound(err)).To(Succeed())
	})

	It("should retain TaskActions with recent completed-time label", func() {
		recentTime := time.Now().UTC().Format(labelTimeFormat)
		createTaskAction(ctx, "gc-recent", map[string]string{
			LabelTerminationStatus: LabelValueTerminated,
			LabelCompletedTime:     recentTime,
		})

		gc := NewGarbageCollector(k8sClient, 1*time.Minute, 1*time.Hour)
		Expect(gc.collect(ctx)).To(Succeed())

		ta := &flyteorgv1.TaskAction{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: "gc-recent", Namespace: "default"}, ta)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should retain non-terminated TaskActions", func() {
		createTaskAction(ctx, "gc-active", nil)

		gc := NewGarbageCollector(k8sClient, 1*time.Minute, 1*time.Hour)
		Expect(gc.collect(ctx)).To(Succeed())

		ta := &flyteorgv1.TaskAction{}
		err := k8sClient.Get(ctx, types.NamespacedName{Name: "gc-active", Namespace: "default"}, ta)
		Expect(err).NotTo(HaveOccurred())
	})

	It("should handle empty list gracefully", func() {
		gc := NewGarbageCollector(k8sClient, 1*time.Minute, 1*time.Hour)
		Expect(gc.collect(ctx)).To(Succeed())
	})

	It("should sweep immediately on Start without waiting a full interval", func() {
		expiredTime := time.Now().UTC().Add(-2 * time.Hour).Format(labelTimeFormat)
		createTaskAction(ctx, "gc-startup", map[string]string{
			LabelTerminationStatus: LabelValueTerminated,
			LabelCompletedTime:     expiredTime,
		})

		// A long interval means the ticker will not fire during the test, so the
		// expired TaskAction can only be deleted by the immediate startup sweep.
		// Before the fix, Start waited a full interval before its first collect.
		gc := NewGarbageCollector(k8sClient, 30*time.Minute, 1*time.Hour)
		startCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go func() { _ = gc.Start(startCtx) }()

		Eventually(func() bool {
			ta := &flyteorgv1.TaskAction{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: "gc-startup", Namespace: "default"}, ta)
			return apierrors.IsNotFound(err)
		}, 10*time.Second, 200*time.Millisecond).Should(BeTrue(), "Start should delete the expired TaskAction via its initial sweep")
	})
})

var _ = Describe("ensureTerminalLabels", func() {
	ctx := context.Background()

	AfterEach(func() {
		deleteTaskAction(ctx, "terminal-labels-test")
	})

	It("should patch completed-time when termination-status is set but completed-time is missing", func() {
		ta := createTaskAction(ctx, "terminal-missing-time", map[string]string{
			LabelTerminationStatus: LabelValueTerminated,
		})
		defer deleteTaskAction(ctx, "terminal-missing-time")

		reconciler := &TaskActionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		Expect(reconciler.ensureTerminalLabels(ctx, ta)).To(Succeed())
		Expect(ta.GetLabels()[LabelTerminationStatus]).To(Equal(LabelValueTerminated))
		Expect(ta.GetLabels()[LabelCompletedTime]).NotTo(BeEmpty())
	})

	It("should be idempotent", func() {
		ta := createTaskAction(ctx, "terminal-labels-test", nil)

		reconciler := &TaskActionReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		// First call should set labels
		Expect(reconciler.ensureTerminalLabels(ctx, ta)).To(Succeed())
		Expect(ta.GetLabels()[LabelTerminationStatus]).To(Equal(LabelValueTerminated))
		Expect(ta.GetLabels()[LabelCompletedTime]).NotTo(BeEmpty())
		firstCompletedTime := ta.GetLabels()[LabelCompletedTime]

		// Re-fetch to get updated resource version
		updated := &flyteorgv1.TaskAction{}
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "terminal-labels-test", Namespace: "default"}, updated)).To(Succeed())

		// Second call should be a no-op (labels already set)
		Expect(reconciler.ensureTerminalLabels(ctx, updated)).To(Succeed())
		Expect(updated.GetLabels()[LabelCompletedTime]).To(Equal(firstCompletedTime))
	})
})
