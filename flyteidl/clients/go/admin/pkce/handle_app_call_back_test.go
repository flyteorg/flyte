package pkce

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	testhttp "github.com/stretchr/testify/http"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
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
	callBackFn = getAuthServerCallbackHandler(testAuthConfig, "", tokenChannel, errorChannel, state, &http.Client{})
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
		assert.True(t, strings.Contains(rw.Output, "Sorry, we can't serve your request"))
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

// Test_RenderCallbackPage simply generates all possible callback HTML pages
// that can be manually opened and viewed in browser
func Test_RenderCallbackPage(t *testing.T) {
	tCases := []struct {
		name string
		data callbackData
	}{
		{
			name: "error",
			data: callbackData{Error: "fail", ErrorDescription: "desc", ErrorHint: "hint"},
		},
		{
			name: "no_code",
			data: callbackData{NoCode: true},
		},
		{
			name: "wrong_state",
			data: callbackData{WrongState: true},
		},
		{
			name: "access_token_error",
			data: callbackData{AccessTokenError: "fail"},
		},
		{
			name: "success",
			data: callbackData{},
		},
	}
	for _, tCase := range tCases {
		t.Run(tCase.name, func(t *testing.T) {
			f, err := os.Create(fmt.Sprintf("%s.html", tCase.name))
			require.NoError(t, err)
			defer f.Close()

			err = callbackTemplate.Execute(f, tCase.data)

			assert.NoError(t, err)
		})
	}
}
