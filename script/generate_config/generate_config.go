package main

import (
	"fmt"
	propellerConfig "github.com/flyteorg/flytepropeller/pkg/controller/config"
	resourceManagerConfig "github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/olekukonko/tablewriter"
	"os"
	"reflect"
	"strings"
)

type example struct {
	Id        int    `json:"id" pflag:",Specifies the project to work on."`
	CreatedAt string `json:"created_at"`
	Tag       string `json:"tag"`
	Text      string `json:"text"`
	AuthorId  int    `json:"author_id"`
}

func PrintConfigTable(b interface{}, name string) {
	if name != "" {
		fmt.Println(name)
		fmt.Println("===============================")
	}
	val := reflect.ValueOf(b)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Description"})
	table.SetAlignment(3)
	table.SetRowLine(true)

	if val.Type().Field(0).Tag.Get("json") == "" || val.Type().Field(0).Tag.Get("pflag") == ""{
		return
	}
	fmt.Println(val.Type().Name())
	fmt.Println("^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^")

	for i := 0; i < val.Type().NumField(); i++ {
		t := val.Type().Field(i)
		fieldName := t.Name
		fieldType := t.Type.String()
		fieldDescription := ""

		if t.Type.Kind() == reflect.Struct{
			defer PrintConfigTable(val.Field(i).Interface(), "")
		}

		if jsonTag := t.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			var commaIdx int
			if commaIdx = strings.Index(jsonTag, ","); commaIdx < 0 {
				commaIdx = len(jsonTag)
			}
			fieldName = jsonTag[:commaIdx]
		}

		if pFlag := t.Tag.Get("pflag"); pFlag != "" && pFlag != "-" {
			var commaIdx int
			if commaIdx = strings.Index(pFlag, ","); commaIdx < 0 {
				commaIdx = -1
			}
			fieldDescription = pFlag[commaIdx+1:]
		}
		data := []string{fieldName, fieldType, fieldDescription}
		table.Append(data)
	}
	table.Render()
	fmt.Println()
}

func main() {
	// fmt.Println("====================")
	fmt.Println(".. _deployment-cluster-config-specification:")
	fmt.Println()

	fmt.Println("Flyte Configuration Specification")
	fmt.Println("---------------------------------------------")
	fmt.Println()

	PrintConfigTable(*storage.GetConfig(), "Storage Configuration")
	PrintConfigTable(*logger.GetConfig(), "Logger Configuration")
	PrintConfigTable(*propellerConfig.GetConfig(), "Propeller Configuration")
	PrintConfigTable(*resourceManagerConfig.GetConfig(), "ResourceManager Configuration")


}
