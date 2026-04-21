package client

import (
	"fmt"
	"net/http"
)

func (c *Client) setOnline(token string) error {
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
		return parseErrorResponse(res)
	}

	return nil
}

func (c *Client) setOffline(token string) error {
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
		return parseErrorResponse(res)
	}

	return nil
}
