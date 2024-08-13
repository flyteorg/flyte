package bubbletea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/flyteorg/flyte/flytectl/pkg/filters"
	"github.com/flyteorg/flyte/flytectl/pkg/printer"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type DataCallback func(filter filters.Filters) ([]proto.Message, error)

type printTableProto struct{ proto.Message }

type direction int

type newDataMsg struct {
	newItems       []proto.Message
	batchIndex     int
	fetchDirection direction
}

const (
	forward direction = iota
	backward
)

const (
	msgPerBatch       = 100 // Please set msgPerBatch as a multiple of msgPerPage
	msgPerPage        = 10
	pagePerBatch      = msgPerBatch / msgPerPage
	prefetchThreshold = pagePerBatch - 1
	localBatchLimit   = 10 // Please set localBatchLimit at least 2
)

var (
	// Callback function used to fetch data from the module that called bubbletea pagination.
	callback   DataCallback
	listHeader []printer.Column
	filter     filters.Filters
	// Record the index of the first and last batch that is in cache
	firstBatchIndex int
	lastBatchIndex  int
	// Record numbers of messages in each batch
	batchLen = make(map[int]int)
	// Avoid fetching back and forward at the same time
	mutex sync.Mutex
	// Used to catch error happened while running paginator
	errMsg error = nil
)

func (p printTableProto) MarshalJSON() ([]byte, error) {
	marshaller := jsonpb.Marshaler{Indent: "\t"}
	buf := new(bytes.Buffer)
	err := marshaller.Marshal(buf, p.Message)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func _max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func _min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getSliceBounds(m *pageModel) (start int, end int) {
	start = (m.paginator.Page - firstBatchIndex*pagePerBatch) * msgPerPage
	end = _min(start+msgPerPage, len(*m.items))
	return start, end
}

func getTable(m *pageModel) (string, error) {
	start, end := getSliceBounds(m)
	curShowMessage := (*m.items)[start:end]
	printTableMessages := make([]*printTableProto, 0, len(curShowMessage))
	for _, m := range curShowMessage {
		printTableMessages = append(printTableMessages, &printTableProto{Message: m})
	}

	jsonRows, err := json.Marshal(printTableMessages)
	if err != nil {
		return "", fmt.Errorf("failed to marshal proto messages")
	}

	var buf strings.Builder
	p := printer.Printer{}
	if err := p.JSONToTable(&buf, jsonRows, listHeader); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func getMessageList(batchIndex int) ([]proto.Message, error) {
	mutex.Lock()
	spin = true
	defer func() {
		spin = false
		mutex.Unlock()
	}()

	msg, err := callback(filters.Filters{
		Limit:  msgPerBatch,
		Page:   int32(batchIndex + 1),
		SortBy: filter.SortBy,
		Asc:    filter.Asc,
	})
	if err != nil {
		return nil, err
	}

	batchLen[batchIndex] = len(msg)

	return msg, nil
}

func fetchDataCmd(batchIndex int, fetchDirection direction) tea.Cmd {
	return func() tea.Msg {
		newItems, err := getMessageList(batchIndex)
		if err != nil {
			errMsg = err
			return err
		}
		msg := newDataMsg{
			newItems:       newItems,
			batchIndex:     batchIndex,
			fetchDirection: fetchDirection,
		}
		return msg
	}
}

func getLastMsgIdx() int {
	sum := 0
	for i := 0; i < lastBatchIndex+1; i++ {
		length, ok := batchLen[i]
		if ok {
			sum += length
		} else {
			sum += msgPerBatch
		}
	}
	return sum
}
