package subcommand

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	cmdUtil "github.com/flyteorg/flytectl/pkg/commandutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"

	"sigs.k8s.io/yaml"
)

// TaskResourceAttrFileConfig shadow config for TaskResourceAttribute.
// The shadow config is not using ProjectDomainAttribute/Workflowattribute directly inorder to simplify the inputs.
// As the same structure is being used for both ProjectDomainAttribute/Workflowattribute
type TaskResourceAttrFileConfig struct {
	Project  string `json:"project"`
	Domain   string `json:"domain"`
	Workflow string `json:"workflow,omitempty"`
	*admin.TaskResourceAttributes
}

// WriteConfigToFile used for marshaling the config to a file which can then be used for update/delete
func (t TaskResourceAttrFileConfig) WriteConfigToFile(fileName string) error {
	d, err := yaml.Marshal(t)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	if _, err = os.Stat(fileName); err == nil {
		if !cmdUtil.AskForConfirmation(fmt.Sprintf("warning file %v will be overwritten", fileName)) {
			return fmt.Errorf("backup the file before continuing")
		}
	}
	return ioutil.WriteFile(fileName, d, 0600)
}

// Dumps the json representation of the TaskResourceAttrFileConfig
func (t TaskResourceAttrFileConfig) String() string {
	tj, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
		return "marshaling error"
	}
	return fmt.Sprintf("%s\n", tj)
}

// ReadConfigFromFile used for unmarshaling the config from a file which is used for update/delete
func (t *TaskResourceAttrFileConfig) ReadConfigFromFile(fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read from %v yaml file", fileName)
	}
	if err = yaml.UnmarshalStrict(data, t); err != nil {
		return err
	}
	return nil
}

// MatchableAttributeDecorator decorator over TaskResourceAttributes. Similar decorator would exist for other MatchingAttributes
func (t *TaskResourceAttrFileConfig) MatchableAttributeDecorator() *admin.MatchingAttributes {
	return &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_TaskResourceAttributes{
			TaskResourceAttributes: t.TaskResourceAttributes,
		},
	}
}

func (t TaskResourceAttrFileConfig) DumpTaskResourceAttr(ctx context.Context, fileName string) {
	// Dump empty task resource attr for editing
	// Write config to file if filename provided in the command
	if len(fileName) > 0 {
		// Read the config from the file and update the TaskResourceAttrFileConfig with the TaskResourceAttrConfig
		if err := t.WriteConfigToFile(fileName); err != nil {
			fmt.Printf("error dumping in file due to %v", err)
			logger.Warnf(ctx, "error dumping in file due to %v", err)
			return
		}
		fmt.Printf("wrote the config to file %v", fileName)
	} else {
		fmt.Printf("%v", t)
	}
}
