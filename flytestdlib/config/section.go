package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/flyteorg/flytestdlib/atomic"

	"github.com/spf13/pflag"
)

type Section interface {
	// Gets a cloned copy of the Config registered to this section. This config instance does not account for any child
	// section registered.
	GetConfig() Config

	// Gets a function pointer to call when the config has been updated.
	GetConfigUpdatedHandler() SectionUpdated

	// Sets the config and sets a bit indicating whether the new config is different when compared to the existing value.
	SetConfig(config Config) error

	// Gets a value indicating whether the config has changed since the last call to GetConfigChangedAndClear and clears
	// the changed bit. This operation is atomic.
	GetConfigChangedAndClear() bool

	// Retrieves the loaded values for section key if one exists, or nil otherwise.
	GetSection(key SectionKey) Section

	// Gets all child config sections.
	GetSections() SectionMap

	// Registers a section with the config manager. Section keys are case insensitive and must be unique.
	// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
	// marshaling. If the section registered gets updated at runtime, the updatesFn will be invoked to handle the propagation
	// of changes.
	RegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) (Section, error)

	// Registers a section with the config manager. Section keys are case insensitive and must be unique.
	// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
	// marshaling. If the section registered gets updated at runtime, the updatesFn will be invoked to handle the propagation
	// of changes.
	MustRegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) Section

	// Registers a section with the config manager. Section keys are case insensitive and must be unique.
	// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
	// marshaling.
	RegisterSection(key SectionKey, configSection Config) (Section, error)

	// Registers a section with the config manager. Section keys are case insensitive and must be unique.
	// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
	// marshaling.
	MustRegisterSection(key SectionKey, configSection Config) Section
}

type Config = interface{}
type SectionKey = string
type SectionMap map[SectionKey]Section

// A section can optionally implements this interface to add its fields as cmdline arguments.
type PFlagProvider interface {
	GetPFlagSet(prefix string) *pflag.FlagSet
}

type SectionUpdated func(ctx context.Context, newValue Config)

// Global section to use with any root-level config sections registered.
var rootSection = NewRootSection()

type section struct {
	config   Config
	handler  SectionUpdated
	isDirty  atomic.Bool
	sections SectionMap
	lockObj  sync.RWMutex
}

// Gets the global root section.
func GetRootSection() Section {
	return rootSection
}

func MustRegisterSection(key SectionKey, configSection Config) Section {
	s, err := RegisterSection(key, configSection)
	if err != nil {
		panic(err)
	}

	return s
}

func (r *section) MustRegisterSection(key SectionKey, configSection Config) Section {
	s, err := r.RegisterSection(key, configSection)
	if err != nil {
		panic(err)
	}

	return s
}

// Registers a section with the config manager. Section keys are case insensitive and must be unique.
// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
// marshaling.
func RegisterSection(key SectionKey, configSection Config) (Section, error) {
	return rootSection.RegisterSection(key, configSection)
}

func (r *section) RegisterSection(key SectionKey, configSection Config) (Section, error) {
	return r.RegisterSectionWithUpdates(key, configSection, nil)
}

func MustRegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) Section {
	s, err := RegisterSectionWithUpdates(key, configSection, updatesFn)
	if err != nil {
		panic(err)
	}

	return s
}

func (r *section) MustRegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) Section {
	s, err := r.RegisterSectionWithUpdates(key, configSection, updatesFn)
	if err != nil {
		panic(err)
	}

	return s
}

// Registers a section with the config manager. Section keys are case insensitive and must be unique.
// The section object must be passed by reference since it'll be used to unmarshal into. It must also support json
// marshaling. If the section registered gets updated at runtime, the updatesFn will be invoked to handle the propagation
// of changes.
func RegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) (Section, error) {
	return rootSection.RegisterSectionWithUpdates(key, configSection, updatesFn)
}

func (r *section) RegisterSectionWithUpdates(key SectionKey, configSection Config, updatesFn SectionUpdated) (Section, error) {
	r.lockObj.Lock()
	defer r.lockObj.Unlock()

	key = strings.ToLower(key)

	if len(key) == 0 {
		return nil, errors.New("key must be a non-zero string")
	}

	if configSection == nil {
		return nil, fmt.Errorf("configSection must be a non-nil pointer. SectionKey: %v", key)
	}

	if reflect.TypeOf(configSection).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("section must be a Pointer. SectionKey: %v", key)
	}

	if _, alreadyExists := r.sections[key]; alreadyExists {
		return nil, fmt.Errorf("key already exists [%v]", key)
	}

	section := NewSection(configSection, updatesFn)
	r.sections[key] = section
	return section, nil
}

// Retrieves the loaded values for section key if one exists, or nil otherwise.
func GetSection(key SectionKey) Section {
	return rootSection.GetSection(key)
}

func (r *section) GetSection(key SectionKey) Section {
	r.lockObj.RLock()
	defer r.lockObj.RUnlock()

	key = strings.ToLower(key)

	if section, alreadyExists := r.sections[key]; alreadyExists {
		return section
	}

	return nil
}

func (r *section) GetSections() SectionMap {
	return r.sections
}

func (r *section) GetConfig() Config {
	r.lockObj.RLock()
	defer r.lockObj.RUnlock()

	return r.config
}

func (r *section) SetConfig(c Config) error {
	r.lockObj.Lock()
	defer r.lockObj.Unlock()

	if reflect.TypeOf(c).Kind() != reflect.Ptr {
		return fmt.Errorf("config must be a Pointer")
	}

	if !DeepEqual(r.config, c) {
		r.config = c
		r.isDirty.Store(true)
	}

	return nil
}

func (r *section) GetConfigUpdatedHandler() SectionUpdated {
	return r.handler
}

func (r *section) GetConfigChangedAndClear() bool {
	return r.isDirty.CompareAndSwap(true, false)
}

func NewSection(configSection Config, updatesFn SectionUpdated) Section {
	return &section{
		config:   configSection,
		handler:  updatesFn,
		isDirty:  atomic.NewBool(false),
		sections: map[SectionKey]Section{},
		lockObj:  sync.RWMutex{},
	}
}

func NewRootSection() Section {
	return NewSection(nil, nil)
}
