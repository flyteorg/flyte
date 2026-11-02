package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

type lokiEntry struct {
	ts   time.Time
	line string
}

// lokiResponse builds a query_range response body with a single stream.
func lokiResponse(entries []lokiEntry) []byte {
	values := make([][]string, 0, len(entries))
	for _, e := range entries {
		values = append(values, []string{strconv.FormatInt(e.ts.UnixNano(), 10), e.line})
	}
	body := map[string]any{
		"status": "success",
		"data": map[string]any{
			"resultType": "streams",
			"result": []map[string]any{
				{"stream": map[string]string{"pod": "my-pod"}, "values": values},
			},
		},
	}
	raw, _ := json.Marshal(body)
	return raw
}

func newLokiTestConfig(url string) config.LokiConfig {
	cfg := config.DefaultLokiConfig()
	cfg.URL = url
	return cfg
}

func newLokiLogContext(start, end *time.Time) *core.LogContext {
	process := &core.ContainerContext_ProcessContext{}
	if start != nil {
		process.ContainerStartTime = timestamppb.New(*start)
	}
	if end != nil {
		process.ContainerEndTime = timestamppb.New(*end)
	}
	return &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{{
			Namespace:            "flyte",
			PodName:              "my-pod",
			PrimaryContainerName: "main",
			Containers: []*core.ContainerContext{
				{ContainerName: "sidecar"},
				{ContainerName: "main", Process: process},
			},
		}},
	}
}

func collectLines(t *testing.T, s *LokiLogStreamer, logCtx *core.LogContext) ([]*dataplane.LogLine, error) {
	t.Helper()
	var got []*dataplane.LogLine
	err := s.fetch(context.Background(), logCtx, func(lines []*dataplane.LogLine) error {
		got = append(got, lines...)
		return nil
	})
	return got, err
}

func TestLokiLogStreamer_QueriesLokiFromLogContext(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Minute)

	var gotReq *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r.Clone(context.Background())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(lokiResponse([]lokiEntry{
			{ts: start.Add(time.Second), line: "hello"},
			{ts: start.Add(2 * time.Second), line: "world"},
		}))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.TenantID = "tenant-a"
	cfg.BasicAuth.Username = "user"
	cfg.BasicAuth.Password = "pass"
	cfg.TimePadding = stdconfig.Duration{Duration: time.Minute}

	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)

	require.NotNil(t, gotReq)
	assert.Equal(t, "/loki/api/v1/query_range", gotReq.URL.Path)
	q := gotReq.URL.Query()
	assert.Equal(t, `{namespace="flyte", pod="my-pod", container="main"}`, q.Get("query"))
	assert.Equal(t, strconv.FormatInt(start.Add(-time.Minute).UnixNano(), 10), q.Get("start"))
	assert.Equal(t, strconv.FormatInt(end.Add(time.Minute).UnixNano(), 10), q.Get("end"))
	assert.Equal(t, "forward", q.Get("direction"))
	assert.Equal(t, strconv.Itoa(cfg.PageSize), q.Get("limit"))
	assert.Equal(t, "tenant-a", gotReq.Header.Get("X-Scope-OrgID"))
	user, pass, ok := gotReq.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "user", user)
	assert.Equal(t, "pass", pass)

	require.Len(t, lines, 2)
	assert.Equal(t, "hello", lines[0].GetMessage())
	assert.Equal(t, start.Add(time.Second), lines[0].GetTimestamp().AsTime())
	assert.Equal(t, dataplane.LogLineOriginator_USER, lines[0].GetOriginator())
	assert.Equal(t, "world", lines[1].GetMessage())
}

func TestLokiLogStreamer_PaginatesUntilShortPage(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)

	var starts []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		starts = append(starts, r.URL.Query().Get("start"))
		var entries []lokiEntry
		switch len(starts) {
		case 1:
			entries = []lokiEntry{{ts: start.Add(1 * time.Second), line: "a"}, {ts: start.Add(2 * time.Second), line: "b"}}
		case 2:
			entries = []lokiEntry{{ts: start.Add(3 * time.Second), line: "c"}}
		default:
			t.Errorf("unexpected extra request %d", len(starts))
		}
		_, _ = w.Write(lokiResponse(entries))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.PageSize = 2
	cfg.TimePadding = stdconfig.Duration{}

	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)

	require.Len(t, starts, 2)
	assert.Equal(t, strconv.FormatInt(start.UnixNano(), 10), starts[0])
	assert.Equal(t, strconv.FormatInt(start.Add(2*time.Second).UnixNano(), 10), starts[1])

	require.Len(t, lines, 3)
	assert.Equal(t, []string{"a", "b", "c"}, []string{lines[0].GetMessage(), lines[1].GetMessage(), lines[2].GetMessage()})
}

