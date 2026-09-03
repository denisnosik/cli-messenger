package server

import (
	"context"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerHealthCheck(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status string `json:"status"`
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := cfg.dbConn.PingContext(ctx); err != nil {
		respondWithJSON(w, http.StatusServiceUnavailable, response{Status: "unhealthy"})
		return
	}

	respondWithJSON(w, http.StatusOK, response{Status: "ok"})
}
