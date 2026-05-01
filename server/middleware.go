package server

import (
	"context"
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
)

type contextKey string

const contextKeyUserID contextKey = "currentUserID"

func (cfg *apiConfig) middlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), contextKeyUserID, currentUserID)
		next(w, r.WithContext(ctx))
	}
}
