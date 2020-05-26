package utils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lyft/flyteidl/clients/go/coreutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRegexCommandArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		inputFileRegex.MatchString("{{ .InputFile }}")
	}
}

type dummyInputReader struct {
	inputPrefix storage.DataReference
	inputPath   storage.DataReference
	inputs      *core.LiteralMap
	inputErr    bool
}

func (d dummyInputReader) GetInputPrefixPath() storage.DataReference {
	return d.inputPrefix
}

func (d dummyInputReader) GetInputPath() storage.DataReference {
	return d.inputPath
}

func (d dummyInputReader) Get(ctx context.Context) (*core.LiteralMap, error) {
	if d.inputErr {
		return nil, fmt.Errorf("expected input fetch error")
	}
	return d.inputs, nil
}

type dummyOutputPaths struct {
	outputPath storage.DataReference
}

func (d dummyOutputPaths) GetRawOutputPrefix() storage.DataReference {
	panic("should not be called")
}

func (d dummyOutputPaths) GetOutputPrefixPath() storage.DataReference {
	return d.outputPath
}

func (d dummyOutputPaths) GetOutputPath() storage.DataReference {
	panic("should not be called")
}

func (d dummyOutputPaths) GetErrorPath() storage.DataReference {
	panic("should not be called")
}

func TestInputRegexMatch(t *testing.T) {
	assert.True(t, inputFileRegex.MatchString("{{$input}}"))
	assert.True(t, inputFileRegex.MatchString("{{ $Input }}"))
	assert.True(t, inputFileRegex.MatchString("{{.input}}"))
	assert.True(t, inputFileRegex.MatchString("{{ .Input }}"))
	assert.True(t, inputFileRegex.MatchString("{{  .Input }}"))
	assert.True(t, inputFileRegex.MatchString("{{       .Input }}"))
	assert.True(t, inputFileRegex.MatchString("{{ .Input}}"))
	assert.True(t, inputFileRegex.MatchString("{{.Input }}"))
	assert.True(t, inputFileRegex.MatchString("--something={{.Input}}"))
	assert.False(t, inputFileRegex.MatchString("{{input}}"), "Missing $")
	assert.False(t, inputFileRegex.MatchString("{$input}}"), "Missing Brace")
}

func TestOutputRegexMatch(t *testing.T) {
	assert.True(t, outputRegex.MatchString("{{.OutputPrefix}}"))
	assert.True(t, outputRegex.MatchString("{{ .OutputPrefix }}"))
	assert.True(t, outputRegex.MatchString("{{  .OutputPrefix }}"))
	assert.True(t, outputRegex.MatchString("{{      .OutputPrefix }}"))
	assert.True(t, outputRegex.MatchString("{{ .OutputPrefix}}"))
	assert.True(t, outputRegex.MatchString("{{.OutputPrefix }}"))
	assert.True(t, outputRegex.MatchString("--something={{.OutputPrefix}}"))
	assert.False(t, outputRegex.MatchString("{{output}}"), "Missing $")
	assert.False(t, outputRegex.MatchString("{.OutputPrefix}}"), "Missing Brace")
}

func TestReplaceTemplateCommandArgs(t *testing.T) {
	t.Run("empty cmd", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(),
			[]string{}, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, []string{}, actual)
	})

	in := dummyInputReader{inputPath: "input/blah"}
	out := dummyOutputPaths{outputPath: "output/blah"}

	t.Run("nothing to substitute", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
		}, in, out)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
		}, actual)
	})

	t.Run("Sub InputFile", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
		}, in, out)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
		}, actual)
	})

	t.Run("Sub Input Prefix", func(t *testing.T) {
		in := dummyInputReader{inputPath: "input/prefix"}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
		}, in, out)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
			"input/prefix",
		}, actual)
	})

	t.Run("Sub Output Prefix", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"output/blah",
		}, actual)
	})

	t.Run("Sub Input Output prefix", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
		}, actual)
	})

	t.Run("Bad input template", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"${{input}}",
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"${{input}}",
			"output/blah",
		}, actual)
	})

	t.Run("Input arg", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": {
					Value: &core.Literal_Collection{
						Collection: &core.LiteralCollection{
							Literals: []*core.Literal{coreutils.MustMakeLiteral("a"), coreutils.MustMakeLiteral("b")},
						},
					},
				},
			},
		}}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [a,b]",
			"output/blah",
		}, actual)
	})

	t.Run("Date", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"date": coreutils.MustMakeLiteral(time.Date(1900, 01, 01, 01, 01, 01, 000000001, time.UTC)),
			},
		}}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.date }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg 1900-01-01T01:01:01.000000001Z",
			"output/blah",
		}, actual)
	})

	t.Run("2d Array arg", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
			},
		}}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [[a,b],[1,2]]",
			"output/blah",
		}, actual)
	})

	t.Run("nil input", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{}}

		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"output/blah",
		}, actual)
	})

	t.Run("multi-input", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"ds":    coreutils.MustMakeLiteral(time.Date(1900, 01, 01, 01, 01, 01, 000000001, time.UTC)),
				"table": coreutils.MustMakeLiteral("my_table"),
				"hr":    coreutils.MustMakeLiteral("hr"),
				"min":   coreutils.MustMakeLiteral(15),
			},
		}}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			`SELECT
        	COUNT(*) as total_count
    	FROM
        	hive.events.{{ .Inputs.table }}
    	WHERE
        	ds = '{{ .Inputs.ds }}' AND hr = '{{ .Inputs.hr }}' AND min = {{ .Inputs.min }}
	    `}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			`SELECT
        	COUNT(*) as total_count
    	FROM
        	hive.events.my_table
    	WHERE
        	ds = '1900-01-01T01:01:01.000000001Z' AND hr = 'hr' AND min = 15
	    `}, actual)
	})

	t.Run("missing input", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
			},
		}}
		_, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.Error(t, err)
	})

	t.Run("bad template", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
			},
		}}
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah blah }}`,
			"{{ .OutputPrefix }}",
		}, in, out)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah blah }}`,
			"output/blah",
		}, actual)
	})
}
