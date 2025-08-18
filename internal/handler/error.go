package handler

import (
	"net/http"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"

	"github.com/labstack/echo/v4"
)

// ========== 핵심 HTTP 상태 응답 함수들 ==========

// BadRequestResponse는 잘못된 요청 응답을 생성합니다 (400).
func BadRequestResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrValidationFailed,
			Message: message,
		},
	})
}

// UnauthorizedResponse는 인증 실패 응답을 생성합니다 (401).
func UnauthorizedResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusUnauthorized, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrInvalidCredentials,
			Message: message,
		},
	})
}

// ForbiddenResponse는 권한 거부 응답을 생성합니다 (403).
func ForbiddenResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusForbidden, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrPermissionDenied,
			Message: message,
		},
	})
}

// NotFoundResponse는 리소스 없음 응답을 생성합니다 (404).
func NotFoundResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusNotFound, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrValidationFailed,
			Message: message,
		},
	})
}

// ConflictResponse는 충돌 응답을 생성합니다 (409).
func ConflictResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusConflict, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrValidationFailed,
			Message: message,
		},
	})
}

// InternalServerErrorResponse는 서버 내부 에러 응답을 생성합니다 (500).
func InternalServerErrorResponse(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, dto.APIResponse{
		Success: false,
		Error: &model.APIError{
			Code:    model.ErrInternalServer,
			Message: message,
		},
	})
}
