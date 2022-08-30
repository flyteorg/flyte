package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	errs "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flytestdlib/ioutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
)

type copyImpl struct {
	rawStore RawStore
	metrics  *copyMetrics
}

type copyMetrics struct {
	CopyLatency                  labeled.StopWatch
	ComputeLengthLatency         labeled.StopWatch
	WriteFailureUnrelatedToCache prometheus.Counter
	ReadFailureUnrelatedToCache  prometheus.Counter
}

// CopyRaw is a naiive implementation for copy that reads all data locally then writes them to destination.
// TODO: We should upstream an API change to stow to implement copy more natively. E.g. Use s3 copy:
// https://docs.aws.amazon.com/AmazonS3/latest/dev/CopyingObjectUsingREST.html
func (c copyImpl) CopyRaw(ctx context.Context, source, destination DataReference, opts Options) error {
	rc, err := c.rawStore.ReadRaw(ctx, source)

	if err != nil && !IsFailedWriteToCache(err) {
		logger.Errorf(ctx, "Failed to read from the raw store when copying [%v] to [%v]. Error: %v", source, destination, err)
		c.metrics.ReadFailureUnrelatedToCache.Inc()
		return errs.Wrap(err, fmt.Sprintf("path:%v", destination))
	}

	length := int64(0)
	if _, isSeeker := rc.(io.Seeker); !isSeeker {
		// If the returned ReadCloser doesn't implement Seeker interface, then the underlying writer won't be able to
		// calculate content length on its own. Some implementations (e.g. S3 Stow Store) will error if it can't.
		var raw []byte
		raw, err = ioutils.ReadAll(rc, c.metrics.ComputeLengthLatency.Start(ctx))
		if err != nil {
			return err
		}

		rc = ioutils.NewBytesReadCloser(raw)
		length = int64(len(raw))
	}

	err = c.rawStore.WriteRaw(ctx, destination, length, Options{}, rc)

	if err != nil && !IsFailedWriteToCache(err) {
		logger.Errorf(ctx, "Failed to write to the raw store when copying. Error: %v", err)
		c.metrics.WriteFailureUnrelatedToCache.Inc()
		return err
	}

	return nil
}

func newCopyMetrics(scope promutils.Scope) *copyMetrics {
	return &copyMetrics{
		CopyLatency:                  labeled.NewStopWatch("overall", "Overall copy latency", time.Millisecond, scope, labeled.EmitUnlabeledMetric),
		ComputeLengthLatency:         labeled.NewStopWatch("length", "Latency involved in computing length of content before writing.", time.Millisecond, scope, labeled.EmitUnlabeledMetric),
		WriteFailureUnrelatedToCache: scope.MustNewCounter("write_failure_unrelated_to_cache", "Raw store write failures that are not caused by ErrFailedToWriteCache"),
		ReadFailureUnrelatedToCache:  scope.MustNewCounter("read_failure_unrelated_to_cache", "Raw store read failures that are not caused by ErrFailedToWriteCache"),
	}
}

func newCopyImpl(store RawStore, metrics *copyMetrics) copyImpl {
	return copyImpl{
		rawStore: store,
		metrics:  metrics,
	}
}
