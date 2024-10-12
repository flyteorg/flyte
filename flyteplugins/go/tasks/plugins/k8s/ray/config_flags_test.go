// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package ray

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsConfig[t]
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

	t.Run("Test_shutdownAfterJobFinishes", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("shutdownAfterJobFinishes", testValue)
			if vBool, err := cmdFlags.GetBool("shutdownAfterJobFinishes"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.ShutdownAfterJobFinishes)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_ttlSecondsAfterFinished", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("ttlSecondsAfterFinished", testValue)
			if vInt32, err := cmdFlags.GetInt32("ttlSecondsAfterFinished"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vInt32), &actual.TTLSecondsAfterFinished)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_serviceType", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("serviceType", testValue)
			if vString, err := cmdFlags.GetString("serviceType"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.ServiceType)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_includeDashboard", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("includeDashboard", testValue)
			if vBool, err := cmdFlags.GetBool("includeDashboard"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.IncludeDashboard)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dashboardHost", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("dashboardHost", testValue)
			if vString, err := cmdFlags.GetString("dashboardHost"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.DashboardHost)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_batchScheduler.scheduler", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("batchScheduler.scheduler", testValue)
			if vString, err := cmdFlags.GetString("batchScheduler.scheduler"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.BatchScheduler.Scheduler)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_batchScheduler.default.Kueue.Priority", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("batchScheduler.default.Kueue.Priority", testValue)
			if vString, err := cmdFlags.GetString("batchScheduler.default.Kueue.Priority"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.BatchScheduler.Default.KueueConfig.PriorityClassName)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_batchScheduler.default.Kueue.Queue", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("batchScheduler.default.Kueue.Queue", testValue)
			if vString, err := cmdFlags.GetString("batchScheduler.default.Kueue.Queue"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.BatchScheduler.Default.KueueConfig.Queue)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_batchScheduler.default.Yunikorn.parameters", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("batchScheduler.default.Yunikorn.parameters", testValue)
			if vString, err := cmdFlags.GetString("batchScheduler.default.Yunikorn.parameters"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.BatchScheduler.Default.YunikornConfig.Parameters)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_batchScheduler.default.Yunikorn.queue", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("batchScheduler.default.Yunikorn.queue", testValue)
			if vString, err := cmdFlags.GetString("batchScheduler.default.Yunikorn.queue"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.BatchScheduler.Default.YunikornConfig.Queue)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_remoteClusterConfig.name", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("remoteClusterConfig.name", testValue)
			if vString, err := cmdFlags.GetString("remoteClusterConfig.name"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.RemoteClusterConfig.Name)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_remoteClusterConfig.endpoint", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("remoteClusterConfig.endpoint", testValue)
			if vString, err := cmdFlags.GetString("remoteClusterConfig.endpoint"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.RemoteClusterConfig.Endpoint)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_remoteClusterConfig.enabled", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("remoteClusterConfig.enabled", testValue)
			if vBool, err := cmdFlags.GetBool("remoteClusterConfig.enabled"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.RemoteClusterConfig.Enabled)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_enableUsageStats", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("enableUsageStats", testValue)
			if vBool, err := cmdFlags.GetBool("enableUsageStats"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vBool), &actual.EnableUsageStats)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_serviceAccount", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("serviceAccount", testValue)
			if vString, err := cmdFlags.GetString("serviceAccount"); err == nil {
				testDecodeJson_Config(t, fmt.Sprintf("%v", vString), &actual.ServiceAccount)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
