package helpers

import (
	"net/http"
	"ssh-key-manager/types"
	"time"

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

// TimestampedResponse 타임스탬프가 포함된 응답
func TimestampedResponse(c echo.Context, statusCode int, success bool, message string, data interface{}) error {
	response := map[string]interface{}{
		"success":   success,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if message != "" {
		response["message"] = message
	}

	if data != nil {
		response["data"] = data
	}

	return c.JSON(statusCode, response)
}

// StandardSuccessResponse 표준 성공 응답
func StandardSuccessResponse(c echo.Context, message string, data interface{}) error {
	return TimestampedResponse(c, http.StatusOK, true, message, data)
}

// StandardErrorResponse 표준 에러 응답
func StandardErrorResponse(c echo.Context, statusCode int, message string) error {
	return TimestampedResponse(c, statusCode, false, message, nil)
}

// APIResponseWithMetadata 메타데이터가 포함된 응답
func APIResponseWithMetadata(c echo.Context, data interface{}, metadata map[string]interface{}) error {
	response := types.APIResponse{
		Success: true,
		Data:    data,
	}

	// 메타데이터 추가
	if metadata != nil {
		responseMap := map[string]interface{}{
			"success":  true,
			"data":     data,
			"metadata": metadata,
		}
		return c.JSON(http.StatusOK, responseMap)
	}

	return c.JSON(http.StatusOK, response)
}

// ConflictResponse는 충돌 응답을 생성합니다 (409)
func ConflictResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusConflict, message)
}

// ForbiddenResponse는 권한 거부 응답을 생성합니다 (403)
func ForbiddenResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusForbidden, message)
}

// TooManyRequestsResponse는 요청 제한 응답을 생성합니다 (429)
func TooManyRequestsResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusTooManyRequests, message)
}

// ServiceUnavailableResponse는 서비스 사용 불가 응답을 생성합니다 (503)
func ServiceUnavailableResponse(c echo.Context, message string) error {
	return ErrorResponse(c, http.StatusServiceUnavailable, message)
}

// NoContentResponse는 내용 없음 응답을 생성합니다 (204)
func NoContentResponse(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// AcceptedResponse는 수락됨 응답을 생성합니다 (202) - 비동기 작업용
func AcceptedResponse(c echo.Context, message string, data interface{}) error {
	return c.JSON(http.StatusAccepted, types.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
