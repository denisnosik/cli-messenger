package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type registerRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type registerResponse struct {
	Nickname string `json:"nickname"`
}

func (c *Client) Register(nickname, password string) (*registerResponse, error) {
	body, err := json.Marshal(registerRequest{
		Nickname: nickname,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/users", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error: %s", res.Status)
	}

	decoder := json.NewDecoder(res.Body)
	var result registerResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}
