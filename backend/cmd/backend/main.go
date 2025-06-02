package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var jwtSecret = []byte("secret")

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type jwtCustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func loginHandler(c echo.Context) error {
	// JSON形式でリクエストボディを取得
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	// ユーザー名とパスワードの検証（ここでは簡単な例として固定値を使用）
	if req.Username != "user" || req.Password != "password" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtCustomClaims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // トークンの有効期限を1時間に設定
			Subject:   req.Username,
		},
	})

	// 認証成功時にJWTトークンを生成
	t, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create token"})
	}
	return c.JSON(http.StatusOK, map[string]string{"token": t})
}

func main() {

	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.POST("/login", loginHandler)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// JWTが必要なグループ
	api := e.Group("/api")
	// jwt認証ミドルウェアを適用
	api.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtCustomClaims)
		},
		SigningKey: jwtSecret,
	}))

	api.GET("/profile", func(c echo.Context) error {
		// JWT認証が成功した場合、ユーザー名を取得
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		claims, ok := token.Claims.(*jwtCustomClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}
		if !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid token"})
		}
		return c.JSON(http.StatusOK, map[string]string{
			"message":  "Welcome to the API",
			"username": claims.Username,
		})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
