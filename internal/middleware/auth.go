package middleware

import (
	"fmt"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// AdminRequired는 관리자 권한을 확인하는 미들웨어입니다.
func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := UserIDFromToken(c)
		if err != nil {
			return c.JSON(401, map[string]interface{}{
				"success": false,
				"error":   "Invalid token",
			})
		}

		// 사용자 권한 확인
		var user model.User
		if err := model.DB.Select("role").First(&user, userID).Error; err != nil {
			return c.JSON(403, map[string]interface{}{
				"success": false,
				"error":   "사용자를 찾을 수 없습니다",
			})
		}

		if user.Role != model.RoleAdmin {
			return c.JSON(403, map[string]interface{}{
				"success": false,
				"error":   "관리자 권한이 필요합니다",
			})
		}

		return next(c)
	}
}

// UserIDFromToken은 JWT 토큰에서 사용자 ID를 추출합니다.
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

	userIDClaim, exists := claims["user_id"]
	if !exists {
		return 0, fmt.Errorf("user_id not found in token")
	}

	var userID uint
	switch v := userIDClaim.(type) {
	case float64:
		userID = uint(v)
	case int:
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
