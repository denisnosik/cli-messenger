package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type registerRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type registerResponse struct {
	Nickname string `json:"nickname"`
}

var testClient = http.Client{Timeout: 5 * time.Second}

const baseURL = "http://localhost:8080"

func TestUserRegister(t *testing.T) {
	// Test: Valid register (handlerCreateUser)
	body, err := json.Marshal(registerRequest{
		Nickname: "test_user1",
		Password: "000000",
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var resultRegister registerResponse
	err = decoder.Decode(&resultRegister)
	require.NoError(t, err)

	assert.Equal(t, "test_user1", resultRegister.Nickname)

	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (request body no nickname)
	body, err = json.Marshal(registerRequest{
		Password: "000000",
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	var errRes errorResponse
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)
	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (request body no password)
	body, err = json.Marshal(registerRequest{
		Nickname: "test_user1",
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)
	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (user already exist)
	body, err = json.Marshal(registerRequest{
		Nickname: "test_user1",
		Password: "000000",
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusConflict, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)
	res.Body.Close()
	// -----------------------------------------
}
func TestUserLogin(t *testing.T) {
	// Test: Valid login user (handlerLoginUser)
	body, err := json.Marshal(registerRequest{
		Nickname: "test_user1",
		Password: "000000",
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/login", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var resultUser User
	err = decoder.Decode(&resultUser)
	require.NoError(t, err)

	assert.Equal(t, "test_user1", resultUser.Nickname)
	assert.NotNil(t, resultUser.ID)
	assert.NotNil(t, resultUser.CreatedAt)
	assert.NotNil(t, resultUser.UpdatedAt)

	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid login user (user doesn't exist)
	body, err = json.Marshal(registerRequest{
		Nickname: "no_such_user",
		Password: "000000",
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/login", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	var errRes errorResponse
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)

	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid login user (wrong password)
	body, err = json.Marshal(registerRequest{
		Nickname: "test_user1",
		Password: "111111",
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/login", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)

	res.Body.Close()
	// -----------------------------------------
}
