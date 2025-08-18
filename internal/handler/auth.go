package handler

import (
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// Register는 새로운 사용자를 등록합니다.
func Register(c echo.Context) error {
	var req dto.AuthRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	user, err := service.C().Auth.RegisterUser(req.Username, req.Password)
	if err != nil {
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserAlreadyExists:
				return ConflictResponse(c, "이미 사용 중인 사용자명입니다")
			case model.ErrWeakPassword:
				return BadRequestResponse(c, "비밀번호는 최소 4자 이상이어야 합니다")
			case model.ErrRequiredField, model.ErrInvalidUsername:
				return BadRequestResponse(c, be.Message)
			default:
				return InternalServerErrorResponse(c, "회원가입 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "회원가입 중 오류가 발생했습니다")
	}

	responseData := map[string]interface{}{
		"message":  "사용자가 성공적으로 등록되었습니다",
		"username": user.Username,
	}

	return CreatedResponse(c, "회원가입이 완료되었습니다", responseData)
}

// Login은 사용자를 인증하고 JWT를 반환합니다.
func Login(c echo.Context) error {
	var req dto.AuthRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	token, user, err := service.C().Auth.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		// 로그인 실패는 대부분 401 Unauthorized
		return UnauthorizedResponse(c, "사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	responseData := map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
		"message":  "로그인이 성공했습니다",
	}

	return SuccessWithMessageResponse(c, "로그인 성공", responseData)
}

// Logout은 사용자를 로그아웃합니다.
func Logout(c echo.Context) error {
	responseData := map[string]interface{}{
		"message": "로그아웃이 완료되었습니다",
	}

	return SuccessWithMessageResponse(c, "로그아웃 완료", responseData)
}

// RefreshToken은 JWT 토큰을 갱신합니다.
func RefreshToken(c echo.Context) error {
	userID, _ := GetUserID(c)

	newToken, err := service.C().Auth.RefreshUserToken(userID)
	if err != nil {
		return InternalServerErrorResponse(c, "토큰 갱신 중 오류가 발생했습니다")
	}

	responseData := map[string]interface{}{
		"token":   newToken,
		"message": "토큰이 성공적으로 갱신되었습니다",
	}

	return SuccessWithMessageResponse(c, "토큰 갱신 완료", responseData)
}

// ValidateToken은 토큰의 유효성을 검사합니다.
func ValidateToken(c echo.Context) error {
	userID, _ := GetUserID(c)

	user, err := service.C().User.GetUserByID(userID)
	if err != nil {
		return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
	}

	responseData := map[string]interface{}{
		"valid":    true,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	return SuccessResponse(c, responseData)
}
