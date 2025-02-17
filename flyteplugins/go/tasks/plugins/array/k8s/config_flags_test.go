// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package k8s

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferenceableKindsConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementConfig(t reflect.Kind) bool {
	_, exists := dereferenceableKindsConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

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

func decode_Config(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_Config(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_Config(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_Config(val, result))
}

func testDecodeRaw_Config(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_Config(vStringSlice, result))
}

func TestConfig_GetPFlagSet(t *testing.T) {
	val := Config{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestConfig_SetFlags(t *testing.T) {
	actual := Config{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_scheduler", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("scheduler", testValue)
			if vString, err := cmdFlags.GetString("scheduler"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.DefaultScheduler)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_maxErrorLength", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("maxErrorLength", testValue)
			if vInt, err := cmdFlags.GetInt("maxErrorLength"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.MaxErrorStringLength)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_maxArrayJobSize", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("maxArrayJobSize", testValue)
			if vInt64, err := cmdFlags.GetInt64("maxArrayJobSize"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt64), &actual.MaxArrayJobSize)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_OutputAssembler.workers", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("OutputAssembler.workers", testValue)
			if vInt, err := cmdFlags.GetInt("OutputAssembler.workers"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.OutputAssembler.Workers)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_OutputAssembler.maxRetries", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("OutputAssembler.maxRetries", testValue)
			if vInt, err := cmdFlags.GetInt("OutputAssembler.maxRetries"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.OutputAssembler.MaxRetries)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_OutputAssembler.maxItems", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("OutputAssembler.maxItems", testValue)
			if vInt, err := cmdFlags.GetInt("OutputAssembler.maxItems"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.OutputAssembler.IndexCacheMaxItems)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_ErrorAssembler.workers", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("ErrorAssembler.workers", testValue)
			if vInt, err := cmdFlags.GetInt("ErrorAssembler.workers"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.ErrorAssembler.Workers)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_ErrorAssembler.maxRetries", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("ErrorAssembler.maxRetries", testValue)
			if vInt, err := cmdFlags.GetInt("ErrorAssembler.maxRetries"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.ErrorAssembler.MaxRetries)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_ErrorAssembler.maxItems", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("ErrorAssembler.maxItems", testValue)
			if vInt, err := cmdFlags.GetInt("ErrorAssembler.maxItems"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt), &actual.ErrorAssembler.IndexCacheMaxItems)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.cloudwatch-enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.cloudwatch-enabled", testValue)
			if vBool, err := cmdFlags.GetBool("logs.config.cloudwatch-enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.LogConfig.Config.IsCloudwatchEnabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.cloudwatch-region", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.cloudwatch-region", testValue)
			if vString, err := cmdFlags.GetString("logs.config.cloudwatch-region"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.CloudwatchRegion)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.cloudwatch-log-group", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.cloudwatch-log-group", testValue)
			if vString, err := cmdFlags.GetString("logs.config.cloudwatch-log-group"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.CloudwatchLogGroup)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.cloudwatch-template-uri", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.cloudwatch-template-uri", testValue)
			if vString, err := cmdFlags.GetString("logs.config.cloudwatch-template-uri"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.CloudwatchTemplateURI)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.kubernetes-enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.kubernetes-enabled", testValue)
			if vBool, err := cmdFlags.GetBool("logs.config.kubernetes-enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.LogConfig.Config.IsKubernetesEnabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.kubernetes-url", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.kubernetes-url", testValue)
			if vString, err := cmdFlags.GetString("logs.config.kubernetes-url"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.KubernetesURL)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.kubernetes-template-uri", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.kubernetes-template-uri", testValue)
			if vString, err := cmdFlags.GetString("logs.config.kubernetes-template-uri"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.KubernetesTemplateURI)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.stackdriver-enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.stackdriver-enabled", testValue)
			if vBool, err := cmdFlags.GetBool("logs.config.stackdriver-enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.LogConfig.Config.IsStackDriverEnabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.gcp-project", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.gcp-project", testValue)
			if vString, err := cmdFlags.GetString("logs.config.gcp-project"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.GCPProjectName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.stackdriver-logresourcename", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.stackdriver-logresourcename", testValue)
			if vString, err := cmdFlags.GetString("logs.config.stackdriver-logresourcename"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.StackdriverLogResourceName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_logs.config.stackdriver-template-uri", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("logs.config.stackdriver-template-uri", testValue)
			if vString, err := cmdFlags.GetString("logs.config.stackdriver-template-uri"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.LogConfig.Config.StackDriverTemplateURI)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
