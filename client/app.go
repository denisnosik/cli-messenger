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

type switchToAppMsg struct{}

type friendsResultMsg struct {
	output string
}

type notificationsResultMsg struct {
	messages []string
	friends  []string
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

	case friendsResultMsg:
		m.output = msg.output
		return m, nil

	case notificationsResultMsg:
		if len(msg.messages) == 0 && len(msg.friends) == 0 {
			m.output = "no notifications"
			return m, nil
		}
		m.notifMessages = msg.messages
		m.notifFriends = msg.friends
		return m, nil

	case appErrMsg:
		m.err = msg.err
		return m, nil

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
		m.err = fmt.Errorf("you must be logged in first")
		return m, nil
	}

	cleanedInput := strings.Fields(m.input)
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
		m.err = fmt.Errorf("you must be logged in first")
		return m, nil
	}

	cleanedInput := strings.Fields(m.input)
	if len(cleanedInput) < 2 {
		m.err = fmt.Errorf("usage: friends <nickname> | --delete <nickname>")
		return m, nil
	}

	flag := strings.ToLower(cleanedInput[1])

	switch flag {
	case "--delete":
		if len(cleanedInput) < 3 {
			m.err = fmt.Errorf("usage: friends --delete <nickname>")
			return m, nil
		}

		target := cleanedInput[2]
		return m, func() tea.Msg {
			err := m.cfg.client.deleteFriendship(target, m.cfg.currentUser.Token)
			if err != nil {
				return appErrMsg{err}
			}

			return friendsResultMsg{fmt.Sprintf("removed %s from friends", target)}
		}

	default:
		target := cleanedInput[1]
		return m, func() tea.Msg {
			result, err := m.cfg.client.sendFriendRequest(target, m.cfg.currentUser.Token)
			if err != nil {
				return appErrMsg{err}
			}
			switch result.Status {
			case "created":
				return friendsResultMsg{fmt.Sprintf("friend request sent to %s", target)}
			case "accepted":
				return friendsResultMsg{fmt.Sprintf("you and %s are now friends", target)}
			case "sent":
				return friendsResultMsg{fmt.Sprintf("friend request to %s already sent", target)}
			case "friends":
				return friendsResultMsg{fmt.Sprintf("you and %s are already friends", target)}
			}
			return nil
		}
	}
}

func commandNotifications(m appModel) (appModel, tea.Cmd) {
	if m.cfg.currentUser.Token == "" {
		m.err = fmt.Errorf("you must be logged in first")
		return m, nil
	}

	return m, func() tea.Msg {
		notifs, err := m.cfg.client.getNotifications(m.cfg.currentUser.Token)
		if err != nil {
			return appErrMsg{err}
		}

		var messages, friends []string

		for _, msg := range notifs.UnreadMessages {
			messages = append(messages, fmt.Sprintf("%d unread from %s", msg.Count, msg.Nickname))
		}

		for _, f := range notifs.FriendRequests {
			if f.SenderNickname == m.cfg.currentUser.Nickname {
				friends = append(friends, fmt.Sprintf("pending request to %s", f.ReceiverNickname))
			} else {
				friends = append(friends, fmt.Sprintf("friend request from %s", f.SenderNickname))
			}
		}

		return notificationsResultMsg{
			messages: messages,
			friends:  friends,
		}
	}
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
