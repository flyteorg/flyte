package client

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"
)

var getCommandResponse = `{
	  "command": {
		"approx_mode": false,
		"approx_aggregations": false,
		"query": "select count(*) as num_rows from miniwikistats;",
		"sample": false
	  },
	  "qbol_session_id": 0,
	  "created_at": "2012-10-11T16:54:57Z",
	  "user_id": 0,
	  "status": "{STATUS}",
	  "command_type": "HiveCommand",
	  "id": 3852,
	  "progress": 100,
	  "throttled": true
	}`

var createCommandResponse = `{
	   "command": {
		 "approx_mode": false,
		 "approx_aggregations": false,
		 "query": "show tables",
		 "sample": false
	   },
	   "qbol_session_id": 0,
	   "created_at": "2012-10-11T16:01:09Z",
	   "user_id": 0,
	   "status": "waiting",
	   "command_type": "HiveCommand",
	   "id": 3850,
	   "progress": 0
	}`

func TestQuboleClient_GetCommandStatus(t *testing.T) {
	tests := []struct {
		Name                 string
		quboleInternalStatus string
		status               QuboleStatus
	}{
		{
			Name:                 "done status",
			quboleInternalStatus: "done",
			status:               QuboleStatusDone,
		},
		{
			Name:                 "unknown status",
			quboleInternalStatus: "bogus",
			status:               QuboleStatusUnknown,
		},
		{
			Name:                 "running status",
			quboleInternalStatus: "running",
			status:               QuboleStatusRunning,
		},
	}

	for _, test := range tests {
		tc := test
		t.Run(tc.Name, func(t *testing.T) {
			client := createQuboleClient(strings.Replace(getCommandResponse, "{STATUS}", tc.quboleInternalStatus, 1))
			status, err := client.GetCommandStatus(context.Background(), "", "")
			assert.NoError(t, err)
			assert.Equal(t, tc.status, status)
		})
	}
}

func TestQuboleClient_ExecuteHiveCommand(t *testing.T) {
	client := createQuboleClient(createCommandResponse)
	details, err := client.ExecuteHiveCommand(context.Background(),
		"", 0, "clusterLabel", "", nil, CommandMetadata{})
	assert.NoError(t, err)
	assert.Equal(t, int64(3850), details.ID)
	assert.Equal(t, QuboleStatusWaiting, details.Status)
}

func TestQuboleClient_KillCommand(t *testing.T) {
	client := createQuboleClient("OK")
	err := client.KillCommand(context.Background(), "", "")
	assert.NoError(t, err)
}

func TestQuboleClient_ExecuteHiveCommandError(t *testing.T) {
	client := createQuboleErrorClient("bad token")
	details, err := client.ExecuteHiveCommand(context.Background(),
		"", 0, "clusterLabel", "", nil, CommandMetadata{})
	assert.Error(t, err)
	assert.Nil(t, details)
}

func TestQuboleClient_GetCommandStatusError(t *testing.T) {
	client := createQuboleErrorClient("bad token")
	details, err := client.GetCommandStatus(context.Background(), "1234", "fake account key")
	assert.Error(t, err)
	assert.Equal(t, QuboleStatusUnknown, details)
}

func createQuboleClient(response string) quboleClient {
	hc := &http.Client{Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		response := &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(response)),
			Header:     make(http.Header),
		}
		response.Header.Set("Content-Type", "application/json")
		return response, nil

	})}

	cfg := config.GetQuboleConfig()
	cmd := cfg.Endpoint.ResolveReference(&cfg.CommandAPIPath.URL)
	analyze := cfg.Endpoint.ResolveReference(&cfg.AnalyzeLinkPath.URL)
	return quboleClient{client: hc, commandURL: cmd, analyzeURL: analyze}
}

func createQuboleErrorClient(errorMsg string) quboleClient {
	hc := &http.Client{Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		response := &http.Response{
			StatusCode: 400,
			Body:       ioutil.NopCloser(bytes.NewBufferString(errorMsg)),
			Header:     make(http.Header),
		}
		response.Header.Set("Content-Type", "application/json")
		return response, nil

	})}

	cfg := config.GetQuboleConfig()
	cmd := cfg.Endpoint.ResolveReference(&cfg.CommandAPIPath.URL)
	analyze := cfg.Endpoint.ResolveReference(&cfg.AnalyzeLinkPath.URL)
	return quboleClient{client: hc, commandURL: cmd, analyzeURL: analyze}
}

type RoundTripFunc func(*http.Request) (*http.Response, error)

func (rt RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return rt(req) }
