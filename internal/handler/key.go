package handler

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// CreateKey는 SSH 키 쌍을 생성합니다.
func CreateKey(c echo.Context) error {
	userID, _ := GetUserID(c)

	sshKey, err := service.C().Key.GenerateSSHKeyPair(userID)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패 (사용자 ID: %d): %v", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			case model.ErrSSHKeyGeneration:
				return InternalServerErrorResponse(c, "SSH 키 생성에 실패했습니다")
			default:
				return InternalServerErrorResponse(c, "SSH 키 생성 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "SSH 키 생성에 실패했습니다")
	}

	log.Printf("✅ SSH 키 생성 성공 (사용자 ID: %d)", userID)
	return SuccessWithMessageResponse(c, "SSH 키가 성공적으로 생성되었습니다", sshKey)
}

// GetKey는 사용자의 SSH 키를 조회합니다.
func GetKey(c echo.Context) error {
	userID, _ := GetUserID(c)

	sshKey, err := service.C().Key.GetUserSSHKey(userID)
	if err != nil {
		log.Printf("❌ SSH 키 조회 실패 (사용자 ID: %d): %v", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrSSHKeyNotFound:
				return NotFoundResponse(c, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
			default:
				return InternalServerErrorResponse(c, "SSH 키 조회 중 오류가 발생했습니다")
			}
		}
		return NotFoundResponse(c, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
	}

	log.Printf("✅ SSH 키 조회 성공 (사용자 ID: %d)", userID)
	return SuccessResponse(c, sshKey)
}

// DeleteKey는 사용자의 SSH 키를 삭제합니다.
func DeleteKey(c echo.Context) error {
	userID, _ := GetUserID(c)

	err := service.C().Key.DeleteUserSSHKey(userID)
	if err != nil {
		log.Printf("❌ SSH 키 삭제 실패 (사용자 ID: %d): %v", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrSSHKeyNotFound:
				return NotFoundResponse(c, "삭제할 SSH 키를 찾을 수 없습니다")
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "SSH 키 삭제 중 오류가 발생했습니다")
			}
		}
		return NotFoundResponse(c, "삭제할 SSH 키를 찾을 수 없습니다")
	}

	log.Printf("✅ SSH 키 삭제 성공 (사용자 ID: %d)", userID)
	return SuccessWithMessageResponse(c, "SSH 키가 성공적으로 삭제되었습니다", nil)
}
