package handler

import (
	"log"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// CreateKey는 SSH 키 쌍을 생성합니다.
func CreateKey(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	sshKey, err := service.GenerateSSHKeyPair(userID)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패 (사용자 ID: %d): %v", userID, err)
		return InternalServerErrorResponse(c, "SSH 키 생성에 실패했습니다")
	}

	log.Printf("✅ SSH 키 생성 성공 (사용자 ID: %d)", userID)
	return SuccessWithMessageResponse(c, "SSH 키가 성공적으로 생성되었습니다", sshKey)
}

// GetKey는 사용자의 SSH 키를 조회합니다.
func GetKey(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	sshKey, err := service.GetUserSSHKey(userID)
	if err != nil {
		log.Printf("❌ SSH 키 조회 실패 (사용자 ID: %d): %v", userID, err)
		return NotFoundResponse(c, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
	}

	log.Printf("✅ SSH 키 조회 성공 (사용자 ID: %d)", userID)
	return SuccessResponse(c, sshKey)
}

// DeleteKey는 사용자의 SSH 키를 삭제합니다.
func DeleteKey(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	err = service.DeleteUserSSHKey(userID)
	if err != nil {
		log.Printf("❌ SSH 키 삭제 실패 (사용자 ID: %d): %v", userID, err)
		return NotFoundResponse(c, "삭제할 SSH 키를 찾을 수 없습니다")
	}

	log.Printf("✅ SSH 키 삭제 성공 (사용자 ID: %d)", userID)
	return SuccessWithMessageResponse(c, "SSH 키가 성공적으로 삭제되었습니다", nil)
}
