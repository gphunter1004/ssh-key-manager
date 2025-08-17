package handler

import (
	"log"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// Register는 새로운 사용자를 등록합니다.
func Register(c echo.Context) error {
	var req model.AuthRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	user, err := service.RegisterUser(req.Username, req.Password)
	if err != nil {
		log.Printf("❌ 회원가입 실패: %v", err)
		return BadRequestResponse(c, err.Error())
	}

	responseData := map[string]interface{}{
		"message":  "사용자가 성공적으로 등록되었습니다",
		"username": user.Username,
	}

	return CreatedResponse(c, "회원가입이 완료되었습니다", responseData)
}

// Login은 사용자를 인증하고 JWT를 반환합니다.
func Login(c echo.Context) error {
	var req model.AuthRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	token, user, err := service.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		log.Printf("❌ 로그인 실패: %v", err)
		return UnauthorizedResponse(c, "사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	responseData := map[string]interface{}{
		"token":    token,
		"username": user.Username,
		"role":     user.Role,
		"message":  "로그인이 성공했습니다",
	}

	return SuccessWithMessageResponse(c, "로그인 성공", responseData)
}

// Logout은 사용자를 로그아웃합니다.
func Logout(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	log.Printf("👤 사용자 로그아웃: ID %d", userID)

	responseData := map[string]interface{}{
		"message": "로그아웃이 완료되었습니다",
	}

	return SuccessWithMessageResponse(c, "로그아웃 완료", responseData)
}

// RefreshToken은 JWT 토큰을 갱신합니다.
func RefreshToken(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	newToken, err := service.RefreshUserToken(userID)
	if err != nil {
		return InternalServerErrorResponse(c, "토큰 갱신 중 오류가 발생했습니다")
	}

	responseData := map[string]interface{}{
		"token":   newToken,
		"message": "토큰이 성공적으로 갱신되었습니다",
	}

	return SuccessWithMessageResponse(c, "토큰 갱신 완료", responseData)
}

// ValidateToken은 토큰의 유효성을 검사합니다.
func ValidateToken(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	user, err := service.GetUserByID(userID)
	if err != nil {
		return UnauthorizedResponse(c, "User not found")
	}

	responseData := map[string]interface{}{
		"valid":    true,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	return SuccessResponse(c, responseData)
}
