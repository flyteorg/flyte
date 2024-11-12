// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package register

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsFilesConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementFilesConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsFilesConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookFilesConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementFilesConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

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

func decode_FilesConfig(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookFilesConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_FilesConfig(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_FilesConfig(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_FilesConfig(val, result))
}

func testDecodeRaw_FilesConfig(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_FilesConfig(vStringSlice, result))
}

func TestFilesConfig_GetPFlagSet(t *testing.T) {
	val := FilesConfig{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestFilesConfig_SetFlags(t *testing.T) {
	actual := FilesConfig{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_version", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("version", testValue)
			if vString, err := cmdFlags.GetString("version"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.Version)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_force", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("force", testValue)
			if vBool, err := cmdFlags.GetBool("force"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vBool), &actual.Force)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_continueOnError", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("continueOnError", testValue)
			if vBool, err := cmdFlags.GetBool("continueOnError"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vBool), &actual.ContinueOnError)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_archive", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("archive", testValue)
			if vBool, err := cmdFlags.GetBool("archive"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vBool), &actual.Archive)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_assumableIamRole", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("assumableIamRole", testValue)
			if vString, err := cmdFlags.GetString("assumableIamRole"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.AssumableIamRole)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_k8sServiceAccount", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("k8sServiceAccount", testValue)
			if vString, err := cmdFlags.GetString("k8sServiceAccount"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.K8sServiceAccount)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_k8ServiceAccount", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("k8ServiceAccount", testValue)
			if vString, err := cmdFlags.GetString("k8ServiceAccount"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.K8ServiceAccount)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_outputLocationPrefix", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("outputLocationPrefix", testValue)
			if vString, err := cmdFlags.GetString("outputLocationPrefix"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.OutputLocationPrefix)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_sourceUploadPath", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("sourceUploadPath", testValue)
			if vString, err := cmdFlags.GetString("sourceUploadPath"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.DeprecatedSourceUploadPath)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_destinationDirectory", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("destinationDirectory", testValue)
			if vString, err := cmdFlags.GetString("destinationDirectory"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vString), &actual.DestinationDirectory)

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
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vBool), &actual.DryRun)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_enableSchedule", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("enableSchedule", testValue)
			if vBool, err := cmdFlags.GetBool("enableSchedule"); err == nil {
				testDecodeJson_FilesConfig(t, fmt.Sprintf("%v", vBool), &actual.EnableSchedule)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
