package containerwatcher

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/sets"
)

func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		if os.IsPermission(err) {
			return false, errors.Wrapf(err, "unable to read file [%s], Flyte does not have permissions", filePath)
		}
		return false, errors.Wrapf(err, "failed to read file")
	}
	return true, nil
}

type fileWatcher struct {
	startWatchDir string
	startFile     string
	exitWatchDir  string
	successFile   string
	errorFile     string
}

func (k fileWatcher) WaitToStart(ctx context.Context) error {
	return wait(ctx, k.startWatchDir, sets.NewString(k.startFile))
}

func (k fileWatcher) WaitToExit(ctx context.Context) error {
	return wait(ctx, k.exitWatchDir, sets.NewString(k.successFile, k.errorFile))
}

func wait(ctx context.Context, watchDir string, s sets.String) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrapf(err, "failed to create watcher")
	}
	defer func() {
		err := w.Close()
		if err != nil {
			logger.Errorf(context.TODO(), "failed to close file watcher")
		}
	}()

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan error)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				done <- nil
				return
			case event, ok := <-w.Events:
				if !ok {
					done <- fmt.Errorf("failed to watch")
					return
				}
				if event.Op == fsnotify.Create || event.Op == fsnotify.Write {
					if s.Has(event.Name) {
						logger.Infof(ctx, "%s file detected", event.Name)
						done <- nil
						return
					}
				}
			case err, ok := <-w.Errors:
				if !ok {
					done <- fmt.Errorf("failed to watch")
					return
				}
				done <- err
				return
			}
		}
	}(childCtx)

	if err := w.Add(watchDir); err != nil {
		return errors.Wrapf(err, "failed to add dir: %s to watcher,", watchDir)
	}

	for _, f := range s.List() {
		if ok, err := FileExists(f); err != nil {
			logger.Errorf(ctx, "Failed to check existence of file %s, err: %s", f, err)
			return err
		} else if ok {
			logger.Infof(ctx, "File %s Already exists", f)
			return nil
		}
	}
	return <-done
}

func NewSuccessFileWatcher(_ context.Context, watchDir, startFileName, successFileName, errorFileName string) (Watcher, error) {
	return fileWatcher{
		startWatchDir: watchDir,
		exitWatchDir:  watchDir,
		startFile:     path.Join(watchDir, startFileName),
		successFile:   path.Join(watchDir, successFileName),
		errorFile:     path.Join(watchDir, errorFileName),
	}, nil
}
