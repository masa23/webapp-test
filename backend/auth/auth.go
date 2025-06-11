package auth

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/masa23/webapp-test/model"
	"gorm.io/gorm"
)

func errorMessage(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, map[string]string{
		"error": message,
	})
}

func unauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, map[string]string{
		"error": "Unauthorized",
	})
}

func findUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("User not found: " + username)
	}
	return &user, nil
}
