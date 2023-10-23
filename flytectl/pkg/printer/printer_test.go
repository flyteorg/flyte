package printer

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
)

type Inner struct {
	X string     `json:"x"`
	Y *time.Time `json:"y"`
}

func LaunchplanToProtoMessages(l []*admin.LaunchPlan) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

func WorkflowToProtoMessages(l []*admin.Workflow) []proto.Message {
	messages := make([]proto.Message, 0, len(l))
	for _, m := range l {
		messages = append(messages, m)
	}
	return messages
}

// TODO Convert this to a Testable Example. For some reason the comparison fails
func TestJSONToTable(t *testing.T) {
	trunc := 5
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	j := []struct {
		A string `json:"a"`
		B int    `json:"b"`
		S *Inner `json:"s"`
	}{
		{"hello", 0, &Inner{"x-hello", nil}},
		{"hello world", 0, &Inner{"x-hello", &d}},
		{"hello", 0, nil},
	}

	b, err := json.Marshal(j)
	assert.NoError(t, err)
	p := Printer{}
	assert.NoError(t, p.JSONToTable(b, []Column{
		{"A", "$.a", &trunc},
		{"S", "$.s.y", nil},
	}))
	// Output:
	// | A     | S                    |
	// ------- ----------------------
	// | hello |                      |
	// | hello | 2020-01-01T00:00:00Z |
	// | hello |                      |
	// 3 rows
}

func TestOutputFormats(t *testing.T) {
	expected := []string{"TABLE", "JSON", "YAML", "DOT", "DOTURL"}
	outputs := OutputFormats()
	assert.Equal(t, 5, len(outputs))
	assert.Equal(t, expected, outputs)
}

func TestOutputFormatString(t *testing.T) {
	o, err := OutputFormatString("JSON")
	assert.Nil(t, err)
	assert.Equal(t, OutputFormat(1), o)
}

func TestOutputFormatStringErr(t *testing.T) {
	o, err := OutputFormatString("FLYTE")
	assert.NotNil(t, err)
	assert.Equal(t, OutputFormat(0), o)
	assert.Equal(t, fmt.Errorf("%s does not belong to OutputFormat values", "FLYTE"), err)
}

func TestIsAOutputFormat(t *testing.T) {
	o := OutputFormat(5)
	check := o.IsAOutputFormat()
	assert.Equal(t, false, check)

	o = OutputFormat(1)
	check = o.IsAOutputFormat()
	assert.Equal(t, true, check)
}

func TestMarshalJson(t *testing.T) {
	o := OutputFormat(1)
	check, err := o.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, []byte(`"JSON"`), check)

	result, err := o.MarshalYAML()
	assert.Nil(t, err)
	assert.Equal(t, "JSON", result)
}

