package template

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/shamaton/msgpack/v2"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flyteidl2/clients/go/coreutils"
	pluginsCoreMocks "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

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
	outputPath          storage.DataReference
	rawOutputDataPrefix storage.DataReference
	prevCheckpointPath  storage.DataReference
	checkpointPath      storage.DataReference
}

func (d dummyOutputPaths) GetDeckPath() storage.DataReference {
	panic("should not be called")
}

func (d dummyOutputPaths) GetPreviousCheckpointsPrefix() storage.DataReference {
	return d.prevCheckpointPath
}

func (d dummyOutputPaths) GetCheckpointPrefix() storage.DataReference {
	return d.checkpointPath
}

func (d dummyOutputPaths) GetRawOutputPrefix() storage.DataReference {
	return d.rawOutputDataPrefix
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

func TestReplaceTemplateCommandArgs(t *testing.T) {
	taskExecutionID := &pluginsCoreMocks.TaskExecutionID{}
	taskExecutionID.EXPECT().GetGeneratedName().Return("per_retry_unique_key")
	taskMetadata := &pluginsCoreMocks.TaskExecutionMetadata{}
	taskMetadata.EXPECT().GetTaskExecutionID().Return(taskExecutionID)
	taskMetadata.EXPECT().GetNamespace().Return("test-namespace")

	t.Run("empty cmd", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{}, Parameters{})
		assert.NoError(t, err)
		assert.Equal(t, []string{}, actual)
	})

	in := dummyInputReader{inputPath: "input/blah"}
	out := dummyOutputPaths{
		outputPath:          "output/blah",
		rawOutputDataPrefix: "s3://custom-bucket",
	}

	params := Parameters{
		TaskExecMetadata: taskMetadata,
		Inputs:           in,
		OutputPath:       out,
		Task:             nil,
	}
	t.Run("nothing to substitute", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
		}, params)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
		}, actual)
	})

	t.Run("Sub InputFile", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
		}, params)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
		}, actual)
	})

	t.Run("Sub Input Prefix", func(t *testing.T) {
		in := dummyInputReader{inputPath: "input/prefix"}
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
		}, params)
		assert.NoError(t, err)

		assert.Equal(t, []string{
			"hello",
			"world",
			"input/prefix",
		}, actual)
	})

	t.Run("Sub Output Prefix", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .OutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"output/blah",
		}, actual)
	})

	t.Run("Sub Input Output prefix", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
		}, actual)
	})

	t.Run("Bad input template", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"${{input}}",
			"{{ .OutputPrefix }}",
			"--switch {{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"${{input}}",
			"output/blah",
			"--switch s3://custom-bucket",
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
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
			"{{ $RawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [a,b]",
			"output/blah",
			"s3://custom-bucket",
		}, actual)
	})

	t.Run("Date", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"date": coreutils.MustMakeLiteral(time.Date(1900, 01, 01, 01, 01, 01, 000000001, time.UTC)),
			},
		}}
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.date }}`,
			"{{ .OutputPrefix }}",
			"{{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg 1900-01-01T01:01:01.000000001Z",
			"output/blah",
			"s3://custom-bucket",
		}, actual)
	})

	t.Run("2d Array arg", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
			},
		}}
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
			"{{ .wrongOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"--someArg [[a,b],[1,2]]",
			"output/blah",
			"{{ .wrongOutputDataPrefix }}",
		}, actual)
	})

	t.Run("nil input", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{}}
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}

		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"{{ .OutputPrefix }}",
			"--raw-data-output-prefix {{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.arr }}`,
			"output/blah",
			"--raw-data-output-prefix s3://custom-bucket",
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
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			`SELECT
        	COUNT(*) as total_count
    	FROM
        	hive.events.{{ .Inputs.table }}
    	WHERE
        	ds = '{{ .Inputs.ds }}' AND hr = '{{ .Inputs.hr }}' AND min = {{ .Inputs.min }}
	    `}, params)
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
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		_, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah }}`,
			"{{ .OutputPrefix }}",
		}, params)
		assert.Error(t, err)
	})

	t.Run("bad template", func(t *testing.T) {
		in := dummyInputReader{inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"arr": coreutils.MustMakeLiteral([]interface{}{[]interface{}{"a", "b"}, []interface{}{1, 2}}),
			},
		}}
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             nil,
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah blah }} {{ .PerretryuNIqueKey }}`,
			"{{ .OutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			`--someArg {{ .Inputs.blah blah }} per_retry_unique_key`,
			"output/blah",
		}, actual)
	})

	t.Run("sub raw output data prefix", func(t *testing.T) {
		actual, err := Render(context.TODO(), []string{
			"hello",
			"{{ .perRetryUniqueKey }}",
			"world",
			"{{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"per_retry_unique_key",
			"world",
			"s3://custom-bucket",
		}, actual)
	})

	t.Run("sub task template happy", func(t *testing.T) {
		ctx := context.TODO()
		tMock := &pluginsCoreMocks.TaskTemplatePath{}
		tMock.EXPECT().Path(ctx).Return("s3://task-path", nil)
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             tMock,
		}

		actual, err := Render(ctx, []string{
			"hello",
			"{{ .perRetryUniqueKey }}",
			"world",
			"{{ .taskTemplatePath }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"per_retry_unique_key",
			"world",
			"s3://task-path",
		}, actual)
	})

	t.Run("sub task template error", func(t *testing.T) {
		ctx := context.TODO()
		tMock := &pluginsCoreMocks.TaskTemplatePath{}
		tMock.EXPECT().Path(ctx).Return("", fmt.Errorf("error"))
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath:       out,
			Task:             tMock,
		}

		_, err := Render(ctx, []string{
			"hello",
			"{{ .perRetryUniqueKey }}",
			"world",
			"{{ .taskTemplatePath }}",
		}, params)
		assert.Error(t, err)
	})

	t.Run("missing checkpoint args", func(t *testing.T) {
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath: dummyOutputPaths{
				outputPath:          out.outputPath,
				rawOutputDataPrefix: out.rawOutputDataPrefix,
				prevCheckpointPath:  "s3://prev-checkpoint/prefix",
				checkpointPath:      "s3://new-checkpoint/prefix",
			},
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
		}, actual)
	})

	t.Run("no prev checkpoint", func(t *testing.T) {
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath: dummyOutputPaths{
				outputPath:          out.outputPath,
				rawOutputDataPrefix: out.rawOutputDataPrefix,
				prevCheckpointPath:  "",
				checkpointPath:      "s3://new-checkpoint/prefix",
			},
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
			"--prev={{ .PrevCheckpointPrefix }}",
			"--checkpoint={{ .CheckpointOutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
			"--prev=\"\"",
			"--checkpoint=s3://new-checkpoint/prefix",
		}, actual)
	})

	t.Run("all checkpoints", func(t *testing.T) {
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath: dummyOutputPaths{
				outputPath:          out.outputPath,
				rawOutputDataPrefix: out.rawOutputDataPrefix,
				prevCheckpointPath:  "s3://prev-checkpoint/prefix",
				checkpointPath:      "s3://new-checkpoint/prefix",
			},
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
			"--prev={{ .PrevCheckpointPrefix }}",
			"--checkpoint={{ .CheckpointOutputPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
			"--prev=s3://prev-checkpoint/prefix",
			"--checkpoint=s3://new-checkpoint/prefix",
		}, actual)
	})

	t.Run("all checkpoints ignore case", func(t *testing.T) {
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath: dummyOutputPaths{
				outputPath:          out.outputPath,
				rawOutputDataPrefix: out.rawOutputDataPrefix,
				prevCheckpointPath:  "s3://prev-checkpoint/prefix",
				checkpointPath:      "s3://new-checkpoint/prefix",
			},
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
			"--prev={{ .prevcheckpointprefix }}",
			"--checkpoint={{ .checkpointoutputprefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
			"--prev=s3://prev-checkpoint/prefix",
			"--checkpoint=s3://new-checkpoint/prefix",
		}, actual)
	})

	t.Run("namespace embedded replacement", func(t *testing.T) {
		params := Parameters{
			TaskExecMetadata: taskMetadata,
			Inputs:           in,
			OutputPath: dummyOutputPaths{
				outputPath:          out.outputPath,
				rawOutputDataPrefix: "s3://raw-data/prefix/{{ .Namespace }}",
				prevCheckpointPath:  "s3://prev-checkpoint/prefix/{{ .Namespace}}",
				checkpointPath:      "s3://new-checkpoint/prefix/{{.namespace}}",
			},
		}
		actual, err := Render(context.TODO(), []string{
			"hello",
			"world",
			"{{ .Input }}",
			"{{ .OutputPrefix }}",
			"--prev={{ .prevcheckpointprefix }}",
			"--checkpoint={{ .checkpointoutputprefix }}",
			"--raw-data-output={{ .rawoutputdataprefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"world",
			"input/blah",
			"output/blah",
			"--prev=s3://prev-checkpoint/prefix/test-namespace",
			"--checkpoint=s3://new-checkpoint/prefix/test-namespace",
			"--raw-data-output=s3://raw-data/prefix/test-namespace",
		}, actual)
	})
}

