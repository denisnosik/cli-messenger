package client

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appModel struct {
	cfg           *config
	input         string
	output        string
	notifMessages []string
	notifFriends  []string
	err           error
}

type command struct {
	name        string
	description string
	callback    func(m appModel) (appModel, tea.Cmd)
}

type appErrMsg struct{ err error }

var (
	appLogoStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("2")).
			Bold(true)

	appWelcomeStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("2")).
			Bold(true)

	appCommandListStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Foreground(lipgloss.Color("15")).
				Bold(true)

	appOutputStyle = lipgloss.NewStyle().
			MarginLeft(2).
			Foreground(lipgloss.Color("15"))

	appLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginLeft(2)

	appInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("255")).
			Padding(0, 1).
			MarginLeft(2).
			Width(40)

	appErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			MarginLeft(2)
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

		case "enter":
			if m.input == "" {
				m.err = fmt.Errorf("command cannot be empty")
				return m, nil
			}

			cleanedInput := strings.Fields(strings.ToLower(m.input))
			commandInput := cleanedInput[0]

			for _, cmd := range getCommands() {
				if commandInput == cmd.name {
					return cmd.callback(m)
				}
			}

			m.err = fmt.Errorf("unknown command")
			return m, nil

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
	b.WriteString(appLogoStyle.Render(logo) + "\n\n")

	b.WriteString(appLabelStyle.Render("Available commands") + "\n")

	b.WriteString("\n")

	for _, command := range getCommands() {
		b.WriteString(appCommandListStyle.Render(fmt.Sprintf("  %-30s %s", command.name, command.description)))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if m.cfg.currentUser.Token != "" {
		b.WriteString(appWelcomeStyle.Render(fmt.Sprintf("Hello, %s!", m.cfg.currentUser.Nickname)))
		b.WriteString("\n\n")
	}

	if m.output != "" {
		b.WriteString(appOutputStyle.Render(m.output))
		b.WriteString("\n\n")
	}

	if len(m.notifMessages) > 0 {
		for _, notif := range m.notifMessages {
			b.WriteString(appOutputStyle.Render(notif))
			b.WriteString("\n")
		}
	}

	if len(m.notifFriends) > 0 {
		for _, notif := range m.notifFriends {
			b.WriteString(appOutputStyle.Render(notif))
			b.WriteString("\n")
		}
	}

	b.WriteString(appLabelStyle.Render("Enter command") + "\n")

	inputBox := appInputStyle.Render("> " + m.input + "█")
	b.WriteString(inputBox + "\n\n")

	if m.err != nil {
		b.WriteString(appErrorStyle.Render(m.err.Error()) + "\n")
	}

	b.WriteString(appLabelStyle.Render("enter — submit  | esc, ctrl+c — quit") + "\n")

	return b.String()
}

func commandRegister(m appModel) (appModel, tea.Cmd) {
	return m, func() tea.Msg {
		return switchToAuthMsg{mode: modeRegister}
	}
}

func commandLogin(m appModel) (appModel, tea.Cmd) {
	return m, func() tea.Msg {
		return switchToAuthMsg{mode: modeLogin}
	}
}

func commandChat(m appModel) (appModel, tea.Cmd) {
	if m.cfg.currentUser.Token == "" {
		m.err = fmt.Errorf("You must be logged in first")
		return m, nil
	}

	cleanedInput := strings.Fields(strings.ToLower(m.input))
	if len(cleanedInput) < 2 {
		m.err = fmt.Errorf("usage: chat <nickname>")
		return m, nil
	}

	nickname := cleanedInput[1]

	return m, func() tea.Msg {
		result, err := m.cfg.client.startChat(nickname, m.cfg.currentUser.Token)
		if err != nil {
			return appErrMsg{err}
		}

		return switchToChatMsg{
			chatID:          result.ChatID,
			token:           m.cfg.currentUser.Token,
			currentNickname: m.cfg.currentUser.Nickname,
		}
	}
}

func commandFriends(m appModel) (appModel, tea.Cmd) {
	if m.cfg.currentUser.Token == "" {
		m.err = fmt.Errorf("You must be logged in first")
		return m, nil
	}

	cleanedInput := strings.Fields(strings.ToLower(m.input))
	if len(cleanedInput) < 2 {
		m.err = fmt.Errorf("usage: friends <nickname> | --delete <nickname>")
		return m, nil
	}

	switch cleanedInput[1] {
	case "--delete":
		if len(cleanedInput) < 3 {
			m.err = fmt.Errorf("usage: friends --delete <nickname>")
			return m, nil
		}

		targetName := cleanedInput[2]

		err := m.cfg.client.deleteFriendship(targetName, m.cfg.currentUser.Token)
		if err != nil {
			m.err = fmt.Errorf("error: %v", err)
			return m, nil
		}

		m.output = fmt.Sprintf("You have successfully removed %s from your friends list.", targetName)
		return m, nil

	default:
		targetNickname := cleanedInput[1]

		result, err := m.cfg.client.sendFriendRequest(targetNickname, m.cfg.currentUser.Token)
		if err != nil {
			m.err = fmt.Errorf("error: %v", err)
			return m, nil
		}

		switch result.Status {
		case "created":
			m.output = fmt.Sprintf("Friend request successfully sent to %s.", targetNickname)
			return m, nil
		case "accepted":
			m.output = fmt.Sprintf("You have successfully added %s as a friend.", targetNickname)
			return m, nil
		case "sent":
			m.output = fmt.Sprintf("Friend request to %s already sent.", targetNickname)
			return m, nil
		case "friends":
			m.output = fmt.Sprintf("You and %s are already friends.", targetNickname)
			return m, nil
		}
	}

	return m, nil
}

func commandNotifications(m appModel) (appModel, tea.Cmd) {
	if m.cfg.currentUser.Token == "" {
		m.err = fmt.Errorf("You must be logged in first")
		return m, nil
	}

	notifications, err := m.cfg.client.getNotifications(m.cfg.currentUser.Token)
	if err != nil {
		m.err = fmt.Errorf("Couldn't get notifications. %v", err)
		return m, nil
	}

	if len(notifications.UnreadMessages) == 0 && len(notifications.FriendRequests) == 0 {
		m.output = "You have no notifications"
		return m, nil
	}

	if len(notifications.UnreadMessages) > 0 {
		for _, msg := range notifications.UnreadMessages {
			m.notifMessages = append(m.notifMessages, fmt.Sprintf("You have %d unread messages from %s\n", msg.Count, msg.Nickname))
		}
	}

	if len(notifications.FriendRequests) > 0 {
		for _, f := range notifications.FriendRequests {
			if f.SenderNickname == m.cfg.currentUser.Nickname {
				m.notifFriends = append(m.notifFriends, fmt.Sprintf("Friend request to %s is pending\n", f.ReceiverNickname))
			} else {
				m.notifFriends = append(m.notifFriends, fmt.Sprintf("Friend request from %s\n", f.SenderNickname))
			}
		}
	}

	return m, nil
}

func getCommands() []command {
	return []command{
		{
			name:        "register",
			description: "Create a new account",
			callback:    commandRegister,
		},
		{
			name:        "login",
			description: "Sign in with you nickname and password",
			callback:    commandLogin,
		},
		{
			name:        "chat",
			description: "Open a chat with a friend. Usage: chat <nickname>",
			callback:    commandChat,
		},
		{
			name:        "friends",
			description: "Send a friend request (friends --help for more)",
			callback:    commandFriends,
		},
		{
			name:        "notifications",
			description: "View unread messages and friend requests",
			callback:    commandNotifications,
		},
	}
}
