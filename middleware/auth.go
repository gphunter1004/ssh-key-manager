package middleware

import (
	"fmt"
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/utils"

	"github.com/labstack/echo/v4"
)

// AdminRequired는 관리자 권한을 확인하는 미들웨어입니다.
func AdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := utils.UserIDFromToken(c)
		if err != nil {
			utils.LogSecurityEvent("관리자 권한 확인 실패", 0, err.Error(), "medium")
			return helpers.UnauthorizedResponse(c, "Invalid token")
		}

		if !services.IsUserAdmin(userID) {
			utils.LogSecurityEvent("권한 없는 관리자 접근 시도", userID, "관리자 권한이 필요한 기능에 접근", "high")
			return helpers.ForbiddenResponse(c, "관리자 권한이 필요합니다")
		}

		utils.LogUserAction(userID, "접근", "관리자 기능", true)
		return next(c)
	}
}

// OptionalAdminRequired는 선택적 관리자 권한을 확인합니다.
// 관리자가 아니어도 접근 가능하지만, 권한에 따라 다른 데이터를 제공할 때 사용
func OptionalAdminRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := utils.UserIDFromToken(c)
		if err != nil {
			return helpers.UnauthorizedResponse(c, "Invalid token")
		}

		// 관리자 여부를 컨텍스트에 저장
		isAdmin := services.IsUserAdmin(userID)
		c.Set("isAdmin", isAdmin)
		c.Set("userID", userID)

		utils.LogUserAction(userID, "접근", "선택적 관리자 기능", true, fmt.Sprintf("관리자: %t", isAdmin))
		return next(c)
	}
}
