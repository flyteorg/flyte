package single

import (
	"path/filepath"
	"net/http"
	"strings"
)

const (
	consoleRoot = "/console"
	consoleStatic = consoleRoot + "/assets/"
	packageDir = "dist"
	indexHTML = "index.html"
)

// GetConsoleFile returns the console file that should be used for the given path.
// For every file add a "dist" as the prefix, as every file is assumed to be packaged in "dist" folder fv
func GetConsoleFile(name string) string {
	// Serve requests for static assets at `/console/assets/<filename>`
	// as-is from `dist`
	if strings.HasPrefix(name, consoleStatic) {
		return filepath.Join(packageDir, strings.TrimPrefix(name, consoleStatic))
	}

	// Send all other requests to `index.html` to be handled by react router
	return filepath.Join(packageDir, indexHTML)
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
	return consoleHandlers
}
