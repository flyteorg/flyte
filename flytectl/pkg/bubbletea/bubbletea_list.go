package bubbletea

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const (
	listHeight   = 17
	defaultWidth = 40
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(0, 0, 0, 0)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := string(i)

	fn := itemStyle.Render

	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type listModel struct {
	quitting bool
	list     list.Model
}

func (m listModel) Init() tea.Cmd {
	return nil
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	var updatedModel tea.Model
	updatedModel, cmd = m.listUpdate(msg)
	m = updatedModel.(listModel)

	return m, cmd
}

func (m listModel) listUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			item, _ := m.list.SelectedItem().(item)
			curArgs = append(curArgs, string(item))
			err := makeListModel(&m, string(item))
			if err != nil || m.quitting {
				listErrMsg = err
				return m, tea.Quit
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) View() string {
	if m.quitting {
		return quitTextStyle.Render("")
	}
	return "\n" + m.list.View()
}

func makeList(items []list.Item, title string) list.Model {
	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	if title != "" {
		l.Title = title
		l.SetShowTitle(true)
		l.Styles.Title = titleStyle
	}
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return l
}

func ShowCmdList(_rootCmd *cobra.Command) {
	rootCmd = _rootCmd

	currentCmd, run, err := checkRunBubbleTea()
	if err != nil || !run {
		return
	}

	initCommandFlagMap()
	err = initCmdCtx()
	if err != nil {
		return
	}

	cmdName := strings.Fields(currentCmd.Use)[0]
	commandMap[cmdName] = Command{
		Cmd:   currentCmd,
		Name:  cmdName,
		Short: currentCmd.Short,
	}

	m := listModel{}
	if err := makeListModel(&m, cmdName); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if !m.quitting {
		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}

	if listErrMsg != nil {
		fmt.Println(listErrMsg)
		os.Exit(1)
	}

	// Originally existed flags need to be append at last, so if any user input is wrong, it can be caught in the main logic
	rootCmd.SetArgs(append(curArgs, existingFlags...))
}
