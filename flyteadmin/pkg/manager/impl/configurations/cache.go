package configurations

import (
	"sync"

	"maps"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type configurationDocumentCache struct {
	configDoc *admin.ConfigurationDocument
	mutex     sync.RWMutex
}

func (m *ConfigurationManager) getCache(version string, mode documentMode) (*admin.ConfigurationDocument, bool) {
	m.cacheConfigDoc.mutex.RLock()
	defer m.cacheConfigDoc.mutex.RUnlock()
	// If the cache is empty or the version is different, return false
	if m.cacheConfigDoc.configDoc == nil || m.cacheConfigDoc.configDoc.Version != version {
		return nil, false
	}
	// If the version is the same and copy is true, return a copy of the cache
	if mode == editable {
		copyDocument := admin.ConfigurationDocument{
			Version:        m.cacheConfigDoc.configDoc.Version,
			Configurations: make(map[string]*admin.Configuration),
		}
		maps.Copy(copyDocument.Configurations, m.cacheConfigDoc.configDoc.Configurations)
		return &copyDocument, true
	}
	// If the version is the same and copy is false, return the cache
	return m.cacheConfigDoc.configDoc, true
}

func (m *ConfigurationManager) setCache(document *admin.ConfigurationDocument) {
	m.cacheConfigDoc.mutex.Lock()
	defer m.cacheConfigDoc.mutex.Unlock()
	m.cacheConfigDoc.configDoc = document
}
