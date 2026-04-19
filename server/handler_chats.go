package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/google/uuid"
)

type Chat struct {
	ID uuid.UUID `json:"id"`
}

func (cfg *apiConfig) handlerChat(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		TargetNickname string `json:"target_nickname"`
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

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode", err)
		return
	}

	targetUser, err := cfg.db.GetUserByNickname(r.Context(), params.TargetNickname)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "user doesn't exist", nil)
		return
	}

	if targetUser.ID == currentUserID {
		respondWithError(w, http.StatusBadRequest, "Can't create chat with yourself", nil)
		return
	}

	chatID, err := cfg.db.GetChatByTwoUsers(r.Context(), database.GetChatByTwoUsersParams{
		UserID:   currentUserID,
		UserID_2: targetUser.ID,
	})
	if err == sql.ErrNoRows {
		chat, err := cfg.db.CreateChat(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create chat", err)
			return
		}

		chatID = chat.ID

		_, err = cfg.db.CreateChatMember(r.Context(), database.CreateChatMemberParams{ChatID: chatID, UserID: currentUserID})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create chat member", err)
			return
		}

		_, err = cfg.db.CreateChatMember(r.Context(), database.CreateChatMemberParams{ChatID: chatID, UserID: targetUser.ID})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create chat member", err)
			return
		}

		respondWithJSON(w, http.StatusCreated, Chat{ID: chatID})
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chat by two users", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chat{ID: chatID})
}
