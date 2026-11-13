package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

const (
	lokiQueryRangePath   = "/loki/api/v1/query_range"
	lokiTenantHeader     = "X-Scope-OrgID"
	lokiMaxResponseBytes = 64 << 20
)

// LokiLogStreamer reads historical task logs from Grafana Loki.
type LokiLogStreamer struct {
	cfg      config.LokiConfig
	client   *http.Client
	queryTpl *template.Template
	baseURL  *url.URL
}

var _ LogStreamer = (*LokiLogStreamer)(nil)

type lokiQueryInput struct {
	Namespace     string
	PodName       string
	ContainerName string
}

func NewLokiLogStreamer(cfg config.LokiConfig) (*LokiLogStreamer, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, fmt.Errorf("loki: url is required")
	}
	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("loki: invalid url %q: %w", cfg.URL, err)
	}
	if cfg.QueryTemplate == "" {
		return nil, fmt.Errorf("loki: queryTemplate is required")
	}
	tpl, err := template.New("loki-query").Option("missingkey=error").Parse(cfg.QueryTemplate)
	if err != nil {
		return nil, fmt.Errorf("loki: invalid queryTemplate: %w", err)
	}
	if cfg.PageSize <= 0 {
		return nil, fmt.Errorf("loki: pageSize must be positive")
	}
	if cfg.BearerToken, err = resolveLokiSecret("bearerToken", cfg.BearerToken, cfg.BearerTokenPath); err != nil {
		return nil, err
	}
	if cfg.BasicAuth.Password, err = resolveLokiSecret("basicAuth.password", cfg.BasicAuth.Password, cfg.BasicAuth.PasswordPath); err != nil {
		return nil, err
	}
	return &LokiLogStreamer{
		cfg:      cfg,
		client:   &http.Client{Timeout: cfg.Timeout.Duration},
		queryTpl: tpl,
		baseURL:  baseURL,
	}, nil
}

func resolveLokiSecret(name, value, path string) (string, error) {
	if path == "" {
		return value, nil
	}
	if value != "" {
		return "", fmt.Errorf("loki: %s and %sPath are mutually exclusive", name, name)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("loki: reading %sPath: %w", name, err)
	}
	return strings.TrimSpace(string(raw)), nil
}

func (s *LokiLogStreamer) TailLogs(ctx context.Context, logContext *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	return s.fetch(ctx, logContext, func(lines []*dataplane.LogLine) error {
		return stream.Send(&dataproxy.TailLogsResponse{
			Logs: []*dataproxy.TailLogsResponse_Logs{{Lines: lines}},
		})
	})
}

func (s *LokiLogStreamer) fetch(ctx context.Context, logContext *core.LogContext, send func([]*dataplane.LogLine) error) error {
	pod, container, err := GetPrimaryPodAndContainer(logContext)
	if err != nil {
		return connect.NewError(connect.CodeNotFound, err)
	}

	query, err := s.renderQuery(pod, container)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	start, end := s.queryWindow(container)
	logger.Debugf(ctx, "loki: querying %q for pod %s/%s container %s from %s to %s",
		query, pod.GetNamespace(), pod.GetPodName(), container.GetContainerName(), start, end)

	// Loki's start is inclusive, so each page is re-queried from the last
	// timestamp sent. Lines already emitted at that instant are dropped by
	// message (with multiplicity), since Loki does not order tied entries
	// consistently across queries.
	sent := 0
	// An empty result is NotFound so that the next streamer is tried.
	done := func() error {
		if sent == 0 {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("loki: no logs for pod %s/%s container %s",
				pod.GetNamespace(), pod.GetPodName(), container.GetContainerName()))
		}
		return nil
	}
	var lastTS time.Time
	seen := map[string]int{}
	seenTotal := 0
	for {
		remaining := s.cfg.MaxLines - sent
		if s.cfg.MaxLines <= 0 {
			remaining = s.cfg.PageSize
		}
		if remaining <= 0 {
			return done()
		}
		limit := min(s.cfg.PageSize, remaining+seenTotal)

		lines, err := s.queryRange(ctx, query, start, end, limit)
		if err != nil {
			return err
		}
		full := len(lines) == limit

		page := make([]*dataplane.LogLine, 0, len(lines))
		consumed := 0
		for _, line := range lines {
			if len(page) == remaining {
				break
			}
			consumed++
			if line.GetTimestamp().AsTime().Equal(lastTS) && seen[line.GetMessage()] > 0 {
				seen[line.GetMessage()]--
				continue
			}
			page = append(page, line)
		}
		if len(page) > 0 {
			if err := send(page); err != nil {
				return err
			}
			sent += len(page)
			lastTS = page[len(page)-1].GetTimestamp().AsTime()
			clear(seen)
			seenTotal = 0
			for i := consumed - 1; i >= 0 && lines[i].GetTimestamp().AsTime().Equal(lastTS); i-- {
				seen[lines[i].GetMessage()]++
				seenTotal++
			}
		}
		if !full {
			return done()
		}

		// A page's worth of lines at one timestamp cannot be paged through
		// with an inclusive start; move past it.
		if len(page) == 0 || seenTotal >= s.cfg.PageSize {
			logger.Warnf(ctx, "loki: %d entries share timestamp %s for pod %s/%s; any beyond pageSize are skipped",
				seenTotal, lastTS, pod.GetNamespace(), pod.GetPodName())
			start = lastTS.Add(time.Nanosecond)
			clear(seen)
			seenTotal = 0
			if !start.Before(end) {
				return done()
			}
			continue
		}
		start = lastTS
	}
}

