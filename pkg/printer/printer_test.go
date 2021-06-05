package printer

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
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

// TODO Convert this to a Testable Example. For some reason the comparison fails
func TestJSONToTable(t *testing.T) {
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	j := []struct {
		A string `json:"a"`
		B int    `json:"b"`
		S *Inner `json:"s"`
	}{
		{"hello", 0, &Inner{"x-hello", nil}},
		{"hello", 0, &Inner{"x-hello", &d}},
		{"hello", 0, nil},
	}

	b, err := json.Marshal(j)
	assert.NoError(t, err)
	p := Printer{}
	assert.NoError(t, p.JSONToTable(b, []Column{
		{"A", "$.a"},
		{"S", "$.s.y"},
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
	expected := []string{"TABLE", "JSON", "YAML"}
	outputs := OutputFormats()
	assert.Equal(t, 3, len(outputs))
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
	o := OutputFormat(4)
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
}
