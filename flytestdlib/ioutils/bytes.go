package ioutils

import (
	"bytes"
	"io"
)

// A Closeable Reader for bytes to mimic stream from inmemory byte storage
type BytesReadCloser struct {
	*bytes.Reader
}

func (*BytesReadCloser) Close() error {
	return nil
}

func NewBytesReadCloser(b []byte) io.ReadCloser {
	return &BytesReadCloser{
		Reader: bytes.NewReader(b),
	}
}