func TestReplaceTemplateCommandArgsSpecialChars(t *testing.T) {
	in := dummyInputReader{inputPath: "input/blah"}
	out := dummyOutputPaths{
		outputPath:          "output/blah",
		rawOutputDataPrefix: "s3://custom-bucket",
	}

	params := Parameters{Inputs: in, OutputPath: out}

	t.Run("dashes are replaced", func(t *testing.T) {
		taskExecutionID := &pluginsCoreMocks.TaskExecutionID{}
		taskExecutionID.EXPECT().GetGeneratedName().Return("per-retry-unique-key")
		taskMetadata := &pluginsCoreMocks.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetTaskExecutionID().Return(taskExecutionID)
		taskMetadata.EXPECT().GetNamespace().Return("my-namespace")

		params.TaskExecMetadata = taskMetadata
		actual, err := Render(context.TODO(), []string{
			"hello",
			"{{ .perRetryUniqueKey }}",
			"world",
			"{{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"per_retry_unique_key",
			"world",
			"s3://custom-bucket",
		}, actual)
	})

	t.Run("non-alphabet leading characters are stripped", func(t *testing.T) {
		var startsWithAlpha = regexp.MustCompile("^[^a-zA-Z_]+")
		taskExecutionID := &pluginsCoreMocks.TaskExecutionID{}
		taskExecutionID.EXPECT().GetGeneratedName().Return("33 per retry-unique-key")
		taskMetadata := &pluginsCoreMocks.TaskExecutionMetadata{}
		taskMetadata.EXPECT().GetTaskExecutionID().Return(taskExecutionID)
		taskMetadata.EXPECT().GetNamespace().Return("my-namespace")

		params.TaskExecMetadata = taskMetadata
		testString := "doesn't start with a number"
		testString2 := "1 does start with a number"
		testString3 := "  1 3 nd spaces "
		assert.Equal(t, testString, startsWithAlpha.ReplaceAllString(testString, "a"))
		assert.Equal(t, "adoes start with a number", startsWithAlpha.ReplaceAllString(testString2, "a"))
		assert.Equal(t, "and spaces ", startsWithAlpha.ReplaceAllString(testString3, "a"))

		actual, err := Render(context.TODO(), []string{
			"hello",
			"{{ .perRetryUniqueKey }}",
			"world",
			"{{ .rawOutputDataPrefix }}",
		}, params)
		assert.NoError(t, err)
		assert.Equal(t, []string{
			"hello",
			"aper_retry_unique_key",
			"world",
			"s3://custom-bucket",
		}, actual)
	})
}

