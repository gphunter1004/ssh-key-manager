package middleware

import (
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// userIDFromToken extracts user ID from the JWT token in the context.
func userIDFromToken(c echo.Context) (uint, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))
	return userID, nil
}

// AdminRequired는 관리자 권한을 확인하는 미들웨어입니다.
func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := userIDFromToken(c)
		if err != nil {
			return helpers.UnauthorizedResponse(c, "Invalid token")
		}

		if !services.IsUserAdmin(userID) {
			return helpers.UnauthorizedResponse(c, "관리자 권한이 필요합니다")
		}

		return next(c)
	}
}

// OptionalAdminRequired는 선택적 관리자 권한을 확인합니다.
// 관리자가 아니어도 접근 가능하지만, 권한에 따라 다른 데이터를 제공할 때 사용
func OptionalAdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := userIDFromToken(c)
		if err != nil {
			return helpers.UnauthorizedResponse(c, "Invalid token")
		}

		// 관리자 여부를 컨텍스트에 저장
		isAdmin := services.IsUserAdmin(userID)
		c.Set("isAdmin", isAdmin)
		c.Set("userID", userID)

		return next(c)
	}
}