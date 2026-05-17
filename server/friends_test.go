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

func TestFriendRequest(t *testing.T) {
	// Test: Valid friend request (handlerFriends)
	user := createAndLoginUser(t)
	assert.NotNil(t, user.nickname)
	assert.NotNil(t, user.token)

	targetNickname := uniqueNickname()
	res := registerUser(t, targetNickname, "000000")
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	res.Body.Close()

	body, err := json.Marshal(friendsRequest{
		TargetNickname: targetNickname,
	})
	require.NoError(t, err)

	req, err := http.NewRequest("POST", baseURL+"/api/friends", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+user.token)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	res.Body.Close()
	// -------------------------------

	// Test: Valid accept friend request
	res = loginUser(t, targetNickname, "000000")

	require.Equal(t, http.StatusOK, res.StatusCode)

	decoder := json.NewDecoder(res.Body)
	var result User
	err = decoder.Decode(&result)
	require.NoError(t, err)

	res.Body.Close()

	body, err = json.Marshal(friendsRequest{
		TargetNickname: user.nickname,
	})
	require.NoError(t, err)

	req, err = http.NewRequest("POST", baseURL+"/api/friends", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer "+result.Token)

	res, err = testClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)

	res.Body.Close()
	// -------------------------------
}
