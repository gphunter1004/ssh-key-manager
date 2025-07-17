package controllers

import (
	"net/http"
	"ssh-key-manager/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetUsers godoc
// @Summary Get all users list
// @Description Get all users with basic information (admin only)
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [get]
func GetUsers(c echo.Context) error {
	users, err := services.GetAllUsers()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"users": users,
		"count": len(users),
	})
}

// GetUserDetail godoc
// @Summary Get user detail with SSH key information
// @Description Get detailed user information including SSH key data
// @Tags users
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /users/{id} [get]
func GetUserDetail(c echo.Context) error {
	// URL 파라미터에서 사용자 ID 추출
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "유효하지 않은 사용자 ID입니다"})
	}

	// 현재 로그인한 사용자 확인 (자신의 정보만 조회 가능하도록 제한할 수도 있음)
	currentUserID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	userDetail, err := services.GetUserDetailWithKey(uint(userID))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	// 개인정보 보호: 자신의 정보만 볼 수 있도록 제한 (선택사항)
	if currentUserID != uint(userID) {
		// 다른 사용자 정보 조회 시 민감한 정보 제거
		userDetail.HasSSHKey = userDetail.SSHKey != nil
		userDetail.SSHKey = nil // 다른 사용자의 키 정보는 숨김
	}

	return c.JSON(http.StatusOK, userDetail)
}

// GetCurrentUser godoc
// @Summary Get current logged-in user information
// @Description Get current user's detailed information including SSH key
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [get]
func GetCurrentUser(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	userDetail, err := services.GetUserDetailWithKey(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, userDetail)
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags users
// @Accept  json
// @Produce  json
// @Param   profile  body   services.UserProfileUpdate  true  "Profile data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [put]
func UpdateUserProfile(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	var updateData services.UserProfileUpdate
	if err := c.Bind(&updateData); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request body"})
	}

	updatedUser, err := services.UpdateUserProfile(userID, updateData)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "프로필이 업데이트되었습니다",
		"user":    updatedUser,
	})
}
