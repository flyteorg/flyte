package storage

import (
	"context"
	"io"
	"time"

	"github.com/lyft/flytestdlib/errors"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/graymeta/stow"
	errs "github.com/pkg/errors"
)

type stowMetrics struct {
	BadReference prometheus.Counter
	BadContainer prometheus.Counter

	HeadFailure prometheus.Counter
	HeadLatency promutils.StopWatch

	ReadFailure     prometheus.Counter
	ReadOpenLatency promutils.StopWatch

	WriteFailure prometheus.Counter
	WriteLatency promutils.StopWatch
}

// Implements DataStore to talk to stow location store.
type StowStore struct {
	stow.Container
	copyImpl
	metrics          *stowMetrics
	containerBaseFQN DataReference
}

type StowMetadata struct {
	exists bool
	size   int64
}

func (s StowMetadata) Size() int64 {
	return s.size
}

func (s StowMetadata) Exists() bool {
	return s.exists
}

func (s *StowStore) getContainer(container string) (c stow.Container, err error) {
	if s.Container.Name() != container {
		s.metrics.BadContainer.Inc()
		return nil, errs.Wrapf(stow.ErrNotFound, "Conf container:%v != Passed Container:%v", s.Container.Name(), container)
	}

	return s.Container, nil
}

func (s *StowStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc()
		return nil, err
	}

	container, err := s.getContainer(c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.HeadLatency.Start()
	item, err := container.Item(k)
	if err == nil {
		if _, err = item.Metadata(); err == nil {
			size, err := item.Size()
			if err == nil {
				t.Stop()
				return StowMetadata{
					exists: true,
					size:   size,
				}, nil
			}
		}
	}
	s.metrics.HeadFailure.Inc()
	if IsNotFound(err) {
		return StowMetadata{exists: false}, nil
	}
	return StowMetadata{exists: false}, errs.Wrapf(err, "path:%v", k)
}

func (s *StowStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc()
		return nil, err
	}

	container, err := s.getContainer(c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.ReadOpenLatency.Start()
	item, err := container.Item(k)
	if err != nil {
		s.metrics.ReadFailure.Inc()
		return nil, err
	}
	t.Stop()

	sizeBytes, err := item.Size()
	if err != nil {
		return nil, err
	}

	if sizeMbs := sizeBytes / MiB; sizeMbs > GetConfig().Limits.GetLimitMegabytes {
		return nil, errors.Errorf(ErrExceedsLimit, "limit exceeded. %vmb > %vmb.", sizeMbs, GetConfig().Limits.GetLimitMegabytes)
	}

	return item.Open()
}

func (s *StowStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc()
		return err
	}

	container, err := s.getContainer(c)
	if err != nil {
		return err
	}

	t := s.metrics.WriteLatency.Start()
	_, err = container.Put(k, raw, size, opts.Metadata)
	if err != nil {
		s.metrics.WriteFailure.Inc()
		return errs.Wrapf(err, "Failed to write data [%vb] to path [%v].", size, k)
	}
	t.Stop()

	return nil
}

func (s *StowStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.containerBaseFQN
}

func NewStowRawStore(containerBaseFQN DataReference, container stow.Container, metricsScope promutils.Scope) (*StowStore, error) {
	self := &StowStore{
		Container:        container,
		containerBaseFQN: containerBaseFQN,
		metrics: &stowMetrics{
			BadReference: metricsScope.MustNewCounter("bad_key", "Indicates the provided storage reference/key is incorrectly formatted"),
			BadContainer: metricsScope.MustNewCounter("bad_container", "Indicates request for a container that has not been initialized"),

			HeadFailure: metricsScope.MustNewCounter("head_failure", "Indicates failure in HEAD for a given reference"),
			HeadLatency: metricsScope.MustNewStopWatch("head", "Indicates time to fetch metadata using the Head API", time.Millisecond),

			ReadFailure:     metricsScope.MustNewCounter("read_failure", "Indicates failure in GET for a given reference"),
			ReadOpenLatency: metricsScope.MustNewStopWatch("read_open", "Indicates time to first byte when reading", time.Millisecond),

			WriteFailure: metricsScope.MustNewCounter("write_failure", "Indicates failure in storing/PUT for a given reference"),
			WriteLatency: metricsScope.MustNewStopWatch("write", "Time to write an object irrespective of size", time.Millisecond),
		},
	}

	self.copyImpl = newCopyImpl(self, metricsScope)

	return self, nil
}
