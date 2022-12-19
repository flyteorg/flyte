// Package template exports the Render method
// Render Evaluates templates in each command with the equivalent value from passed args. Templates are case-insensitive
// Supported templates are:
//   - {{ .InputFile }} to receive the input file path. The protocol used will depend on the underlying system
//     configuration. E.g. s3://bucket/key/to/file.pb or /var/run/local.pb are both valid.
//   - {{ .OutputPrefix }} to receive the path prefix for where to store the outputs.
//   - {{ .Inputs.myInput }} to receive the actual value of the input passed. See docs on LiteralMapToTemplateArgs for how
//     what to expect each literal type to be serialized as.
//   - {{ .RawOutputDataPrefix }} to receive a path where the raw output data should be ideally written. It is guaranteed
//     to be unique per retry and finally one will be saved as the output path
//   - {{ .PerRetryUniqueKey }} A key/id/str that is generated per retry and is guaranteed to be unique. Useful in query
//     manipulations
//   - {{ .TaskTemplatePath }} A path in blobstore/metadata store (e.g. s3, gcs etc) to where an offloaded version of the
//     task template exists and can be accessed by the container / task execution environment. The template is a
//     a serialized protobuf
//   - {{ .PrevCheckpointPrefix }} A path to the checkpoint directory for the previous attempt. If this is the first attempt
//     then this is replaced by an empty string
//   - {{ .CheckpointOutputPrefix }} A Flyte aware path where the current execution should write the checkpoints.
package template

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/ptypes"
	"github.com/pkg/errors"
)

var alphaNumericOnly = regexp.MustCompile("[^a-zA-Z0-9_]+")
var startsWithAlpha = regexp.MustCompile("^[^a-zA-Z_]+")

// Regexes for Supported templates
var inputFileRegex = regexp.MustCompile(`(?i){{\s*[\.$]Input\s*}}`)
var inputPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]InputPrefix\s*}}`)
var outputRegex = regexp.MustCompile(`(?i){{\s*[\.$]OutputPrefix\s*}}`)
var inputVarRegex = regexp.MustCompile(`(?i){{\s*[\.$]Inputs\.(?P<input_name>[^}\s]+)\s*}}`)
var rawOutputDataPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]RawOutputDataPrefix\s*}}`)
var perRetryUniqueKey = regexp.MustCompile(`(?i){{\s*[\.$]PerRetryUniqueKey\s*}}`)
var taskTemplateRegex = regexp.MustCompile(`(?i){{\s*[\.$]TaskTemplatePath\s*}}`)
var prevCheckpointPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]PrevCheckpointPrefix\s*}}`)
var currCheckpointPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]CheckpointOutputPrefix\s*}}`)

type ErrorCollection struct {
	Errors []error
}

func (e ErrorCollection) Error() string {
	sb := strings.Builder{}
	for idx, err := range e.Errors {
		sb.WriteString(fmt.Sprintf("%v: %v\r\n", idx, err))
	}

	return sb.String()
}

// Parameters struct is used by the Templating Engine to replace the templated parameters
type Parameters struct {
	TaskExecMetadata core.TaskExecutionMetadata
	Inputs           io.InputReader
	OutputPath       io.OutputFilePaths
	Task             core.TaskTemplatePath
}

// Render Evaluates templates in each command with the equivalent value from passed args. Templates are case-insensitive
// If a command isn't a valid template or failed to evaluate, it'll be returned as is.
// Refer to the package docs for a list of supported templates
// NOTE: I wanted to do in-place replacement, until I realized that in-place replacement will alter the definition of the
// graph. This is not desirable, as we may have to retry and in that case the replacement will not work and we want
// to create a new location for outputs
func Render(ctx context.Context, inputTemplate []string, params Parameters) ([]string, error) {
	if len(inputTemplate) == 0 {
		return []string{}, nil
	}

	// TODO: Change GetGeneratedName to follow these conventions
	var perRetryUniqueKey = params.TaskExecMetadata.GetTaskExecutionID().GetGeneratedName()
	perRetryUniqueKey = startsWithAlpha.ReplaceAllString(perRetryUniqueKey, "a")
	perRetryUniqueKey = alphaNumericOnly.ReplaceAllString(perRetryUniqueKey, "_")

	logger.Debugf(ctx, "Using [%s] from [%s]", perRetryUniqueKey, params.TaskExecMetadata.GetTaskExecutionID().GetGeneratedName())
	if params.Inputs == nil || params.OutputPath == nil {
		return nil, fmt.Errorf("input reader and output path cannot be nil")
	}
	res := make([]string, 0, len(inputTemplate))
	for _, t := range inputTemplate {
		updated, err := render(ctx, t, params, perRetryUniqueKey)
		if err != nil {
			return res, err
		}

		res = append(res, updated)
	}

	return res, nil
}

