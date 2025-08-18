package handler

import (
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// GetCurrentUser는 현재 로그인한 사용자 정보를 반환합니다.
func GetCurrentUser(c echo.Context) error {
	userID, _ := GetUserID(c)

	user, err := service.C().User.GetUserByID(userID)
	if err != nil {
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "사용자 정보 조회 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "사용자 정보 조회 중 오류가 발생했습니다")
	}

	// SSH 키 존재 여부 확인
	hasSSHKey := service.C().Key.HasUserSSHKey(userID)

	responseData := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"has_ssh_key": hasSSHKey,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,
	}

	return SuccessResponse(c, responseData)
}

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(c echo.Context) error {
	userID, _ := GetUserID(c)

	var req dto.UserUpdateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	user, err := service.C().User.UpdateUserProfile(userID, req)
	if err != nil {
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			case model.ErrUserAlreadyExists:
				return ConflictResponse(c, "이미 사용 중인 사용자명입니다")
			case model.ErrWeakPassword:
				return BadRequestResponse(c, "비밀번호는 최소 4자 이상이어야 합니다")
			default:
				return InternalServerErrorResponse(c, "프로필 업데이트 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "프로필 업데이트 중 오류가 발생했습니다")
	}

	return SuccessWithMessageResponse(c, "프로필이 업데이트되었습니다", user)
}
