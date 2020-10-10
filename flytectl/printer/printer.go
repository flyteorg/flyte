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

type Column struct {
	Header   string
	JSONPath string
}

type Printer struct{}

const (
	empty = ""
	tab   = "\t"
)

// Projects the columns in one row of data from the given JSON using the []Column map
func extractRow(data interface{}, columns []Column) []string {
	if columns == nil || data == nil {
		return nil
	}
	tableData := make([]string, 0, len(columns))

	for _, c := range columns {
		out, err := jsonpath.Read(data, c.JSONPath)
		if err != nil {
			out = ""
		}
		tableData = append(tableData, fmt.Sprintf("%s", out))
	}
	return tableData
}

// Projects the columns from the given list of JSON elements using the []Column map
// Potential performance problem, as it returns all the rows in memory.
// We could use the render row, but that may lead to misalignment.
// TODO figure out a more optimal way
func projectColumns(input []interface{}, column []Column) ([][]string, error) {
	responses := make([][]string, 0, len(input))
	for _, data := range input {
		responses = append(responses, extractRow(data, column))
	}
	return responses, nil
}

func JSONToTable(b []byte, columns []Column) error {
	var jsonRows []interface{}
	err := json.Unmarshal(b, &jsonRows)
	if err != nil {
		return err
	}
	if jsonRows == nil {
		return nil
	}
	rows, err := projectColumns(jsonRows, columns)
	if err != nil {
		return err
	}
	printer := tableprinter.New(os.Stdout)
	// TODO make this configurable
	printer.AutoWrapText = false
	printer.BorderLeft = true
	printer.BorderRight = true
	printer.ColumnSeparator = "|"
	printer.HeaderBgColor = tablewriter.BgHiWhiteColor
	headers := make([]string, 0, len(columns))
	positions := make([]int, 0, len(columns))
	for _, c := range columns {
		headers = append(headers, c.Header)
		positions = append(positions, 30)
	}
	if r := printer.Render(headers, rows, positions, true); r == -1 {
		return fmt.Errorf("failed to render table")
	}
	fmt.Printf("%d rows\n", len(rows))
	return nil
}

func (p Printer) Print(format OutputFormat, i interface{}, columns []Column) error {

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
		return JSONToTable(buf.Bytes(), columns)
	}
	return nil
}
