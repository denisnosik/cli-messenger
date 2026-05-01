package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

type model struct {
	messages        []string
	input           string
	msgChan         chan wsMessage
	conn            *websocket.Conn
	ctx             context.Context
	currentNickname string
	width           int
	height          int
	scrollOffset    int
	msgBoxHeight    int
	err             error
}

const maxInputLen = 150

var (
	containerStyle = lipgloss.NewStyle().
			Padding(1, 2)

	timeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	nicknameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true)

	youStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	counterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

type incomingMsg wsMessage

func (c *Client) connectToChat(chatID uuid.UUID, token string, currentNickname string) error {
	cutPrefixURL, _ := strings.CutPrefix(baseURL, "http://")
	wsURL := fmt.Sprintf("ws://%s/api/chats/ws?chat_id=%s&token=%s", cutPrefixURL, chatID, url.QueryEscape(token))

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Couldn't connect to chat:", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan wsMessage, 50)

	// messages reader
	go func() {
		defer cancel()

		send := func(msg wsMessage) bool {
			select {
			case msgChan <- msg:
				return true
			case <-ctx.Done():
				return false
			}
		}

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				send(wsMessage{Content: "Disconnected"})
				return
			}

			wsMsg := wsMessage{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				msgChan <- wsMessage{Content: string(msg)}
				continue
			}

			if !send(wsMsg) {
				return
			}
		}
	}()

	// throttle markAsRead
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := c.markAsRead(chatID, token); err != nil {
					log.Printf("Couldn't mark as read: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	m := model{
		msgChan:         msgChan,
		conn:            conn,
		ctx:             ctx,
		currentNickname: currentNickname,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return err
	}

	if err := conn.Close(); err != nil {
		log.Printf("couldn't close connection: %v", err)
	}

	return nil
}

func (m model) Init() tea.Cmd {
	return waitForMessage(m.ctx, m.msgChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		const reservedLines = 7

		m.msgBoxHeight = msg.Height - reservedLines
		if m.msgBoxHeight < 3 {
			m.msgBoxHeight = 3
		}

		return m, nil

	case incomingMsg:
		formatted := formatMessage(wsMessage(msg), m.currentNickname)
		m.messages = append(m.messages, formatted)
		return m, waitForMessage(m.ctx, m.msgChan)

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input != "" {
				if err := m.conn.WriteMessage(websocket.TextMessage, []byte(m.input)); err != nil {
					m.err = fmt.Errorf("couldn't send message: %w", err)
					return m, nil
				}

				formatted := formatMessage(wsMessage{
					Nickname:  m.currentNickname,
					CreatedAt: time.Now(),
					Content:   m.input,
				}, m.currentNickname)

				m.messages = append(m.messages, formatted)
				m.input = ""
			}

		case "backspace":
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
			}

		case "up":
			if m.scrollOffset < len(m.messages)-m.msgBoxHeight {
				m.scrollOffset++
			}

		case "down":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}

		case "ctrl+c", "esc":
			return m, tea.Quit

		default:
			if len([]rune(m.input)) < maxInputLen && len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Render("Error: " + m.err.Error())
	}

	width := m.width
	if width < 40 {
		width = 80
	}

	innerWidth := width - 8

	const reservedLines = 7
	msgBoxHeight := m.height - reservedLines
	if msgBoxHeight < 3 {
		msgBoxHeight = 3
	}

	total := len(m.messages)
	end := total - m.scrollOffset
	if end < 0 {
		end = 0
	}

	start := end - msgBoxHeight
	if start < 0 {
		start = 0
	}

	visible := m.messages[start:end]

	var msgBuilder strings.Builder
	for _, msg := range visible {
		msgBuilder.WriteString(msg + "\n")
	}

	msgs := msgBuilder.String()
	if msgs == "" {
		msgs = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Render("No messages yet...")
	}

	scrollIndicator := ""
	if m.scrollOffset > 0 {
		scrollIndicator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render(fmt.Sprintf(" ↑ %d more", m.scrollOffset))
	}

	messagesBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("255")).
		Padding(0, 1).
		Width(innerWidth).
		Height(msgBoxHeight).
		Render(msgs)

	counter := fmt.Sprintf("%d/%d", len([]rune(m.input)), maxInputLen)
	counterLine := lipgloss.NewStyle().
		Width(innerWidth - 2).
		Align(lipgloss.Right).
		Render(counterStyle.Render(counter))

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("255")).
		Padding(0, 1).
		Width(innerWidth).
		Render("> " + m.input + "█\n" + counterLine)

	help := helpStyle.Render("enter — send  |  ↑↓ — scroll  |  esc/ctrl+c — quit")

	return containerStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			messagesBox,
			scrollIndicator,
			"",
			inputBox,
			help,
		),
	)
}

func formatMessage(msg wsMessage, currentNickname string) string {
	ts := timeStyle.Render("[" + msg.CreatedAt.Format("02 Jan 15:04") + "]")

	isYou := msg.Nickname == currentNickname

	var nick string
	if isYou {
		nick = youStyle.Render("[You]")
	} else {
		nick = nicknameStyle.Render("[" + msg.Nickname + "]")
	}

	return ts + " " + nick + " " + msg.Content
}

func waitForMessage(ctx context.Context, ch chan wsMessage) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			return incomingMsg(msg)
		case <-ctx.Done():
			return nil
		}
	}
}
