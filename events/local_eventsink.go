package events

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

type localSink struct {
	mu     sync.Mutex
	writer writer
}

func (s *localSink) Sink(ctx context.Context, message proto.Message) error {
	s.mu.Lock()
	defer s.writer.Flush()
	defer s.mu.Unlock()

	var eventOutput string
	switch e := message.(type) {
	case *event.WorkflowExecutionEvent:
		eventOutput = fmt.Sprintf("[--WF EVENT--] %s, Phase: %s, OccuredAt: %s\n",
			e.ExecutionId, e.Phase, ptypes.TimestampString(e.OccurredAt))
	case *event.NodeExecutionEvent:
		eventOutput = fmt.Sprintf("[--NODE EVENT--] %s, Phase: %s, OccuredAt: %s\n",
			e.Id, e.Phase, ptypes.TimestampString(e.OccurredAt))
	case *event.TaskExecutionEvent:
		eventOutput = fmt.Sprintf("[--TASK EVENT--] %s,%s, Phase: %s, OccuredAt: %s\n",
			e.TaskId, e.ParentNodeExecutionId, e.Phase, ptypes.TimestampString(e.OccurredAt))
	}

	return s.writer.Write(ctx, eventOutput)
}

func (s *localSink) Close() error {
	return nil
}

// EventSink will sink events to a writer that puts the events somewhere depending on how it was configured
type writer interface {
	Write(ctx context.Context, content string) error
	Flush() error
}

func NewLogSink() (EventSink, error) {
	return &localSink{writer: &LogWriter{}}, nil
}

func NewStdoutSink() (EventSink, error) {
	return &localSink{writer: &StdWriter{}}, nil
}

// TODO this will cause multiple handles to the same file if we open multiple syncs. Maybe we should remove this
func NewFileSink(path string) (EventSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(f)

	if err != nil {
		return nil, err
	}
	return &localSink{writer: &FileWriter{ioWriter: w}}, nil
}
