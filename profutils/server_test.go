package profutils

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/version"

	"github.com/flyteorg/flytestdlib/internal/utils"

	"github.com/stretchr/testify/assert"
)

type MockResponseWriter struct {
	Status  int
	Headers http.Header
	Body    []byte
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	m.Body = b
	return 1, nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.Status = statusCode
}

func (m *MockResponseWriter) Header() http.Header {
	return m.Headers
}

type TestObj struct {
	X int `json:"x"`
}

func init() {
	if err := configureGlobalHTTPHandler(nil); err != nil {
		panic(err)
	}
}

func TestWriteJsonResponse(t *testing.T) {
	m := &MockResponseWriter{Headers: http.Header{}}
	assert.NoError(t, WriteJSONResponse(m, http.StatusOK, TestObj{10}))
	assert.Equal(t, http.StatusOK, m.Status)
	assert.Equal(t, http.Header{contentTypeHeader: []string{contentTypeJSON}}, m.Headers)
	assert.Equal(t, `{"x":10}`, string(m.Body))
}

func TestWriteStringResponse(t *testing.T) {
	m := &MockResponseWriter{Headers: http.Header{}}
	assert.NoError(t, WriteStringResponse(m, http.StatusOK, "OK"))
	assert.Equal(t, http.StatusOK, m.Status)
	assert.Equal(t, "OK", string(m.Body))
}

func TestConfigHandler(t *testing.T) {
	writer := &MockResponseWriter{Headers: http.Header{}}
	testURL := utils.MustParseURL(configPath)
	request := &http.Request{
		URL: &testURL,
	}

	http.DefaultServeMux.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusOK, writer.Status)

	m := map[string]interface{}{}
	assert.NoError(t, json.Unmarshal(writer.Body, &m))
	assert.Equal(t, map[string]interface{}{
		"logger": map[string]interface{}{
			"show-source": false,
			"mute":        false,
			"level":       float64(3),
			"formatter": map[string]interface{}{
				"type": "json",
			},
		},
	}, m)
}

func TestVersionHandler(t *testing.T) {
	writer := &MockResponseWriter{Headers: http.Header{}}
	testURL := utils.MustParseURL(versionPath)
	request := &http.Request{
		URL: &testURL,
	}

	version.BuildTime = time.Now().String()

	http.DefaultServeMux.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusOK, writer.Status)
	assert.NotNil(t, writer.Body)
	bv := BuildVersion{}
	assert.NoError(t, json.Unmarshal(writer.Body, &bv))
	assert.Equal(t, bv.Timestamp, version.BuildTime)
	assert.Equal(t, bv.Build, version.Build)
	assert.Equal(t, bv.Version, version.Version)
}

func TestHealthcheckHandler(t *testing.T) {
	writer := &MockResponseWriter{Headers: http.Header{}}
	testURL := utils.MustParseURL(healthcheck)
	request := &http.Request{
		URL: &testURL,
	}

	http.DefaultServeMux.ServeHTTP(writer, request)
	assert.Equal(t, http.StatusOK, writer.Status)
	assert.Equal(t, `OK`, string(writer.Body))
}
