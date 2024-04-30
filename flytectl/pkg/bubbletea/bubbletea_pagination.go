package bubbletea

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/golang/protobuf/proto"

	tea "github.com/charmbracelet/bubbletea"
)

type pageModel struct {
	items     []proto.Message
	paginator paginator.Model
}

func newModel(initMsg []proto.Message) pageModel {
	p := paginator.New()
	p.PerPage = msgPerPage
	p.SetTotalPages(len(initMsg))

	return pageModel{
		paginator: p,
		items:     initMsg,
	}
}

func (m pageModel) Init() tea.Cmd {
	return nil
}

func (m pageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	m.paginator, cmd = m.paginator.Update(msg)
	preFetchBatch(&m)
	return m, cmd
}

func (m pageModel) View() string {
	var b strings.Builder
	table, err := getTable(&m)
	if err != nil {
		return ""
	}
	b.WriteString(table)
	b.WriteString(fmt.Sprintf("  PAGE - %d\n", m.paginator.Page+1))
	b.WriteString("\n\n  h/l ←/→ page • q: quit\n")
	return b.String()
}

func Paginator(_listHeader []printer.Column, _callback DataCallback) {
	listHeader = _listHeader
	callback = _callback

	var msg []proto.Message
	for i := firstBatchIndex; i < lastBatchIndex+1; i++ {
		msg = append(msg, getMessageList(i)...)
	}

	p := tea.NewProgram(newModel(msg))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
