package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type chatRequest struct {
	TargetNickname string `json:"target_nickname"`
}

type chatResponse struct {
	ChatID uuid.UUID `json:"id"`
}

type wsMessage struct {
	Nickname string `json:"nickname"`
	Content  string `json:"content"`
}

func (c *Client) StartChat(targetNickname string, currentUser CurrentUser) (*chatResponse, error) {
	body, err := json.Marshal(chatRequest{
		TargetNickname: targetNickname,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/chats", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+currentUser.Token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error: %s", res.Status)
	}

	decoder := json.NewDecoder(res.Body)
	var result chatResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}

func (c *Client) ConnectToChat(chatID uuid.UUID, token string) error {
	wsURL := fmt.Sprintf("ws://localhost:8080/api/chats/ws?chat_id=%s&token=%s", chatID, url.QueryEscape(token))
	fmt.Println("Connecting to:", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Couldn't connect to chat:", err)
		return err
	}
	defer conn.Close()

	done := make(chan struct{})
	// reader for msgs
	go func() {
		defer close(done)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("\nDisconnected from chat")
				return
			}
			wsMsg := wsMessage{}
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				fmt.Printf("\n%s\n> ", string(msg))
				continue
			}

			fmt.Printf("\n[%s] %s\n> ", wsMsg.Nickname, wsMsg.Content)
		}
	}()

	fmt.Println("Connected! Type your message (or 'exit' to leave):")

	// write and send
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-done:
			return nil
		default:
		}
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		text := strings.TrimSpace(scanner.Text())
		if text == "exit" {
			break
		}
		if text == "" {
			continue
		}

		err := conn.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			fmt.Println("Couldn't send message:", err)
			break
		}
	}

	err = conn.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		fmt.Println("Couldn't close message:", err)
		return err
	}
	return nil
}
