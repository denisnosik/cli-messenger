package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"uuid"

	"github.com/denisnosik/dedachat/internal/database"
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

	currentUserID := r.Context().Value(contextKeyUserID).(uuid.UUID)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode", nil)
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
	currentUserID := r.Context().Value(contextKeyUserID).(uuid.UUID)

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

func (cfg *apiConfig) handlerDeleteFriend(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		TargetNickname string `json:"target_nickname"`
	}

	currentUserID := r.Context().Value(contextKeyUserID).(uuid.UUID)

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
		respondWithError(w, http.StatusBadRequest, "You can't use your nickname", nil)
		return
	}

	friendshipStatus, err := cfg.db.GetFriendshipStatus(r.Context(), database.GetFriendshipStatusParams{
		UserID:   currentUserID,
		FriendID: targetUser.ID,
	})
	if err == sql.ErrNoRows || friendshipStatus.RequestStatus == "pending" {
		respondWithError(w, http.StatusNotFound, "You are not friends", nil)
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get friendship status", nil)
		return
	}

	err = cfg.db.DeleteFriendship(r.Context(), database.DeleteFriendshipParams{
		UserID:   currentUserID,
		FriendID: targetUser.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete friendship", nil)
		return
	}

	chatID, err := cfg.db.GetChatByTwoUsers(r.Context(), database.GetChatByTwoUsersParams{
		UserID:   currentUserID,
		UserID_2: targetUser.ID,
	})
	if err == nil {
		err = cfg.db.DeleteChat(r.Context(), chatID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't delete chat", nil)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
