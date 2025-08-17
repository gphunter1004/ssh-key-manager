package controllers

import (
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/utils"

	"github.com/labstack/echo/v4"
)

// ValidateToken은 토큰의 유효성을 검사합니다.
func ValidateToken(c echo.Context) error {
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		utils.LogSecurityEvent("토큰 검증 실패", 0, err.Error(), "medium")
		return helpers.UnauthorizedResponse(c, "invalid or expired jwt")
	}

	// 사용자 존재 여부 확인
	userDetail, err := services.GetUserDetailWithKey(userID)
	if err != nil {
		utils.LogUserAction(userID, "조회", "사용자 정보", false, err.Error())
		return helpers.UnauthorizedResponse(c, "user not found")
	}

	// 간단한 사용자 정보만 반환 (보안상 민감한 정보 제외)
	responseData := map[string]interface{}{
		"valid":    true,
		"user_id":  userID,
		"username": userDetail.Username,
		"role":     userDetail.Role,
	}

	utils.LogUserAction(userID, "검증", "토큰", true)
	return helpers.SuccessResponse(c, responseData)
}

// CreateKey는 SSH 키 쌍을 생성하거나 재생성합니다.
func CreateKey(c echo.Context) error {
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var key interface{}
	err = utils.LogOperation("SSH 키 생성", func() error {
		utils.LogServiceCall("KeyService", "GenerateSSHKeyPair", userID)
		var generateErr error
		key, generateErr = services.GenerateSSHKeyPair(userID)
		return generateErr
	})

	if err != nil {
		utils.LogUserAction(userID, "생성", "SSH 키", false, err.Error())
		return helpers.InternalServerErrorResponse(c, "Failed to generate key pair")
	}

	utils.LogUserAction(userID, "생성", "SSH 키", true)
	return helpers.SuccessWithMessageResponse(c, "SSH 키가 성공적으로 생성되었습니다", key)
}

// GetKey는 사용자의 SSH 키 쌍을 조회합니다.
func GetKey(c echo.Context) error {
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	utils.LogServiceCall("KeyService", "GetKeyByUserID", userID)
	key, err := services.GetKeyByUserID(userID)
	if err != nil {
		utils.LogUserAction(userID, "조회", "SSH 키", false, err.Error())
		return helpers.NotFoundResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "조회", "SSH 키", true)
	return helpers.SuccessResponse(c, key)
}

// DeleteKey는 사용자의 SSH 키 쌍을 삭제합니다.
func DeleteKey(c echo.Context) error {
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	err = utils.LogOperation("SSH 키 삭제", func() error {
		utils.LogServiceCall("KeyService", "DeleteKeyByUserID", userID)
		return services.DeleteKeyByUserID(userID)
	})

	if err != nil {
		utils.LogUserAction(userID, "삭제", "SSH 키", false, err.Error())
		return helpers.NotFoundResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "삭제", "SSH 키", true)
	utils.LogSecurityEvent("SSH 키 삭제", userID, "사용자가 SSH 키를 삭제했습니다", "low")
	return helpers.SuccessWithMessageResponse(c, "SSH 키가 성공적으로 삭제되었습니다", nil)
}
