package handler

import (
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/middleware"
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
		return HandleBusinessError(c, err)
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
		if be, ok := err.(*model.BusinessError); ok {
			return BusinessErrorResponse(c, mapBusinessErrorToHTTPStatus(be.Code), be)
		}
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
	_, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	responseData := map[string]interface{}{
		"message": "로그아웃이 완료되었습니다",
	}

	return SuccessWithMessageResponse(c, "로그아웃 완료", responseData)
}

// RefreshToken은 JWT 토큰을 갱신합니다.
func RefreshToken(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	newToken, err := service.C().Auth.RefreshUserToken(userID)
	if err != nil {
		if be, ok := err.(*model.BusinessError); ok {
			return BusinessErrorResponse(c, mapBusinessErrorToHTTPStatus(be.Code), be)
		}
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
	// 미들웨어에서 이미 토큰 검증됨
	userID, _ := GetUserID(c)

	user, err := service.C().User.GetUserByID(userID)
	if err != nil {
		return HandleBusinessError(c, err)
	}

	responseData := map[string]interface{}{
		"valid":    true,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	return SuccessResponse(c, responseData)
}

// mapBusinessErrorToHTTPStatus는 비즈니스 에러를 HTTP 상태 코드로 매핑합니다.
func mapBusinessErrorToHTTPStatus(code model.ErrorCode) int {
	switch code {
	case model.ErrInvalidCredentials, model.ErrTokenExpired, model.ErrInvalidToken, model.ErrInvalidJWT:
		return 401
	case model.ErrPermissionDenied:
		return 403
	case model.ErrUserNotFound, model.ErrServerNotFound, model.ErrSSHKeyNotFound, model.ErrDepartmentNotFound:
		return 404
	case model.ErrUserAlreadyExists, model.ErrServerExists, model.ErrDepartmentExists, model.ErrSSHKeyExists:
		return 409
	case model.ErrValidationFailed, model.ErrInvalidInput, model.ErrWeakPassword, model.ErrRequiredField,
		model.ErrInvalidFormat, model.ErrInvalidRange, model.ErrInvalidUsername, model.ErrInvalidServerID,
		model.ErrInvalidDeptID, model.ErrCannotDeleteSelf, model.ErrLastAdmin, model.ErrDepartmentHasUsers,
		model.ErrDepartmentHasChild, model.ErrInvalidParentDept, model.ErrServerNotOwned, model.ErrInvalidSSHKey:
		return 400
	case model.ErrConnectionFailed:
		return 502
	case model.ErrSSHKeyGeneration, model.ErrSSHKeyDeployment:
		return 500
	case model.ErrInternalServer, model.ErrDatabaseError, model.ErrConfigError, model.ErrFileSystemError:
		return 500
	default:
		return 400
	}
}
