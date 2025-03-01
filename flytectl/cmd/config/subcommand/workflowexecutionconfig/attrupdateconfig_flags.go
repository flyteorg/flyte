// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package workflowexecutionconfig

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (AttrUpdateConfig) elemValueOrNil(v interface{}) interface{} {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).IsNil() {
			return reflect.Zero(t.Elem()).Interface()
		} else {
			return reflect.ValueOf(v).Interface()
		}
	} else if v == nil {
		return reflect.Zero(t).Interface()
	}

	return v
}

func (AttrUpdateConfig) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (AttrUpdateConfig) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in AttrUpdateConfig and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg AttrUpdateConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("AttrUpdateConfig", pflag.ExitOnError)
    cmdFlags.StringSliceVar(&DefaultUpdateConfig.AttrFile, fmt.Sprintf("%v%v", prefix, "attrFile"), DefaultUpdateConfig.AttrFile, "attribute file name to be used for updating attribute for the resource type.")
	cmdFlags.BoolVar(&DefaultUpdateConfig.DryRun, fmt.Sprintf("%v%v", prefix, "dryRun"), DefaultUpdateConfig.DryRun, "execute command without making any modifications.")
	cmdFlags.BoolVar(&DefaultUpdateConfig.Force, fmt.Sprintf("%v%v", prefix, "force"), DefaultUpdateConfig.Force, "do not ask for an acknowledgement during updates.")
	return cmdFlags
}