func (s *LokiLogStreamer) renderQuery(pod *core.PodLogContext, container *core.ContainerContext) (string, error) {
	var buf bytes.Buffer
	err := s.queryTpl.Execute(&buf, lokiQueryInput{
		Namespace:     pod.GetNamespace(),
		PodName:       pod.GetPodName(),
		ContainerName: container.GetContainerName(),
	})
	if err != nil {
		return "", fmt.Errorf("loki: rendering query template: %w", err)
	}
	return buf.String(), nil
}

func (s *LokiLogStreamer) queryWindow(container *core.ContainerContext) (time.Time, time.Time) {
	now := time.Now()
	padding := s.cfg.TimePadding.Duration

	process := container.GetProcess()
	if process.GetContainerStartTime() == nil {
		return now.Add(-s.cfg.Lookback.Duration), now
	}

	start := process.GetContainerStartTime().AsTime().Add(-padding)
	end := now
	if process.GetContainerEndTime() != nil {
		end = process.GetContainerEndTime().AsTime().Add(padding)
		if end.After(now) {
			end = now
		}
	}
	return start, end
}

type lokiQueryRangeResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			// Each value is [timestamp, line] with an optional trailing
			// structured-metadata object.
			Values [][]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func (s *LokiLogStreamer) queryRange(ctx context.Context, query string, start, end time.Time, limit int) ([]*dataplane.LogLine, error) {
	u := *s.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + lokiQueryRangePath
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.UnixNano(), 10))
	params.Set("end", strconv.FormatInt(end.UnixNano(), 10))
	params.Set("limit", strconv.Itoa(limit))
	params.Set("direction", "forward")
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("loki: building request: %w", err))
	}
	s.setAuth(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("loki: request failed: %w", err))
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(io.LimitReader(resp.Body, lokiMaxResponseBytes))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("loki: reading response: %w", err))
	}
	if resp.StatusCode != http.StatusOK {
		return nil, connect.NewError(connect.CodeUnavailable,
			fmt.Errorf("loki: query_range returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body))))
	}

	var parsed lokiQueryRangeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("loki: decoding response: %w", err))
	}
	if parsed.Status != "success" {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("loki: query status %q", parsed.Status))
	}
	if parsed.Data.ResultType != "" && parsed.Data.ResultType != "streams" {
		return nil, connect.NewError(connect.CodeInternal,
			fmt.Errorf("loki: unexpected result type %q; the query must select log streams, not a metric", parsed.Data.ResultType))
	}

	lines := make([]*dataplane.LogLine, 0, limit)
	for _, stream := range parsed.Data.Result {
		for _, value := range stream.Values {
			if len(value) < 2 {
				continue
			}
			var tsRaw, line string
			if err := json.Unmarshal(value[0], &tsRaw); err != nil {
				logger.Warnf(ctx, "loki: skipping entry with bad timestamp %s: %v", value[0], err)
				continue
			}
			if err := json.Unmarshal(value[1], &line); err != nil {
				logger.Warnf(ctx, "loki: skipping entry with bad line at %s: %v", tsRaw, err)
				continue
			}
			ns, err := strconv.ParseInt(tsRaw, 10, 64)
			if err != nil {
				logger.Warnf(ctx, "loki: skipping entry with bad timestamp %q: %v", tsRaw, err)
				continue
			}
			lines = append(lines, toLogLine(time.Unix(0, ns), line))
		}
	}
	// Entries are ordered per stream; merging several streams needs a sort.
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].GetTimestamp().AsTime().Before(lines[j].GetTimestamp().AsTime())
	})
	if len(lines) > limit {
		lines = lines[:limit]
	}
	return lines, nil
}

func (s *LokiLogStreamer) setAuth(req *http.Request) {
	if s.cfg.TenantID != "" {
		req.Header.Set(lokiTenantHeader, s.cfg.TenantID)
	}
	switch {
	case s.cfg.BearerToken != "":
		req.Header.Set("Authorization", "Bearer "+s.cfg.BearerToken)
	case s.cfg.BasicAuth.Username != "" || s.cfg.BasicAuth.Password != "":
		req.SetBasicAuth(s.cfg.BasicAuth.Username, s.cfg.BasicAuth.Password)
	}
}

func toLogLine(ts time.Time, line string) *dataplane.LogLine {
	return &dataplane.LogLine{
		Originator: dataplane.LogLineOriginator_USER,
		Timestamp:  timestamppb.New(ts),
		Message:    strings.TrimRight(line, "\r\n"),
	}
}
