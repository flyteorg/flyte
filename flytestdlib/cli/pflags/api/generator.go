package api

import (
	"context"
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/lyft/flytestdlib/logger"

	"go/importer"

	"github.com/ernesto-jimenez/gogen/gogenutil"
)

const (
	indent = "  "
)

// PFlagProviderGenerator parses and generates GetPFlagSet implementation to add PFlags for a given struct's fields.
type PFlagProviderGenerator struct {
	pkg        *types.Package
	st         *types.Named
	defaultVar *types.Var
}

// This list is restricted because that's the only kinds viper parses out, otherwise it assumes strings.
// github.com/spf13/viper/viper.go:1016
var allowedKinds = []types.Type{
	types.Typ[types.Int],
	types.Typ[types.Int8],
	types.Typ[types.Int16],
	types.Typ[types.Int32],
	types.Typ[types.Int64],
	types.Typ[types.Bool],
	types.Typ[types.String],
}

type SliceOrArray interface {
	Elem() types.Type
}

func capitalize(s string) string {
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-'a'+'A') + s[1:]
	}

	return s
}

func buildFieldForSlice(ctx context.Context, t SliceOrArray, name, goName, usage, defaultValue string) (FieldInfo, error) {
	strategy := SliceRaw
	FlagMethodName := "StringSlice"
	typ := types.NewSlice(types.Typ[types.String])
	emptyDefaultValue := `[]string{}`
	if b, ok := t.Elem().(*types.Basic); !ok {
		logger.Infof(ctx, "Elem of type [%v] is not a basic type. It must be json unmarshalable or generation will fail.", t.Elem())
		if !isJSONUnmarshaler(t.Elem()) {
			return FieldInfo{},
				fmt.Errorf("slice of type [%v] is not supported. Only basic slices or slices of json-unmarshalable types are supported",
					t.Elem().String())
		}
	} else {
		logger.Infof(ctx, "Elem of type [%v] is a basic type. Will use a pflag as a Slice.", b)
		strategy = SliceJoined
		FlagMethodName = fmt.Sprintf("%vSlice", capitalize(b.Name()))
		typ = types.NewSlice(b)
		emptyDefaultValue = fmt.Sprintf(`[]%v{}`, b.Name())
	}

	testValue := defaultValue
	if len(defaultValue) == 0 {
		defaultValue = emptyDefaultValue
		testValue = `"1,1"`
	}

	return FieldInfo{
		Name:           name,
		GoName:         goName,
		Typ:            typ,
		FlagMethodName: FlagMethodName,
		DefaultValue:   defaultValue,
		UsageString:    usage,
		TestValue:      testValue,
		TestStrategy:   strategy,
	}, nil
}

// Appends field accessors using "." as the delimiter.
// e.g. appendAccessors("var1", "field1", "subField") will output "var1.field1.subField"
func appendAccessors(accessors ...string) string {
	sb := strings.Builder{}
	switch len(accessors) {
	case 0:
		return ""
	case 1:
		return accessors[0]
	}

	for _, s := range accessors {
		if len(s) > 0 {
			if sb.Len() > 0 {
				if _, err := sb.WriteString("."); err != nil {
					fmt.Printf("Failed to writeString, error: %v", err)
					return ""
				}
			}

			if _, err := sb.WriteString(s); err != nil {
				fmt.Printf("Failed to writeString, error: %v", err)
				return ""
			}
		}
	}

	return sb.String()
}

