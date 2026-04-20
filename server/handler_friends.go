package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
	"github.com/denisnosik/cli-messenger/internal/database"
)

type FriendshipStatus struct {
	Status string `json:"friendship_status"`
}

type AllFriends struct {
	Friends []string `json:"friends"`
}

func (cfg *apiConfig) handlerFriends(w http.ResponseWriter, r *http.Request) {
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
	if err == sql.ErrNoRows {
		respondWithError(w, http.StatusNotFound, "User doesn't exist", nil)
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user", nil)
		return
	}

	if targetUser.ID == currentUserID {
		respondWithError(w, http.StatusBadRequest, "Can't send friend request to yourself", nil)
		return
	}

	friendshipStatus, err := cfg.db.GetFriendshipStatus(r.Context(), database.GetFriendshipStatusParams{
		UserID:   currentUserID,
		FriendID: targetUser.ID,
	})
	if err == sql.ErrNoRows {
		_, err := cfg.db.CreateFriendRequest(r.Context(), database.CreateFriendRequestParams{
			UserID:   currentUserID,
			FriendID: targetUser.ID,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create friend request", nil)
			return
		}
		respondWithJSON(w, http.StatusCreated, FriendshipStatus{Status: "created"})
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get friendship status from db", nil)
		return
	}

	switch friendshipStatus.RequestStatus {
	case "pending":
		if friendshipStatus.UserID == targetUser.ID {
			err := cfg.db.AcceptFriendRequest(r.Context(), database.AcceptFriendRequestParams{
				UserID:   targetUser.ID,
				FriendID: currentUserID,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't accept friendship", nil)
				return
			}
			err = cfg.db.CreateFriendship(r.Context(), database.CreateFriendshipParams{
				UserID:   currentUserID,
				FriendID: targetUser.ID,
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't create friendship", nil)
				return
			}
			respondWithJSON(w, http.StatusOK, FriendshipStatus{Status: "accepted"})
			return
		} else {
			respondWithJSON(w, http.StatusOK, FriendshipStatus{Status: "sent"})
			return
		}

	case "accepted":
		respondWithJSON(w, http.StatusOK, FriendshipStatus{Status: "friends"})
		return
	}
}

func (cfg *apiConfig) handlerGetFriends(w http.ResponseWriter, r *http.Request) {
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

	friends, err := cfg.db.GetAllFriendsForUser(r.Context(), currentUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get friends", nil)
		return
	}

	if len(friends) == 0 {
		respondWithError(w, http.StatusNotFound, "You have no friends", nil)
		return
	}

	type friendWithStatus struct {
		Nickname string `json:"nickname"`
		Online   bool   `json:"online"`
	}

	var friendsWithStatus []friendWithStatus
	for _, f := range friends {
		online := cfg.hub.IsOnline(f.FriendID)
		friendsWithStatus = append(friendsWithStatus, friendWithStatus{
			Nickname: f.FriendNickname,
			Online:   online,
		})
	}

	respondWithJSON(w, http.StatusOK, friendsWithStatus)
}
