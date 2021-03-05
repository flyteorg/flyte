package events

import (
	"bufio"
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/logger"
)

// Log Writer writes to the log output
type LogWriter struct{}

func (w *LogWriter) Write(ctx context.Context, content string) error {
	logger.Info(ctx, content)
	return nil
}

func (w *LogWriter) Flush() error {
	return nil
}

// File Writer is just a thin wrapper around io.Writer
type FileWriter struct {
	ioWriter *bufio.Writer
}

func (fw *FileWriter) Write(ctx context.Context, content string) error {
	_, err := fw.ioWriter.WriteString(content)
	return err
}

func (fw *FileWriter) Flush() error {
	return fw.ioWriter.Flush()
}

// Std Writer is just the default standard if no sink type is provided
type StdWriter struct{}

func (s *StdWriter) Write(ctx context.Context, content string) error {
	_, err := fmt.Println(content)
	return err
}

func (s *StdWriter) Flush() error {
	return nil
}
