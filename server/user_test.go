package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
	"uuid"

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
	doRequestExpect(t, registerUser(t, nickname, "000000"), http.StatusCreated)

	// login
	res := loginUser(t, nickname, "000000")
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	var result User
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))

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
	return fmt.Sprintf("test_user_%d", uuid.New())
}

func TestUserRegister(t *testing.T) {
	withLeakCheck(t)

	t.Run("register new user", func(t *testing.T) {
		nickname := uniqueNickname()

		res := registerUser(t, nickname, "000000")
		defer res.Body.Close()
		require.Equal(t, http.StatusCreated, res.StatusCode)

		var result registerResponse
		require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
		require.Equal(t, nickname, result.Nickname)
	})

	t.Run("register new user missing nickname", func(t *testing.T) {
		doRequestExpect(t, registerUser(t, "", "000000"), http.StatusBadRequest)
	})

	t.Run("register new user missing password", func(t *testing.T) {
		doRequestExpect(t, registerUser(t, uniqueNickname(), ""), http.StatusBadRequest)
	})

	t.Run("register new user bad request no nickname and no password", func(t *testing.T) {
		doRequestExpect(t, registerUser(t, "", ""), http.StatusBadRequest)
	})

	t.Run("register new user status conflict", func(t *testing.T) {
		nickname := uniqueNickname()

		doRequestExpect(t, registerUser(t, nickname, "000000"), http.StatusCreated)
		doRequestExpect(t, registerUser(t, nickname, "000000"), http.StatusConflict)
	})

	t.Run("register new user malformed body", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURL+"/api/register", bytes.NewBufferString("{invalid json"))
		require.NoError(t, err)

		res, err := testClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestUserLogin(t *testing.T) {
	withLeakCheck(t)

	t.Run("login user", func(t *testing.T) {
		user := createAndLoginUser(t)

		require.NotEmpty(t, user.nickname)
		require.NotEmpty(t, user.token)
	})

	t.Run("login user status unauthorized user doesn't exist", func(t *testing.T) {
		doRequestExpectError(t, loginUser(t, uniqueNickname(), "000000"), http.StatusUnauthorized)
	})

	t.Run("login user wrong password", func(t *testing.T) {
		nickname := uniqueNickname()

		doRequestExpect(t, registerUser(t, nickname, "000000"), http.StatusCreated)

		doRequestExpectError(t, loginUser(t, nickname, "wrong-password"), http.StatusUnauthorized)
	})

	t.Run("login user missing password", func(t *testing.T) {
		doRequestExpectError(t, loginUser(t, uniqueNickname(), ""), http.StatusBadRequest)
	})

	t.Run("login user missing nickname", func(t *testing.T) {
		doRequestExpectError(t, loginUser(t, "", "000000"), http.StatusBadRequest)
	})

	t.Run("login user malformed body", func(t *testing.T) {
		req, err := http.NewRequest("POST", baseURL+"/api/login", bytes.NewBufferString("{invalid json"))
		require.NoError(t, err)

		res, err := testClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
