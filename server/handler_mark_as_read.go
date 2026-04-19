package server

import (
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerMarkAsRead(w http.ResponseWriter, r *http.Request) {
	chatID := r.PathValue("chat_id")
	if chatID == "" {
		respondWithError(w, http.StatusBadRequest, "chat_id required", nil)
		return
	}

	parsedChatID, err := uuid.Parse(chatID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chat_id", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get JWT", err)
		return
	}

	currentUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	_, err = cfg.db.GetChatMember(r.Context(), database.GetChatMemberParams{
		ChatID: parsedChatID,
		UserID: currentUserID,
	})
	if err != nil {
		respondWithError(w, http.StatusForbidden, "not a member of this chat", nil)
		return
	}

	err = cfg.db.MarkMessagesAsRead(r.Context(), database.MarkMessagesAsReadParams{
		ChatID:   parsedChatID,
		SenderID: currentUserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't mark messages as read", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
