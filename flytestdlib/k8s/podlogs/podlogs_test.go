package podlogs

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

func TestParseLine_WithTimestamp(t *testing.T) {
	line := "2024-01-15T10:30:00.123456789Z Hello, world!"
	logLine := ParseLine(line)

	assert.Equal(t, "Hello, world!", logLine.Message)
	require.NotNil(t, logLine.Timestamp)
	expected := time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)
	assert.Equal(t, expected, logLine.Timestamp.AsTime())
	assert.Equal(t, dataplane.LogLineOriginator_USER, logLine.Originator)
}

func TestParseLine_WithoutTimestamp(t *testing.T) {
	logLine := ParseLine("just a plain log message")
	assert.Equal(t, "just a plain log message", logLine.Message)
	assert.Nil(t, logLine.Timestamp)
	assert.Equal(t, dataplane.LogLineOriginator_USER, logLine.Originator)
}

func TestParseLine_MalformedTimestamp(t *testing.T) {
	logLine := ParseLine("not-a-timestamp some message")
	assert.Equal(t, "not-a-timestamp some message", logLine.Message)
	assert.Nil(t, logLine.Timestamp)
}

func TestParseLine_EmptyMessage(t *testing.T) {
	logLine := ParseLine("2024-01-15T10:30:00Z ")
	assert.Equal(t, "", logLine.Message)
	require.NotNil(t, logLine.Timestamp)
}

func TestStream_BatchesAndFinalFlush(t *testing.T) {
	input := "line-1\nline-2\nline-3\n"
	var batches [][]*dataplane.LogLine
	err := Stream(context.Background(), strings.NewReader(input), 2, func(lines []*dataplane.LogLine) error {
		batches = append(batches, lines)
		return nil
	})
	require.NoError(t, err)
	require.Len(t, batches, 2)
	assert.Len(t, batches[0], 2)
	assert.Len(t, batches[1], 1)
	assert.Equal(t, "line-3", batches[1][0].Message)
}

func TestStream_PropagatesFlushError(t *testing.T) {
	err := Stream(context.Background(), strings.NewReader("a\nb\n"), 1, func(lines []*dataplane.LogLine) error {
		return io.ErrClosedPipe
	})
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}
