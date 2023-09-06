package profutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	// #nosec G108
	_ "net/http/pprof" // Import for pprof server
)

const (
	healthcheck = "/healthcheck"
	metricsPath = "/metrics"
	versionPath = "/version"
	configPath  = "/config"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json; charset=utf-8"
)

type BuildVersion struct {
	Build     string `json:"build"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp,string"`
}

// Writes a string to the Http output stream
func WriteStringResponse(resp http.ResponseWriter, code int, body string) error {
	resp.WriteHeader(code)
	_, err := resp.Write([]byte(body))
	return err
}

// Writes a JSON to the http output stream
func WriteJSONResponse(resp http.ResponseWriter, code int, body interface{}) error {
	resp.Header().Set(contentTypeHeader, contentTypeJSON)
	resp.WriteHeader(code)
	j, err := json.Marshal(body)
	if err != nil {
		return WriteStringResponse(resp, http.StatusInternalServerError, err.Error())
	}
	return WriteStringResponse(resp, http.StatusOK, string(j))
}

// Simple healthcheck module that returns OK and provides a simple L7 healthcheck
// TODO we may want to provide a simple function that returns a bool, where users could provide custom healthchecks
func healtcheckHandler(w http.ResponseWriter, req *http.Request) {
	err := WriteStringResponse(w, http.StatusOK, http.StatusText(http.StatusOK))
	if err != nil {
		panic(err)
	}
}

// Handler that returns a JSON response indicating the Build Version information (refer to #version module)
func versionHandler(w http.ResponseWriter, req *http.Request) {
	err := WriteJSONResponse(
		w,
		http.StatusOK,
		BuildVersion{
			Build:     version.Build,
			Version:   version.Version,
			Timestamp: version.BuildTime,
		})
	if err != nil {
		panic(err)
	}
}

// Provides a handler that dumps the config information as a string
func configHandler(w http.ResponseWriter, req *http.Request) {
	m, err := config.AllConfigsAsMap(config.GetRootSection())
	if err != nil {
		err = WriteStringResponse(w, http.StatusInternalServerError, err.Error())
		if err != nil {
			logger.Errorf(context.TODO(), "Failed to write error response. Error: %v", err)
			panic(err)
		}
	}

	if err := WriteJSONResponse(w, http.StatusOK, m); err != nil {
		panic(err)
	}
}

// Starts an http server on the given port
func StartProfilingServer(ctx context.Context, pprofPort int) error {
	logger.Infof(ctx, "Starting profiling server on port [%v]", pprofPort)
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         fmt.Sprintf(":%d", pprofPort),
	}

	e := srv.ListenAndServe()
	if e != nil {
		logger.Errorf(ctx, "Failed to start profiling server. Error: %v", e)
		return fmt.Errorf("failed to start profiling server, %s", e)
	}

	return nil
}

func configureGlobalHTTPHandler(handlers map[string]http.Handler) error {
	if handlers == nil {
		handlers = map[string]http.Handler{}
	}
	handlers[metricsPath] = promhttp.Handler()
	handlers[healthcheck] = http.HandlerFunc(healtcheckHandler)
	handlers[versionPath] = http.HandlerFunc(versionHandler)
	handlers[configPath] = http.HandlerFunc(configHandler)

	for p, h := range handlers {
		http.Handle(p, h)
	}

	return nil
}

// Forwards the call to StartProfilingServer
// Also registers:
// 1. the prometheus HTTP handler on '/metrics' path shared with the profiling server.
// 2. A healthcheck (L7) handler on '/healthcheck'.
// 3. A version handler on '/version' provides information about the specific build.
// 4. A config handler on '/config' provides a dump of the currently loaded config.
func StartProfilingServerWithDefaultHandlers(ctx context.Context, pprofPort int, handlers map[string]http.Handler) error {
	if err := configureGlobalHTTPHandler(handlers); err != nil {
		return err
	}

	return StartProfilingServer(ctx, pprofPort)
}
