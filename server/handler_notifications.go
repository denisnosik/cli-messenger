package server

import (
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
)

type notification struct {
	Nickname string `json:"Nickname"`
	Count    int64  `json:"Count"`
}

func (cfg *apiConfig) handlerNotifications(w http.ResponseWriter, r *http.Request) {
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

	dbNotifications, err := cfg.db.GetUnreadMessages(r.Context(), currentUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get notifications", err)
		return
	}

	var notifications []notification
	for _, n := range dbNotifications {
		notifications = append(notifications, notification{
			Nickname: n.Nickname,
			Count:    n.Count,
		})
	}

	respondWithJSON(w, http.StatusOK, notifications)
}
