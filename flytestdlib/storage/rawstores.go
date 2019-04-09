package storage

import (
	"fmt"

	"github.com/lyft/flytestdlib/promutils"
)

type dataStoreCreateFn func(cfg *Config, metricsScope promutils.Scope) (RawStore, error)

var stores = map[string]dataStoreCreateFn{
	TypeMemory: NewInMemoryRawStore,
	TypeLocal:  newLocalRawStore,
	TypeMinio:  newS3RawStore,
	TypeS3:     newS3RawStore,
}

// Creates a new Data Store with the supplied config.
func NewDataStore(cfg *Config, metricsScope promutils.Scope) (s *DataStore, err error) {
	var rawStore RawStore
	if fn, found := stores[cfg.Type]; found {
		rawStore, err = fn(cfg, metricsScope)
		if err != nil {
			return &emptyStore, err
		}

		protoStore := NewDefaultProtobufStore(newCachedRawStore(cfg, rawStore, metricsScope), metricsScope)
		return NewCompositeDataStore(URLPathConstructor{}, protoStore), nil
	}

	return &emptyStore, fmt.Errorf("type is of an invalid value [%v]", cfg.Type)
}

// Composes a new DataStore.
func NewCompositeDataStore(refConstructor ReferenceConstructor, composedProtobufStore ComposedProtobufStore) *DataStore {
	return &DataStore{
		ReferenceConstructor:  refConstructor,
		ComposedProtobufStore: composedProtobufStore,
	}
}
