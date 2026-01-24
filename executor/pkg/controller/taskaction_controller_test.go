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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

var _ = Describe("TaskAction Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
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
						Org:           "test-org",
						Project:       "test-project",
						Domain:        "test-domain",
						ActionName:    "test-action",
						InputURI:      "/tmp/input",
						RunOutputBase: "/tmp/output",
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
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that the TaskAction was updated with a condition
			updatedTaskAction := &flyteorgv1.TaskAction{}
			err = k8sClient.Get(ctx, typeNamespacedName, updatedTaskAction)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedTaskAction.Status.Conditions).NotTo(BeEmpty())
		})
	})
})
