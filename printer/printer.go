package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/yalp/jsonpath"
)

//go:generate enumer --type=OutputFormat -json -yaml -trimprefix=OutputFormat
type OutputFormat uint8

const (
	OutputFormatTABLE OutputFormat = iota
	OutputFormatJSON
	OutputFormatYAML
)

func OutputFormats() []string {
	var v []string
	for _, o := range OutputFormatValues() {
		v = append(v, o.String())
	}
	return v
}

type Printer struct{}

const (
	empty = ""
	tab   = "\t"
)

func (p Printer) projectColumns(input []interface{}, column map[string]string, printTransform func(data []byte) (interface{}, error)) ([]interface{}, error) {
	responses := make([]interface{}, 0, len(input))
	for _, data := range input {
		tableData := make(map[string]interface{})

		for k := range column {
			out, err := jsonpath.Read(data, column[k])
			if err != nil {
				out = nil
			}
			tableData[k] = out
		}
		jsonbody, err := json.Marshal(tableData)
		if err != nil {
			return responses, err
		}
		response, err := printTransform(jsonbody)
		if err != nil {
			return responses, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}

func (p Printer) Print(format OutputFormat, i interface{}, column map[string]string, printTransform func(data []byte) (interface{}, error)) error {

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(i)
	if err != nil {
		return err
	}

	// Factory Method for all printer
	switch format {
	case OutputFormatJSON: // Print protobuf to json
		fmt.Println(buf.String())
	case OutputFormatYAML:
		v, err := yaml.JSONToYAML(buf.Bytes())
		if err != nil {
			return err
		}
		fmt.Println(string(v))
	default: // Print table
		var rows []interface{}
		err := json.Unmarshal(buf.Bytes(), &rows)
		if err != nil {
			return err
		}
		if rows == nil {
			return nil
		}
		response, err := p.projectColumns(rows, column, printTransform)
		if err != nil {
			return err
		}
		printer := tableprinter.New(os.Stdout)
		printer.AutoWrapText = false
		printer.BorderLeft = true
		printer.BorderRight = true
		printer.ColumnSeparator = "|"
		printer.HeaderBgColor = tablewriter.BgHiWhiteColor
		if printer.Print(response) == -1 {
			return fmt.Errorf("failed to print table data")
		}
		fmt.Printf("%d rows\n", len(rows))
	}
	return nil
}
