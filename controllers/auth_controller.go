package controllers

import (
	"net/http"
	"ssh-key-manager/services"

	"github.com/labstack/echo/v4"
)

type authRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with username and password
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user  body   authRequest  true  "User Info"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /register [post]
func Register(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	if err := services.RegisterUser(req.Username, req.Password); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "User registered successfully"})
}

// Login godoc
// @Summary Log in a user
// @Description Log in with username and password to get a JWT
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   user  body   authRequest  true  "Credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /login [post]
func Login(c echo.Context) error {
	var req authRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	token, err := services.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid credentials"})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
