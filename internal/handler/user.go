package handler

import (
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// GetCurrentUser는 현재 로그인한 사용자 정보를 반환합니다.
func GetCurrentUser(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	user, err := service.GetUserByID(userID)
	if err != nil {
		return NotFoundResponse(c, err.Error())
	}

	// SSH 키 존재 여부 확인
	hasSSHKey := service.HasUserSSHKey(userID)

	responseData := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"has_ssh_key": hasSSHKey,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	return SuccessResponse(c, responseData)
}

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	var req model.UserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	user, err := service.UpdateUserProfile(userID, req)
	if err != nil {
		return BadRequestResponse(c, err.Error())
	}

	return SuccessWithMessageResponse(c, "프로필이 업데이트되었습니다", user)
}
