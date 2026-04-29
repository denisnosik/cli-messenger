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
	Nickname string `json:"nickname"`
	Online   bool   `json:"online"`
}

func (c *Client) sendFriendRequest(targetNickname string, token string) (*friendsResponse, error) {
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
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
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

func (c *Client) getFriends(token string) ([]allFriendsResponse, error) {
	req, err := http.NewRequest("GET", baseURL+"/api/friends", nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("server is unavailable")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(res)
	}

	decoder := json.NewDecoder(res.Body)
	var result []allFriendsResponse
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return result, nil
}

func (c *Client) deleteFriendship(targetNickname string, token string) error {
	body, err := json.Marshal(friendsRequest{
		TargetNickname: targetNickname,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", baseURL+"/api/friends", bytes.NewBuffer(body))
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
