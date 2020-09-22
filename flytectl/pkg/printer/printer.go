package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/landoop/tableprinter"
	"github.com/yalp/jsonpath"
	"os"
)

type Printer struct{}

const (
	empty = ""
	tab   = "\t"
)

func (p Printer) PrintOutput(output string, i interface{}) {
	// Factory Method for all printer
	switch output {
	case "json": // Print protobuf to json
		buffer := new(bytes.Buffer)
		encoder := json.NewEncoder(buffer)
		encoder.SetIndent(empty, tab)

		err := encoder.Encode(i)
		if err != nil {
			os.Exit(1)
		}

		fmt.Println(buffer.String())
		break
	default: // Print table

		printer := tableprinter.New(os.Stdout)
		printer.Print(i)
		break
	}
}

func(p Printer) BuildOutput(input []interface{},column map[string]string,printTransform func(data []byte)(interface{},error)) ([]interface{},error) {
	responses := make([]interface{}, 0, len(input))
	for _, data := range input {
		tableData := make(map[string]interface{})

		for k := range column {
			out, _ := jsonpath.Read(data, column[k])
			tableData[k] = out.(string)
		}
		jsonbody, err := json.Marshal(tableData)
		if err != nil {
			return responses,err
		}
		response,err := printTransform(jsonbody)
		if err != nil {
			return responses,err
		}
		responses = append(responses, response)
	}
	return responses,nil
}

func (p Printer) Print(output string, i interface{},column map[string]string,printTransform func(data []byte)(interface{},error)) {

	var data interface{}
	byte, _ := json.Marshal(i)
	_ = json.Unmarshal(byte, &data)
	if data == nil {
		os.Exit(1)
	}
	input := data.([]interface{})
	response,err := p.BuildOutput(input,column,printTransform)
	if err != nil {
		os.Exit(1)
	}
	p.PrintOutput(output, response)
}
