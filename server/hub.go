package server

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	userID uuid.UUID
	chatID uuid.UUID
	send   chan []byte
}

type Hub struct {
	chats      map[uuid.UUID][]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.Mutex
}

type Message struct {
	chatID  uuid.UUID
	payload []byte
}

func NewHub() *Hub {
	return &Hub{
		chats:      make(map[uuid.UUID][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.chats[client.chatID] = append(h.chats[client.chatID], client)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			clients := h.chats[client.chatID]
			for i, c := range clients {
				if c == client {
					h.chats[client.chatID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.Lock()
			for _, client := range h.chats[msg.chatID] {
				client.send <- msg.payload
			}
			h.mu.Unlock()
		}
	}
}
