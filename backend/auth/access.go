package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type jwtCustomClaims struct {
	jwt.RegisteredClaims
}

func NewJWTClaims(c echo.Context) jwt.Claims {
	return new(jwtCustomClaims)
}

// JWT: Context からユーザーIDを取得
func JWTAuth(c echo.Context) (*uint64, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, unauthorized(c)
	}
	return extractUserIDFromClaims(c, token.Claims)
}

// JWT: トークンとシークレットからユーザーIDを取得
func JWTTokenAuth(c echo.Context, tokenStr, jwtSecret string) (*uint64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, unauthorized(c)
	}
	return extractUserIDFromClaims(c, token.Claims)
}

func extractUserIDFromClaims(c echo.Context, claims jwt.Claims) (*uint64, error) {
	jwtClaims, ok := claims.(*jwtCustomClaims)
	if !ok || jwtClaims.Subject == "" {
		return nil, unauthorized(c)
	}
	id, err := strconv.ParseUint(jwtClaims.Subject, 10, 64)
	if err != nil {
		return nil, unauthorized(c)
	}
	return &id, nil
}

func createJWTToken(userID uint64, username string, secret []byte, expired time.Duration) (string, error) {
	claims := jwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expired)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GenerateJWTToken(c echo.Context, userID uint64, username string, jwtSecret []byte, expired time.Duration) (string, error) {
	token, err := createJWTToken(userID, username, jwtSecret, expired)
	if err != nil {
		return "", errorMessage(c, "Failed to create JWT token: "+err.Error())
	}

	// JWTトークンをコンテキストに設定
	c.Set("user", &jwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: username,
		},
	})

	return token, nil
}
