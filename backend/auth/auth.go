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

// JWTでの認証を行う
// Authorization: Bearer <token>
func JWTAuth(c echo.Context) (userId *uint64, err error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}
	claims, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}
	if claims.Subject == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}
	// ユーザーIDを取得
	userIdUint, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}
	return &userIdUint, nil
}

func JWTTokenAuth(token string, jwtSecret string) (userId *uint64, err error) {
	t, err := jwt.ParseWithClaims(token, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方法の検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}

	// クレームを取得
	claims, ok := t.Claims.(*jwtCustomClaims)
	if !ok || !t.Valid {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}

	if claims.Subject == "" {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}

	// ユーザーIDを取得
	userIdUint, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid JWT token")
	}
	return &userIdUint, nil
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// login処理をする JWTトークンを生成して返す
func Login(c echo.Context, db *gorm.DB, jwtSecret []byte) error {
	var loginReq LoginRequest
	if err := c.Bind(&loginReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// ユーザー名とパスワードでユーザーを検索
	var user model.User
	if err := db.Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// パスワードをハッシュ化して比較
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid username or password")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Error comparing password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // トークンの有効期限を1時間に設定
			Subject:   strconv.FormatUint(user.ID, 10),                   // ユーザーIDをサブジェクトとして設定
			IssuedAt:  jwt.NewNumericDate(time.Now()),                    // トークンの発行日時を設定
			ID:        user.Username,                                     // ユーザー名をIDとして設定
		},
	})

	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not create token")
	}

	return c.JSON(http.StatusOK, map[string]string{"token": t})
}

// APIキーを取得する
// Authorization: ApiKey <access_token>:<secret_token>
func getAPIToken(c echo.Context) (accessToken, secretToken string, err error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Authorization header is required")
	}

	// AuthorizationヘッダーからAPIキーを取得
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "apikey" {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
	}
	parts = strings.SplitN(parts[1], ":", 2)
	if len(parts) != 2 {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
	}
	accessToken = parts[0]
	secretToken = parts[1]
	if accessToken == "" || secretToken == "" {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Access token and secret token are required")
	}
	if len(accessToken) < MinAccessTokenLength || len(secretToken) < MinSecretTokenLength {
		return "", "", echo.NewHTTPError(http.StatusUnauthorized, "Invalid token length")
	}
	return accessToken, secretToken, nil
}

// APIキーでの認証を行う
func APIKeyAuth(c echo.Context, db *gorm.DB) (userId *uint64, err error) {
	accessToken, secretToken, err := getAPIToken(c)
	if err != nil {
		return nil, err
	}

	apiTokenRecord := &model.APIToken{}
	if err := db.Where("access_token = ?", accessToken).First(apiTokenRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// APIキーのシークレットトークンを検証
	err = bcrypt.CompareHashAndPassword([]byte(apiTokenRecord.SecretToken), []byte(secretToken))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid secret token")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Error comparing secret token")
	}

	// APIキーに紐づくユーザーを取得
	user := &model.User{}
	if err := db.Where("api_token_id = ?", apiTokenRecord.ID).First(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "User not found for API key")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	// 認証成功時はユーザー情報を返す
	return &user.ID, nil
}
