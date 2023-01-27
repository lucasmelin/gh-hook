package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func Input(password bool, prompt string) (string, error) {
	i := textinput.New()
	i.Focus()
	i.Prompt = prompt
	i.Placeholder = "Type something..."
	i.Width = 50
	i.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	i.CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	i.CharLimit = 0

	if password {
		i.EchoMode = textinput.EchoPassword
		i.EchoCharacter = '•'
	}

	p := tea.NewProgram(inputModel{
		textinput:   i,
		cancelled:   false,
		header:      "",
		headerStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}, tea.WithOutput(os.Stderr))
	tm, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run input: %w", err)
	}
	m := tm.(inputModel)

	if m.cancelled {
		return "", fmt.Errorf("cancelled")
	}

	return m.textinput.Value(), nil
}

type inputModel struct {
	header      string
	headerStyle lipgloss.Style
	textinput   textinput.Model
	quitting    bool
	cancelled   bool
}

func (m inputModel) Init() tea.Cmd { return textinput.Blink }
func (m inputModel) View() string {
	if m.quitting {
		return ""
	}

	if m.header != "" {
		header := m.headerStyle.Render(m.header)
		return lipgloss.JoinVertical(lipgloss.Left, header, m.textinput.View())
	}

	return m.textinput.View()
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}
