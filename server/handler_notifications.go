package server

import (
	"net/http"
	"uuid"
)

type unreadMessages struct {
	Nickname string `json:"nickname"`
	Count    int64  `json:"count"`
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
	currentUserID := r.Context().Value(contextKeyUserID).(uuid.UUID)

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

	unreadMsgs := make([]unreadMessages, 0, len(dbUnreadMessages))
	for _, msg := range dbUnreadMessages {
		unreadMsgs = append(unreadMsgs, unreadMessages{
			Nickname: msg.Nickname,
			Count:    msg.Count,
		})
	}

	friendReqs := make([]friendRequest, 0, len(dbFriendRequest))
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
