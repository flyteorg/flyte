package subcommand

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"sigs.k8s.io/yaml"
)

// WriteConfigToFile used for marshaling the Config to a file which can then be used for update/delete
func WriteConfigToFile(matchableAttrConfig interface{}, fileName string) error {
	d, err := yaml.Marshal(matchableAttrConfig)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	if _, err = os.Stat(fileName); err == nil {
		if !cmdUtil.AskForConfirmation(fmt.Sprintf("warning file %v will be overwritten", fileName), os.Stdin) {
			return fmt.Errorf("backup the file before continuing")
		}
	}
	return ioutil.WriteFile(fileName, d, 0600)
}

// String Dumps the json representation of the TaskResourceAttrFileConfig
func String(matchableAttrConfig interface{}) string {
	tj, err := json.Marshal(matchableAttrConfig)
	if err != nil {
		fmt.Println(err)
		return "marshaling error"
	}
	return fmt.Sprintf("%s\n", tj)
}

// ReadConfigFromFile used for unmarshaling the Config from a file which is used for update/delete
func ReadConfigFromFile(matchableAttrConfig interface{}, fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read from %v yaml file", fileName)
	}
	if err = yaml.UnmarshalStrict(data, matchableAttrConfig); err != nil {
		return err
	}
	return nil
}

func DumpTaskResourceAttr(matchableAttrConfig interface{}, fileName string) error {
	// Write Config to file if filename provided in the command
	if len(fileName) > 0 {
		if err := WriteConfigToFile(matchableAttrConfig, fileName); err != nil {
			return fmt.Errorf("error dumping in file due to %v", err)
		}
		fmt.Printf("wrote the config to file %v", fileName)
	} else {
		fmt.Printf("%v", String(matchableAttrConfig))
	}
	return nil
}