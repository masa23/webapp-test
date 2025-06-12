package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// 共通関数
func authenticatedUser(c echo.Context) (*model.User, error) {
	userId, err := auth.JWTAuth(c)
	if err != nil || userId == nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
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

func getServerFromParam(c echo.Context) (*model.Server, error) {
	idStr := c.Param("id")
	if idStr == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Server ID is required")
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid Server ID format")
	}
	var sv model.Server
	if err := db.First(&sv, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Database error")
	}
	return &sv, nil
}

func checkOwnership(user *model.User, sv *model.Server) error {
	if user.OrganizationID != sv.OrganizationID {
		return echo.NewHTTPError(http.StatusForbidden, "Permission denied")
	}
	return nil
}

func serverActionHandler(action func(model.Server) error, successMsg string) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := authenticatedUser(c)
		if err != nil {
			return err
		}
		sv, err := getServerFromParam(c)
		if err != nil {
			return err
		}
		if err := checkOwnership(user, sv); err != nil {
			return err
		}
		if err := action(*sv); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to execute action")
		}
		return c.JSON(http.StatusOK, map[string]string{"message": successMsg})
	}
}

// ハンドラ群
func loginHandler(c echo.Context) error {
	return auth.Login(c, db, conf.RefreshToken.Duration)
}

func logoutHandler(c echo.Context) error {
	return auth.Logout(c, db)
}

func refreshHandler(c echo.Context) error {
	return auth.Refresh(c, db, conf.AccessToken.JWTSecret, conf.AccessToken.Duration)
}

func helloWorldHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "Hello, World!"})
}

func profileHandler(c echo.Context) error {
	user, err := authenticatedUser(c)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, user)
}

func getServersHandler(c echo.Context) error {
	user, err := authenticatedUser(c)
	if err != nil {
		return err
	}
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	search := c.QueryParam("search")

	resp, err := server.GetServersByOrganizationIDAndSearch(db, user.OrganizationID, search, page, pageSize)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve servers")
	}
	return c.JSON(http.StatusOK, resp)
}

func getServerHandler(c echo.Context) error {
	user, err := authenticatedUser(c)
	if err != nil {
		return err
	}
	svResp, err := server.GetServerByID(db, parseUintParam(c, "id"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Server not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve server")
	}
	if err := checkOwnership(user, &svResp.Server); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, svResp)
}

func parseUintParam(c echo.Context, name string) uint64 {
	idStr := c.Param(name)
	id, _ := strconv.ParseUint(idStr, 10, 64)
	return id
}

type wsReader struct {
	conn *websocket.Conn
}

func (r *wsReader) Read(p []byte) (int, error) {
	mt, msg, err := r.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	if mt != websocket.BinaryMessage {
		return 0, io.EOF
	}
	n := copy(p, msg)
	return n, nil
}

type wsWriter struct {
	conn *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func getServerVNCHandler(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Token is required")
	}

	userId, err := auth.JWTTokenAuth(c, token, conf.AccessToken.JWTSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	sv, err := getServerFromParam(c)
	if err != nil {
		return err
	}

	var user model.User
	if err := db.First(&user, *userId).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}
	if err := checkOwnership(&user, sv); err != nil {
		return err
	}

	port, err := server.ServerDomDisplay(*sv)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get VNC port")
	}

	vncConn, err := net.Dial("tcp", sv.HostName+":"+strconv.Itoa(port))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to VNC server")
	}

	wsConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		vncConn.Close()
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to upgrade to WebSocket")
	}

	defer func() {
		wsConn.Close()
		vncConn.Close()
	}()

	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	wsR := &wsReader{conn: wsConn}
	wsW := &wsWriter{conn: wsConn}

	go func() {
		defer cancel()
		io.Copy(vncConn, wsR) // WebSocket → VNC
	}()

	go func() {
		defer cancel()
		io.Copy(wsW, vncConn) // VNC → WebSocket
	}()

	<-ctx.Done()

	return nil
}

func main() {
	var confPath string
	flag.StringVar(&confPath, "config", "config.yaml", "Path to config file")
	flag.Parse()

	var err error
	conf, err = config.Load(confPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	db, err = gorm.Open(sqlite.Open("test.sqlite3"), &gorm.Config{})
	if err != nil {
		e.Logger.Fatal("DB connection failed:", err)
	}
	if err := model.Migrate(db); err != nil {
		e.Logger.Fatal("Migration failed:", err)
	}

	// ルーティング
	e.POST("/auth/login", loginHandler)
	e.GET("/auth/refresh", refreshHandler)
	e.POST("/auth/logout", logoutHandler)
	e.GET("/", helloWorldHandler)
	e.GET("/ws/server/:id/vnc", getServerVNCHandler)

	api := e.Group("/api")
	api.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: auth.NewJWTClaims,
		SigningKey:    []byte(conf.AccessToken.JWTSecret),
	}))

	api.GET("/profile", profileHandler)
	api.GET("/servers", getServersHandler)
	api.GET("/server/:id", getServerHandler)
	api.POST("/server/:id/power/off", serverActionHandler(server.ServerPowerOff, "Server powered off successfully"))
	api.POST("/server/:id/power/on", serverActionHandler(server.ServerPowerOn, "Server powered on successfully"))
	api.POST("/server/:id/power/reboot", serverActionHandler(server.ServerReboot, "Server rebooted successfully"))
	api.POST("/server/:id/power/force-reboot", serverActionHandler(server.ServerForceReboot, "Server force rebooted successfully"))
	api.POST("/server/:id/power/force-off", serverActionHandler(server.ServerForcePowerOff, "Server force powered off successfully"))

	e.Logger.Fatal(e.Start(":8080"))
}