func BenchmarkRegexCommandArgs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		inputFileRegex.MatchString("{{ .InputFile }}")
	}
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

func getBlobLiteral(uri string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Blob{
					Blob: &core.Blob{
						Metadata: nil,
						Uri:      uri,
					},
				},
			},
		},
	}
}

func getSchemaLiteral(uri string) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Schema{
					Schema: &core.Schema{Type: nil, Uri: uri},
				},
			},
		},
	}
}

func TestSerializeLiteral(t *testing.T) {
	ctx := context.Background()

	t.Run("serialize blob", func(t *testing.T) {
		b := getBlobLiteral("asdf fdsa")
		interpolated, err := serializeLiteral(ctx, b)
		assert.NoError(t, err)
		assert.Equal(t, "asdf fdsa", interpolated)
	})

	t.Run("serialize blob", func(t *testing.T) {
		s := getSchemaLiteral("s3://some-bucket/fdsa/x.parquet")
		interpolated, err := serializeLiteral(ctx, s)
		assert.NoError(t, err)
		assert.Equal(t, "s3://some-bucket/fdsa/x.parquet", interpolated)
	})
}

func TestSerializeLiteralScalar_BinaryMessagePack(t *testing.T) {
	// Create a simple map to be serialized into MessagePack format
	testMap := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": 1.1,
		"d": "string",
	}

	// Serialize the map using MessagePack
	encodedData, err := msgpack.Marshal(testMap)
	assert.NoError(t, err)

	// Create the core.Scalar_Binary with the encoded MessagePack data and MESSAGEPACK tag
	binaryScalar := &core.Scalar{
		Value: &core.Scalar_Binary{
			Binary: &core.Binary{
				Value: encodedData,
				Tag:   coreutils.MESSAGEPACK,
			},
		},
	}

	// Call the function we want to test
	result, err := serializeLiteralScalar(binaryScalar)
	assert.NoError(t, err)

	// Since the map should be decoded back, we expect a simple string representation of the map
	expectedResult := "map[a:1 b:true c:1.1 d:string]"
	assert.Equal(t, expectedResult, result)
}

func TestSerializeLiteralScalar_BinaryUnsupportedTag(t *testing.T) {
	// Create some binary data for testing
	binaryData := []byte{0x01, 0x02, 0x03}

	// Create a core.Scalar_Binary with an unsupported tag
	binaryScalar := &core.Scalar{
		Value: &core.Scalar_Binary{
			Binary: &core.Binary{
				Value: binaryData,
				Tag:   "unsupported-tag",
			},
		},
	}

	// Call the function and expect an error because the tag is unsupported
	_, err := serializeLiteralScalar(binaryScalar)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported binary tag")
}
