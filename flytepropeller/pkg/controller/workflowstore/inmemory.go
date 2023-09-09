package workflowstore

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
)

type InmemoryWorkflowStore struct {
	store map[string]map[string]*v1alpha1.FlyteWorkflow
}

func (i *InmemoryWorkflowStore) Create(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
	if w != nil {
		if w.Name != "" && w.Namespace != "" {
			if _, ok := i.store[w.Namespace]; !ok {
				i.store[w.Namespace] = map[string]*v1alpha1.FlyteWorkflow{}
			}
			i.store[w.Namespace][w.Name] = w
			return nil
		}
	}
	return kubeerrors.NewBadRequest(fmt.Sprintf("Workflow object with Namespace [%v] & Name [%v] is required", w.Namespace, w.Name))
}

func (i *InmemoryWorkflowStore) Delete(ctx context.Context, namespace, name string) error {
	if m, ok := i.store[namespace]; ok {
		if _, ok := m[name]; ok {
			delete(m, name)
			return nil
		}
	}
	return nil
}

func (i *InmemoryWorkflowStore) Get(ctx context.Context, namespace, name string) (*v1alpha1.FlyteWorkflow, error) {
	if m, ok := i.store[namespace]; ok {
		if v, ok := m[name]; ok {
			return v, nil
		}
	}
	return nil, ErrWorkflowNotFound
}

func (i *InmemoryWorkflowStore) UpdateStatus(ctx context.Context, w *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	if w != nil {
		if w.Name != "" && w.Namespace != "" {
			if m, ok := i.store[w.Namespace]; ok {
				if _, ok := m[w.Name]; ok {
					newW := w.DeepCopy()
					// Appends to the resource version to ensure that resource version has been updated
					newW.ResourceVersion = w.ResourceVersion + "x"
					m[w.Name] = newW
					return newW, nil
				}
			}

			return nil, kubeerrors.NewNotFound(v1alpha1.Resource(v1alpha1.FlyteWorkflowKind), w.Name)
		}
	}

	return nil, kubeerrors.NewBadRequest("Workflow object with Namespace & Name is required")
}

func (i *InmemoryWorkflowStore) Update(ctx context.Context, w *v1alpha1.FlyteWorkflow, priorityClass PriorityClass) (
	newWF *v1alpha1.FlyteWorkflow, err error) {
	return i.UpdateStatus(ctx, w, priorityClass)
}

// Returns an inmemory store, that will update the resource version for every update automatically. This is a good
// idea to test that resource version checks work, but does not represent that the object was actually updated or not
// The resource version is ALWAYS updated.
func NewInMemoryWorkflowStore() *InmemoryWorkflowStore {
	return &InmemoryWorkflowStore{
		store: map[string]map[string]*v1alpha1.FlyteWorkflow{},
	}
}
