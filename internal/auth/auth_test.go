package auth

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthPassword(t *testing.T) {
	// Test: Valid password
	hashPwd, err := HashPassword("HelloWorld123")
	require.NoError(t, err)
	require.NotNil(t, hashPwd)
	match, err := CheckPasswordHash("HelloWorld123", hashPwd)
	require.NoError(t, err)
	require.True(t, match)

	// Test: Wrong password no errors
	hashPwd, err = HashPassword("HelloWorld123")
	require.NoError(t, err)
	require.NotNil(t, hashPwd)
	match, err = CheckPasswordHash("helloworld321", hashPwd)
	require.NoError(t, err)
	require.False(t, match)
}

func TestAuthJWT(t *testing.T) {
	// Test: Valid JWT
	id := uuid.NewV4()
	jwt, err := MakeJWT(id, "secret", 10*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, jwt)
	idFromJWT, err := ValidateJWT(jwt, "secret")
	require.NoError(t, err)
	require.NotNil(t, idFromJWT)
	assert.Equal(t, id, idFromJWT)

	// Test: Wrong JWT secret
	id = uuid.NewV4()
	jwt, err = MakeJWT(id, "secret", 10*time.Minute)
	require.NoError(t, err)
	require.NotNil(t, jwt)
	idFromJWT, err = ValidateJWT(jwt, "wrong_secret")
	require.Error(t, err)
	assert.NotEqual(t, id, idFromJWT)

	// Test: JWT expired
	id = uuid.NewV4()
	jwt, err = MakeJWT(id, "secret", 1*time.Millisecond)
	require.NoError(t, err)
	require.NotNil(t, jwt)
	time.Sleep(5 * time.Millisecond)
	idFromJWT, err = ValidateJWT(jwt, "secret")
	require.Error(t, err)
	assert.NotEqual(t, id, idFromJWT)
}

func TestGetBearerToken(t *testing.T) {
	// Test: Valid bearer token
	headers := http.Header{}
	headers.Set("Authorization", "Bearer cool.jwt.token")
	token, err := GetBearerToken(headers)
	require.NoError(t, err)
	require.NotNil(t, token)
	assert.Equal(t, token, "cool.jwt.token")

	// Test: Empty headers
	headers = http.Header{}
	token, err = GetBearerToken(headers)
	require.Error(t, err)
	assert.Equal(t, token, "")

	// Test: Empty token
	headers = http.Header{}
	headers.Set("Authorization", "Bearer ")
	token, err = GetBearerToken(headers)
	require.Error(t, err)
	assert.Equal(t, token, "")
}
