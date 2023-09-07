package ioutils

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

// A Buffered outputWriter just records the io.OutputReader and can be accessed using special methods.
type BufferedOutputWriter struct {
	io.OutputFilePaths
	outReader io.OutputReader
}

func (o *BufferedOutputWriter) Put(ctx context.Context, reader io.OutputReader) error {
	o.outReader = reader
	return nil
}

func (o *BufferedOutputWriter) GetReader() io.OutputReader {
	return o.outReader
}

// Returns a new object of type BufferedOutputWriter
func NewBufferedOutputWriter(ctx context.Context, paths io.OutputFilePaths) *BufferedOutputWriter {
	return &BufferedOutputWriter{
		OutputFilePaths: paths,
	}
}
