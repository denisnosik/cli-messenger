package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type chatRequest struct {
	TargetNickname string `json:"target_nickname"`
}

type chatResponse struct {
	ChatID uuid.UUID `json:"id"`
}

func (c *Client) StartChat(targetNickname string, token string) (*chatResponse, error) {
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
		var errResponse struct {
			Error string `json:"error"`
		}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return nil, fmt.Errorf("%s", errResponse.Error)
	}

	decoder := json.NewDecoder(res.Body)
	var result chatResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}
