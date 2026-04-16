package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const RefreshTokenExp = 60 * 24 * time.Hour // 60 days
const JWTExp = 1 * time.Hour                // 1 hour for JTW tokens

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "cli-messenger-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}
	if !token.Valid {
		return uuid.UUID{}, errors.New("invalid token")
	}

	idStr, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuid.Parse(idStr)
}

func GetBearerToken(headers http.Header) (string, error) {
	if headers == nil {
		return "", errors.New("header doesn't exist")
	}
	bearerTokenStr := headers.Get("Authorization")
	tokenStr, _ := strings.CutPrefix(bearerTokenStr, "Bearer ")
	if tokenStr == "" {
		return "", errors.New("token doesn't exist")
	}

	return tokenStr, nil
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)

	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string, error) {
	if headers == nil {
		return "", errors.New("header doesn't exist")
	}
	bearerAPIKey := headers.Get("Authorization")
	apiKeyStr, _ := strings.CutPrefix(bearerAPIKey, "ApiKey ")
	if apiKeyStr == "" {
		return "", errors.New("key doesn't exist")
	}

	return apiKeyStr, nil
}
