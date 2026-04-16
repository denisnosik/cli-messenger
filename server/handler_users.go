package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode.", err)
		return
	}

	hashPwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot hash the password", err)
		return
	}

	dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Nickname:       params.Nickname,
		HashedPassword: hashPwd,
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			respondWithError(w, http.StatusConflict, "User already exists", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "DB error", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		Nickname:  dbUser.Nickname,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}

	respondWithJSON(w, http.StatusCreated, user)
}
