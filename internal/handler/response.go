package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// APIResponse는 표준 API 응답 구조체입니다.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SuccessResponse는 성공 응답을 생성합니다.
func SuccessResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessageResponse는 메시지와 함께 성공 응답을 생성합니다.
func SuccessWithMessageResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse는 생성 성공 응답을 생성합니다.
func CreatedResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse는 에러 응답을 생성합니다.
func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   message,
	})
}

// BadRequestResponse는 잘못된 요청 응답을 생성합니다.
func BadRequestResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusBadRequest, message)
}

// UnauthorizedResponse는 인증 실패 응답을 생성합니다.
func UnauthorizedResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, message)
}

// NotFoundResponse는 리소스 없음 응답을 생성합니다.
func NotFoundResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusNotFound, message)
}

// InternalServerErrorResponse는 서버 에러 응답을 생성합니다.
func InternalServerErrorResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, message)
}

// ForbiddenResponse는 권한 거부 응답을 생성합니다.
func ForbiddenResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, message)
}

// ConflictResponse는 충돌 응답을 생성합니다.
func ConflictResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusConflict, message)
}