// Traverses fields in type and follows recursion tree to discover all fields. It stops when one of two conditions is
// met; encountered a basic type (e.g. string, int... etc.) or the field type implements UnmarshalJSON.
// If passed a non-empty defaultValueAccessor, it'll be used to fill in default values instead of any default value
// specified in pflag tag.
func discoverFieldsRecursive(ctx context.Context, typ *types.Named, defaultValueAccessor, fieldPath string) ([]FieldInfo, error) {
	logger.Printf(ctx, "Finding all fields in [%v.%v.%v]",
		typ.Obj().Pkg().Path(), typ.Obj().Pkg().Name(), typ.Obj().Name())

	ctx = logger.WithIndent(ctx, indent)

	st := typ.Underlying().(*types.Struct)
	fields := make([]FieldInfo, 0, st.NumFields())
	for i := 0; i < st.NumFields(); i++ {
		v := st.Field(i)
		if !v.IsField() {
			continue
		}

		// Parses out the tag if one exists.
		tag, err := ParseTag(st.Tag(i))
		if err != nil {
			return nil, err
		}

		if len(tag.Name) == 0 {
			tag.Name = v.Name()
		}

		if tag.DefaultValue == "-" {
			logger.Infof(ctx, "Skipping field [%s], as '-' value detected", tag.Name)
			continue
		}

		typ := v.Type()
		ptr, isPtr := typ.(*types.Pointer)
		if isPtr {
			typ = ptr.Elem()
		}

		switch t := typ.(type) {
		case *types.Basic:
			if len(tag.DefaultValue) == 0 {
				tag.DefaultValue = fmt.Sprintf("*new(%v)", typ.String())
			}

			logger.Infof(ctx, "[%v] is of a basic type with default value [%v].", tag.Name, tag.DefaultValue)

			isAllowed := false
			for _, k := range allowedKinds {
				if t.String() == k.String() {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				return nil, fmt.Errorf("only these basic kinds are allowed. given [%v] (Kind: [%v]. expected: [%+v]",
					t.String(), t.Kind(), allowedKinds)
			}

			defaultValue := tag.DefaultValue
			if len(defaultValueAccessor) > 0 {
				defaultValue = appendAccessors(defaultValueAccessor, fieldPath, v.Name())

				if isPtr {
					defaultValue = fmt.Sprintf("%s.elemValueOrNil(%s).(%s)", defaultValueAccessor, defaultValue, t.Name())
				}
			}

			fields = append(fields, FieldInfo{
				Name:           tag.Name,
				GoName:         v.Name(),
				Typ:            t,
				FlagMethodName: camelCase(t.String()),
				DefaultValue:   defaultValue,
				UsageString:    tag.Usage,
				TestValue:      `"1"`,
				TestStrategy:   JSON,
			})
		case *types.Named:
			if _, isStruct := t.Underlying().(*types.Struct); !isStruct {
				// TODO: Add a more descriptive error message.
				return nil, fmt.Errorf("invalid type. it must be struct, received [%v] for field [%v]", t.Underlying().String(), tag.Name)
			}

			// If the type has json unmarshaler, then stop the recursion and assume the type is string. config package
			// will use json unmarshaler to fill in the final config object.
			jsonUnmarshaler := isJSONUnmarshaler(t)

			defaultValue := tag.DefaultValue
			if len(defaultValueAccessor) > 0 {
				defaultValue = appendAccessors(defaultValueAccessor, fieldPath, v.Name())
				if isStringer(t) {
					defaultValue = defaultValue + ".String()"
				} else {
					logger.Infof(ctx, "Field [%v] of type [%v] does not implement Stringer interface."+
						" Will use %s.mustMarshalJSON() to get its default value.", defaultValueAccessor, v.Name(), t.String())
					defaultValue = fmt.Sprintf("%s.mustMarshalJSON(%s)", defaultValueAccessor, defaultValue)
				}
			}

			testValue := defaultValue
			if len(testValue) == 0 {
				testValue = `"1"`
			}

			logger.Infof(ctx, "[%v] is of a Named type (struct) with default value [%v].", tag.Name, tag.DefaultValue)

			if jsonUnmarshaler {
				logger.Infof(logger.WithIndent(ctx, indent), "Type is json unmarhslalable.")

				fields = append(fields, FieldInfo{
					Name:           tag.Name,
					GoName:         v.Name(),
					Typ:            types.Typ[types.String],
					FlagMethodName: "String",
					DefaultValue:   defaultValue,
					UsageString:    tag.Usage,
					TestValue:      testValue,
					TestStrategy:   JSON,
				})
			} else {
				logger.Infof(ctx, "Traversing fields in type.")

				nested, err := discoverFieldsRecursive(logger.WithIndent(ctx, indent), t, defaultValueAccessor, appendAccessors(fieldPath, v.Name()))
				if err != nil {
					return nil, err
				}

				for _, subField := range nested {
					fields = append(fields, FieldInfo{
						Name:           fmt.Sprintf("%v.%v", tag.Name, subField.Name),
						GoName:         fmt.Sprintf("%v.%v", v.Name(), subField.GoName),
						Typ:            subField.Typ,
						FlagMethodName: subField.FlagMethodName,
						DefaultValue:   subField.DefaultValue,
						UsageString:    subField.UsageString,
						TestValue:      subField.TestValue,
						TestStrategy:   subField.TestStrategy,
					})
				}
			}
		case *types.Slice:
			logger.Infof(ctx, "[%v] is of a slice type with default value [%v].", tag.Name, tag.DefaultValue)

			f, err := buildFieldForSlice(logger.WithIndent(ctx, indent), t, tag.Name, v.Name(), tag.Usage, tag.DefaultValue)
			if err != nil {
				return nil, err
			}

			fields = append(fields, f)
		case *types.Array:
			logger.Infof(ctx, "[%v] is of an array with default value [%v].", tag.Name, tag.DefaultValue)

			f, err := buildFieldForSlice(logger.WithIndent(ctx, indent), t, tag.Name, v.Name(), tag.Usage, tag.DefaultValue)
			if err != nil {
				return nil, err
			}

			fields = append(fields, f)
		default:
			return nil, fmt.Errorf("unexpected type %v", t.String())
		}
	}

	return fields, nil
}

// NewGenerator initializes a PFlagProviderGenerator for pflags files for targetTypeName struct under pkg. If pkg is not filled in,
// it's assumed to be current package (which is expected to be the common use case when invoking pflags from
// go:generate comments)
func NewGenerator(pkg, targetTypeName, defaultVariableName string) (*PFlagProviderGenerator, error) {
	ctx := context.Background()
	var err error
	// Resolve package path
	if pkg == "" || pkg[0] == '.' {
		pkg, err = filepath.Abs(filepath.Clean(pkg))
		if err != nil {
			return nil, err
		}
		pkg = gogenutil.StripGopath(pkg)
	}

	targetPackage, err := importer.ForCompiler(token.NewFileSet(), "source", nil).Import(pkg)
	if err != nil {
		return nil, err
	}

	obj := targetPackage.Scope().Lookup(targetTypeName)
	if obj == nil {
		return nil, fmt.Errorf("struct %s missing", targetTypeName)
	}

	var st *types.Named
	switch obj.Type().Underlying().(type) {
	case *types.Struct:
		st = obj.Type().(*types.Named)
	default:
		return nil, fmt.Errorf("%s should be an struct, was %s", targetTypeName, obj.Type().Underlying())
	}

	var defaultVar *types.Var
	obj = targetPackage.Scope().Lookup(defaultVariableName)
	if obj != nil {
		defaultVar = obj.(*types.Var)
	}

	if defaultVar != nil {
		logger.Infof(ctx, "Using default variable with name [%v] to assign all default values.", defaultVariableName)
	} else {
		logger.Infof(ctx, "Using default values defined in tags if any.")
	}

	return &PFlagProviderGenerator{
		st:         st,
		pkg:        targetPackage,
		defaultVar: defaultVar,
	}, nil
}

func (g PFlagProviderGenerator) GetTargetPackage() *types.Package {
	return g.pkg
}

func (g PFlagProviderGenerator) Generate(ctx context.Context) (PFlagProvider, error) {
	defaultValueAccessor := ""
	if g.defaultVar != nil {
		defaultValueAccessor = g.defaultVar.Name()
	}

	fields, err := discoverFieldsRecursive(ctx, g.st, defaultValueAccessor, "")
	if err != nil {
		return PFlagProvider{}, err
	}

	return newPflagProvider(g.pkg, g.st.Obj().Name(), fields), nil
}
