package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type errMsg error

type model struct {
	spinner    spinner.Model
	textinput  textinput.Model
	choice     int
	chosen     bool
	quitting   bool
	listOffset int
	err        error
}

type task struct {
	title string
	desc  string
	done  bool
}

type termSize struct {
	width  int
	height int
}

func (t task) Title() string       { return t.title }
func (t task) Description() string { return t.desc }
func (t task) FilterValue() string { return t.title }
func (t task) IsDone() bool        { return t.done }

var rawItems = []task{
	task{title: "task 1", desc: "penis", done: false},
	task{title: "task 2", desc: "penis2", done: false},
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var tabKeys = key.NewBinding(
	key.WithKeys("tab"),
	key.WithHelp("", "press tab to switch contexts"),
)

var enterKey = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "add task"),
)

var escKey = key.NewBinding(
	key.WithKeys("esc"),
)

var size = termSize{width: 0, height: 0}

var defaultStyles = list.NewDefaultItemStyles()
var activeForeground = lipgloss.NewStyle().Bold(true).Foreground(defaultStyles.SelectedTitle.GetForeground())
var titleForeground = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("#7D56F4"))

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	t := textinput.New()
	t.Placeholder = "enter a task ..."
	return model{spinner: s, textinput: t}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tea.EnterAltScreen)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit
		}
		if m.textinput.Focused() {
			if key.Matches(msg, enterKey) {
				rawItems = append(rawItems, task{title: m.textinput.Value(), desc: ""})
				m.textinput.Reset()
			} else if key.Matches(msg, escKey) {
				m.textinput.Blur()
			} else {
				m.textinput, cmd = m.textinput.Update(msg)
				cmds = append(cmds, cmd)
			}
			break
		}
		switch msg.String() {
		case "j", "down":
			m.choice++
			if m.choice >= len(rawItems) {
				m.choice = len(rawItems) - 1
			}
			if m.choice >= size.height-8+m.listOffset && m.choice < len(rawItems)-1 {
				m.listOffset += 1
			}
		case "k", "up":
			m.choice--
			if m.choice < 0 {
				m.choice = 0
			}
			if m.choice < m.listOffset && m.listOffset > 0 {
				m.listOffset -= 1
			}
		case "r", "backspace":
			if m.choice == len(rawItems)-1 {
				rawItems = rawItems[:m.choice]
				if m.choice > 0 {
					m.choice -= 1
				}
			} else {
				rawItems = append(rawItems[:m.choice], rawItems[m.choice+1:]...)
			}
		case "enter":
			m.chosen = true
			rawItems[m.choice].done = !rawItems[m.choice].done
		case "esc":
			m.textinput.Focus()
		}
		break
	case tea.WindowSizeMsg:
		size.width = msg.Width
		size.height = msg.Height
		m.listOffset = 0
		m.choice = 0
		break
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

var term = termenv.EnvColorProfile()

func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

var subtle = makeFgStyle("241")

var dot = colorFg(" • ", "236")

func choice(subject task, active bool) string {
	var res string
	cursor := " "
	if active {
		cursor = "⭢"
	}
	checked := "[ ]"
	if subject.IsDone() {
		checked = "[✓]"
	}
	if active {
		res = fmt.Sprintf("%s %s %s", cursor, checked, subject.Title())
		res = activeForeground.Render(res) + "\n"
	} else {
		res = fmt.Sprintf("%s %s %s\n", cursor, checked, subject.Title())
	}
	return res
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	str := "  " + titleForeground.Render("TODOs") + "\n\n"
	str += m.textinput.View() + "\n\n"
	for i := m.listOffset; i < len(rawItems)+m.listOffset; i++ {
		if i-m.listOffset == size.height-7 {
			break
		}
		str += choice(rawItems[i], i == m.choice)
	}
	for i := 0; i < size.height-len(rawItems)-7; i++ { // -7 because of the title, input field and margin
		str += "\n"
	}
	if m.listOffset+(size.height-7) >= len(rawItems) {
		str += "\n"
	} else {
		str += subtle("  ...\n")
	}
	str += subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("r: remove") + dot + subtle("q : quit")

	return str
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
