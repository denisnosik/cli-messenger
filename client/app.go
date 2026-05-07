package client

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appModel struct {
	input string
	err   error
}

type command struct {
	name        string
	description string
	callback    tea.Cmd
}

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

	appInputStyle = lipgloss.NewStyle().
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
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
			}
		default:
			m.err = nil
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

func (m appModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(appWelcomeStyle.Render(logo) + "\n\n")

	b.WriteString(appLabelStyle.Render("Available commands") + "\n")

	b.WriteString("\n")

	for _, command := range getCommands() {
		b.WriteString(appCommandListStyle.Render(fmt.Sprintf("  %-30s %s", command.name, command.description)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	b.WriteString(appLabelStyle.Render("Enter command") + "\n")

	inputBox := appInputStyle.Render("> " + m.input + "█")
	b.WriteString(inputBox + "\n\n")

	b.WriteString(appLabelStyle.Render("enter — submit  | esc, ctrl+c — quit") + "\n")

	return b.String()
}

func startApp() {
	if _, err := tea.NewProgram(appModel{}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getCommands() []command {
	return []command{
		{
			name:        "register",
			description: "Create a new account",
			//callback:    commandRegister,
		},
		{
			name:        "login",
			description: "Sign in with you nickname and password",
			//callback:    commandLogin,
		},
		{
			name:        "chat <nickname>",
			description: "Open a chat with a friend",
			//callback:    commandChat,
		},
		{
			name:        "friends <nickname>",
			description: "Send a friend request (> friends --help for more)",
			//callback:    commandFriends,
		},
		{
			name:        "notifications",
			description: "View unread messages and friend requests",
			//callback:    commandNotifications,
		},
	}
}

// func (m appModel) runCommand() tea.Msg {

// }
