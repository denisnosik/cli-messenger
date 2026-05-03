package client

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type authStep int

const (
	stepNickname authStep = iota
	stepPassword
)

type authModel struct {
	cfg      *config
	step     authStep
	nickname string
	password string
	err      error
}

type loginSuccessMsg struct {
	nickname string
	token    string
}

type loginErrMsg struct{ err error }

var (
	authWelcomeStyle = lipgloss.NewStyle().
				MarginLeft(2).
				Foreground(lipgloss.Color("2")).
				Bold(true)

	authLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginLeft(2)

	authActiveInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("255")).
				Padding(0, 1).
				MarginLeft(2).
				Width(40)

	authInactiveInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("238")).
				Padding(0, 1).
				MarginLeft(2).
				Width(40)

	authErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			MarginLeft(2)
)

func (m authModel) Init() tea.Cmd {
	return nil
}

func (m authModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case loginSuccessMsg:
		m.cfg.currentUser.Nickname = msg.nickname
		m.cfg.currentUser.Token = msg.token
		return m, tea.Quit

	case loginErrMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "enter":
			if m.step == stepNickname {
				if m.nickname == "" {
					m.err = fmt.Errorf("nickname cannot be empty")
					return m, nil
				}
				m.step = stepPassword
			} else {
				if m.password == "" {
					m.err = fmt.Errorf("password cannot be empty")
					return m, nil
				}

				return m, m.doLogin
			}

		case "backspace":
			if m.step == stepNickname {
				runes := []rune(m.nickname)
				if len(runes) > 0 {
					m.nickname = string(runes[:len(runes)-1])
				}
			} else {
				runes := []rune(m.password)
				if len(runes) > 0 {
					m.password = string(runes[:len(runes)-1])
				}
			}

		default:
			m.err = nil
			if len(msg.String()) == 1 {
				if m.step == stepNickname {
					m.nickname += msg.String()
				} else {
					m.password += msg.String()
				}
			}
		}
	}

	return m, nil
}

func (m authModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(authWelcomeStyle.Render("Welcome") + "\n\n")

	nicknameBox := ""
	if m.step == stepNickname {
		nicknameBox = authActiveInputStyle.Render(m.nickname + "█")
	} else {
		nicknameBox = authInactiveInputStyle.Render(m.nickname)
	}

	b.WriteString(authLabelStyle.Render("Nickname") + "\n")
	b.WriteString(nicknameBox + "\n\n")

	if m.step == stepPassword {
		masked := strings.Repeat("*", len([]rune(m.password)))
		passwordBox := authActiveInputStyle.Render(masked + "█")

		b.WriteString(authLabelStyle.Render("Password") + "\n")
		b.WriteString(passwordBox + "\n\n")
	}

	if m.err != nil {
		b.WriteString(authErrorStyle.Render(m.err.Error()) + "\n")
	}

	b.WriteString(authLabelStyle.Render("enter — next  |  esc — quit") + "\n")

	return b.String()
}

func (m authModel) doLogin() tea.Msg {
	result, err := m.cfg.client.login(m.nickname, m.password)
	if err != nil {
		return loginErrMsg{fmt.Errorf("couldn't login")}
	}

	if err := m.cfg.client.setOnline(result.Token); err != nil {
		return loginErrMsg{fmt.Errorf("couldn't set online")}
	}

	return loginSuccessMsg{
		nickname: result.Nickname,
		token:    result.Token,
	}
}

func startAuth(cfg *config) {
	if _, err := tea.NewProgram(authModel{cfg: cfg}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
