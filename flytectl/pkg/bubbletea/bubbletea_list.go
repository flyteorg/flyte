package bubbletea

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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

type viewType int

const (
	listView viewType = iota
	inputView
)

func initListModel() listModel {
	ti := textinput.New()
	ti.Placeholder = "Type in here"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return listModel{
		view:      listView,
		textInput: ti,
	}
}

type listModel struct {
	quitting   bool
	view       viewType
	list       list.Model
	textInput  textinput.Model
	inputTitle string
}

func (m listModel) Init() tea.Cmd {
	return textinput.Blink
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

		case "enter":
			var err error
			if m.view == inputView {
				m.view = listView
				args = append(args, m.textInput.Value())
				m.textInput.SetValue("")
				err = genListModel(&m, "")
			} else if m.view == listView {
				item, _ := m.list.SelectedItem().(item)
				args = append(args, string(item))
				err = genListModel(&m, string(item))
			}
			if err != nil || m.quitting {
				return m, tea.Quit
			}
			return m, nil

		}
	}

	var cmd tea.Cmd
	if m.view == inputView {
		m.textInput, cmd = m.textInput.Update(msg)
	} else if m.view == listView {
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m listModel) View() string {
	if m.quitting {
		return quitTextStyle.Render("")
	}

	if m.view == inputView {
		return fmt.Sprintf(
			m.inputTitle+"\n\n%s\n\n%s",
			m.textInput.View(),
			"(press q to quit)",
		) + "\n"
	}

	return "\n" + m.list.View()
}

func genList(items []list.Item, title string) list.Model {
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

	currentCmd, run, err := ifRunBubbleTea()
	if err != nil || !run {
		return
	}

	InitCommandFlagMap()

	cmdName := strings.Fields(currentCmd.Use)[0]
	nameToCommand[cmdName] = Command{
		Cmd:   currentCmd,
		Name:  cmdName,
		Short: currentCmd.Short,
	}

	m := initListModel()
	if err := genListModel(&m, cmdName); err != nil {
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

	// fmt.Println(append(newArgs, existingFlags...))
	// Originally existed flags need to be append at last, so if any user input is wrong, it can be caught in the main logic
	rootCmd.SetArgs(append(args, existingFlags...))
}