func TestPrint(t *testing.T) {
	p := Printer{}
	lp := []Column{
		{Header: "Version", JSONPath: "$.id.version"},
		{Header: "Name", JSONPath: "$.id.name"},
	}
	launchPlan := &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "launchplan1",
			Version: "v2",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt:      &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{},
		},
	}
	launchPlans := []*admin.LaunchPlan{launchPlan}
	err := p.Print(OutputFormat(0), lp, LaunchplanToProtoMessages(launchPlans)...)
	assert.Nil(t, err)
	err = p.Print(OutputFormat(1), lp, LaunchplanToProtoMessages(launchPlans)...)
	assert.Nil(t, err)
	err = p.Print(OutputFormat(2), lp, LaunchplanToProtoMessages(launchPlans)...)
	assert.Nil(t, err)
	err = p.Print(OutputFormat(3), lp, LaunchplanToProtoMessages(launchPlans)...)
	assert.NotNil(t, err)
	err = p.Print(OutputFormat(4), lp, LaunchplanToProtoMessages(launchPlans)...)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("visualization is only supported on workflows"), err)

	sortedListLiteralType := core.Variable{
		Type: &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	variableMap := map[string]*core.Variable{
		"sorted_list1": &sortedListLiteralType,
		"sorted_list2": &sortedListLiteralType,
	}

	var compiledTasks []*core.CompiledTask
	compiledTasks = append(compiledTasks, &core.CompiledTask{
		Template: &core.TaskTemplate{
			Id: &core.Identifier{
				Project: "dummyProject",
				Domain:  "dummyDomain",
				Name:    "dummyName",
				Version: "dummyVersion",
			},
			Interface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: variableMap,
				},
			},
		},
	})

	workflow1 := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Tasks: compiledTasks,
			},
		},
	}

	workflows := []*admin.Workflow{workflow1}

	err = p.Print(OutputFormat(3), lp, WorkflowToProtoMessages(workflows)...)
	assert.Nil(t, err)
	workflows = []*admin.Workflow{}
	err = p.Print(OutputFormat(3), lp, WorkflowToProtoMessages(workflows)...)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("atleast one workflow required for visualization"), err)
	var badCompiledTasks []*core.CompiledTask
	badCompiledTasks = append(badCompiledTasks, &core.CompiledTask{
		Template: &core.TaskTemplate{},
	})
	badWorkflow := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Tasks: badCompiledTasks,
			},
		},
	}
	workflows = []*admin.Workflow{badWorkflow}
	err = p.Print(OutputFormat(3), lp, WorkflowToProtoMessages(workflows)...)
	assert.NotNil(t, err)

	assert.Equal(t, fmt.Errorf("no template found in the workflow task template:<> "), errors.Unwrap(err))

	badWorkflow2 := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt:        &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: nil,
		},
	}
	workflows = []*admin.Workflow{badWorkflow2}
	err = p.Print(OutputFormat(3), lp, WorkflowToProtoMessages(workflows)...)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("empty workflow closure"), errors.Unwrap(err))

	var badSubWorkflow []*core.CompiledWorkflow
	badSubWorkflow = append(badSubWorkflow, &core.CompiledWorkflow{
		Template: &core.WorkflowTemplate{},
	})

	badWorkflow3 := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Tasks:        compiledTasks,
				SubWorkflows: badSubWorkflow,
			},
		},
	}
	workflows = []*admin.Workflow{badWorkflow3}
	err = p.Print(OutputFormat(3), lp, WorkflowToProtoMessages(workflows)...)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("no template found in the sub workflow template:<> "), errors.Unwrap(err))
}

func TestGetTruncatedLine(t *testing.T) {
	testStrings := map[string]string{
		"foo":                        "foo",
		"":                           "",
		"short description":          "short description",
		"1234567890123456789012345":  "1234567890123456789012345",
		"12345678901234567890123456": "1234567890123456789012...",
		"long description probably needs truncate": "long description proba...",
	}
	for k, v := range testStrings {
		assert.Equal(t, v, getTruncatedLine(k))
	}
}

func TestFormatVariableDescriptions(t *testing.T) {
	fooVar := &core.Variable{
		Description: "foo",
	}
	barVar := &core.Variable{
		Description: "bar",
	}
	variableMap := map[string]*core.Variable{
		"var1": fooVar,
		"var2": barVar,
		"foo":  fooVar,
		"bar":  barVar,
	}
	FormatVariableDescriptions(variableMap)
	assert.Equal(t, "bar\nfoo\nvar1: foo\nvar2: bar", variableMap[DefaultFormattedDescriptionsKey].Description)
}

func TestFormatParameterDescriptions(t *testing.T) {
	fooParam := &core.Parameter{
		Var: &core.Variable{
			Description: "foo",
		},
	}
	barParam := &core.Parameter{
		Var: &core.Variable{
			Description: "bar",
		},
	}
	emptyParam := &core.Parameter{}
	paramMap := map[string]*core.Parameter{
		"var1":  fooParam,
		"var2":  barParam,
		"foo":   fooParam,
		"bar":   barParam,
		"empty": emptyParam,
	}
	FormatParameterDescriptions(paramMap)
	assert.Equal(t, "bar\nfoo\nvar1: foo\nvar2: bar", paramMap[DefaultFormattedDescriptionsKey].Var.Description)
}
