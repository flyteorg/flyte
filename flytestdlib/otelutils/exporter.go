// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelutils

import (
	"context"
	"encoding/hex"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ trace.SpanExporter = &Exporter{}

var (
	traceIDRegexp = regexp.MustCompile(`\"traceId\":\"([^"]*)\"`)
	spanIDRegexp = regexp.MustCompile(`\"spanId\":\"([^"]*)\"`)
	parentSpanIDRegexp = regexp.MustCompile(`\"parentSpanId\":\"([^"]*)\"`)
	randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func NewExporter(writer io.Writer) (*Exporter, error) {
	return &Exporter{
		idMap:  make(map[string]string),
		writer: writer,
	}, nil
}

// Exporter is an implementation of trace.SpanSyncer that writes spans to stdout.
type Exporter struct {
	idMap       map[string]string
	stopped     bool
	stoppedLock sync.RWMutex
	writer      io.Writer
	writerLock  sync.Mutex
}

// ExportSpans writes spans in json format to stdout.
func (e *Exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	e.stoppedLock.RLock()
	stopped := e.stopped
	e.stoppedLock.RUnlock()
	if stopped {
		return nil
	}

	if len(spans) == 0 {
		return nil
	}

	//stubs := tracetest.SpanStubsFromReadOnlySpans(spans)
    protoSpans := Spans(spans)
	exportTraceServiceRequest := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}

	msg := protojson.MarshalOptions{Multiline: false}.Format(exportTraceServiceRequest)

	e.writerLock.Lock()
	defer e.writerLock.Unlock()

	// TODO @hamersaw - regex find 'traceId' and 'spanId' and replace with hex encoded strings rather than protobuf []byte encoding
	uniqueIDs := make(map[string]string)

	traceIDs := traceIDRegexp.FindAllStringSubmatch(msg, -1)
	e.mapUniqueIDs(traceIDs, uniqueIDs, 32)

	spanIDs := spanIDRegexp.FindAllStringSubmatch(msg, -1)
	e.mapUniqueIDs(spanIDs, uniqueIDs, 16)

	parentSpanIDs := parentSpanIDRegexp.FindAllStringSubmatch(msg, -1)
	e.mapUniqueIDs(parentSpanIDs, uniqueIDs, 16)

	// TODO @hamersaw - parentSpanIds are not matching up with spanIds, need to figure out why
	// TODO @hamersaw - gorm migrations are not working from formatting - ex `FROM \"migrations\"`

	for id, hexID := range uniqueIDs {
		msg = strings.ReplaceAll(msg, id, hexID)
	}

	msg += "\n"
	e.writer.Write([]byte(msg))
	return nil
}

// Shutdown is called to stop the exporter, it performs no action.
func (e *Exporter) Shutdown(ctx context.Context) error {
	e.stoppedLock.Lock()
	e.stopped = true
	e.stoppedLock.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return nil
}

// MarshalLog is the marshaling function used by the logging system to represent this exporter.
func (e *Exporter) MarshalLog() interface{} {
	return struct {
		Type string
	}{
		Type: "oteljson",
	}
}

func (e *Exporter) mapUniqueIDs(ids [][]string, uniqueIDs map[string]string, idLength int) {
	for _, id := range ids {
		if _, exists := uniqueIDs[id[1]]; !exists {
			hexTraceID, globalExists := e.idMap[id[1]]
			if !globalExists {
				hexTraceID = getRandomHexString(idLength)
				e.idMap[id[1]] = hexTraceID
			}

			uniqueIDs[id[1]] = hexTraceID
		}
	}
}


func getRandomHexString(n int) string {
    b := make([]byte, (n+1)/2) // can be simplified to n/2 if n is always even
    if _, err := randomGenerator.Read(b); err != nil {
            panic(err)
    }

    return hex.EncodeToString(b)[:n]
}
