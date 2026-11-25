package profutils

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // #nosec G108 Import for pprof server
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// Starts an http server on the given port
func StartProfilingServer(ctx context.Context, pprofPort int) error {
	logger.Infof(ctx, "Starting profiling server on port [%v]", pprofPort)
	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         fmt.Sprintf(":%d", pprofPort),
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Errorf(ctx, "Failed to gracefully shutdown profiling server. Error: %v", err)
		}
	}()

	return srv.ListenAndServe()
}
