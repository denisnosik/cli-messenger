package server

import (
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
)

func (cfg *apiConfig) handlerSetOnline(w http.ResponseWriter, r *http.Request) {
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

	cfg.hub.SetOnline(currentUserID)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerSetOffline(w http.ResponseWriter, r *http.Request) {
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

	cfg.hub.SetOffline(currentUserID)
	w.WriteHeader(http.StatusNoContent)
}
