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

	noStyle       = lipgloss.NewStyle()
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
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
	quitting          bool
	curView           viewType
	list              list.Model
	textInputs        []textinput.Model
	pendingInputFlags []string
	focusIndex        int
	textInputTitle    string
}

type viewType int

const (
	listView viewType = iota
	inputView
)

func makeTextInputModel(flagList []string) []textinput.Model {
	inputs := make([]textinput.Model, len(flagList))

	var t textinput.Model
	for i := range inputs {
		t = textinput.New()
		t.CharLimit = 32

		t.Placeholder = flagList[i]
		if i == 0 {
			t.Focus()
		}
		t.PromptStyle = focusedStyle
		t.TextStyle = focusedStyle
		inputs[i] = t
	}

	return inputs
}

func initListModel() listModel {

	return listModel{
		curView: listView,
	}
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
		}
	}

	var cmd tea.Cmd
	var updatedModel tea.Model
	if m.curView == inputView {
		updatedModel, cmd = m.textInputUpdate(msg)
	} else if m.curView == listView {
		updatedModel, cmd = m.listUpdate(msg)
	}
	m = updatedModel.(listModel)

	return m, cmd
}

func (m listModel) listUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			item, _ := m.list.SelectedItem().(item)
			args = append(args, string(item))
			err := makeListModel(&m, string(item))
			if err != nil || m.quitting {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) textInputUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		// Set focus to next input
		case tea.KeyEnter, tea.KeyUp, tea.KeyDown:
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			// If so, exit. //TODO save to args
			if s == "enter" && m.focusIndex == len(m.textInputs) {
				m.curView = listView
				for i := range m.pendingInputFlags {
					args = append(args, m.pendingInputFlags[i])
					args = append(args, m.textInputs[i].Value())
				}
				err := makeListModel(&m, "")
				if err != nil || m.quitting {
					return m, tea.Quit
				}
				return m, nil
			}

			// Cycle indexes
			if s == "up" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.textInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.textInputs)
			}

			cmds := make([]tea.Cmd, len(m.textInputs))
			for i := 0; i <= len(m.textInputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.textInputs[i].Focus()
					m.textInputs[i].PromptStyle = focusedStyle
					m.textInputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.textInputs[i].Blur()
				m.textInputs[i].PromptStyle = noStyle
				m.textInputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *listModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m listModel) View() string {
	if m.quitting {
		return quitTextStyle.Render("")
	}

	if m.curView == inputView {
		var b strings.Builder
		b.WriteString(m.textInputTitle + "\n\n")
		for i := range m.textInputs {
			b.WriteString(m.textInputs[i].View())
			if i < len(m.textInputs)-1 {
				b.WriteRune('\n')
			}
		}

		button := &blurredButton
		if m.focusIndex == len(m.textInputs) {
			button = &focusedButton
		}
		fmt.Fprintf(&b, "\n\n%s\n\n", *button)

		return b.String()
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

	fmt.Println(append(args, existingFlags...))
	// Originally existed flags need to be append at last, so if any user input is wrong, it can be caught in the main logic
	rootCmd.SetArgs(append(args, existingFlags...))
}
