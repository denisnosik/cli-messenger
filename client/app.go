package client

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appModel struct {
	command string
	err     error
}

type commandList map[string]func(cfg *config)

var (
	appWelcomeStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("2")).
			Bold(true)

	appCommandListStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Foreground(lipgloss.Color("15")).
				Bold(true)

	appLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginLeft(2)

	appCommandInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("255")).
				Padding(0, 1).
				MarginLeft(2).
				Width(40)
)

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// case "enter":
		// 	if m.command == "" {
		// 		m.err = fmt.Errorf("command cannot be empty")
		// 		return m, nil
		// 	}
		// 	return m, m.runCommand

		case "backspace":
			runes := []rune(m.command)
			if len(runes) > 0 {
				m.command = string(runes[:len(runes)-1])
			}
		default:
			m.err = nil
			if len(msg.String()) == 1 {
				m.command += msg.String()
			}
		}
	}

	return m, nil
}

func (m appModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(appWelcomeStyle.Render(logo) + "\n\n")

	b.WriteString(appCommandListStyle.Render())

	b.WriteString(appLabelStyle.Render("Enter command...") + "\n")

	commandBox := appCommandInputStyle.Render(m.command + "█")
	b.WriteString(commandBox + "\n\n")

	return b.String()
}

func startApp() {
	if _, err := tea.NewProgram(appModel{}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// func (m appModel) runCommand() tea.Msg {

// }
