package flytek8s

import (
	"context"
	"sync"

	"github.com/flyteorg/flytestdlib/logger"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var DefaultPodTemplateStore PodTemplateStore = NewPodTemplateStore()

// PodTemplateStore maintains a thread-safe mapping of active PodTemplates with their associated
// namespaces.
type PodTemplateStore struct {
	*sync.Map
	defaultNamespace string
}

// NewPodTemplateStore initializes a new PodTemplateStore
func NewPodTemplateStore() PodTemplateStore {
	return PodTemplateStore{
		Map: &sync.Map{},
	}
}

// Delete removes the specified PodTemplate from the store.
func (p *PodTemplateStore) Delete(podTemplate *v1.PodTemplate) {
	if value, ok := p.Load(podTemplate.Name); ok {
		podTemplates := value.(*sync.Map)
		podTemplates.Delete(podTemplate.Namespace)
		logger.Debugf(context.Background(), "deleted PodTemplate '%s:%s' from store", podTemplate.Namespace, podTemplate.Name)

		// we specifically are not deleting empty maps from the store because this may introduce race
		// conditions where a PodTemplate is being added to the 2nd dimension map while the top level map
		// is concurrently being deleted.
	}
}

// LoadOrDefault returns the PodTemplate with the specified name in the given namespace. If one
// does not exist it attempts to retrieve the one associated with the defaultNamespace.
func (p *PodTemplateStore) LoadOrDefault(namespace string, podTemplateName string) *v1.PodTemplate {
	if value, ok := p.Load(podTemplateName); ok {
		podTemplates := value.(*sync.Map)
		if podTemplate, ok := podTemplates.Load(namespace); ok {
			return podTemplate.(*v1.PodTemplate)
		}

		if podTemplate, ok := podTemplates.Load(p.defaultNamespace); ok {
			return podTemplate.(*v1.PodTemplate)
		}
	}

	return nil
}

// SetDefaultNamespace sets the default namespace for the PodTemplateStore.
func (p *PodTemplateStore) SetDefaultNamespace(namespace string) {
	p.defaultNamespace = namespace
}

// Store loads the specified PodTemplate into the store.
func (p *PodTemplateStore) Store(podTemplate *v1.PodTemplate) {
	value, _ := p.LoadOrStore(podTemplate.Name, &sync.Map{})
	podTemplates := value.(*sync.Map)
	podTemplates.Store(podTemplate.Namespace, podTemplate)
	logger.Debugf(context.Background(), "registered PodTemplate '%s:%s' in store", podTemplate.Namespace, podTemplate.Name)
}

// GetPodTemplateUpdatesHandler returns a new ResourceEventHandler which adds / removes
// PodTemplates to / from the provided PodTemplateStore.
func GetPodTemplateUpdatesHandler(store *PodTemplateStore) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if podTemplate, ok := obj.(*v1.PodTemplate); ok {
				store.Store(podTemplate)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			if podTemplate, ok := new.(*v1.PodTemplate); ok {
				store.Store(podTemplate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if podTemplate, ok := obj.(*v1.PodTemplate); ok {
				store.Delete(podTemplate)
			}
		},
	}
}
