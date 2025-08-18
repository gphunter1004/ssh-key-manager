package handler

import (
	"fmt"
	"log"
	"ssh-key-manager/internal/middleware"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetUserID는 미들웨어에서 설정된 사용자 ID를 Context에서 추출합니다.
func GetUserID(c echo.Context) (uint, error) {
	return middleware.GetUserID(c)
}

// ParseIDParam은 URL 파라미터에서 ID를 추출하고 검증합니다.
func ParseIDParam(c echo.Context, paramName string) (uint, error) {
	param := c.Param(paramName)
	if param == "" {
		return 0, fmt.Errorf("%s 파라미터가 없습니다", paramName)
	}

	id, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("유효하지 않은 %s ID입니다", paramName)
	}

	if id == 0 {
		return 0, fmt.Errorf("%s ID는 0일 수 없습니다", paramName)
	}

	return uint(id), nil
}

// ParseUserIDParam은 사용자 ID 파라미터를 추출합니다.
func ParseUserIDParam(c echo.Context) (uint, error) {
	return ParseIDParam(c, "사용자")
}

// ParseServerIDParam은 서버 ID 파라미터를 추출합니다.
func ParseServerIDParam(c echo.Context) (uint, error) {
	return ParseIDParam(c, "서버")
}

// ParseDepartmentIDParam은 부서 ID 파라미터를 추출합니다.
func ParseDepartmentIDParam(c echo.Context) (uint, error) {
	return ParseIDParam(c, "부서")
}

// LogSuccess는 성공 로그를 출력합니다.
func LogSuccess(action string, userID uint, details ...interface{}) {
	if len(details) > 0 {
		log.Printf("✅ %s 성공 (사용자 ID: %d): %v", action, userID, details[0])
	} else {
		log.Printf("✅ %s 성공 (사용자 ID: %d)", action, userID)
	}
}

// LogError는 에러 로그를 출력합니다.
func LogError(action string, userID uint, err error) {
	log.Printf("❌ %s 실패 (사용자 ID: %d): %v", action, userID, err)
}

// LogAdminAction은 관리자 액션 로그를 출력합니다.
func LogAdminAction(action string, adminID, targetID uint, details ...interface{}) {
	if len(details) > 0 {
		log.Printf("✅ %s 성공 (관리자 ID: %d, 대상 ID: %d): %v", action, adminID, targetID, details[0])
	} else {
		log.Printf("✅ %s 성공 (관리자 ID: %d, 대상 ID: %d)", action, adminID, targetID)
	}
}

// LogAdminError는 관리자 액션 에러 로그를 출력합니다.
func LogAdminError(action string, adminID, targetID uint, err error) {
	log.Printf("❌ %s 실패 (관리자 ID: %d, 대상 ID: %d): %v", action, adminID, targetID, err)
}

// ValidateJSONRequest는 JSON 요청을 바인딩하고 검증합니다.
func ValidateJSONRequest(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}
	return nil
}
