package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type notificationsResponse struct {
	UnreadMessages []unreadMessages `json:"unread_messages"`
	FriendRequests []friendRequest  `json:"friend_requests"`
}

type unreadMessages struct {
	Nickname string `json:"nickname"`
	Count    int64  `json:"count"`
}

type friendRequest struct {
	SenderNickname   string `json:"sender_nickname"`
	ReceiverNickname string `json:"receiver_nickname"`
}

func (c *Client) getNotifications(token string) (notificationsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/notifications", nil)
	if err != nil {
		return notificationsResponse{}, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return notificationsResponse{}, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return notificationsResponse{}, parseErrorResponse(res)
	}

	decoder := json.NewDecoder(res.Body)
	var notifications notificationsResponse
	if err := decoder.Decode(&notifications); err != nil {
		return notificationsResponse{}, fmt.Errorf("decode failed: %w", err)
	}

	return notifications, nil
}
