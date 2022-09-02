package pkce

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"

	"github.com/stretchr/testify/assert"
	testhttp "github.com/stretchr/testify/http"
	"golang.org/x/oauth2"
)

var (
	rw         *testhttp.TestResponseWriter
	req        *http.Request
	callBackFn func(rw http.ResponseWriter, req *http.Request)
)

func HandleAppCallBackSetup(t *testing.T, state string) (tokenChannel chan *oauth2.Token, errorChannel chan error) {
	var testAuthConfig *oauth.Config
	errorChannel = make(chan error, 1)
	tokenChannel = make(chan *oauth2.Token)
	testAuthConfig = &oauth.Config{Config: &oauth2.Config{}, DeviceEndpoint: "dummyDeviceEndpoint"}
	callBackFn = getAuthServerCallbackHandler(testAuthConfig, "", tokenChannel, errorChannel, state)
	assert.NotNil(t, callBackFn)
	req = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "http",
			Host:     "dummyHost",
			Path:     "dummyPath",
			RawQuery: "&error=invalid_request",
		},
	}
	rw = &testhttp.TestResponseWriter{}
	return
}

func TestHandleAppCallBackWithErrorInRequest(t *testing.T) {
	tokenChannel, errorChannel := HandleAppCallBackSetup(t, "")
	req = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "http",
			Host:     "dummyHost",
			Path:     "dummyPath",
			RawQuery: "&error=invalid_request",
		},
	}
	callBackFn(rw, req)
	var errorValue error
	select {
	case errorValue = <-errorChannel:
		assert.NotNil(t, errorValue)
		assert.True(t, strings.Contains(rw.Output, "invalid_request"))
		assert.Equal(t, errors.New("error on callback during authorization due to invalid_request"), errorValue)
	case <-tokenChannel:
		assert.Fail(t, "received a token for a failed test")
	}
}

func TestHandleAppCallBackWithCodeNotFound(t *testing.T) {
	tokenChannel, errorChannel := HandleAppCallBackSetup(t, "")
	req = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "http",
			Host:     "dummyHost",
			Path:     "dummyPath",
			RawQuery: "",
		},
	}
	callBackFn(rw, req)
	var errorValue error
	select {
	case errorValue = <-errorChannel:
		assert.NotNil(t, errorValue)
		assert.True(t, strings.Contains(rw.Output, "Could not find the authorize code"))
		assert.Equal(t, errors.New("could not find the authorize code"), errorValue)
	case <-tokenChannel:
		assert.Fail(t, "received a token for a failed test")
	}
}

func TestHandleAppCallBackCsrfAttach(t *testing.T) {
	tokenChannel, errorChannel := HandleAppCallBackSetup(t, "the real state")
	req = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "http",
			Host:     "dummyHost",
			Path:     "dummyPath",
			RawQuery: "&code=dummyCode&state=imposterString",
		},
	}
	callBackFn(rw, req)
	var errorValue error
	select {
	case errorValue = <-errorChannel:
		assert.NotNil(t, errorValue)
		assert.True(t, strings.Contains(rw.Output, "Sorry we can't serve your request"))
		assert.Equal(t, errors.New("possibly a csrf attack"), errorValue)
	case <-tokenChannel:
		assert.Fail(t, "received a token for a failed test")
	}
}

func TestHandleAppCallBackFailedTokenExchange(t *testing.T) {
	tokenChannel, errorChannel := HandleAppCallBackSetup(t, "realStateString")
	req = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme:   "http",
			Host:     "dummyHost",
			Path:     "dummyPath",
			RawQuery: "&code=dummyCode&state=realStateString",
		},
	}
	rw = &testhttp.TestResponseWriter{}
	callBackFn(rw, req)
	var errorValue error
	select {
	case errorValue = <-errorChannel:
		assert.NotNil(t, errorValue)
		assert.True(t, strings.Contains(errorValue.Error(), "error while exchanging auth code due"))
	case <-tokenChannel:
		assert.Fail(t, "received a token for a failed test")
	}
}
