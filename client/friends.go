package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type friendsRequest struct {
	TargetNickname string `json:"target_nickname"`
}

type friendsResponse struct {
	Status string `json:"friendship_status"`
}

type allFriendsResponse struct {
	Friends []string `json:"friends"`
}

func (c *Client) handlerFriends(targetNickname string, currentUser CurrentUser) (*friendsResponse, error) {
	body, err := json.Marshal(friendsRequest{
		TargetNickname: targetNickname,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+"/api/friends", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+currentUser.Token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("error: %s", res.Status)
	}

	decoder := json.NewDecoder(res.Body)
	var result friendsResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}

func (c *Client) handlerGetFriends(currentUser CurrentUser) (*allFriendsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/friends", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+currentUser.Token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errResponse struct {
			Error string `json:"error"`
		}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return nil, fmt.Errorf("%s", errResponse.Error)
	}

	decoder := json.NewDecoder(res.Body)
	var result allFriendsResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return &result, nil
}
