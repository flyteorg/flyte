package storage

import (
	"context"
	"io"
	"time"

	"github.com/lyft/flytestdlib/ioutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
)

type copyImpl struct {
	rawStore RawStore
	metrics  copyMetrics
}

type copyMetrics struct {
	CopyLatency          labeled.StopWatch
	ComputeLengthLatency labeled.StopWatch
}

// A naiive implementation for copy that reads all data locally then writes them to destination.
// TODO: We should upstream an API change to stow to implement copy more natively. E.g. Use s3 copy:
// 	https://docs.aws.amazon.com/AmazonS3/latest/dev/CopyingObjectUsingREST.html
func (c copyImpl) CopyRaw(ctx context.Context, source, destination DataReference, opts Options) error {
	rc, err := c.rawStore.ReadRaw(ctx, source)
	if err != nil {
		return err
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

	return c.rawStore.WriteRaw(ctx, destination, length, Options{}, rc)
}

func newCopyMetrics(scope promutils.Scope) copyMetrics {
	return copyMetrics{
		CopyLatency:          labeled.NewStopWatch("overall", "Overall copy latency", time.Millisecond, scope, labeled.EmitUnlabeledMetric),
		ComputeLengthLatency: labeled.NewStopWatch("length", "Latency involved in computing length of content before writing.", time.Millisecond, scope, labeled.EmitUnlabeledMetric),
	}
}

func newCopyImpl(store RawStore, metricsScope promutils.Scope) copyImpl {
	return copyImpl{
		rawStore: store,
		metrics:  newCopyMetrics(metricsScope.NewSubScope("copy")),
	}
}
