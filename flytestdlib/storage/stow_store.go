package storage

import (
	"context"
	"io"
	"time"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytestdlib/errors"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/graymeta/stow"
	errs "github.com/pkg/errors"
)

const (
	FailureTypeLabel contextutils.Key = "failure_type"
)

type stowMetrics struct {
	BadReference labeled.Counter
	BadContainer labeled.Counter

	HeadFailure labeled.Counter
	HeadLatency labeled.StopWatch

	ReadFailure     labeled.Counter
	ReadOpenLatency labeled.StopWatch

	WriteFailure labeled.Counter
	WriteLatency labeled.StopWatch
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

func (s *StowStore) getContainer(ctx context.Context, container string) (c stow.Container, err error) {
	if s.Container.Name() != container {
		s.metrics.BadContainer.Inc(ctx)
		return nil, errs.Wrapf(stow.ErrNotFound, "Conf container:%v != Passed Container:%v", s.Container.Name(), container)
	}

	return s.Container, nil
}

func (s *StowStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return nil, err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.HeadLatency.Start(ctx)
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

	if IsNotFound(err) {
		return StowMetadata{exists: false}, nil
	}

	incFailureCounterForError(ctx, s.metrics.HeadFailure, err)
	return StowMetadata{exists: false}, errs.Wrapf(err, "path:%v", k)
}

func (s *StowStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return nil, err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.ReadOpenLatency.Start(ctx)
	item, err := container.Item(k)
	if err != nil {
		incFailureCounterForError(ctx, s.metrics.ReadFailure, err)
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
		s.metrics.BadReference.Inc(ctx)
		return err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return err
	}

	t := s.metrics.WriteLatency.Start(ctx)
	_, err = container.Put(k, raw, size, opts.Metadata)
	if err != nil {
		incFailureCounterForError(ctx, s.metrics.WriteFailure, err)
		return errs.Wrapf(err, "Failed to write data [%vb] to path [%v].", size, k)
	}

	t.Stop()

	return nil
}

func (s *StowStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.containerBaseFQN
}

func NewStowRawStore(containerBaseFQN DataReference, container stow.Container, metricsScope promutils.Scope) (*StowStore, error) {
	failureTypeOption := labeled.AdditionalLabelsOption{Labels: []string{FailureTypeLabel.String()}}
	self := &StowStore{
		Container:        container,
		containerBaseFQN: containerBaseFQN,
		metrics: &stowMetrics{
			BadReference: labeled.NewCounter("bad_key", "Indicates the provided storage reference/key is incorrectly formatted", metricsScope, labeled.EmitUnlabeledMetric),
			BadContainer: labeled.NewCounter("bad_container", "Indicates request for a container that has not been initialized", metricsScope, labeled.EmitUnlabeledMetric),

			HeadFailure: labeled.NewCounter("head_failure", "Indicates failure in HEAD for a given reference", metricsScope, labeled.EmitUnlabeledMetric),
			HeadLatency: labeled.NewStopWatch("head", "Indicates time to fetch metadata using the Head API", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),

			ReadFailure:     labeled.NewCounter("read_failure", "Indicates failure in GET for a given reference", metricsScope, labeled.EmitUnlabeledMetric, failureTypeOption),
			ReadOpenLatency: labeled.NewStopWatch("read_open", "Indicates time to first byte when reading", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),

			WriteFailure: labeled.NewCounter("write_failure", "Indicates failure in storing/PUT for a given reference", metricsScope, labeled.EmitUnlabeledMetric, failureTypeOption),
			WriteLatency: labeled.NewStopWatch("write", "Time to write an object irrespective of size", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),
		},
	}

	self.copyImpl = newCopyImpl(self, metricsScope)

	return self, nil
}
