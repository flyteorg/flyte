package mocks

import (
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
)
import "context"
import "k8s.io/apimachinery/pkg/runtime"
import "k8s.io/apimachinery/pkg/types"

type Cache struct {
	informertest.FakeInformers
}

// Get on this mock will always return object not found.
func (_m *Cache) Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	return errors.NewNotFound(v1.Resource("pod"), accessor.GetName())
}
