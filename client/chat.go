package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

type model struct {
	messages []string
	input    string
	msgChan  chan wsMessage
	conn     *websocket.Conn
}

type incomingMsg wsMessage

func (c *Client) connectToChat(chatID uuid.UUID, token string) error {
	cutPrefixURL, _ := strings.CutPrefix(baseURL, "http://")
	wsURL := fmt.Sprintf("ws://%s/api/chats/ws?chat_id=%s&token=%s", cutPrefixURL, chatID, url.QueryEscape(token))

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Couldn't connect to chat:", err)
		return err
	}
	defer conn.Close()

	msgChan := make(chan wsMessage)

	done := make(chan struct{})
	closeOnce := sync.Once{}

	closeDone := func() {
		closeOnce.Do(func() { close(done) })
	}

	// messages reader
	go func() {
		defer closeDone()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				msgChan <- wsMessage{Content: "Disconnected"}
				return
			}

			wsMsg := wsMessage{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				msgChan <- wsMessage{Content: string(msg)}
				continue
			}

			msgChan <- wsMsg
		}
	}()

	m := model{
		msgChan: msgChan,
		conn:    conn,
	}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		return err
	}

	// throttle markAsRead
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.markAsRead(chatID, token)
			case <-done:
				return
			}
		}
	}()

	return nil
}

func (m model) Init() tea.Cmd {
	return waitForMessage(m.msgChan)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case incomingMsg:
		formatted := fmt.Sprintf("[%s] [%s] %s",
			msg.CreatedAt.Format("02 Jan 15:04"),
			msg.Nickname,
			msg.Content,
		)
		m.messages = append(m.messages, formatted)

		return m, waitForMessage(m.msgChan)

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.input != "" {
				m.conn.WriteMessage(websocket.TextMessage, []byte(m.input))
				formatted := fmt.Sprintf("[%s] [You] %s",
					time.Now().Format("02 Jan 15:04"),
					m.input,
				)
				m.messages = append(m.messages, formatted)
				m.input = ""
			}

		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		case "ctrl+c", "esc":
			return m, tea.Quit

		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	for _, msg := range m.messages {
		b.WriteString(msg + "\n")
	}

	b.WriteString("\n> " + m.input)

	return b.String()
}

func waitForMessage(ch chan wsMessage) tea.Cmd {
	return func() tea.Msg {
		return incomingMsg(<-ch)
	}
}
