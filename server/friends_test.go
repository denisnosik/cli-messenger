package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestFriend(t *testing.T) {
	// Test: Valid friend request (handlerFriends)
	user := createAndLoginUser(t)
	assert.NotNil(t, user.nickname)
	assert.NotNil(t, user.token)

	targetNickname := uniqueNickname()
	res := registerUser(t, targetNickname, "000000")
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	res.Body.Close()

	res = sendFriendRequest(t, targetNickname, user.token)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	res.Body.Close()
	// -------------------------------

	// Test: Valid accept friend request
	res = loginUser(t, targetNickname, "000000")

	require.Equal(t, http.StatusOK, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var result User
	err := decoder.Decode(&result)
	require.NoError(t, err)
	res.Body.Close()

	res = sendFriendRequest(t, user.nickname, result.Token)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	res.Body.Close()
	// -------------------------------

	// Test: Invalid user doesn't exist
	user = createAndLoginUser(t)
	assert.NotNil(t, user.nickname)
	assert.NotNil(t, user.token)

	targetNickname = uniqueNickname()

	res = sendFriendRequest(t, targetNickname, user.token)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	decoder = json.NewDecoder(res.Body)
	var errRes errorResponse
	err = decoder.Decode(&errRes)
	require.NoError(t, err)

	assert.NotNil(t, errRes.Error)

	res.Body.Close()
	// -------------------------------
}
