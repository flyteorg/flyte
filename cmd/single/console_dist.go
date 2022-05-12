//go:build console
// +build console

package single

import (
	"embed"
	"net/http"
)

//go:embed dist/*
var console embed.FS

var consoleHandler = http.FileServer(consoleFS{fs: http.FS(console)})

type handlerFunc = func(http.ResponseWriter, *http.Request)

var consoleHandlers = map[string]handlerFunc{
	consoleRoot: func(writer http.ResponseWriter, request *http.Request) {
		consoleHandler.ServeHTTP(writer, request)
	},
	consoleRoot + "/": func(writer http.ResponseWriter, request *http.Request) {
		consoleHandler.ServeHTTP(writer, request)
	},
}
