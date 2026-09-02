package auth

import (
	"net/http"
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("hash password valid password", func(t *testing.T) {
		hashPwd, err := HashPassword("000000")
		require.NoError(t, err)
		require.NotEmpty(t, hashPwd)

		match, err := CheckPasswordHash("000000", hashPwd)
		require.NoError(t, err)
		require.True(t, match)
	})

	t.Run("auth wrong password but no errors", func(t *testing.T) {
		hashPwd, err := HashPassword("000000")
		require.NoError(t, err)
		require.NotEmpty(t, hashPwd)

		match, err := CheckPasswordHash("111111", hashPwd)
		require.NoError(t, err)
		require.False(t, match)
	})
}

func TestAuthJWT(t *testing.T) {
	t.Run("auth valid jwt", func(t *testing.T) {
		id := uuid.NewV4()
		jwt, err := MakeJWT(id, "secret", 10*time.Minute)
		require.NoError(t, err)
		require.NotEmpty(t, jwt)

		idFromJWT, err := ValidateJWT(jwt, "secret")
		require.NoError(t, err)
		require.Equal(t, id, idFromJWT)
	})

	t.Run("auth wrong jwt secret", func(t *testing.T) {
		id := uuid.NewV4()
		jwt, err := MakeJWT(id, "secret", 10*time.Minute)
		require.NoError(t, err)

		_, err = ValidateJWT(jwt, "wrong_secret")
		require.Error(t, err)
	})

	t.Run("auth jwt expired", func(t *testing.T) {
		id := uuid.NewV4()
		jwt, err := MakeJWT(id, "secret", 1*time.Millisecond)
		require.NoError(t, err)

		time.Sleep(50 * time.Millisecond)

		_, err = ValidateJWT(jwt, "secret")
		require.Error(t, err)
	})
}

func TestBearerToken(t *testing.T) {
	t.Run("bearer token valid", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer cool.jwt.token")

		token, err := GetBearerToken(headers)
		require.NoError(t, err)
		require.Equal(t, "cool.jwt.token", token)
	})

	t.Run("bearer token empty headers", func(t *testing.T) {
		headers := http.Header{}
		token, err := GetBearerToken(headers)
		require.Error(t, err)
		require.Equal(t, "", token)
	})

	t.Run("bearer token empty token", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Authorization", "Bearer ")
		token, err := GetBearerToken(headers)
		require.Error(t, err)
		require.Equal(t, "", token)
	})
}
