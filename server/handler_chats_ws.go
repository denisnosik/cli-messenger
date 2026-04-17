package server

import (
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (cfg *apiConfig) handlerChatWS(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chat_id")
	parsedChatID, err := uuid.Parse(chatID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chat_id", err)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get JWT", nil)
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
	defer conn.Close()

	client := &Client{
		conn:   conn,
		userID: currentUserID,
		chatID: parsedChatID,
		send:   make(chan []byte, 256),
	}

	cfg.hub.register <- client
	defer func() {
		cfg.hub.unregister <- client
	}()

	// write to client from send chan
	go func() {
		defer conn.Close()
		for msg := range client.send {
			err := conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				break
			}
		}
	}()

	// read msgs from client (main goroutine)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break // client disconnected
		}

		// save to DB
		cfg.db.CreateMessage(r.Context(), database.CreateMessageParams{
			ChatID:   parsedChatID,
			SenderID: currentUserID,
			Content:  string(msg),
		})

		// broadcast to everyone in the chat
		cfg.hub.broadcast <- &Message{
			chatID:  parsedChatID,
			payload: msg,
		}
	}

}
