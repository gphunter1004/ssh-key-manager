package controllers

import (
	"net/http"
	"ssh-key-manager/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// userIDFromToken extracts user ID from the JWT token in the context.
func userIDFromToken(c echo.Context) (uint, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	return userID, nil
}

// CreateKey creates or regenerates an SSH key pair for the authenticated user.
func CreateKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	key, err := services.GenerateSSHKeyPair(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate key pair"})
	}

	return c.JSON(http.StatusOK, key)
}

// GetKey retrieves the SSH key pair for the authenticated user.
func GetKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	key, err := services.GetKeyByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, key)
}

// DeleteKey deletes the SSH key pair for the authenticated user.
func DeleteKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	if err := services.DeleteKeyByUserID(userID); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Key deleted successfully"})
}
