package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/lyft/flytestdlib/errors"
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
		if err != nil || out == nil {
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
func projectColumns(rows []interface{}, column []Column) ([][]string, error) {
	responses := make([][]string, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, extractRow(row, column))
	}
	return responses, nil
}

func (p Printer) JSONToTable(jsonRows []byte, columns []Column) error {
	var rawRows []interface{}
	if err := json.Unmarshal(jsonRows, &rawRows); err != nil {
		return errors.Wrapf("JSONUnmarshalFailure", err, "failed to unmarshal into []interface{} from json")
	}
	if rawRows == nil {
		return errors.Errorf("JSONUnmarshalNil", "expected one row or empty rows, received nil")
	}
	rows, err := projectColumns(rawRows, columns)
	if err != nil {
		return err
	}
	printer := tableprinter.New(os.Stdout)
	// TODO make this configurable
	printer.AutoWrapText = false
	printer.BorderLeft = true
	printer.BorderRight = true
	printer.BorderBottom = true
	printer.BorderTop = true
	printer.RowLine = true
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

func (p Printer) Print(format OutputFormat, columns []Column, messages ...proto.Message) error {

	printableMessages := make([]*PrintableProto, 0, len(messages))
	for _, m := range messages {
		printableMessages = append(printableMessages, &PrintableProto{Message: m})
	}

	// Factory Method for all printer
	switch format {
	case OutputFormatJSON, OutputFormatYAML: // Print protobuf to json
		buf := new(bytes.Buffer)
		encoder := json.NewEncoder(buf)
		encoder.SetIndent(empty, tab)
		var err error
		if len(printableMessages) == 1 {
			err = encoder.Encode(printableMessages[0])
		} else {
			err = encoder.Encode(printableMessages)
		}
		if err != nil {
			return err
		}
		if format == OutputFormatJSON {
			fmt.Println(buf.String())
		} else {
			v, err := yaml.JSONToYAML(buf.Bytes())
			if err != nil {
				return err
			}
			fmt.Println(string(v))
		}
	default: // Print table
		rows, err := json.Marshal(printableMessages)
		if err != nil {
			return errors.Wrapf("ProtoToJSONFailure", err, "failed to marshal proto messages")
		}
		return p.JSONToTable(rows, columns)
	}
	return nil
}

type PrintableProto struct {
	proto.Message
}

var marshaller = jsonpb.Marshaler{
	Indent: tab,
}

func (p PrintableProto) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := marshaller.Marshal(buf, p.Message)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
