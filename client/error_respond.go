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
	json.NewDecoder(res.Body).Decode(&errResponse)
	if errResponse.Error != "" {
		return fmt.Errorf("%s", errResponse.Error)
	}
	return fmt.Errorf("error: %s", res.Status)
}