func TestLokiLogStreamer_StopsAtMaxLines(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		entries := []lokiEntry{
			{ts: start.Add(time.Duration(requests*2-1) * time.Second), line: "x"},
			{ts: start.Add(time.Duration(requests*2) * time.Second), line: "y"},
		}
		_, _ = w.Write(lokiResponse(entries))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.PageSize = 2
	cfg.MaxLines = 3

	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)
	assert.Len(t, lines, 3)
	assert.Equal(t, 2, requests)
}

func TestLokiLogStreamer_UsesLookbackWhenNoStartTime(t *testing.T) {
	var q map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q = r.URL.Query()
		_, _ = w.Write(lokiResponse(nil))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.Lookback = stdconfig.Duration{Duration: 2 * time.Hour}

	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	before := time.Now()
	lines, err := collectLines(t, streamer, newLokiLogContext(nil, nil))
	after := time.Now()
	require.NoError(t, err)
	assert.Empty(t, lines)

	startNs, err := strconv.ParseInt(q["start"][0], 10, 64)
	require.NoError(t, err)
	endNs, err := strconv.ParseInt(q["end"][0], 10, 64)
	require.NoError(t, err)
	assert.True(t, startNs >= before.Add(-2*time.Hour).UnixNano(), "start should be ~now-lookback")
	assert.True(t, startNs <= after.Add(-2*time.Hour).UnixNano(), "start should be ~now-lookback")
	assert.True(t, endNs >= before.UnixNano() && endNs <= after.UnixNano(), "end should be ~now")
}

func TestLokiLogStreamer_ServerErrorIsUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "too many outstanding requests", http.StatusTooManyRequests)
	}))
	defer server.Close()

	streamer, err := NewLokiLogStreamer(newLokiTestConfig(server.URL))
	require.NoError(t, err)

	start := time.Now().Add(-time.Minute)
	_, err = collectLines(t, streamer, newLokiLogContext(&start, nil))
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnavailable, connect.CodeOf(err))
	assert.Contains(t, err.Error(), "429")
}

func TestLokiLogStreamer_CustomQueryTemplate(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query().Get("query")
		_, _ = w.Write(lokiResponse(nil))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.QueryTemplate = `{k8s_namespace_name="{{ .Namespace }}", k8s_pod_name="{{ .PodName }}"} | k8s_container_name="{{ .ContainerName }}"`

	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	start := time.Now().Add(-time.Minute)
	_, err = collectLines(t, streamer, newLokiLogContext(&start, nil))
	require.NoError(t, err)
	assert.Equal(t, `{k8s_namespace_name="flyte", k8s_pod_name="my-pod"} | k8s_container_name="main"`, query)
}

func TestNewLokiLogStreamer_RejectsBadConfig(t *testing.T) {
	cfg := newLokiTestConfig("")
	_, err := NewLokiLogStreamer(cfg)
	assert.Error(t, err, "missing URL")

	cfg = newLokiTestConfig("http://loki:3100")
	cfg.QueryTemplate = "{{ .Unclosed "
	_, err = NewLokiLogStreamer(cfg)
	assert.Error(t, err, "bad template")
}

func TestLokiLogStreamer_MissingPrimaryPodIsNotFound(t *testing.T) {
	streamer, err := NewLokiLogStreamer(newLokiTestConfig("http://loki:3100"))
	require.NoError(t, err)

	_, err = collectLines(t, streamer, &core.LogContext{})
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestLokiLogStreamer_MergesMultipleStreamsInOrder(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := fmt.Sprintf(`{"status":"success","data":{"resultType":"streams","result":[
			{"stream":{"stream":"stderr"},"values":[["%d","err-2"]]},
			{"stream":{"stream":"stdout"},"values":[["%d","out-1"],["%d","out-3"]]}
		]}}`, start.Add(2*time.Second).UnixNano(), start.Add(1*time.Second).UnixNano(), start.Add(3*time.Second).UnixNano())
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	streamer, err := NewLokiLogStreamer(newLokiTestConfig(server.URL))
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)
	require.Len(t, lines, 3)
	assert.Equal(t, []string{"out-1", "err-2", "out-3"}, []string{lines[0].GetMessage(), lines[1].GetMessage(), lines[2].GetMessage()})
}

