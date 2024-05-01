package bubbletea

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flyteorg/flytectl/pkg/filters"
	"github.com/flyteorg/flytectl/pkg/printer"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type DataCallback func(filter filters.Filters) []proto.Message

type PrintableProto struct{ proto.Message }

const (
	msgPerBatch  = 100 // Please set msgPerBatch as a multiple of msgPerPage
	msgPerPage   = 10
	pagePerBatch = msgPerBatch / msgPerPage
)

var (
	// Used for indexing local stored rows
	localPageIndex int
	// Recording batch index fetched from admin
	firstBatchIndex int32 = 1
	lastBatchIndex  int32 = 10
	batchLen              = make(map[int32]int)
	// Callback function used to fetch data from the module that called bubbletea pagination.
	callback DataCallback
	// The header of the table
	listHeader []printer.Column

	marshaller = jsonpb.Marshaler{
		Indent: "\t",
	}
)

func (p PrintableProto) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := marshaller.Marshal(buf, p.Message)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getSliceBounds(idx int, length int) (start int, end int) {
	start = idx * msgPerPage
	end = min(idx*msgPerPage+msgPerPage, length)
	return start, end
}

func getTable(m *pageModel) (string, error) {
	start, end := getSliceBounds(localPageIndex, len(m.items))
	curShowMessage := m.items[start:end]
	printableMessages := make([]*PrintableProto, 0, len(curShowMessage))
	for _, m := range curShowMessage {
		printableMessages = append(printableMessages, &PrintableProto{Message: m})
	}

	jsonRows, err := json.Marshal(printableMessages)
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

func getMessageList(batchIndex int32) []proto.Message {
	msg := callback(filters.Filters{
		Limit:  msgPerBatch,
		Page:   batchIndex,
		SortBy: "created_at",
		Asc:    false,
	})
	batchLen[batchIndex] = len(msg)

	return msg
}

func countTotalPages() int {
	sum := 0
	for _, l := range batchLen {
		sum += l
	}
	return sum
}

// Only (lastBatchIndex-firstBatchIndex)*msgPerBatch of rows are stored in local memory.
// When user tries to get rows out of this range, this function will be triggered.
func preFetchBatch(m *pageModel) {
	localPageIndex = m.paginator.Page - int(firstBatchIndex-1)*pagePerBatch

	// Triggers when user is at the last local page
	if localPageIndex+1 == len(m.items)/msgPerPage {
		newMessages := getMessageList(lastBatchIndex + 1)
		m.paginator.SetTotalPages(countTotalPages())
		if len(newMessages) != 0 {
			lastBatchIndex++
			m.items = append(m.items, newMessages...)
			m.items = m.items[batchLen[firstBatchIndex]:] // delete the msgs in the "firstBatchIndex" batch
			localPageIndex -= batchLen[firstBatchIndex] / msgPerPage
			firstBatchIndex++
		}
		return
	}
	// Triggers when user is at the first local page
	if localPageIndex == 0 && firstBatchIndex > 1 {
		newMessages := getMessageList(firstBatchIndex - 1)
		m.paginator.SetTotalPages(countTotalPages())
		firstBatchIndex--
		m.items = append(newMessages, m.items...)
		m.items = m.items[:len(m.items)-batchLen[lastBatchIndex]] // delete the msgs in the "lastBatchIndex" batch
		localPageIndex += batchLen[firstBatchIndex] / msgPerPage
		lastBatchIndex--
		return
	}
}
