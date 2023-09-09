package config

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"

	stdLibErrs "github.com/flyteorg/flytestdlib/errors"
)

// Uses Json marshal/unmarshal to make a deep copy of a config object.
func DeepCopyConfig(config Config) (Config, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	t := reflect.TypeOf(config)
	ptrValue := reflect.New(t)
	newObj := ptrValue.Interface()
	if err = json.Unmarshal(raw, newObj); err != nil {
		return nil, err
	}

	return ptrValue.Elem().Interface(), nil
}

func DeepEqual(config1, config2 Config) bool {
	return reflect.DeepEqual(config1, config2)
}

func toInterface(config Config) (interface{}, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	var m interface{}
	err = json.Unmarshal(raw, &m)
	return m, err
}

// Builds a generic map out of the root section config and its sub-sections configs.
func AllConfigsAsMap(root Section) (m map[string]interface{}, err error) {
	errs := stdLibErrs.ErrorCollection{}
	allConfigs := make(map[string]interface{}, len(root.GetSections()))
	if root.GetConfig() != nil {
		rootConfig, err := toInterface(root.GetConfig())
		if !errs.Append(err) {
			if asMap, isCasted := rootConfig.(map[string]interface{}); isCasted {
				allConfigs = asMap
			} else {
				allConfigs[""] = rootConfig
			}
		}
	}

	for k, section := range root.GetSections() {
		if _, alreadyExists := allConfigs[k]; alreadyExists {
			errs.Append(errors.Wrap(ErrChildConfigOverridesConfig,
				fmt.Sprintf("section key [%v] overrides an existing native config property", k)))
		}

		allConfigs[k], err = AllConfigsAsMap(section)
		errs.Append(err)
	}

	return allConfigs, errs.ErrorOrDefault()
}
