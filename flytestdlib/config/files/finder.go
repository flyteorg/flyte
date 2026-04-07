package files

import (
	"os"
	"path/filepath"
)

const (
	configFileType = "yaml"
	configFileName = "config"
)

var configLocations = [][]string{
	{"."},
	{"/etc", "flyte", "config"},
	{os.ExpandEnv("$GOPATH"), "src", "github.com", "lyft", "flytestdlib"},
}

// Check if File / Directory Exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func isFile(path string) (bool, error) {
	s, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return !s.IsDir(), nil
}

func contains(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}

	return false
}

// Finds config files in search paths. If searchPaths is empty, it'll look in default locations (see configLocations above)
// If searchPaths is not empty but no configs are found there, it'll still look into configLocations.
// If it found any config file in searchPaths, it'll stop the search.
// searchPaths can contain patterns to match (behavior is OS-dependent). And it'll try to Glob the pattern for any matching
// files.
func FindConfigFiles(searchPaths []string) []string {
	res := make([]string, 0, 1)

	for _, location := range searchPaths {
		matchedFiles, err := filepath.Glob(location)
		if err != nil {
			continue
		}

		for _, matchedFile := range matchedFiles {
			if file, err := isFile(matchedFile); err == nil && file && !contains(res, matchedFile) {
				res = append(res, matchedFile)
			}
		}
	}

	if len(res) == 0 {
		for _, location := range configLocations {
			pathToTest := filepath.Join(append(location, configFileName+"."+configFileType)...)
			if b, err := exists(pathToTest); err == nil && b && !contains(res, pathToTest) {
				res = append(res, pathToTest)
			}
		}
	}

	return res
}
