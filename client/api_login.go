package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type loginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type loginResponse struct {
	Nickname string `json:"nickname"`
	Token    string `json:"token"`
}

func (c *Client) login(nickname, password string) (*loginResponse, error) {
	body, err := json.Marshal(loginRequest{
		Nickname: nickname,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/login", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(res)
	}

	decoder := json.NewDecoder(res.Body)
	var result loginResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}
