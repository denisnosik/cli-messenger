package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) SetOnline(token string) error {
	req, err := http.NewRequest("POST", baseURL+"/api/online", nil)
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
		var errResponse struct {
			Error string `json:"error"`
		}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return fmt.Errorf("%s", errResponse.Error)
	}

	return nil
}

func (c *Client) SetOffline(token string) error {
	req, err := http.NewRequest("POST", baseURL+"/api/offline", nil)
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
		var errResponse struct {
			Error string `json:"error"`
		}
		json.NewDecoder(res.Body).Decode(&errResponse)
		return fmt.Errorf("%s", errResponse.Error)
	}

	return nil
}
