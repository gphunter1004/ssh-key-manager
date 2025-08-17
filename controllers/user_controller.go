package controllers

import (
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/types"
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
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.ListResponse(c, users, len(users))
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
		return helpers.BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 현재 로그인한 사용자 확인
	currentUserID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 관리자인지 확인
	isAdmin := services.IsUserAdmin(currentUserID)

	userDetail, err := services.GetUserDetailWithKey(uint(userID))
	if err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	// 개인정보 보호: 자신의 정보이거나 관리자만 전체 정보 조회 가능
	if currentUserID != uint(userID) && !isAdmin {
		// 다른 사용자 정보 조회 시 민감한 정보 제거 (비관리자)
		userDetail.HasSSHKey = userDetail.SSHKey != nil
		userDetail.SSHKey = nil // 다른 사용자의 키 정보는 숨김
	}

	return helpers.SuccessResponse(c, userDetail)
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
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	userDetail, err := services.GetUserDetailWithKey(userID)
	if err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, userDetail)
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update current user's profile information
// @Tags users
// @Accept  json
// @Produce  json
// @Param   profile  body   types.UserProfileUpdate  true  "Profile data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /users/me [put]
func UpdateUserProfile(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var updateData types.UserProfileUpdate
	if err := c.Bind(&updateData); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	updatedUser, err := services.UpdateUserProfile(userID, updateData)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "프로필이 업데이트되었습니다", updatedUser)
}

// GetUserStats godoc
// @Summary Get user statistics
// @Description Get overall user statistics including SSH key coverage
// @Tags users
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/stats [get]
func GetUserStats(c echo.Context) error {
	// 권한 확인 (관리자 전용 기능으로 확장 가능)
	_, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	stats, err := services.GetUserStats()
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, stats)
}

// AdminRequired는 관리자 권한을 확인하는 미들웨어입니다.
func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := userIDFromToken(c)
		if err != nil {
			return helpers.UnauthorizedResponse(c, "Invalid token")
		}

		if !services.IsUserAdmin(userID) {
			return helpers.UnauthorizedResponse(c, "관리자 권한이 필요합니다")
		}

		return next(c)
	}
}

// UpdateUserRole godoc
// @Summary Update user role (Admin only)
// @Description Update a user's role - only accessible by administrators
// @Tags admin
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "User ID"
// @Param   role body      types.UserRoleUpdateRequest  true  "Role data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id}/role [put]
func UpdateUserRole(c echo.Context) error {
	adminUserID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 대상 사용자 ID 추출
	userIDParam := c.Param("id")
	targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 요청 바디 파싱
	var req types.UserRoleUpdateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	// 권한 변경 실행
	if err := services.UpdateUserRole(adminUserID, uint(targetUserID), req.Role); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	// 변경된 사용자 정보 조회
	userDetail, err := services.GetUserDetailWithKey(uint(targetUserID))
	if err != nil {
		return helpers.InternalServerErrorResponse(c, "사용자 정보 조회 실패")
	}

	return helpers.SuccessWithMessageResponse(c, "사용자 권한이 변경되었습니다", userDetail)
}

// GetAdminStats godoc
// @Summary Get admin statistics
// @Description Get comprehensive system statistics (Admin only)
// @Tags admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/stats [get]
func GetAdminStats(c echo.Context) error {
	stats, err := services.GetAdminStats()
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, stats)
}

// GetAllUsersAdmin godoc
// @Summary Get all users with full details (Admin only)
// @Description Get all users including role information - admin version with more details
// @Tags admin
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /admin/users [get]
func GetAllUsersAdmin(c echo.Context) error {
	users, err := services.GetAllUsers()
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	// 관리자용 추가 정보 포함
	responseData := map[string]interface{}{
		"users": users,
		"count": len(users),
		"summary": map[string]interface{}{
			"total":  len(users),
			"admins": countUsersByRole(users, "admin"),
			"users":  countUsersByRole(users, "user"),
		},
	}

	return helpers.SuccessResponse(c, responseData)
}

// DeleteUser godoc
// @Summary Delete a user (Admin only)
// @Description Delete a user account - only accessible by administrators
// @Tags admin
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/users/{id} [delete]
func DeleteUser(c echo.Context) error {
	adminUserID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 대상 사용자 ID 추출
	userIDParam := c.Param("id")
	targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 자신을 삭제하려는지 확인
	if adminUserID == uint(targetUserID) {
		return helpers.BadRequestResponse(c, "자신의 계정은 삭제할 수 없습니다")
	}

	// 사용자 삭제 실행
	if err := services.DeleteUser(uint(targetUserID)); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "사용자가 삭제되었습니다", nil)
}

// countUsersByRole은 특정 권한을 가진 사용자 수를 계산합니다.
func countUsersByRole(users []types.UserInfo, role string) int {
	count := 0
	for _, user := range users {
		if user.Role == role {
			count++
		}
	}
	return count
}
