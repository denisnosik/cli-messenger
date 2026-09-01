package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"uuid"
)

type chatRequest struct {
	TargetNickname string `json:"target_nickname"`
}

type chatResponse struct {
	ChatID uuid.UUID `json:"id"`
}

func (c *Client) startChat(targetNickname string, token string) (*chatResponse, error) {
	body, err := json.Marshal(chatRequest{
		TargetNickname: targetNickname,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/chats", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, parseErrorResponse(res)
	}

	decoder := json.NewDecoder(res.Body)
	var result chatResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}

func (c *Client) markAsRead(chatID uuid.UUID, token string) error {
	url := fmt.Sprintf("%s/api/chats/%s/read", baseURL, chatID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		return parseErrorResponse(res)
	}

	return nil
}
