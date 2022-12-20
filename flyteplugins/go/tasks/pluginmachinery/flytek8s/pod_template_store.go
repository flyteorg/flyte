package flytek8s

import (
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

var DefaultPodTemplateStore PodTemplateStore = NewPodTemplateStore()

// PodTemplateStore maintains a thread-safe mapping of active PodTemplates with their associated
// namespaces.
type PodTemplateStore struct {
	sync.Map
	defaultNamespace string
}

// NewPodTemplateStore initializes a new PodTemplateStore
func NewPodTemplateStore() PodTemplateStore {
	return PodTemplateStore{}
}

// LoadOrDefautl returns the PodTemplate associated with the given namespace. If one does not exist
// it attempts to retrieve the one associated with the defaultNamespace parameter.
func (p *PodTemplateStore) LoadOrDefault(namespace string) *v1.PodTemplate {
	if podTemplate, ok := p.Load(namespace); ok {
		return podTemplate.(*v1.PodTemplate)
	}

	if podTemplate, ok := p.Load(p.defaultNamespace); ok {
		return podTemplate.(*v1.PodTemplate)
	}

	return nil
}

// SetDefaultNamespace sets the default namespace for the PodTemplateStore.
func (p *PodTemplateStore) SetDefaultNamespace(namespace string) {
	p.defaultNamespace = namespace
}

// GetPodTemplateUpdatesHandler returns a new ResourceEventHandler which adds / removes
// PodTemplates with the associated podTemplateName to / from the provided PodTemplateStore.
func GetPodTemplateUpdatesHandler(store *PodTemplateStore, podTemplateName string) cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			podTemplate, ok := obj.(*v1.PodTemplate)
			if ok && podTemplate.Name == podTemplateName {
				store.Store(podTemplate.Namespace, podTemplate)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			podTemplate, ok := new.(*v1.PodTemplate)
			if ok && podTemplate.Name == podTemplateName {
				store.Store(podTemplate.Namespace, podTemplate)
			}
		},
		DeleteFunc: func(obj interface{}) {
			podTemplate, ok := obj.(*v1.PodTemplate)
			if ok && podTemplate.Name == podTemplateName {
				store.Delete(podTemplate.Namespace)
			}
		},
	}
}
