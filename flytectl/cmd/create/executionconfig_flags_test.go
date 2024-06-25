// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package create

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsExecutionConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementExecutionConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsExecutionConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookExecutionConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementExecutionConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

		raw, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Failed to marshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		res := reflect.New(to).Interface()
		err = json.Unmarshal(raw, &res)
		if err != nil {
			fmt.Printf("Failed to umarshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		return res, nil
	}

	return data, nil
}

func decode_ExecutionConfig(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookExecutionConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_ExecutionConfig(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_ExecutionConfig(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_ExecutionConfig(val, result))
}

func testDecodeRaw_ExecutionConfig(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_ExecutionConfig(vStringSlice, result))
}

func TestExecutionConfig_GetPFlagSet(t *testing.T) {
	val := ExecutionConfig{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestExecutionConfig_SetFlags(t *testing.T) {
	actual := ExecutionConfig{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_execFile", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("execFile", testValue)
			if vString, err := cmdFlags.GetString("execFile"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.ExecFile)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_targetDomain", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("targetDomain", testValue)
			if vString, err := cmdFlags.GetString("targetDomain"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.TargetDomain)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_targetProject", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("targetProject", testValue)
			if vString, err := cmdFlags.GetString("targetProject"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.TargetProject)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_targetExecutionCluster", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("targetExecutionCluster", testValue)
			if vString, err := cmdFlags.GetString("targetExecutionCluster"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.TargetExecutionCluster)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_kubeServiceAcct", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("kubeServiceAcct", testValue)
			if vString, err := cmdFlags.GetString("kubeServiceAcct"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.KubeServiceAcct)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_iamRoleARN", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("iamRoleARN", testValue)
			if vString, err := cmdFlags.GetString("iamRoleARN"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.IamRoleARN)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_relaunch", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("relaunch", testValue)
			if vString, err := cmdFlags.GetString("relaunch"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.Relaunch)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_recover", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("recover", testValue)
			if vString, err := cmdFlags.GetString("recover"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.Recover)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dryRun", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("dryRun", testValue)
			if vBool, err := cmdFlags.GetBool("dryRun"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vBool), &actual.DryRun)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_version", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("version", testValue)
			if vString, err := cmdFlags.GetString("version"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.Version)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_clusterPool", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("clusterPool", testValue)
			if vString, err := cmdFlags.GetString("clusterPool"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.ClusterPool)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_overwriteCache", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("overwriteCache", testValue)
			if vBool, err := cmdFlags.GetBool("overwriteCache"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vBool), &actual.OverwriteCache)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_workflow", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("workflow", testValue)
			if vString, err := cmdFlags.GetString("workflow"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.Workflow)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_task", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("task", testValue)
			if vString, err := cmdFlags.GetString("task"); err == nil {
				testDecodeJson_ExecutionConfig(t, fmt.Sprintf("%v", vString), &actual.Task)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
