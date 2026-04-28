// Package podlogs provides shared helpers for streaming Kubernetes pod logs
// and parsing the RFC3339-prefixed log lines emitted when Timestamps=true.
package podlogs

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

// DefaultBatchSize is the default number of log lines sent per batch.
const DefaultBatchSize = 100

// ParseLine splits a K8s log line into timestamp and message.
// K8s timestamped lines are formatted as: "2006-01-02T15:04:05.999999999Z message".
// If parsing fails, the full line is returned as the message with no timestamp.
func ParseLine(line string) *dataplane.LogLine {
	if idx := strings.IndexByte(line, ' '); idx > 0 {
		if t, err := time.Parse(time.RFC3339Nano, line[:idx]); err == nil {
			return &dataplane.LogLine{
				Originator: dataplane.LogLineOriginator_USER,
				Timestamp:  timestamppb.New(t),
				Message:    line[idx+1:],
			}
		}
	}
	return &dataplane.LogLine{
		Originator: dataplane.LogLineOriginator_USER,
		Message:    line,
	}
}

// Stream reads lines from r, groups them into batches of up to batchSize, and
// invokes flush for each batch. Batches are also flushed when the underlying
// reader has no buffered data available, so clients aren't stuck waiting on a
// partially-filled batch when the pod is idle.
//
// Returns io.EOF as nil. A non-EOF read error is returned only when ctx has not
// been cancelled — callers typically treat ctx cancellation as a clean close.
func Stream(ctx context.Context, r io.Reader, batchSize int, flush func([]*dataplane.LogLine) error) error {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	reader := bufio.NewReader(r)
	lines := make([]*dataplane.LogLine, 0, batchSize)

	doFlush := func() error {
		if len(lines) == 0 {
			return nil
		}
		if err := flush(lines); err != nil {
			return err
		}
		lines = make([]*dataplane.LogLine, 0, batchSize)
		return nil
	}

	var readErr error
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			lines = append(lines, ParseLine(strings.TrimRight(line, "\r\n")))
			if len(lines) >= batchSize {
				if err := doFlush(); err != nil {
					return err
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				readErr = err
			}
			break
		}
		if len(lines) > 0 && reader.Buffered() == 0 {
			if err := doFlush(); err != nil {
				return err
			}
		}
	}

	if err := doFlush(); err != nil {
		return err
	}
	if readErr != nil && ctx.Err() == nil {
		return readErr
	}
	return nil
}
