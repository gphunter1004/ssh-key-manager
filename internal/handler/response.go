package handler

import (
	"net/http"
	"ssh-key-manager/internal/dto"

	"github.com/labstack/echo/v4"
)

// SuccessResponse는 성공 응답을 생성합니다.
func SuccessResponse(c echo.Context, data interface{}) error {
	return c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMessageResponse는 메시지와 함께 성공 응답을 생성합니다.
func SuccessWithMessageResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusOK, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse는 생성 성공 응답을 생성합니다.
func CreatedResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusCreated, dto.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
