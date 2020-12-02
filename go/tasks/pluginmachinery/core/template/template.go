package template

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/lyft/flytestdlib/logger"

	"reflect"

	"github.com/golang/protobuf/ptypes"
	idlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/pkg/errors"
)

var alphaNumericOnly = regexp.MustCompile("[^a-zA-Z0-9_]+")
var startsWithAlpha = regexp.MustCompile("^[^a-zA-Z_]+")

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

// Evaluates templates in each command with the equivalent value from passed args. Templates are case-insensitive
// Supported templates are:
// - {{ .InputFile }} to receive the input file path. The protocol used will depend on the underlying system
// 		configuration. E.g. s3://bucket/key/to/file.pb or /var/run/local.pb are both valid.
// - {{ .OutputPrefix }} to receive the path prefix for where to store the outputs.
// - {{ .Inputs.myInput }} to receive the actual value of the input passed. See docs on LiteralMapToTemplateArgs for how
// 		what to expect each literal type to be serialized as.
// If a command isn't a valid template or failed to evaluate, it'll be returned as is.
// NOTE: I wanted to do in-place replacement, until I realized that in-place replacement will alter the definition of the
// graph. This is not desirable, as we may have to retry and in that case the replacement will not work and we want
// to create a new location for outputs
func ReplaceTemplateCommandArgs(ctx context.Context, tExecMeta core.TaskExecutionMetadata, command []string, in io.InputReader,
	out io.OutputFilePaths) ([]string, error) {

	// TODO: Change GetGeneratedName to follow these conventions
	var perRetryUniqueKey = tExecMeta.GetTaskExecutionID().GetGeneratedName()
	perRetryUniqueKey = startsWithAlpha.ReplaceAllString(perRetryUniqueKey, "a")
	perRetryUniqueKey = alphaNumericOnly.ReplaceAllString(perRetryUniqueKey, "_")

	logger.Debugf(ctx, "Using [%s] from [%s]", perRetryUniqueKey, tExecMeta.GetTaskExecutionID().GetGeneratedName())

	if len(command) == 0 {
		return []string{}, nil
	}
	if in == nil || out == nil {
		return nil, fmt.Errorf("input reader and output path cannot be nil")
	}
	res := make([]string, 0, len(command))
	for _, commandTemplate := range command {
		updated, err := replaceTemplateCommandArgs(ctx, perRetryUniqueKey, commandTemplate, in, out)
		if err != nil {
			return res, err
		}

		res = append(res, updated)
	}

	return res, nil
}

var inputFileRegex = regexp.MustCompile(`(?i){{\s*[\.$]Input\s*}}`)
var inputPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]InputPrefix\s*}}`)
var outputRegex = regexp.MustCompile(`(?i){{\s*[\.$]OutputPrefix\s*}}`)
var inputVarRegex = regexp.MustCompile(`(?i){{\s*[\.$]Inputs\.(?P<input_name>[^}\s]+)\s*}}`)
var rawOutputDataPrefixRegex = regexp.MustCompile(`(?i){{\s*[\.$]RawOutputDataPrefix\s*}}`)
var perRetryUniqueKey = regexp.MustCompile(`(?i){{\s*[\.$]PerRetryUniqueKey\s*}}`)

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

func replaceTemplateCommandArgs(ctx context.Context, perRetryKey string, commandTemplate string,
	in io.InputReader, out io.OutputFilePaths) (string, error) {

	val := inputFileRegex.ReplaceAllString(commandTemplate, in.GetInputPath().String())
	val = outputRegex.ReplaceAllString(val, out.GetOutputPrefixPath().String())
	val = inputPrefixRegex.ReplaceAllString(val, in.GetInputPrefixPath().String())
	val = rawOutputDataPrefixRegex.ReplaceAllString(val, out.GetRawOutputPrefix().String())
	val = perRetryUniqueKey.ReplaceAllString(val, perRetryKey)

	inputs, err := in.Get(ctx)
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
