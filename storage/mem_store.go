package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/flyteorg/flytestdlib/promutils"
)

type rawFile = []byte

type InMemoryStore struct {
	copyImpl
	cache map[DataReference]rawFile
}

type MemoryMetadata struct {
	exists bool
	size   int64
}

func (m MemoryMetadata) Size() int64 {
	return m.size
}

func (m MemoryMetadata) Exists() bool {
	return m.exists
}

func (s *InMemoryStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	data, found := s.cache[reference]
	return MemoryMetadata{exists: found, size: int64(len(data))}, nil
}

func (s *InMemoryStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	if raw, found := s.cache[reference]; found {
		return ioutil.NopCloser(bytes.NewReader(raw)), nil
	}

	return nil, os.ErrNotExist
}

func (s *InMemoryStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) (
	err error) {

	rawBytes, err := ioutil.ReadAll(raw)
	if err != nil {
		return err
	}

	s.cache[reference] = rawBytes
	return nil
}

func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.cache = map[DataReference]rawFile{}
	return nil
}

func (s *InMemoryStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return DataReference("")
}

// CreateSignedURL creates a signed url with the provided properties.
func (s *InMemoryStore) CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error) {
	return SignedURLResponse{}, fmt.Errorf("unsupported")
}

func NewInMemoryRawStore(_ *Config, scope promutils.Scope) (RawStore, error) {
	self := &InMemoryStore{
		cache: map[DataReference]rawFile{},
	}

	self.copyImpl = newCopyImpl(self, scope)
	return self, nil
}
