package single

import (
	"embed"
	"fmt"
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
	fs := http.FS(console)
	rawFS := http.FileServer(fs)
	consoleFS := http.StripPrefix("/console/", rawFS)
	handlers["/console/assets/"] = func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf("all files under console returning, %s\n", request.URL.Path)
		consoleFS.ServeHTTP(writer, request)
	}

	handlers["/console/"] = func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf("all files under console returning, %s\n", request.URL.Path)
		newPath := strings.TrimLeft(request.URL.Path, "/console")
		if strings.Contains(newPath, "/") {
			WriteIndex(writer)
		} else {
			consoleFS.ServeHTTP(writer, request)
		}
	}
	handlers["/console"] = func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("returning index.html")
		WriteIndex(writer)
	}

	return handlers
}