func render(ctx context.Context, inputTemplate string, params Parameters, perRetryKey string) (string, error) {

	val := inputFileRegex.ReplaceAllString(inputTemplate, params.Inputs.GetInputPath().String())
	val = outputRegex.ReplaceAllString(val, params.OutputPath.GetOutputPrefixPath().String())
	val = inputPrefixRegex.ReplaceAllString(val, params.Inputs.GetInputPrefixPath().String())
	val = rawOutputDataPrefixRegex.ReplaceAllString(val, params.OutputPath.GetRawOutputPrefix().String())
	prevCheckpoint := params.OutputPath.GetPreviousCheckpointsPrefix().String()
	if prevCheckpoint == "" {
		prevCheckpoint = "\"\""
	}
	val = prevCheckpointPrefixRegex.ReplaceAllString(val, prevCheckpoint)
	val = currCheckpointPrefixRegex.ReplaceAllString(val, params.OutputPath.GetCheckpointPrefix().String())
	val = perRetryUniqueKey.ReplaceAllString(val, perRetryKey)

	// For Task template, we will replace only if there is a match. This is because, task template replacement
	// may be expensive, as we may offload
	if taskTemplateRegex.MatchString(val) {
		p, err := params.Task.Path(ctx)
		if err != nil {
			logger.Debugf(ctx, "Failed to substitute Task Template reference - reason %s", err)
			return "", err
		}
		val = taskTemplateRegex.ReplaceAllString(val, p.String())
	}

	inputs, err := params.Inputs.Get(ctx)
	if err != nil {
		return val, errors.Wrapf(err, "unable to read inputs")
	}
	if inputs == nil || inputs.Literals == nil {
		return val, nil
	}

	var errs ErrorCollection
	val = inputVarRegex.ReplaceAllStringFunc(val, func(s string) string {
		matches := inputVarRegex.FindAllStringSubmatch(s, 1)
		varName := matches[0][1]
		replaced, err := transformVarNameToStringVal(ctx, varName, inputs)
		if err != nil {
			errs.Errors = append(errs.Errors, errors.Wrapf(err, "input template [%s]", s))
			return ""
		}
		return replaced
	})

	if len(errs.Errors) > 0 {
		return "", errs
	}

	return val, nil
}

func transformVarNameToStringVal(ctx context.Context, varName string, inputs *idlCore.LiteralMap) (string, error) {
	inputVal, exists := inputs.Literals[varName]
	if !exists {
		return "", fmt.Errorf("requested input is not found [%s]", varName)
	}

	v, err := serializeLiteral(ctx, inputVal)
	if err != nil {
		return "", errors.Wrapf(err, "failed to bind a value to inputName [%s]", varName)
	}
	return v, nil
}

func serializePrimitive(p *idlCore.Primitive) (string, error) {
	switch o := p.Value.(type) {
	case *idlCore.Primitive_Integer:
		return fmt.Sprintf("%v", o.Integer), nil
	case *idlCore.Primitive_Boolean:
		return fmt.Sprintf("%v", o.Boolean), nil
	case *idlCore.Primitive_Datetime:
		return ptypes.TimestampString(o.Datetime), nil
	case *idlCore.Primitive_Duration:
		return o.Duration.String(), nil
	case *idlCore.Primitive_FloatValue:
		return fmt.Sprintf("%v", o.FloatValue), nil
	case *idlCore.Primitive_StringValue:
		return o.StringValue, nil
	default:
		return "", fmt.Errorf("received an unexpected primitive type [%v]", reflect.TypeOf(p.Value))
	}
}

func serializeLiteralScalar(l *idlCore.Scalar) (string, error) {
	switch o := l.Value.(type) {
	case *idlCore.Scalar_Primitive:
		return serializePrimitive(o.Primitive)
	case *idlCore.Scalar_Blob:
		return o.Blob.Uri, nil
	case *idlCore.Scalar_Schema:
		return o.Schema.Uri, nil
	default:
		return "", fmt.Errorf("received an unexpected scalar type [%v]", reflect.TypeOf(l.Value))
	}
}

func serializeLiteral(ctx context.Context, l *idlCore.Literal) (string, error) {
	switch o := l.Value.(type) {
	case *idlCore.Literal_Collection:
		res := make([]string, 0, len(o.Collection.Literals))
		for _, sub := range o.Collection.Literals {
			s, err := serializeLiteral(ctx, sub)
			if err != nil {
				return "", err
			}

			res = append(res, s)
		}

		return fmt.Sprintf("[%v]", strings.Join(res, ",")), nil
	case *idlCore.Literal_Scalar:
		return serializeLiteralScalar(o.Scalar)
	default:
		logger.Debugf(ctx, "received unexpected primitive type")
		return "", fmt.Errorf("received an unexpected primitive type [%v]", reflect.TypeOf(l.Value))
	}
}
