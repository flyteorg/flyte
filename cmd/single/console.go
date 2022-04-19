package single

import (
	"embed"
	"net/http"
	"strings"
)

//go:embed dist/*
var console embed.FS

const consoleRoot = "/console"
const assetsDir = "assets"
const packageDir = "dist"
const indexHTML = "/index.html"

// GetConsoleFile returns the console file that should be used for the given path.
// Every path has a '/console' as the prefix. After dropping this prefix, the following 3 rules are checked
// Rule 1: If path is now "" or "/" then return index.html
// Rule 2: If path contains no "assets" sub-string and has an additional substring "/", then return "index.html".
//         This is to allow vanity urls in React
// Rule 3: Finally return every file path as is
// For every file add a "dist" as the prefix, as every file is assumed to be packaged in "dist" folder fv
func GetConsoleFile(name string) string {
	name = strings.TrimPrefix(name, consoleRoot)
	if name == "" || name == "/" {
		name = indexHTML
	} else if !strings.Contains(name, assetsDir) {
		if strings.Contains(name[1:], "/") {
			name = indexHTML
		}
	}
	return packageDir + name
}

type consoleFS struct {
	fs http.FileSystem
}

// Open Implements a specific handler for Console - SinglePage React App
// the path re-writing is critical
func (f consoleFS) Open(name string) (http.File, error) {
	return f.fs.Open(GetConsoleFile(name))
}

// GetConsoleHandlers returns a set of handlers that can be added to the Server mux and can handle all console related
// requests
func GetConsoleHandlers() map[string]func(http.ResponseWriter, *http.Request) {
	handlers := make(map[string]func(http.ResponseWriter, *http.Request))
	// Serves console
	consoleHandler := http.FileServer(consoleFS{fs: http.FS(console)})

	// This is the base handle for "/console"
	handlers[consoleRoot] = func(writer http.ResponseWriter, request *http.Request) {
		consoleHandler.ServeHTTP(writer, request)
	}

	// this is the root handler for any pattern that matches "/console/
	// http.mux needs a trailing "/" to allow longest pattern matching.
	// For the previous handler "/console" the mux will only consider exact match
	handlers[consoleRoot+"/"] = func(writer http.ResponseWriter, request *http.Request) {
		consoleHandler.ServeHTTP(writer, request)
	}
	return handlers
}
