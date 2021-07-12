package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"
	hversion "github.com/hashicorp/go-version"
)

const (
	HTTPRequestErrorMessage = "something went wrong. Received status code [%v] while sending a request to [%s]"
)

type githubversion struct {
	TagName string `json:"tag_name"`
}

func GetRequest(baseURL, url string) ([]byte, error) {
	response, err := http.Get(fmt.Sprintf("%s%s", baseURL, url))
	if err != nil {
		return []byte(""), err
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return []byte(""), err
		}
		return data, nil
	}
	return []byte(""), fmt.Errorf(HTTPRequestErrorMessage, response.StatusCode, fmt.Sprintf("%s%s", baseURL, url))
}

func ParseGithubTag(data []byte) (string, error) {
	var result = githubversion{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result.TagName, nil
}

func WriteIntoFile(data []byte, file string) error {
	err := ioutil.WriteFile(file, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// SetupFlyteDir will create .flyte dir if not exist
func SetupFlyteDir() error {
	if err := os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func IsVersionGreaterThan(version1, version2 string) (bool, error) {
	semanticVersion1, err := hversion.NewVersion(version1)
	if err != nil {
		return false, err
	}
	semanticVersion2, err := hversion.NewVersion(version2)
	if err != nil {
		return false, err
	}
	return semanticVersion2.LessThanOrEqual(semanticVersion1), nil
}
