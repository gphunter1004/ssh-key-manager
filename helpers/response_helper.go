package helpers

import (
	"net/http"
	"ssh-key-manager/types"

	"github.com/labstack/echo/v4"
)

// SuccessResponse는 성공 응답을 생성합니다.
func SuccessResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessageResponse는 메시지와 함께 성공 응답을 생성합니다.
func SuccessWithMessageResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, types.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse는 생성 성공 응답을 생성합니다.
func CreatedResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, types.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ErrorResponse는 에러 응답을 생성합니다.
func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, types.APIResponse{
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

// ListResponse는 목록 조회 응답을 생성합니다.
func ListResponse(c echo.Context, items interface{}, count int) error {
	return SuccessResponse(c, types.ListResponse{
		Items: items,
		Count: count,
	})
}

// PaginatedListResponse는 페이징된 목록 응답을 생성합니다.
func PaginatedListResponse(c echo.Context, items interface{}, count, page, limit, total int) error {
	return SuccessResponse(c, types.ListResponse{
		Items: items,
		Count: count,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// ValidationErrorResponse는 검증 에러 응답을 생성합니다.
func ValidationErrorResponse(c echo.Context, errors []types.ValidationError) error {
	return c.JSON(http.StatusBadRequest, types.ValidationErrorResponse{
		Error:  "입력값 검증에 실패했습니다",
		Errors: errors,
	})
}