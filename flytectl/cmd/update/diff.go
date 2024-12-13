package update

import (
	"encoding/json"
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"gopkg.in/yaml.v3"
)

const (
	diffPathBefore = "before"
	diffPathAfter  = "after"
)

// DiffAsYaml marshals both objects as YAML and returns differences
// between marshalled values in unified format. Marshalling respects
// JSON field annotations.
func DiffAsYaml(path1, path2 string, object1, object2 any) (string, error) {
	yaml1, err := marshalToYamlString(object1)
	if err != nil {
		return "", fmt.Errorf("diff as yaml: %w", err)
	}

	yaml2, err := marshalToYamlString(object2)
	if err != nil {
		return "", fmt.Errorf("diff as yaml: %w", err)
	}

	patch := diffStrings(path1, path2, yaml1, yaml2)
	return patch, nil
}

// marshalToYamlString marshals value to a YAML string, while respecting
// JSON field annotations.
func marshalToYamlString(value any) (string, error) {
	jsonText, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshalling object to json: %w", err)
	}

	var jsonObject interface{}
	if err := yaml.Unmarshal(jsonText, &jsonObject); err != nil {
		return "", fmt.Errorf("unmarshalling yaml to object: %w", err)
	}

	data, err := yaml.Marshal(jsonObject)
	if err != nil {
		return "", fmt.Errorf("marshalling object to yaml: %w", err)
	}

	return string(data), nil
}

// diffStrings returns differences between two strings in unified format.
// An empty string will be returned if both strings are equal.
func diffStrings(path1, path2, s1, s2 string) string {
	// We add new lines at the end of each string to avoid
	// "\ No newline at end of file" appended to each diff.
	s1 += "\n"
	s2 += "\n"

	edits := myers.ComputeEdits("", s1, s2)
	diff := fmt.Sprint(gotextdiff.ToUnified(path1, path2, s1, edits))

	return diff
}
