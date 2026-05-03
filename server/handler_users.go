package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/denisnosik/dedachat/internal/auth"
	"github.com/denisnosik/dedachat/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode.", nil)
		return
	}

	hashPwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Cannot hash the password", nil)
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
		respondWithError(w, http.StatusInternalServerError, "DB error", nil)
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

func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode", nil)
		return
	}

	dbUser, err := cfg.db.GetUserByNickname(r.Context(), params.Nickname)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect nickname or password", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect nickname or password", err)
		return
	}

	user := User{
		ID:        dbUser.ID,
		Nickname:  dbUser.Nickname,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, auth.JWTExp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't make JWT", nil)
		return
	}

	user.Token = token

	respondWithJSON(w, http.StatusOK, user)
}
