// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package launchplan

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (AdminConfig) elemValueOrNil(v interface{}) interface{} {
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

func (AdminConfig) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (AdminConfig) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in AdminConfig and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg AdminConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("AdminConfig", pflag.ExitOnError)
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "tps"), defaultAdminConfig.TPS, "The maximum number of transactions per second to flyte admin from this client.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "burst"), defaultAdminConfig.Burst, "Maximum burst for throttle")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "cacheSize"), defaultAdminConfig.MaxCacheSize, "Maximum cache in terms of number of items stored.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "workers"), defaultAdminConfig.Workers, "Number of parallel workers to work on the queue.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "cache-resync-duration"), defaultAdminConfig.CacheResyncDuration.String(), "Frequency of re-syncing launchplans within the auto refresh cache.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "watchConfig.enabled"), defaultAdminConfig.WatchConfig.Enabled, "True when propeller calls Watch API to populate auto-refresh cache with execution status updates")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "watchConfig.freshnessDuration"), defaultAdminConfig.WatchConfig.FreshnessDuration.String(), "How long a cache item should be used as-is without syncing")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "watchConfig.reconnectDelay"), defaultAdminConfig.WatchConfig.ReconnectDelay.String(), "How long to wait before attempting to connection to Watch API")
	return cmdFlags
}
