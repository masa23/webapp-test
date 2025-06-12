package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/masa23/webapp-test/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ログインしてRefreshトークンを生成してクッキーに設定
func Login(c echo.Context, db *gorm.DB) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errorMessage(c, "Invalid request format")
	}

	user, err := findUserByUsername(db, req.Username)
	if err != nil {
		return errorMessage(c, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return unauthorized(c)
	}

	// リフレッシュトークンを生成
	rt, err := GenerateRefreshToken(c, user.ID, db)
	if err != nil {
		return errorMessage(c, "Failed to generate refresh token: "+err.Error())
	}

	return c.JSON(http.StatusOK, rt)
}

func Logout(c echo.Context, db *gorm.DB) error {
	// リフレッシュトークンを削除
	if err := revokedRefreshToken(c, db); err != nil {
		return errorMessage(c, "Failed to logout: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"` // UNIXタイムスタンプ
}

func Refresh(c echo.Context, db *gorm.DB, jwtSecret string) error {
	rt, err := CheckRefreshToken(c, db)
	if err != nil {
		return errorMessage(c, "Invalid or expired refresh token: "+err.Error())
	}

	// expired
	expired := time.Second * 30

	// 新しいアクセストークンを生成
	newAccessToken, err := GenerateJWTToken(c, rt.UserID, rt.Token, []byte(jwtSecret), expired)
	if err != nil {
		return errorMessage(c, "Failed to generate access token: "+err.Error())
	}

	return c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken: newAccessToken,
		ExpiresAt:   time.Now().Unix() + int64(expired.Seconds()),
	})
}

func setRefreshTokenCookie(c echo.Context, token string) {
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = token
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	c.SetCookie(cookie)
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// URLセーフなBase64文字列に変換
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}

func generateRefreshToken(userID uint64, db *gorm.DB) (*model.RefreshToken, error) {
	// 32バイトのランダム値 → Base64で約43文字（URLセーフ）
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	rt := &model.RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7日間
	}

	if err := db.Create(rt).Error; err != nil {
		return nil, err
	}
	return rt, nil
}

func GenerateRefreshToken(c echo.Context, userID uint64, db *gorm.DB) (*model.RefreshToken, error) {
	rt, err := generateRefreshToken(userID, db)
	if err != nil {
		return nil, err
	}

	// クッキーにリフレッシュトークンを設定
	setRefreshTokenCookie(c, rt.Token)

	return rt, nil
}

func CheckRefreshToken(c echo.Context, db *gorm.DB) (*model.RefreshToken, error) {
	tokenStr, err := c.Cookie("refresh_token")
	if err != nil {
		return nil, err
	}

	var rt model.RefreshToken
	if err := db.Where("token = ?", tokenStr.Value).First(&rt).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	if rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	return &rt, nil
}

func deleteRefreshTokenCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = "refresh_token"
	cookie.Value = ""
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.Expires = time.Unix(0, 0) // 期限切れに設定
	c.SetCookie(cookie)
}

func revokedRefreshToken(c echo.Context, db *gorm.DB) error {
	tokenStr, err := c.Cookie("refresh_token")
	if err != nil {
		return err
	}

	// リフレッシュトークンをデータベースから削除
	if err := db.Where("token = ?", tokenStr.Value).Delete(&model.RefreshToken{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorMessage(c, "Refresh token not found")
		}
		return errorMessage(c, "Failed to revoke refresh token: "+err.Error())
	}

	// クッキーを削除
	deleteRefreshTokenCookie(c)

	return c.NoContent(http.StatusNoContent)
}
