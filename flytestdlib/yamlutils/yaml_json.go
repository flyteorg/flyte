package yamlutils

import (
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
)

func ReadYamlFileAsJSON(path string) ([]byte, error) {
	r, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return yaml.YAMLToJSON(r)
}
