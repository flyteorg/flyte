package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytectl/pkg/visualize"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/errors"

	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"github.com/pkg/browser"
	"github.com/yalp/jsonpath"
)

//go:generate enumer --type=OutputFormat -json -yaml -trimprefix=OutputFormat
type OutputFormat uint8

const (
	OutputFormatTABLE OutputFormat = iota
	OutputFormatJSON
	OutputFormatYAML
	OutputFormatDOT
	OutputFormatDOTURL
)

// Set implements PFlag's Value interface to attempt to set the value of the flag from string.
func (i *OutputFormat) Set(val string) error {
	policy, err := OutputFormatString(val)
	if err != nil {
		return err
	}

	*i = policy
	return nil
}

// Type implements PFlag's Value interface to return type name.
func (i OutputFormat) Type() string {
	return "OutputFormat"
}

const GraphVisualizationServiceURL = "http://graph.flyte.org/#"

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
	// Optional Truncation directive to limit content. This will simply truncate the string output.
	TruncateTo *int
}

type Printer struct{}

const (
	empty                           = ""
	tab                             = "\t"
	DefaultFormattedDescriptionsKey = "_formatted_descriptions"
	defaultLineWidth                = 25
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
		s := fmt.Sprintf("%s", out)
		if c.TruncateTo != nil {
			t := *c.TruncateTo
			if len(s) > t {
				s = s[:t]
			}
		}
		tableData = append(tableData, s)
	}
	return tableData
}

// Projects the columns from the given list of JSON elements using the []Column map
// Potential performance problem, as it returns all the rows in memory.
// We could use the render row, but that may lead to misalignment.
// TODO figure out a more optimal way
func projectColumns(rows []interface{}, column []Column) [][]string {
	responses := make([][]string, 0, len(rows))
	for _, row := range rows {
		responses = append(responses, extractRow(row, column))
	}
	return responses
}

func (p Printer) JSONToTable(jsonRows []byte, columns []Column) error {
	var rawRows []interface{}
	if err := json.Unmarshal(jsonRows, &rawRows); err != nil {
		return errors.Wrapf("JSONUnmarshalFailure", err, "failed to unmarshal into []interface{} from json")
	}
	if rawRows == nil {
		return errors.Errorf("JSONUnmarshalNil", "expected one row or empty rows, received nil")
	}
	rows := projectColumns(rawRows, columns)

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

func (p Printer) PrintInterface(format OutputFormat, columns []Column, v interface{}) error {
	jsonRows, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// Factory Method for all printer
	switch format {
	case OutputFormatJSON, OutputFormatYAML:
		return printJSONYaml(format, v)
	default: // Print table
		return p.JSONToTable(jsonRows, columns)
	}
}

// printJSONYaml internal function for printing
func printJSONYaml(format OutputFormat, v interface{}) error {
	if format != OutputFormatJSON && format != OutputFormatYAML {
		return fmt.Errorf("this function should be called only for json/yaml printing")
	}
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent(empty, tab)

	err := encoder.Encode(v)
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
	return nil
}

func FormatVariableDescriptions(variableMap map[string]*core.Variable) {
	keys := make([]string, 0, len(variableMap))
	// sort the keys for testing and consistency with other output formats
	for k := range variableMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var descriptions []string
	for _, k := range keys {
		v := variableMap[k]
		// a: a isn't very helpful
		if k != v.Description {
			descriptions = append(descriptions, getTruncatedLine(fmt.Sprintf("%s: %s", k, v.Description)))
		} else {
			descriptions = append(descriptions, getTruncatedLine(k))
		}

	}
	variableMap[DefaultFormattedDescriptionsKey] = &core.Variable{Description: strings.Join(descriptions, "\n")}
}

func FormatParameterDescriptions(parameterMap map[string]*core.Parameter) {
	keys := make([]string, 0, len(parameterMap))
	// sort the keys for testing and consistency with other output formats
	for k := range parameterMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var descriptions []string
	for _, k := range keys {
		v := parameterMap[k]
		if v.Var == nil {
			continue
		}
		// a: a isn't very helpful
		if k != v.Var.Description {
			descriptions = append(descriptions, getTruncatedLine(fmt.Sprintf("%s: %s", k, v.Var.Description)))
		} else {
			descriptions = append(descriptions, getTruncatedLine(k))
		}
	}
	parameterMap[DefaultFormattedDescriptionsKey] = &core.Parameter{Var: &core.Variable{Description: strings.Join(descriptions, "\n")}}
}

func getTruncatedLine(line string) string {
	// TODO: maybe add width to function signature later
	width := defaultLineWidth
	if len(line) > width {
		return line[:width-3] + "..."
	}
	return line
}

func (p Printer) Print(format OutputFormat, columns []Column, messages ...proto.Message) error {

	printableMessages := make([]*PrintableProto, 0, len(messages))
	for _, m := range messages {
		printableMessages = append(printableMessages, &PrintableProto{Message: m})
	}

	// Factory Method for all printer
	switch format {
	case OutputFormatJSON, OutputFormatYAML: // Print protobuf to json
		var v interface{}
		if len(printableMessages) == 1 {
			v = printableMessages[0]
		} else {
			v = printableMessages
		}
		return printJSONYaml(format, v)
	case OutputFormatDOT, OutputFormatDOTURL:
		var workflows []*admin.Workflow
		for _, m := range messages {
			if w, ok := m.(*admin.Workflow); ok {
				workflows = append(workflows, w)
			} else {
				return fmt.Errorf("visualization is only supported on workflows")
			}
		}
		if len(workflows) == 0 {
			return fmt.Errorf("atleast one workflow required for visualization")
		}
		workflow := workflows[0]
		graphStr, err := visualize.RenderWorkflow(workflow.Closure.CompiledWorkflow)
		if err != nil {
			return errors.Wrapf("VisualizationError", err, "failed to visualize workflow")
		}
		if format == OutputFormatDOTURL {
			urlToOpen := GraphVisualizationServiceURL + url.PathEscape(graphStr)
			fmt.Println("Opening the browser at " + urlToOpen)
			return browser.OpenURL(urlToOpen)
		}
		fmt.Println(graphStr)
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
