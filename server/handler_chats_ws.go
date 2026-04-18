package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	hub    *Hub
	userID uuid.UUID
	chatID uuid.UUID
	send   chan []byte
}

type Message struct {
	chatID  uuid.UUID
	payload []byte
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func (cfg *apiConfig) handlerChatWS(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		respondWithError(w, http.StatusBadRequest, "chat_id required", nil)
		return
	}

	parsedChatID, err := uuid.Parse(chatID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse chat id", err)
		return
	}

	token := r.URL.Query().Get("token")
	log.Printf("Received token: %q", token)
	if token == "" {
		respondWithError(w, http.StatusBadRequest, "token required", nil)
		return
	}

	currentUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't upgrade", err)
		return
	}

	client := &Client{
		conn:   conn,
		hub:    cfg.hub,
		userID: currentUserID,
		chatID: parsedChatID,
		send:   make(chan []byte, 256),
	}

	cfg.hub.register <- client

	go client.writeToClient()
	go client.readFromClient(cfg)
}

func (c *Client) writeToClient() {
	ticker := time.NewTicker(pingPeriod)

	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readFromClient(cfg *apiConfig) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		_, err = cfg.db.CreateMessage(context.Background(), database.CreateMessageParams{
			ChatID:   c.chatID,
			SenderID: c.userID,
			Content:  string(msg),
		})
		if err != nil {
			log.Printf("db error: %v", err)
			continue
		}

		c.hub.broadcast <- &Message{
			chatID:  c.chatID,
			payload: msg,
		}
	}
}
