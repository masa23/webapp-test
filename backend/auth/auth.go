package auth

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/masa23/webapp-test/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	MinAccessTokenLength = 64
	MinSecretTokenLength = 72
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
		return nil, unauthorized()
	}
	return extractUserIDFromClaims(token.Claims)
}

// JWT: トークンとシークレットからユーザーIDを取得
func JWTTokenAuth(tokenStr, jwtSecret string) (*uint64, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, unauthorized()
	}
	return extractUserIDFromClaims(token.Claims)
}

func extractUserIDFromClaims(claims jwt.Claims) (*uint64, error) {
	c, ok := claims.(*jwtCustomClaims)
	if !ok || c.Subject == "" {
		return nil, unauthorized()
	}
	id, err := strconv.ParseUint(c.Subject, 10, 64)
	if err != nil {
		return nil, unauthorized()
	}
	return &id, nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ログインしてJWTを返す
func Login(c echo.Context, db *gorm.DB, jwtSecret []byte) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return badRequest("Invalid request format")
	}

	user, err := findUserByUsername(db, req.Username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return unauthorized()
	}

	token, err := createJWTToken(user.ID, user.Username, jwtSecret)
	if err != nil {
		return internalError("Could not create token")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func createJWTToken(userID uint64, username string, secret []byte) (string, error) {
	claims := jwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// APIキー認証: ヘッダーから認証情報取得・検証
func APIKeyAuth(c echo.Context, db *gorm.DB) (*uint64, error) {
	accessToken, secretToken, err := parseAPIKeyHeader(c)
	if err != nil {
		return nil, err
	}

	// APIトークン取得
	apiToken := &model.APIToken{}
	if err := db.Where("access_token = ?", accessToken).First(apiToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// シークレットトークン検証
	if err := bcrypt.CompareHashAndPassword([]byte(apiToken.SecretToken), []byte(secretToken)); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid secret token")
	}

	// ユーザー取得
	user := &model.User{}
	if err := db.Where("api_token_id = ?", apiToken.ID).First(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found for API key")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return &user.ID, nil
}

func parseAPIKeyHeader(c echo.Context) (string, string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", "", unauthorized()
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "apikey" {
		return "", "", unauthorized()
	}

	tokens := strings.SplitN(parts[1], ":", 2)
	if len(tokens) != 2 || len(tokens[0]) < MinAccessTokenLength || len(tokens[1]) < MinSecretTokenLength {
		return "", "", unauthorized()
	}
	return tokens[0], tokens[1], nil
}

func findUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, dbError(err)
	}
	return &user, nil
}

// エラーヘルパー
func unauthorized() error {
	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
}

func badRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func internalError(msg string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, msg)
}

func dbError(err error) error {
	if err == gorm.ErrRecordNotFound {
		return unauthorized()
	}
	return internalError("Database error")
}
