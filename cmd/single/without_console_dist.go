//go:build !console
// +build !console

package single

import (
	"io/fs"
	"net/http"
)

var console fs.FS

var consoleHandlers = make(map[string]func(http.ResponseWriter, *http.Request))
