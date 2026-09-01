package server

import (
	"net/http"
	"uuid"

	"github.com/denisnosik/dedachat/internal/database"
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

	currentUserID := r.Context().Value(contextKeyUserID).(uuid.UUID)

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
