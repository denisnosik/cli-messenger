package server

import (
	"bytes"
	"encoding/json/v2"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func createChatRequest(t *testing.T, targetNickname, userToken string) *http.Response {
	t.Helper()

	body, err := json.Marshal(struct {
		TargetNickname string `json:"target_nickname"`
	}{
		TargetNickname: targetNickname,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/chats", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+userToken)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	return res
}

func makeFriends(t *testing.T, a, b testUser) {
	t.Helper()
	doRequestExpect(t, sendFriendRequest(t, b.nickname, a.token), http.StatusCreated)
	doRequestExpect(t, sendFriendRequest(t, a.nickname, b.token), http.StatusOK)
}

func TestChat(t *testing.T) {
	withLeakCheck(t)

	t.Run("create chat with friend", func(t *testing.T) {
		user := createAndLoginUser(t)
		friend := createAndLoginUser(t)
		makeFriends(t, user, friend)

		doRequestExpect(t, createChatRequest(t, friend.nickname, user.token), http.StatusCreated)
	})

	t.Run("get existing chat returns ok", func(t *testing.T) {
		user := createAndLoginUser(t)
		friend := createAndLoginUser(t)
		makeFriends(t, user, friend)

		doRequestExpect(t, createChatRequest(t, friend.nickname, user.token), http.StatusCreated)
		doRequestExpect(t, createChatRequest(t, friend.nickname, user.token), http.StatusOK)
	})

	t.Run("chat with self", func(t *testing.T) {
		user := createAndLoginUser(t)
		doRequestExpectError(t, createChatRequest(t, user.nickname, user.token), http.StatusBadRequest)
	})

	t.Run("chat with nonexistent user", func(t *testing.T) {
		user := createAndLoginUser(t)
		doRequestExpectError(t, createChatRequest(t, uniqueNickname(), user.token), http.StatusNotFound)
	})

	t.Run("chat without friendship", func(t *testing.T) {
		user := createAndLoginUser(t)
		stranger := createAndLoginUser(t)
		doRequestExpectError(t, createChatRequest(t, stranger.nickname, user.token), http.StatusBadRequest)
	})

	t.Run("chat with pending friend request", func(t *testing.T) {
		user := createAndLoginUser(t)
		friend := createAndLoginUser(t)
		doRequestExpect(t, sendFriendRequest(t, friend.nickname, user.token), http.StatusCreated)

		doRequestExpectError(t, createChatRequest(t, friend.nickname, user.token), http.StatusBadRequest)
	})

	t.Run("create chat malformed body", func(t *testing.T) {
		user := createAndLoginUser(t)

		req, err := http.NewRequest("POST", baseURL+"/api/chats", bytes.NewBufferString("{invalid json"))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+user.token)

		res, err := testClient.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		require.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
