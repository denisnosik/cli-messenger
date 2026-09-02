package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type friendsRequest struct {
	TargetNickname string `json:"target_nickname"`
}

func sendFriendRequest(t *testing.T, targetNickname, userToken string) *http.Response {
	t.Helper()

	body, err := json.Marshal(friendsRequest{
		TargetNickname: targetNickname,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/friends", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+userToken)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	return res
}

func deleteFriendshipRequest(t *testing.T, targetNickname, userToken string) *http.Response {
	t.Helper()

	body, err := json.Marshal(friendsRequest{
		TargetNickname: targetNickname,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("DELETE", baseURL+"/api/friends", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+userToken)

	res, err := testClient.Do(req)
	require.NoError(t, err)

	return res
}

func doRequestExpect(t *testing.T, res *http.Response, wantStatus int) {
	t.Helper()
	defer res.Body.Close()
	require.Equal(t, wantStatus, res.StatusCode)
}

func doRequestExpectError(t *testing.T, res *http.Response, wantStatus int) errorResponse {
	t.Helper()
	defer res.Body.Close()
	require.Equal(t, wantStatus, res.StatusCode)

	var errRes errorResponse
	require.NoError(t, json.NewDecoder(res.Body).Decode(&errRes))
	require.NotEmpty(t, errRes.Error)
	return errRes
}

func TestFriend(t *testing.T) {
	withLeakCheck(t)

	t.Run("send friend request", func(t *testing.T) {
		user := createAndLoginUser(t)
		require.NotEmpty(t, user.nickname)
		require.NotEmpty(t, user.token)

		target := uniqueNickname()

		doRequestExpect(t, registerUser(t, target, "000000"), http.StatusCreated)
		doRequestExpect(t, sendFriendRequest(t, target, user.token), http.StatusCreated)
	})

	t.Run("accept friend request", func(t *testing.T) {
		user := createAndLoginUser(t)
		target := createAndLoginUser(t)

		doRequestExpect(t, sendFriendRequest(t, target.nickname, user.token), http.StatusCreated)
		doRequestExpect(t, sendFriendRequest(t, user.nickname, target.token), http.StatusOK)
	})

	t.Run("delete friendship", func(t *testing.T) {
		user := createAndLoginUser(t)
		target := createAndLoginUser(t)

		doRequestExpect(t, sendFriendRequest(t, target.nickname, user.token), http.StatusCreated)
		doRequestExpect(t, sendFriendRequest(t, user.nickname, target.token), http.StatusOK)

		doRequestExpect(t, deleteFriendshipRequest(t, target.nickname, user.token), http.StatusNoContent)
	})

	t.Run("delete friendship user not found", func(t *testing.T) {
		user := createAndLoginUser(t)

		doRequestExpect(t, deleteFriendshipRequest(t, uniqueNickname(), user.token), http.StatusNotFound)
	})

	t.Run("send friend request user not found", func(t *testing.T) {
		user := createAndLoginUser(t)

		doRequestExpectError(t, sendFriendRequest(t, uniqueNickname(), user.token), http.StatusNotFound)
	})
}
