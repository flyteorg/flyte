package app

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// App is the shared entry-point skeleton for Flyte services.
type App struct {
	// Name is the cobra command name (e.g. "flyte", "runs-service").
	Name string

	// Short is the short description shown in --help.
	Short string

	// Setup is called after config initialisation. It should populate the
	// SetupContext with shared resources and register HTTP handlers / workers.
	Setup func(ctx context.Context, sc *SetupContext) error
}

var (
	cfgFile        string
	configAccessor config.Accessor
)

// Run builds the cobra command and executes it.
func (a *App) Run() error {
	return a.Command().Execute()
}

// Command returns the root cobra.Command with config init wired in.
func (a *App) Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          a.Name,
		Short:        a.Short,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.serve(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	configAccessor = viper.NewAccessor(config.Options{StrictMode: false})
	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return rootCmd
}

// serve is the main run loop.
func (a *App) serve(ctx context.Context) error {
	// 1. Logger
	if err := logger.SetConfig(logger.GetConfig()); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Infof(ctx, "Starting %s", a.Name)

	// 2. Build SetupContext with defaults
	sc := &SetupContext{
		Host: "0.0.0.0",
		Port: 8080,
		Mux:  http.NewServeMux(),
	}

	// 3. Let the caller populate resources & register handlers
	if err := a.Setup(ctx, sc); err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1+len(sc.workers))

	// 4. Start HTTP server (skipped when Port == 0, i.e. worker-only mode)
	var server *http.Server
	if sc.Port > 0 {
		sc.Mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})
		sc.Mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			for _, check := range sc.readyChecks {
				if err := check(r); err != nil {
					w.WriteHeader(http.StatusServiceUnavailable)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		var handler http.Handler = sc.Mux
		if sc.Middleware != nil {
			handler = sc.Middleware(handler)
		}
		handler = requestGzipDecompressMiddleware(handler)

		addr := fmt.Sprintf("%s:%d", sc.Host, sc.Port)
		server = &http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(handler, &http2.Server{}),
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Infof(ctx, "%s listening on %s", a.Name, addr)
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("server error: %w", err)
			}
		}()
	}

	// 5. Start background workers
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	for _, w := range sc.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Infof(ctx, "Starting worker: %s", w.name)
			if err := w.fn(workerCtx); err != nil {
				errCh <- fmt.Errorf("worker %s error: %w", w.name, err)
			}
		}()
	}

	// 6. Wait for signal or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Infof(ctx, "Received signal %v, shutting down gracefully...", sig)
	case err := <-errCh:
		logger.Errorf(ctx, "Fatal error: %v", err)
		workerCancel()
		return err
	}

	// 7. Graceful shutdown with overall deadline.
	// A second SIGINT/SIGTERM forces immediate exit.
	const shutdownTimeout = 30 * time.Second
	forceCh := make(chan os.Signal, 1)
	signal.Notify(forceCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-forceCh
		logger.Warnf(ctx, "Received second signal, forcing immediate exit")
		os.Exit(1)
	}()

	workerCancel()
	var shutdownErr error
	if server != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Errorf(ctx, "Server shutdown error: %v", err)
			shutdownErr = err
		}
	}

	// Wait for all workers to finish, but enforce a deadline so we don't hang.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		logger.Infof(ctx, "%s stopped", a.Name)
	case <-time.After(shutdownTimeout):
		logger.Warnf(ctx, "%s: graceful shutdown timed out after %v, forcing exit", a.Name, shutdownTimeout)
	}
	return shutdownErr
}

// requestGzipDecompressMiddleware pre-decompresses request bodies that carry
// Content-Encoding: gzip before they reach the connect-rpc handler.
//
// Some HTTP clients (e.g. pyqwest used by the Python connectrpc SDK) compress
// the request body at the application level and set Content-Encoding: gzip,
// but also use chunked transfer encoding for larger payloads.  Go's h2c
// framing delivers the body in chunks, so by the time connect-go calls
// gzip.NewReader on the raw body stream the first bytes it receives may be a
// chunk-size line rather than the gzip magic bytes, causing "gzip: invalid
// header".  Pre-reading and fully decompressing the body here, before h2c
// framing is involved, sidesteps the problem entirely.
func requestGzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err == nil {
				r2 := r.Clone(r.Context())
				r2.Body = struct {
					io.Reader
					io.Closer
				}{gr, r.Body}
				r2.Header.Del("Content-Encoding")
				r = r2
			}
		}
		next.ServeHTTP(w, r)
	})
}

func initConfig(cmd *cobra.Command) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return configAccessor.UpdateConfig(context.Background())
}
