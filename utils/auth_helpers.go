package utils

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// UserIDFromToken은 JWT 토큰에서 사용자 ID를 추출합니다.
// controllers와 middleware에서 공통으로 사용하는 함수입니다.
func UserIDFromToken(c echo.Context) (uint, error) {
	user := c.Get("user")
	if user == nil {
		return 0, fmt.Errorf("token not found in context")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return 0, fmt.Errorf("invalid token type")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	// user_id 클레임 확인
	userIDClaim, exists := claims["user_id"]
	if !exists {
		return 0, fmt.Errorf("user_id not found in token")
	}

	// 타입 검증 및 변환
	var userID uint
	switch v := userIDClaim.(type) {
	case float64:
		userID = uint(v)
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("invalid user_id format: %v", err)
		}
		userID = uint(id)
	default:
		return 0, fmt.Errorf("unsupported user_id type: %T", v)
	}

	// 만료 시간 확인
	if exp, ok := claims["exp"]; ok {
		if expTime, ok := exp.(float64); ok {
			if time.Now().Unix() > int64(expTime) {
				return 0, fmt.Errorf("token has expired")
			}
		}
	}

	return userID, nil
}

// ParseUintParam은 URL 파라미터를 uint로 파싱합니다.
func ParseUintParam(c echo.Context, paramName string) (uint, error) {
	param := c.Param(paramName)
	if param == "" {
		return 0, fmt.Errorf("%s 파라미터가 필요합니다", paramName)
	}

	id, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("유효하지 않은 %s입니다", paramName)
	}

	return uint(id), nil
}
