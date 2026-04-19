package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type notificationResponse struct {
	Nickname string `json:"Nickname"`
	Count    int64  `json:"Count"`
}

func (c *Client) GetNotifications(token string) ([]notificationResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/notifications", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var notifications []notificationResponse
	if err := decoder.Decode(&notifications); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return notifications, nil
}
