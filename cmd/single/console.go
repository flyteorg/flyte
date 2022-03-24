package single

import (
	"context"
	"embed"
	"github.com/flyteorg/flytestdlib/logger"
	"net/http"
	"strings"
)

//go:embed dist/index.html dist/*.html dist/assets
var console embed.FS

func WriteIndex(writer http.ResponseWriter) {
	b, err := console.ReadFile("dist/index.html")
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = writer.Write(b)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func GetConsoleHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	handlers := make(map[string]func(http.ResponseWriter, *http.Request))
	// Serves console
	rawFS := http.FileServer(http.FS(console))
	consoleFS := http.StripPrefix("/console/", rawFS)
	handlers["/console/assets/"] = func(writer http.ResponseWriter, request *http.Request) {
		logger.Infof(context.TODO(), "Returning assets, %s", request.URL.Path)
		consoleFS.ServeHTTP(writer, request)
	}

	handlers["/console/"] = func(writer http.ResponseWriter, request *http.Request) {
		newPath := strings.TrimLeft(request.URL.Path, "/console")
		if strings.Contains(newPath, "/") {
			logger.Infof(context.TODO(), "Redirecting request to index.html, %s", request.URL.Path)
			WriteIndex(writer)
		} else {
			consoleFS.ServeHTTP(writer, request)
		}
	}
	handlers["/console"] = func(writer http.ResponseWriter, request *http.Request) {
		logger.Infof(context.TODO(), "Returning index.html")
		WriteIndex(writer)
	}

	return handlers
}

