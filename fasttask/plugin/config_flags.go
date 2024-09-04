// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package plugin

import (
	"encoding/json"
	"reflect"

	"fmt"

	"github.com/spf13/pflag"
)

// If v is a pointer, it will get its element value or the zero value of the element type.
// If v is not a pointer, it will return it as is.
func (Config) elemValueOrNil(v interface{}) interface{} {
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

func (Config) mustJsonMarshal(v interface{}) string {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

func (Config) mustMarshalJSON(v json.Marshaler) string {
	raw, err := v.MarshalJSON()
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// GetPFlagSet will return strongly types pflags for all fields in Config and its nested types. The format of the
// flags is json-name.json-sub-name... etc.
func (cfg Config) GetPFlagSet(prefix string) *pflag.FlagSet {
	cmdFlags := pflag.NewFlagSet("Config", pflag.ExitOnError)
	cmdFlags.StringSlice(fmt.Sprintf("%v%v", prefix, "additional-worker-args"), defaultConfig.AdditionalWorkerArgs, "Additional arguments to pass to the fasttask worker binary.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "callback-uri"), defaultConfig.CallbackURI, "Fasttask gRPC service URI that fasttask workers will connect to.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "default-ttl"), defaultConfig.DefaultTTL.String(), "Default TTL for environments.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "endpoint"), defaultConfig.Endpoint, "Fasttask gRPC service endpoint.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "env-detect-orphan-interval"), defaultConfig.EnvDetectOrphanInterval.String(), "Frequency that orphaned environments detection is performed.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "env-gc-interval"), defaultConfig.EnvGCInterval.String(), "Frequency that environments are GCed in case of TTL expirations.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "env-repair-interval"), defaultConfig.EnvRepairInterval.String(), "Frequency that environments are repaired in case of external modifications (ex. pod deletion).")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "grace-period-status-not-found"), defaultConfig.GracePeriodStatusNotFound.String(), "The grace period for a task status to be reported before the task is considered failed.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "heartbeat-buffer-size"), defaultConfig.HeartbeatBufferSize, "The size of the heartbeat buffer for each worker.")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "logs.cloudwatch-enabled"), defaultConfig.Logs.IsCloudwatchEnabled, "Enable Cloudwatch Logging")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.cloudwatch-region"), defaultConfig.Logs.CloudwatchRegion, "AWS region in which Cloudwatch logs are stored.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.cloudwatch-log-group"), defaultConfig.Logs.CloudwatchLogGroup, "Log group to which streams are associated.")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.cloudwatch-template-uri"), defaultConfig.Logs.CloudwatchTemplateURI, "Template Uri to use when building cloudwatch log links")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "logs.kubernetes-enabled"), defaultConfig.Logs.IsKubernetesEnabled, "Enable Kubernetes Logging")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.kubernetes-url"), defaultConfig.Logs.KubernetesURL, "Console URL for Kubernetes logs")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.kubernetes-template-uri"), defaultConfig.Logs.KubernetesTemplateURI, "Template Uri to use when building kubernetes log links")
	cmdFlags.Bool(fmt.Sprintf("%v%v", prefix, "logs.stackdriver-enabled"), defaultConfig.Logs.IsStackDriverEnabled, "Enable Log-links to stackdriver")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.gcp-project"), defaultConfig.Logs.GCPProjectName, "Name of the project in GCP")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.stackdriver-logresourcename"), defaultConfig.Logs.StackdriverLogResourceName, "Name of the logresource in stackdriver")
	cmdFlags.String(fmt.Sprintf("%v%v", prefix, "logs.stackdriver-template-uri"), defaultConfig.Logs.StackDriverTemplateURI, "Template Uri to use when building stackdriver log links")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "nonce-length"), defaultConfig.NonceLength, "The length of the nonce value to uniquely link a fasttask replica to the environment instance,  ensuring fast turnover of environments regardless of cache freshness.")
	cmdFlags.Int(fmt.Sprintf("%v%v", prefix, "task-status-buffer-size"), defaultConfig.TaskStatusBufferSize, "The size of the task status buffer for each task.")
	return cmdFlags
}
