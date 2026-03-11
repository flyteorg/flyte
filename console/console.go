//go:build console

package console

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	pathpkg "path"
	"strings"

	"github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

//go:embed dist
var distFS embed.FS

func Setup(ctx context.Context, sc *app.SetupContext) error {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return err
	}
	fileServer := http.FileServer(http.FS(sub))
	handler := spaHandler(sub, fileServer)
	sc.Mux.HandleFunc("/v2", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v2/", http.StatusMovedPermanently)
	})
	sc.Mux.Handle("/v2/", http.StripPrefix("/v2", handler))
	logger.Infof(ctx, "Mounted Console at /v2/")
	return nil
}

// spaHandler wraps fileServer with SPA fallback: if the file doesn't exist,
// it serves index.html so client-side routing can take over.
func spaHandler(root fs.FS, fileServer http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(pathpkg.Clean(r.URL.Path), "/")
		if name == "" {
			name = "."
		}
		f, err := root.Open(name)
		if err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/index.html"
			fileServer.ServeHTTP(w, r2)
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
