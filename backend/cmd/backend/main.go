package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/masa23/webapp-test/auth"
	"github.com/masa23/webapp-test/config"
	"github.com/masa23/webapp-test/model"
	"github.com/masa23/webapp-test/server"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var conf *config.Config

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(c echo.Context) error {
	return auth.Login(c, db, []byte(conf.JWTSecret))
}

func helloWorldHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Hello, World!",
	})
}

func authcationUser(c echo.Context) (*model.User, error) {
	userId, err := auth.JWTAuth(c)
	if err != nil {
		return nil, err
	}
	if userId == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in JWT token")
	}

	var user model.User
	if err := db.First(&user, *userId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	return &user, nil
}

func getServersHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	//  pageとpageSizeはクエリパラメータから取得
	page := c.QueryParam("page")
	pageSize := c.QueryParam("pageSize")
	if page == "" {
		page = "1" // デフォルトページ
	}
	if pageSize == "" {
		pageSize = "10" // デフォルトページサイズ
	}
	// クエリパラメータを整数に変換
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page number")
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid page size")
	}

	response, err := server.GetServersByOrganizationID(db, user.OrganizationID, pageInt, pageSizeInt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve servers")
	}

	return c.JSON(http.StatusOK, response)
}

func getServerHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	response, err := server.GetServerByID(db, serverId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve server")
	}

	if response.Server.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to access this server")
	}

	return c.JSON(http.StatusOK, response)
}

func profileHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func postServerPowerOffHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	var sv model.Server
	if err := db.First(&sv, serverId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if sv.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to power off this server")
	}

	if err := server.ServerPowerOff(sv); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to power off server")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Server powered off successfully"})
}

func postServerPowerOnHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	var sv model.Server
	if err := db.First(&sv, serverId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if sv.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to power on this server")
	}

	if err := server.ServerPowerOn(sv); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to power on server")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Server powered on successfully"})
}

func postServerPowerRebootHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	var sv model.Server
	if err := db.First(&sv, serverId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if sv.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to reboot this server")
	}

	if err := server.ServerReboot(sv); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reboot server")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Server rebooted successfully"})
}

func postServerPowerForceRebootHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	var sv model.Server
	if err := db.First(&sv, serverId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if sv.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to force reboot this server")
	}

	if err := server.ServerForceReboot(sv); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to force reboot server")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Server force rebooted successfully"})
}

func postServerPowerForceOffHandler(c echo.Context) error {
	user, err := authcationUser(c)
	if err != nil {
		return err
	}

	serverIdStr := c.Param("id")
	if serverIdStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}

	serverId, err := strconv.ParseUint(serverIdStr, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}

	var sv model.Server
	if err := db.First(&sv, serverId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}

	if sv.OrganizationID != user.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to force off this server")
	}

	if err := server.ServerForcePowerOff(sv); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to force power off server")
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Server force powered off successfully"})
}

func main() {
	var err error
	var confPath string
	flag.StringVar(&confPath, "config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	conf, err = config.Load(confPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	e := echo.New()

	// middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// DB接続の初期化
	db, err = gorm.Open(sqlite.Open("test.sqlite3"), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal("Failed to connect to database:", err)
	}
	// マイグレーション
	if err := model.Migrate(db); err != nil {
		e.Logger.Fatal("Failed to migrate database:", err)
	}

	e.POST("/login", loginHandler)
	e.GET("/", helloWorldHandler)

	// JWTが必要なグループ
	api := e.Group("/api")
	// jwt認証ミドルウェアを適用
	api.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: auth.NewJWTClaims,
		SigningKey:    []byte(conf.JWTSecret),
	}))

	api.GET("/profile", profileHandler)
	api.GET("/servers", getServersHandler)
	api.GET("/server/:id", getServerHandler)
	api.POST("/server/:id/power/off", postServerPowerOffHandler)
	api.POST("/server/:id/power/on", postServerPowerOnHandler)
	api.POST("/server/:id/power/reboot", postServerPowerRebootHandler)
	api.POST("/server/:id/power/force-reboot", postServerPowerForceRebootHandler)
	api.POST("/server/:id/power/force-off", postServerPowerForceOffHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
