package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type errMsg error

type model struct {
	spinner  spinner.Model
	list     list.Model
	quitting bool
	err      error
}

type task struct {
	title string
}

func (t task) FilterValue() string { return t.title }

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		break
	case spinner.TickMsg:
		newSpinner, spinner_cmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		cmds = append(cmds, spinner_cmd)
		break
	case errMsg:
		m.err = msg
		return m, nil
	}

	//newList, list_cmd := m.list.Update(msg)
	//m.list = newList
	//cmds = append(cmds, list_cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	return docStyle.Render(m.list.View())

	/*str := fmt.Sprintf("\n\n   %s Loading forever... %s\n\n", m.spinner.View(), quitKeys.Help().Desc)
	if m.quitting {
		return str + "\n"
	}
	return str*/
}

func main() {
	initialItems := []list.Item{
		task{title: "task 1"},
		task{title: "task 2"},
	}
	m := model{list: list.New(initialItems, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "TODOs"

	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
