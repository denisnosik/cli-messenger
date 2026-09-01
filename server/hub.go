package server

import (
	"uuid"
)

type Hub struct {
	clients    map[*Client]bool
	online     map[uuid.UUID]int
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	onlineReq  chan onlineRequest
	setOnline  chan uuid.UUID
	setOffline chan uuid.UUID
}

type onlineRequest struct {
	userID uuid.UUID
	res    chan bool
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		online:     make(map[uuid.UUID]int),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		onlineReq:  make(chan onlineRequest),
		setOnline:  make(chan uuid.UUID),
		setOffline: make(chan uuid.UUID),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.broadcast:
			for client := range h.clients {
				if client.chatID != msg.chatID {
					continue
				}

				if client.userID == msg.senderID {
					continue
				}

				select {
				case client.send <- msg.payload:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case req := <-h.onlineReq:
			req.res <- h.online[req.userID] > 0

		case userID := <-h.setOnline:
			h.online[userID]++

		case userID := <-h.setOffline:
			h.online[userID]--
			if h.online[userID] < 0 {
				h.online[userID] = 0
			}
		}
	}
}

func (h *Hub) IsOnline(userID uuid.UUID) bool {
	res := make(chan bool)
	h.onlineReq <- onlineRequest{
		userID: userID,
		res:    res,
	}
	return <-res
}

func (h *Hub) SetOnline(userID uuid.UUID) {
	h.setOnline <- userID
}

func (h *Hub) SetOffline(userID uuid.UUID) {
	h.setOffline <- userID
}
