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

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// fakeEventsClient is a no-op implementation of EventsProxyServiceClient for tests.
type fakeEventsClient struct{}

func (f *fakeEventsClient) Record(_ context.Context, _ *connect.Request[workflow.RecordRequest]) (*connect.Response[workflow.RecordResponse], error) {
	return connect.NewResponse(&workflow.RecordResponse{}), nil
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
						Org:           "test-org",
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
})
