package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/lyft/flytestdlib/logger"
)

const url = "https://api.qubole.com/api"
const apiPath = "/v1.2/commands"
const QuboleLogLinkFormat = "https://api.qubole.com/v2/analyze?command_id=%s"

const tokenKeyForAth = "X-AUTH-TOKEN"
const acceptHeaderKey = "Accept"
const hiveCommandType = "HiveCommand"
const killStatus = "kill"
const httpRequestTimeoutSecs = 30
const host = "api.qubole.com"
const hostHeaderKey = "Host"
const HeaderContentType = "Content-Type"
const ContentTypeJSON = "application/json"

// QuboleClient CommandStatus Response Format, only used to unmarshall the response
type quboleCmdDetailsInternal struct {
	ID     int64
	Status string
}

type QuboleCommandDetails struct {
	ID     int64
	Status QuboleStatus
}

// QuboleClient API Request Body, meant to be passed into JSON.marshal
// Any nil, 0 or "" fields will not be marshaled
type RequestBody struct {
	Query        string   `json:"query,omitempty"`
	ClusterLabel string   `json:"label,omitempty"`
	CommandType  string   `json:"command_type,omitempty"`
	Retry        uint32   `json:"retry,omitempty"`
	Status       string   `json:"status,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Timeout      uint32   `json:"timeout,omitempty"`
	InlineScript string   `json:"inline,omitempty"`
	Files        string   `json:"files,omitempty"`
}

//go:generate mockery -name QuboleClient

// Interface to interact with QuboleClient for hive tasks
type QuboleClient interface {
	ExecuteHiveCommand(ctx context.Context, commandStr string, timeoutVal uint32, clusterLabel string,
		accountKey string, tags []string) (*QuboleCommandDetails, error)
	KillCommand(ctx context.Context, commandID string, accountKey string) error
	GetCommandStatus(ctx context.Context, commandID string, accountKey string) (QuboleStatus, error)
}

// TODO: The Qubole client needs a rate limiter
type quboleClient struct {
	client *http.Client
}

func (q *quboleClient) getHeaders(accountKey string) http.Header {
	headers := make(http.Header)
	headers.Set(tokenKeyForAth, accountKey)
	headers.Set(HeaderContentType, ContentTypeJSON)
	headers.Set(acceptHeaderKey, ContentTypeJSON)
	headers.Set(hostHeaderKey, host)

	return headers
}

// no-op closer for in-memory buffers used as io.Reader
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func addJSONBody(req *http.Request, body interface{}) error {
	// marshals body into JSON and set the request body
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req.Header.Add(HeaderContentType, ContentTypeJSON)
	req.Body = &nopCloser{bytes.NewReader(js)}
	return nil
}

func unmarshalBody(res *http.Response, t interface{}) error {
	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(bts, t)
}

func closeBody(ctx context.Context, response *http.Response) {
	_, err := io.Copy(ioutil.Discard, response.Body)
	if err != nil {
		logger.Errorf(ctx, "unexpected failure writing to devNull: %v", err)
	}
	err = response.Body.Close()
	if err != nil {
		logger.Warnf(ctx, "failure closing response body: %v", err)
	}
}

// Helper method to execute the requests
func (q *quboleClient) executeRequest(ctx context.Context, method string, path string, body *RequestBody, accountKey string) (*http.Response, error) {
	var req *http.Request
	var err error
	path = url + "/" + path

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest("GET", path, nil)
	case http.MethodPost:
		req, err = http.NewRequest("POST", path, nil)
		req.Header = q.getHeaders(accountKey)
	case http.MethodPut:
		req, err = http.NewRequest("PUT", path, nil)
	}

	if err != nil {
		return nil, err
	}

	if body != nil {
		err := addJSONBody(req, body)
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf(ctx, "qubole endpoint: %v", path)
	req.Header = q.getHeaders(accountKey)
	return q.client.Do(req)
}

/*
	Execute Hive Command on the QuboleClient Hive Cluster and return the CommandId
	param: context.Context ctx: The default go context.
	param: string commandStr: the query to execute
	param: uint32 timeoutVal: timeout for the query to execute in seconds
	param: string ClusterLabel: label for cluster on which to execute the Hive Command.
	return: *int64: CommandId for the command executed
	return: error: error in-case of a failure
*/
func (q *quboleClient) ExecuteHiveCommand(
	ctx context.Context,
	commandStr string,
	timeoutVal uint32,
	clusterLabel string,
	accountKey string,
	tags []string) (*QuboleCommandDetails, error) {

	requestBody := RequestBody{
		CommandType:  hiveCommandType,
		Query:        commandStr,
		Timeout:      timeoutVal,
		ClusterLabel: clusterLabel,
		Tags:         tags,
	}
	response, err := q.executeRequest(ctx, http.MethodPost, apiPath, &requestBody, accountKey)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, response)

	if response.StatusCode != 200 {
		bts, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(fmt.Sprintf("Bad response from Qubole creating query: %d %s",
			response.StatusCode, string(bts)))
	}

	var cmd quboleCmdDetailsInternal
	if err = unmarshalBody(response, &cmd); err != nil {
		return nil, err
	}

	status := newQuboleStatus(ctx, cmd.Status)
	return &QuboleCommandDetails{ID: cmd.ID, Status: status}, nil
}

/*
	Terminate a QuboleClient command
	param: context.Context ctx: The default go context.
	param: string CommandId: the CommandId to terminate.
	return: error: error in-case of a failure
*/
func (q *quboleClient) KillCommand(ctx context.Context, commandID string, accountKey string) error {
	killPath := apiPath + "/" + commandID
	requestBody := RequestBody{Status: killStatus}

	response, err := q.executeRequest(ctx, http.MethodPut, killPath, &requestBody, accountKey)
	defer closeBody(ctx, response)
	return err
}

/*
	Get the status of a QuboleClient command
	param: context.Context ctx: The default go context.
	param: string CommandId: the CommandId to fetch the status for
	return: *string: commandStatus for the CommandId passed
	return: error: error in-case of a failure
*/
func (q *quboleClient) GetCommandStatus(ctx context.Context, commandID string, accountKey string) (QuboleStatus, error) {
	statusPath := apiPath + "/" + commandID
	response, err := q.executeRequest(ctx, http.MethodGet, statusPath, nil, accountKey)
	if err != nil {
		return QuboleStatusUnknown, err
	}
	defer closeBody(ctx, response)

	if response.StatusCode != 200 {
		bts, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return QuboleStatusUnknown, err
		}
		return QuboleStatusUnknown, errors.New(fmt.Sprintf("Bad response from Qubole getting command status: %d %s",
			response.StatusCode, string(bts)))
	}

	var cmd quboleCmdDetailsInternal
	if err = unmarshalBody(response, &cmd); err != nil {
		return QuboleStatusUnknown, err
	}

	cmdStatus := newQuboleStatus(ctx, cmd.Status)
	return cmdStatus, nil
}

func NewQuboleClient() QuboleClient {
	return &quboleClient{
		client: &http.Client{Timeout: httpRequestTimeoutSecs * time.Second},
	}
}
