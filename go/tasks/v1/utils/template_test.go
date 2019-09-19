package utils

import (
	"bytes"
	"context"
	"testing"
	"text/template"
	"time"

	"github.com/lyft/flyteidl/clients/go/coreutils"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRegexCommandArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		inputFileRegex.MatchString("{{ .InputFile }}")
	}
}

// Benchmark results:
// Regex_replacement-8         	 3000000	       583 ns/op
// NotCompiled-8               	  100000	     14684 ns/op
// Precompile/Execute-8        	  500000	      2706 ns/op
func BenchmarkReplacements(b *testing.B) {
	cmd := `abc {{ .Inputs.x }} `
	cmdTemplate := `abc {{ index .Inputs "x" }}`
	cmdArgs := CommandLineTemplateArgs{
		Input: "inputfile.pb",
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
			},
		},
	}

	b.Run("NotCompiled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			t, err := template.New("NotCompiled").Parse(cmdTemplate)
			assert.NoError(b, err)
			var buf bytes.Buffer
			err = t.Execute(&buf, cmdArgs)
			assert.NoError(b, err)
		}
	})

	b.Run("Precompile", func(b *testing.B) {
		t, err := template.New("NotCompiled").Parse(cmdTemplate)
		assert.NoError(b, err)

		b.Run("Execute", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				err = t.Execute(&buf, cmdArgs)
				assert.NoError(b, err)
			}
		})
	})

	b.Run("Regex replacement", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			inputVarRegex.FindAllStringSubmatchIndex(cmd, -1)
		}
	})
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
			[]string{},
			CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
		assert.NoError(t, err)
		assert.Equal(t, []string{}, actual)
	})

	t.Run("nothing to substitute", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
		}, CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
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
		}, CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
		}, actual)
	})

	t.Run("Sub Output Prefix", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			"{{ .OutputPrefix }}",
		}, CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
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
		}, CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
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
		}, CommandLineTemplateArgs{Input: "input/blah", OutputPrefix: "output/blah"})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"${{input}}",
			"output/blah",
		}, actual)
	})

	t.Run("Input arg", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, CommandLineTemplateArgs{
			Input:        "input/blah",
			OutputPrefix: "output/blah",
			Inputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"arr": {
						Value: &core.Literal_Collection{
							Collection: &core.LiteralCollection{
								Literals: []*core.Literal{coreutils.MustMakeLiteral("a"), coreutils.MustMakeLiteral("b")},
							},
						},
					},
				},
			}})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [a,b]",
			"output/blah",
		}, actual)
	})

	t.Run("Date", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.date }}`,
			"{{ .OutputPrefix }}",
		}, CommandLineTemplateArgs{
			Input:        "input/blah",
			OutputPrefix: "output/blah",
			Inputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"date": coreutils.MustMakeLiteral(time.Date(1900, 01, 01, 01, 01, 01, 000000001, time.UTC)),
				},
			}})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg 1900-01-01T01:01:01.000000001Z",
			"output/blah",
		}, actual)
	})

	t.Run("2d Array arg", func(t *testing.T) {
		actual, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, CommandLineTemplateArgs{
			Input:        "input/blah",
			OutputPrefix: "output/blah",
			Inputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{
					"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
				},
			}})
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [[a,b],[1,2]]",
			"output/blah",
		}, actual)
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := ReplaceTemplateCommandArgs(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
		}, CommandLineTemplateArgs{
			Input:        "input/blah",
			OutputPrefix: "output/blah",
			Inputs:       &core.LiteralMap{Literals: nil}})
		assert.Error(t, err)
	})
}
