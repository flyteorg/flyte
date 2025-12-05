package tasklog

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	defaultName       = "Azure Logs"
	defaulQueryFormat = `let StartTime = datetime_add('hour', -1, datetime("{{.podRFC3339StartTime}}"));
let FinishTime = datetime_add('hour', 1, datetime("{{.podRFC3339FinishTime}}"));
ContainerLogV2
| where TimeGenerated between (StartTime .. FinishTime)
 and ContainerName == "{{.containerName}}"
 and PodName == "{{.podName}}"
 and PodNamespace == "{{.namespace}}"`
)

// Azure Logs specific templater.
// Azure encodes two parts of the URI in two distinct ways.
// The first half the URI is usually composed of Azure tenant ID, subscription ID, resource group name, etc.
// The second half is the query itself, which is gzipped, base64, and then URL encoded.
type AzureLogsTemplatePlugin struct {
	TemplateLogPlugin `json:",squash"` //nolint

	QueryFormat *string `json:"queryFormat" pflag:",The plain text query to use for Azure Logs."`
}

func (t AzureLogsTemplatePlugin) GetTaskLogs(input Input) (Output, error) {
	taskLogs := make([]*core.TaskLog, 0)

	var rawQuery string
	if t.QueryFormat == nil {
		rawQuery = defaulQueryFormat
	} else {
		rawQuery = *t.QueryFormat
	}

	query := replaceAll(rawQuery, input.templateVars())
	var compressedBuffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuffer)
	_, err := gzipWriter.Write([]byte(query))
	if err != nil {
		return Output{TaskLogs: taskLogs}, fmt.Errorf("gzip compression failed: %v", err)
	}
	err = gzipWriter.Close()
	if err != nil {
		return Output{TaskLogs: taskLogs}, fmt.Errorf("gzip writer close failed: %v", err)
	}

	// Base64 and URL encoding
	base64Encoded := base64.StdEncoding.EncodeToString(compressedBuffer.Bytes())
	urlEncoded := url.QueryEscape(base64Encoded)

	for _, baseURL := range t.TemplateURIs {
		completeURL := fmt.Sprintf("%s%s", baseURL, urlEncoded)
		taskLogs = append(taskLogs, &core.TaskLog{
			Name:          t.DisplayName + input.LogName,
			Uri:           completeURL,
			MessageFormat: core.TaskLog_JSON,
		})
	}

	return Output{TaskLogs: taskLogs}, nil
}
