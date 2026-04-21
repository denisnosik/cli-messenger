package server

import (
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerSetOnline(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("currentUserID").(uuid.UUID)

	cfg.hub.SetOnline(currentUserID)
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerSetOffline(w http.ResponseWriter, r *http.Request) {
	currentUserID := r.Context().Value("currentUserID").(uuid.UUID)

	cfg.hub.SetOffline(currentUserID)
	w.WriteHeader(http.StatusNoContent)
}
