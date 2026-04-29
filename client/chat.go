package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type wsMessage struct {
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	Content   string    `json:"content"`
}

func (c *Client) connectToChat(chatID uuid.UUID, token string) error {
	cutPrefixURL, _ := strings.CutPrefix(baseURL, "http://")
	wsURL := fmt.Sprintf("ws://%s/api/chats/ws?chat_id=%s&token=%s", cutPrefixURL, chatID, url.QueryEscape(token))

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Couldn't connect to chat:", err)
		return err
	}
	defer conn.Close()

	done := make(chan struct{})
	closeOnce := sync.Once{}

	closeDone := func() {
		closeOnce.Do(func() { close(done) })
	}

	// reader for msgs
	go func() {
		defer closeDone()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("\nDisconnected from chat")
				return
			}
			wsMsg := wsMessage{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				fmt.Printf("\n%s\n", string(msg))
				continue
			}

			fmt.Printf("\n[%s] [%s] %s\n", wsMsg.CreatedAt.Format("02 Jan 15:04"), wsMsg.Nickname, wsMsg.Content)
		}
	}()

	fmt.Println("Connected! Type your message (or '/exit' to leave):")

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

	// main write and send
	for {
		select {
		case <-done:
			return nil
		default:
		}

		line, err := rl.Readline()
		if err != nil {
			break
		}

		text := strings.TrimSpace(line)
		if text == "/exit" {
			closeDone()
			break
		}
		if text == "" {
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			fmt.Println("Couldn't send message:", err)
			break
		}
	}

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		fmt.Println("Couldn't close message:", err)
		return err
	}
	return nil
}
