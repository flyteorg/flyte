package flytek8s

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPodTemplateStore(t *testing.T) {
	ctx := context.TODO()

	podTemplate := &v1.PodTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "defaultPodTemplate",
			Namespace: "defaultNamespace",
		},
		Template: v1.PodTemplateSpec{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Command: []string{"flytepropeller"},
						Args:    []string{"--config", "/etc/flyte/config/*.yaml"},
					},
				},
			},
		},
	}

	store := NewPodTemplateStore()
	store.SetDefaultNamespace(podTemplate.Namespace)

	kubeClient := fake.NewSimpleClientset()
	informerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClient, 30*time.Second)

	updateHandler := GetPodTemplateUpdatesHandler(&store)
	informerFactory.Core().V1().PodTemplates().Informer().AddEventHandler(updateHandler)
	go informerFactory.Start(ctx.Done())

	// create the podTemplate
	_, err := kubeClient.CoreV1().PodTemplates(podTemplate.Namespace).Create(ctx, podTemplate, metav1.CreateOptions{})
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	createPodTemplate := store.LoadOrDefault(podTemplate.Namespace, podTemplate.Name)
	assert.NotNil(t, createPodTemplate)
	assert.True(t, reflect.DeepEqual(podTemplate, createPodTemplate))

	// non-default namespace podTemplate does not exist
	newNamespacePodTemplate := podTemplate.DeepCopy()
	newNamespacePodTemplate.Namespace = "foo"

	nonDefaultNamespacePodTemplate := store.LoadOrDefault(newNamespacePodTemplate.Namespace, newNamespacePodTemplate.Name)
	assert.NotNil(t, nonDefaultNamespacePodTemplate)
	assert.True(t, reflect.DeepEqual(podTemplate, nonDefaultNamespacePodTemplate))

	// non-default name podTemplate does not exist
	newNamePodTemplate := podTemplate.DeepCopy()
	newNamePodTemplate.Name = "foo"

	nonDefaultNamePodTemplate := store.LoadOrDefault(newNamePodTemplate.Namespace, newNamePodTemplate.Name)
	assert.Nil(t, nonDefaultNamePodTemplate)

	// non-default namespace podTemplate exists
	_, err = kubeClient.CoreV1().PodTemplates(newNamespacePodTemplate.Namespace).Create(ctx, newNamespacePodTemplate, metav1.CreateOptions{})
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	createNewNamespacePodTemplate := store.LoadOrDefault(newNamespacePodTemplate.Namespace, newNamespacePodTemplate.Name)
	assert.NotNil(t, createNewNamespacePodTemplate)
	assert.True(t, reflect.DeepEqual(newNamespacePodTemplate, createNewNamespacePodTemplate))

	// update the podTemplate
	updatedPodTemplate := podTemplate.DeepCopy()
	updatedPodTemplate.Template.Spec.RestartPolicy = v1.RestartPolicyNever
	_, err = kubeClient.CoreV1().PodTemplates(podTemplate.Namespace).Update(ctx, updatedPodTemplate, metav1.UpdateOptions{})
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	updatePodTemplate := store.LoadOrDefault(podTemplate.Namespace, podTemplate.Name)
	assert.NotNil(t, updatePodTemplate)
	assert.True(t, reflect.DeepEqual(updatedPodTemplate, updatePodTemplate))

	// delete the podTemplate in the namespace
	err = kubeClient.CoreV1().PodTemplates(podTemplate.Namespace).Delete(ctx, podTemplate.Name, metav1.DeleteOptions{})
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	deletePodTemplate := store.LoadOrDefault(podTemplate.Namespace, podTemplate.Name)
	assert.Nil(t, deletePodTemplate)
}
