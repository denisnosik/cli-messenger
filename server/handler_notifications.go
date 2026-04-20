package server

import (
	"net/http"

	"github.com/denisnosik/cli-messenger/internal/auth"
)

type unreadMessages struct {
	Nickname string `json:"Nickname"`
	Count    int64  `json:"Count"`
}

type friendRequest struct {
	SenderNickname   string `json:"sender_nickname"`
	ReceiverNickname string `json:"receiver_nickname"`
}

type notificationsResponse struct {
	UnreadMessages []unreadMessages `json:"unread_messages"`
	FriendRequests []friendRequest  `json:"friend_requests"`
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

	dbUnreadMessages, err := cfg.db.GetUnreadMessages(r.Context(), currentUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get unread messages", err)
		return
	}

	dbFriendRequest, err := cfg.db.GetFriendRequestsForUser(r.Context(), currentUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get friend requests", err)
		return
	}

	var unreadMsgs []unreadMessages
	for _, msg := range dbUnreadMessages {
		unreadMsgs = append(unreadMsgs, unreadMessages{
			Nickname: msg.Nickname,
			Count:    msg.Count,
		})
	}

	var friendReqs []friendRequest
	for _, f := range dbFriendRequest {
		friendReqs = append(friendReqs, friendRequest{
			SenderNickname:   f.SenderNickname,
			ReceiverNickname: f.ReceiverNickname,
		})
	}

	respondWithJSON(w, http.StatusOK, notificationsResponse{
		UnreadMessages: unreadMsgs,
		FriendRequests: friendReqs,
	})
}
