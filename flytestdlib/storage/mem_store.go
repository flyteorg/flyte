package storage

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type rawFile = []byte

type InMemoryStore struct {
	copyImpl
	cache   map[DataReference]rawFile
	rwMutex sync.RWMutex
}

type MemoryMetadata struct {
	exists     bool
	size       int64
	etag       string
	contentMD5 string
}

func (m MemoryMetadata) Size() int64 {
	return m.size
}

func (m MemoryMetadata) Exists() bool {
	return m.exists
}

func (m MemoryMetadata) Etag() string {
	return m.etag
}

func (m MemoryMetadata) ContentMD5() string {
	return m.contentMD5
}

func (s *InMemoryStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	data, found := s.cache[reference]
	var hash [md5.Size]byte
	if found {
		hash = md5.Sum(data) // #nosec
	}

	return MemoryMetadata{
		exists: found, size: int64(len(data)),
		etag: hex.EncodeToString(hash[:]),
	}, nil
}

func (s *InMemoryStore) List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) ([]DataReference, Cursor, error) {
	var items []DataReference
	prefix := strings.TrimSuffix(string(reference), "/") + "/"

	for ref := range s.cache {
		if strings.HasPrefix(ref.String(), prefix) {
			items = append(items, ref)
		}
	}

	if len(items) == 0 {
		return nil, NewCursorAtEnd(), os.ErrNotExist
	}

	return items, NewCursorAtEnd(), nil
}

func (s *InMemoryStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()

	if raw, found := s.cache[reference]; found {
		return ioutil.NopCloser(bytes.NewReader(raw)), nil
	}

	return nil, os.ErrNotExist
}

// Delete removes the referenced data from the cache map.
func (s *InMemoryStore) Delete(ctx context.Context, reference DataReference) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	if _, found := s.cache[reference]; !found {
		return os.ErrNotExist
	}

	delete(s.cache, reference)

	return nil
}

func (s *InMemoryStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) (
	err error) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	rawBytes, err := ioutil.ReadAll(raw)
	if err != nil {
		return err
	}

	s.cache[reference] = rawBytes
	return nil
}

func (s *InMemoryStore) Clear(ctx context.Context) error {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

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

func NewInMemoryRawStore(_ context.Context, _ *Config, metrics *dataStoreMetrics) (RawStore, error) {
	self := &InMemoryStore{
		cache: map[DataReference]rawFile{},
	}

	self.copyImpl = newCopyImpl(self, metrics.copyMetrics)
	return self, nil
}
