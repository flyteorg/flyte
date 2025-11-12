package yamlutils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"
)

func ReadYamlFileAsJSON(path string) ([]byte, error) {
	r, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return yaml.YAMLToJSON(r)
}