func TestLokiLogStreamer_KeepsTiedTimestampsAcrossPages(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	t1, t2, t3 := start.Add(1*time.Second), start.Add(2*time.Second), start.Add(3*time.Second)
	all := []lokiEntry{{ts: t1, line: "a"}, {ts: t2, line: "b"}, {ts: t2, line: "c"}, {ts: t2, line: "d"}, {ts: t3, line: "e"}}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		from, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		var entries []lokiEntry
		for _, e := range all {
			if e.ts.UnixNano() >= from && len(entries) < limit {
				entries = append(entries, e)
			}
		}
		_, _ = w.Write(lokiResponse(entries))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.PageSize = 3
	cfg.TimePadding = stdconfig.Duration{}
	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)

	got := make([]string, 0, len(lines))
	for _, l := range lines {
		got = append(got, l.GetMessage())
	}
	assert.Equal(t, []string{"a", "b", "c", "d", "e"}, got)
	assert.LessOrEqual(t, requests, 3)
}

func TestLokiLogStreamer_SkipsPastTimestampGroupLargerThanPage(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	t1, t2 := start.Add(1*time.Second), start.Add(2*time.Second)
	all := []lokiEntry{{ts: t1, line: "a"}, {ts: t1, line: "b"}, {ts: t1, line: "c"}, {ts: t2, line: "d"}}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		from, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		var entries []lokiEntry
		for _, e := range all {
			if e.ts.UnixNano() >= from && len(entries) < limit {
				entries = append(entries, e)
			}
		}
		_, _ = w.Write(lokiResponse(entries))
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.PageSize = 2
	cfg.TimePadding = stdconfig.Duration{}
	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)

	got := make([]string, 0, len(lines))
	for _, l := range lines {
		got = append(got, l.GetMessage())
	}
	// "c" is unreachable with pageSize 2; the streamer must move on rather than loop.
	assert.Equal(t, []string{"a", "b", "d"}, got)
	assert.LessOrEqual(t, requests, 4)
}

func TestLokiLogStreamer_KeepsMessagePayloadVerbatim(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	lokiTS := start.Add(5 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(lokiResponse([]lokiEntry{{ts: lokiTS, line: "2026-11-01T10:00:00Z task failed\n"}}))
	}))
	defer server.Close()

	streamer, err := NewLokiLogStreamer(newLokiTestConfig(server.URL))
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "2026-11-01T10:00:00Z task failed", lines[0].GetMessage())
	assert.True(t, lokiTS.Equal(lines[0].GetTimestamp().AsTime()))
}

func TestLokiLogStreamer_TiedTimestampsAcrossStreamsReorderedBetweenPages(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	t1, t2 := start.Add(1*time.Second), start.Add(2*time.Second)
	type entry struct {
		lokiEntry
		stream string
	}
	all := []entry{
		{lokiEntry{ts: t1, line: "a"}, "x"},
		{lokiEntry{ts: t2, line: "b"}, "y"},
		{lokiEntry{ts: t2, line: "c"}, "x"},
	}

	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		from, _ := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		byStream := map[string][][]string{}
		n := 0
		for _, e := range all {
			if e.ts.UnixNano() >= from && n < limit {
				byStream[e.stream] = append(byStream[e.stream], []string{strconv.FormatInt(e.ts.UnixNano(), 10), e.line})
				n++
			}
		}
		// Loki does not guarantee stream order between queries; flip it each time.
		order := []string{"y", "x"}
		if requests%2 == 0 {
			order = []string{"x", "y"}
		}
		var result []map[string]any
		for _, s := range order {
			if vals, ok := byStream[s]; ok {
				result = append(result, map[string]any{"stream": map[string]string{"stream": s}, "values": vals})
			}
		}
		raw, _ := json.Marshal(map[string]any{"status": "success", "data": map[string]any{"resultType": "streams", "result": result}})
		_, _ = w.Write(raw)
	}))
	defer server.Close()

	cfg := newLokiTestConfig(server.URL)
	cfg.PageSize = 2
	cfg.TimePadding = stdconfig.Duration{}
	streamer, err := NewLokiLogStreamer(cfg)
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)

	got := make([]string, 0, len(lines))
	for _, l := range lines {
		got = append(got, l.GetMessage())
	}
	assert.ElementsMatch(t, []string{"a", "b", "c"}, got)
}

func TestLokiLogStreamer_AcceptsStructuredMetadata(t *testing.T) {
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	ts := strconv.FormatInt(start.Add(time.Second).UnixNano(), 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"status":"success","data":{"resultType":"streams","result":[{"stream":{"pod":"my-pod"},"values":[[%q,"hello",{"trace_id":"abc"}]]}]}}`, ts)
	}))
	defer server.Close()

	streamer, err := NewLokiLogStreamer(newLokiTestConfig(server.URL))
	require.NoError(t, err)

	lines, err := collectLines(t, streamer, newLokiLogContext(&start, &end))
	require.NoError(t, err)
	require.Len(t, lines, 1)
	assert.Equal(t, "hello", lines[0].GetMessage())
}
