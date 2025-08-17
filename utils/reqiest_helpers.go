package utils

import (
	"ssh-key-manager/helpers"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// BindAndValidate 요청 바인딩 및 기본 검증
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}
	return nil
}

// HandleServiceError 서비스 에러 처리 표준화
func HandleServiceError(c echo.Context, err error, operation string) error {
	// 사용자 친화적 에러 메시지들
	userFriendlyMessages := []string{
		"이미 사용 중인",
		"찾을 수 없습니다",
		"유효하지 않은",
		"입력해주세요",
		"이상이어야 합니다",
		"까지 가능합니다",
		"권한이 없습니다",
		"선택해주세요",
		"최소",
		"최대",
		"필수",
		"올바르지 않습니다",
	}

	errMsg := err.Error()
	for _, msg := range userFriendlyMessages {
		if strings.Contains(errMsg, msg) {
			return helpers.BadRequestResponse(c, err.Error())
		}
	}

	return helpers.InternalServerErrorResponse(c, operation+" 중 오류가 발생했습니다")
}

// ExtractPaginationParams 페이징 파라미터 추출
func ExtractPaginationParams(c echo.Context) (page, limit int) {
	page = 1
	limit = 20

	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	return page, limit
}

// ExtractSortParams 정렬 파라미터 추출
func ExtractSortParams(c echo.Context, defaultSort string) (sortBy, sortOrder string) {
	sortBy = c.QueryParam("sort_by")
	if sortBy == "" {
		sortBy = defaultSort
	}

	sortOrder = c.QueryParam("sort_order")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return sortBy, sortOrder
}

// ExtractSearchParams 검색 파라미터 추출
func ExtractSearchParams(c echo.Context) (query string, fields []string) {
	query = strings.TrimSpace(c.QueryParam("q"))

	fieldsParam := c.QueryParam("fields")
	if fieldsParam != "" {
		fields = strings.Split(fieldsParam, ",")
		for i, field := range fields {
			fields[i] = strings.TrimSpace(field)
		}
	}

	return query, fields
}

// GetClientIP 클라이언트 IP 추출
func GetClientIP(c echo.Context) string {
	// X-Forwarded-For 헤더 확인 (프록시 환경)
	if xff := c.Request().Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// X-Real-IP 헤더 확인
	if xri := c.Request().Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// RemoteAddr 사용
	return c.Request().RemoteAddr
}

// GetUserAgent 사용자 에이전트 추출
func GetUserAgent(c echo.Context) string {
	return c.Request().Header.Get("User-Agent")
}
