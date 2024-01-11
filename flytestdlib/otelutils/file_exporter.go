package otelutils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

var _ trace.SpanExporter = &Exporter{}

var (
	traceIDRegexp      = regexp.MustCompile(`\"traceId\":\"([^"]*)\"`)
	spanIDRegexp       = regexp.MustCompile(`\"spanId\":\"([^"]*)\"`)
	parentSpanIDRegexp = regexp.MustCompile(`\"parentSpanId\":\"([^"]*)\"`)
)

type Exporter struct {
	formatter    spanFormatter
	shutdown     bool
	shutdownLock sync.RWMutex
	writer       io.Writer
	writerLock   sync.Mutex
}

func (e *Exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	e.shutdownLock.RLock()
	shutdown := e.shutdown
	e.shutdownLock.RUnlock()
	if shutdown {
		return nil
	}

	if len(spans) == 0 {
		return nil
	}

	e.writerLock.Lock()
	defer e.writerLock.Unlock()
	msg, err := e.formatter.format(spans)
	if err != nil {
		return err
	}
	msg += "\n"

	// if not all bytes are written an error is returned per docs, so not necessary to check
	// the number of bytes written
	_, err = e.writer.Write([]byte(msg))
	return err
}

func (e *Exporter) Shutdown(ctx context.Context) error {
	e.shutdownLock.Lock()
	e.shutdown = true
	e.shutdownLock.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return nil
}

func (e *Exporter) MarshalLog() interface{} {
	return struct {
		Type string
	}{
		Type: "oteljson",
	}
}

func NewExporter(formatter spanFormatter, writer io.Writer) (*Exporter, error) {
	return &Exporter{
		formatter: formatter,
		writer:    writer,
	}, nil
}

func getRandomHexString(n int) string {
	b := make([]byte, (n+1)/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)[:n]
}

type spanFormatter interface {
	format(spans []trace.ReadOnlySpan) (string, error)
}

type OtelSpanFormatter struct{}

func (*OtelSpanFormatter) format(spans []trace.ReadOnlySpan) (string, error) {
	msg := ""
	stubs := tracetest.SpanStubsFromReadOnlySpans(spans)
	for _, stub := range stubs {
		jsonData, err := json.Marshal(stub)
		if err != nil {
			return "", err
		}

		// currently converting to string, if this becomes a performance issue
		// we can use a byte buffer and write to it
		if len(msg) > 0 {
			msg += "\n"
		}
		msg += string(jsonData)
	}

	return msg, nil
}

func NewOTELSpanFormatter() *OtelSpanFormatter {
	return &OtelSpanFormatter{}
}

type OtelCollectorSpanFormatter struct {
	byteToStringIDMap *lru.Cache[string, string]
}

func (o *OtelCollectorSpanFormatter) format(spans []trace.ReadOnlySpan) (string, error) {
	protoSpans := Spans(spans)
	exportTraceServiceRequest := &coltracepb.ExportTraceServiceRequest{
		ResourceSpans: protoSpans,
	}

	msg := protojson.MarshalOptions{Multiline: false}.Format(exportTraceServiceRequest)

	// use regex to replace ids with hex encoded strings rather than protobuf []byte encodings
	uniqueIDs := make(map[string]string)

	traceIDs := traceIDRegexp.FindAllStringSubmatch(msg, -1)
	o.mapUniqueIDs(traceIDs, uniqueIDs, 32)

	spanIDs := spanIDRegexp.FindAllStringSubmatch(msg, -1)
	o.mapUniqueIDs(spanIDs, uniqueIDs, 16)

	parentSpanIDs := parentSpanIDRegexp.FindAllStringSubmatch(msg, -1)
	o.mapUniqueIDs(parentSpanIDs, uniqueIDs, 16)

	// TODO @hamersaw - gorm migrations are not working from formatting - ex `FROM \"migrations\"`

	for id, hexID := range uniqueIDs {
		msg = strings.ReplaceAll(msg, id, hexID)
	}

	return msg, nil
}

func (o *OtelCollectorSpanFormatter) mapUniqueIDs(ids [][]string, uniqueIDs map[string]string, idLength int) {
	for _, id := range ids {
		if _, exists := uniqueIDs[id[1]]; !exists {
			hexTraceID, globalExists := o.byteToStringIDMap.Get(id[1])
			if !globalExists {
				hexTraceID = getRandomHexString(idLength)
				o.byteToStringIDMap.Add(id[1], hexTraceID)
			}

			uniqueIDs[id[1]] = hexTraceID
		}
	}
}

func NewOTELCollectorSpanFormatter() (*OtelCollectorSpanFormatter, error) {
	byteToStringIDMap, err := lru.New[string, string](4096)
	if err != nil {
		return nil, err
	}

	return &OtelCollectorSpanFormatter{
		byteToStringIDMap: byteToStringIDMap,
	}, nil
}
