package server

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type testUser struct {
	nickname string
	token    string
}

const baseURL = "http://localhost:8080"

func createAndLoginUser(t *testing.T) testUser {
	t.Helper()
	nickname := uniqueNickname()

	// register
	res := registerUser(t, nickname, "000000")
	res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	// login
	res = loginUser(t, nickname, "000000")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var result User
	err := decoder.Decode(&result)
	require.NoError(t, err)

	return testUser{
		nickname: result.Nickname,
		token:    result.Token,
	}
}

func registerUser(t *testing.T, nickname, password string) *http.Response {
	t.Helper()

	body, err := json.Marshal(registerRequest{
		Nickname: nickname,
		Password: password,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/register", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	return res
}

func loginUser(t *testing.T, nickname, password string) *http.Response {
	t.Helper()

	body, err := json.Marshal(registerRequest{
		Nickname: nickname,
		Password: password,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/login", bytes.NewBuffer(body))
	require.NoError(t, err)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	return res
}

func uniqueNickname() string {
	return fmt.Sprintf("test_user_%d", time.Now().UnixNano())
}

func TestUserRegister(t *testing.T) {
	// Test: Valid register (handlerCreateUser)
	nickname := uniqueNickname()
	res := registerUser(t, nickname, "000000")

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var resultRegister registerResponse
	err := decoder.Decode(&resultRegister)
	require.NoError(t, err)

	assert.Equal(t, nickname, resultRegister.Nickname)
	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (no nickname)
	res = registerUser(t, "", "000000")

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	var errRes errorResponse
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)
	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (no password)
	nickname = uniqueNickname()
	res = registerUser(t, nickname, "")

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)
	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid register (user already exist)
	nickname = uniqueNickname()
	res = registerUser(t, nickname, "000000")
	res.Body.Close()

	res = registerUser(t, nickname, "000000")

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
	user := createAndLoginUser(t)

	assert.NotNil(t, user.nickname)
	assert.NotNil(t, user.token)
	// -----------------------------------------

	// Test: Invalid login user (user doesn't exist)
	nickname := uniqueNickname()
	res := loginUser(t, nickname, "000000")

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var errRes errorResponse
	err := decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)

	res.Body.Close()
	// -----------------------------------------

	// Test: Invalid login user (wrong password)
	nickname = uniqueNickname()
	res = registerUser(t, nickname, "000000")
	res.Body.Close()
	require.Equal(t, http.StatusCreated, res.StatusCode)

	res = loginUser(t, nickname, "wrong-password")

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)

	res.Body.Close()
	// -----------------------------------------
}
