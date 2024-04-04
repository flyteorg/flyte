// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package configs

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (CacheServiceConfig) elemValueOrNil(v interface{}) interface{} {
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

func (CacheServiceConfig) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (CacheServiceConfig) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in CacheServiceConfig and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg CacheServiceConfig) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("CacheServiceConfig", pflag.ExitOnError)
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "storage-prefix"), defaultConfig.StoragePrefix, "StoragePrefix specifies the prefix where CacheService stores offloaded output in CloudStorage. If not ...")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "metrics-scope"), defaultConfig.MetricsScope, "Scope that the metrics will record under.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "profiler-port"), defaultConfig.ProfilerPort, "Port that the profiling service is listening on.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "heartbeat-grace-period-multiplier"), defaultConfig.HeartbeatGracePeriodMultiplier, "Number of heartbeats before a reservation expires without an extension.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "max-reservation-heartbeat"), defaultConfig.MaxReservationHeartbeat.String(), "The maximum available reservation extension heartbeat interval.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "data-store-type"), defaultConfig.OutputDataStoreType, "Cache storage implementation to use")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "reservation-data-store-type"), defaultConfig.ReservationDataStoreType, "Reservation storage implementation to use")
	cmdFlags.Int64(fmt.Sprintf("%v%v", prefix, "maxInlineSizeBytes"), defaultConfig.MaxInlineSizeBytes, "The maximum size that an output literal will be stored in line. Default 0 means everything will be offloaded to blob storage.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "aws-region"), defaultConfig.AwsRegion, "Region to connect to.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "redis-address"), defaultConfig.RedisAddress, "Address of the Redis server.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "redis-username"), defaultConfig.RedisUsername, "Username for the Redis server.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "redis-password"), defaultConfig.RedisPassword, "Password for the Redis server.")
	return cmdFlags
}
