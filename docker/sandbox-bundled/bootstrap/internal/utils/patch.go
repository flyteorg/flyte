package utils

import (
	"encoding/json"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

func sanitizePatch(m map[string]interface{}, filterNil, filterEmpty bool) {
	val := reflect.ValueOf(m)
	for _, e := range val.MapKeys() {
		v := val.MapIndex(e)
		if v.IsNil() {
			if filterNil {
				delete(m, e.String())
			}
			continue
		}
		switch t := v.Interface().(type) {
		case map[string]interface{}:
			sanitizePatch(t, filterNil, filterEmpty)
			if filterEmpty && len(t) == 0 {
				delete(m, e.String())
			}
		case []interface{}:
			if filterEmpty && len(t) == 0 {
				delete(m, e.String())
			}
		}
	}
}

func MarshalPatch(o interface{}, filterNil, filterEmpty bool) ([]byte, error) {
	m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(o)
	if err != nil {
		return nil, err
	}
	sanitizePatch(m, filterNil, filterEmpty)

	j, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	y, err := yaml.JSONToYAML(j)
	if err != nil {
		return nil, err
	}

	return y, nil
}
