package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func parseErrorResponse(res *http.Response) error {
	var errResponse struct {
		Error string `json:"error"`
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&errResponse); err != nil {
		return fmt.Errorf("Couldn't decode: %w", err)
	}

	if errResponse.Error != "" {
		return fmt.Errorf("%s", errResponse.Error)
	}

	return fmt.Errorf("error: %s", res.Status)
}
