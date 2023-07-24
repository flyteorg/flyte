package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	errors2 "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"
)

const (
	logLinkFormat          = "?command_id=%s"
	acceptHeaderKey        = "Accept"
	hiveCommandType        = "HiveCommand"
	killStatus             = "kill"
	httpRequestTimeoutSecs = 30
	hostHeaderKey          = "Host"
	HeaderContentType      = "Content-Type"
	ContentTypeJSON        = "application/json"
	ErrRequestFailed       = "REQUEST_FAILED"
)

// QuboleClient CommandStatus Response Format, only used to unmarshal the response
type quboleCmdDetailsInternal struct {
	ID     int64
	Status string
}

type QuboleCommandDetails struct {
	ID     int64
	Status QuboleStatus
	URI    url.URL
}

type CommandMetadata struct {
	TaskName            string
	Domain              string
	Project             string
	Labels              map[string]string
	AttemptNumber       uint32
	MaxAttempts         uint32
	WorkflowID          string
	WorkflowExecutionID string
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

//go:generate mockery -all -case=snake

// Interface to interact with QuboleClient for hive tasks
type QuboleClient interface {
	ExecuteHiveCommand(ctx context.Context, commandStr string, timeoutVal uint32, clusterPrimaryLabel string,
		accountKey string, tags []string, commandMetadata CommandMetadata) (*QuboleCommandDetails, error)
	KillCommand(ctx context.Context, commandID string, accountKey string) error
	GetCommandStatus(ctx context.Context, commandID string, accountKey string) (QuboleStatus, error)
}

// TODO: The Qubole client needs a rate limiter
type quboleClient struct {
	client     *http.Client
	commandURL *url.URL
	analyzeURL *url.URL
}

func (q *quboleClient) getHeaders(url *url.URL, accountKey string) http.Header {
	headers := make(http.Header)
	headers.Set("X-AUTH-TOKEN", accountKey)
	headers.Set(HeaderContentType, ContentTypeJSON)
	headers.Set(acceptHeaderKey, ContentTypeJSON)
	headers.Set(hostHeaderKey, url.Host)

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
func (q *quboleClient) executeRequest(ctx context.Context, method string, u *url.URL,
	body *RequestBody, accountKey string) (*http.Response, error) {
	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest("GET", u.String(), nil)
	case http.MethodPost:
		req, err = http.NewRequest("POST", u.String(), nil)
	case http.MethodPut:
		req, err = http.NewRequest("PUT", u.String(), nil)
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

	logger.Debugf(ctx, "Qubole endpoint: %v", u.String())
	req.Header = q.getHeaders(u, accountKey)
	return q.client.Do(req)
}

/*
Execute Hive Command on the QuboleClient Hive Cluster and return the CommandID
param: context.Context ctx: The default go context.
param: string commandStr: the query to execute
param: uint32 timeoutVal: timeout for the query to execute in seconds
param: string ClusterLabel: label for cluster on which to execute the Hive Command.
param: CommandMetadata _: additional labels for the command
return: *int64: CommandID for the command executed
return: error: error in-case of a failure
*/
func (q *quboleClient) ExecuteHiveCommand(
	ctx context.Context,
	commandStr string,
	timeoutVal uint32,
	clusterPrimaryLabel string,
	accountKey string,
	tags []string,
	_ CommandMetadata) (*QuboleCommandDetails, error) {

	requestBody := RequestBody{
		CommandType:  hiveCommandType,
		Query:        commandStr,
		Timeout:      timeoutVal,
		ClusterLabel: clusterPrimaryLabel,
		Tags:         tags,
	}

	response, err := q.executeRequest(ctx, http.MethodPost, q.commandURL, &requestBody, accountKey)
	if err != nil {
		return nil, err
	}
	defer closeBody(ctx, response)

	if response.StatusCode != 200 {
		bts, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, errors2.Wrapf(ErrRequestFailed, err, "request failed. Response code [%v]", response.StatusCode)
		}

		return nil, fmt.Errorf("bad response from Qubole creating query: %d %s",
			response.StatusCode, string(bts))
	}

	var cmd quboleCmdDetailsInternal
	if err = unmarshalBody(response, &cmd); err != nil {
		return nil, err
	}

	u, err := q.GetLogLinkPath(ctx, strconv.FormatInt(cmd.ID, 10))
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, fmt.Errorf("failed to build log link for command [%v]", cmd.ID)
	}

	status := NewQuboleStatus(ctx, cmd.Status)
	return &QuboleCommandDetails{
		ID:     cmd.ID,
		Status: status,
		URI:    *u,
	}, nil
}

/*
Terminate a QuboleClient command
param: context.Context ctx: The default go context.
param: string CommandID: the CommandID to terminate.
return: error: error in-case of a failure
*/
func (q *quboleClient) KillCommand(ctx context.Context, commandID string, accountKey string) error {
	commandStatus, err := url.Parse(commandID)
	if err != nil {
		return err
	}
	killPath := q.commandURL.ResolveReference(commandStatus)
	requestBody := RequestBody{Status: killStatus}
	response, err := q.executeRequest(ctx, http.MethodPut, killPath, &requestBody, accountKey)
	defer closeBody(ctx, response)
	return err
}

/*
Get the status of a QuboleClient command
param: context.Context ctx: The default go context.
param: string CommandID: the CommandID to fetch the status for
return: *string: commandStatus for the CommandID passed
return: error: error in-case of a failure
*/
func (q *quboleClient) GetCommandStatus(ctx context.Context, commandID string, accountKey string) (QuboleStatus, error) {
	commandStatus, err := url.Parse(commandID)
	if err != nil {
		return QuboleStatusUnknown, err
	}
	statusPath := q.commandURL.ResolveReference(commandStatus)
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

		return QuboleStatusUnknown, fmt.Errorf("bad response from Qubole getting command status: %d %s, path: %s, %s",
			response.StatusCode, string(bts), statusPath, q.commandURL.String())
	}

	var cmd quboleCmdDetailsInternal
	if err = unmarshalBody(response, &cmd); err != nil {
		return QuboleStatusUnknown, err
	}

	cmdStatus := NewQuboleStatus(ctx, cmd.Status)
	return cmdStatus, nil
}

func (q *quboleClient) GetLogLinkPath(ctx context.Context, commandID string) (*url.URL, error) {
	logLink := fmt.Sprintf(logLinkFormat, commandID)
	l, err := url.Parse(logLink)
	if err != nil {
		return nil, err
	}
	return q.analyzeURL.ResolveReference(l), nil
}

func NewQuboleClient(cfg *config.Config) QuboleClient {
	return &quboleClient{
		client:     &http.Client{Timeout: httpRequestTimeoutSecs * time.Second},
		commandURL: cfg.Endpoint.ResolveReference(&cfg.CommandAPIPath.URL),
		analyzeURL: cfg.Endpoint.ResolveReference(&cfg.AnalyzeLinkPath.URL),
	}
}
