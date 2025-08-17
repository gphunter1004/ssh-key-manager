package controllers

import (
	"fmt"
	"log"
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// userIDFromToken은 JWT 토큰에서 사용자 ID를 추출합니다.
func userIDFromToken(c echo.Context) (uint, error) {
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

	log.Printf("🔍 토큰에서 사용자 ID 추출 성공: %d", userID)
	return userID, nil
}

// ValidateToken은 토큰의 유효성을 검사합니다.
func ValidateToken(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		log.Printf("❌ 토큰 검증 실패: %v", err)
		return helpers.UnauthorizedResponse(c, "invalid or expired jwt")
	}

	// 사용자 존재 여부 확인
	userDetail, err := services.GetUserDetailWithKey(userID)
	if err != nil {
		log.Printf("❌ 사용자 조회 실패: %v", err)
		return helpers.UnauthorizedResponse(c, "user not found")
	}

	// 간단한 사용자 정보만 반환 (보안상 민감한 정보 제외)
	responseData := map[string]interface{}{
		"valid":    true,
		"user_id":  userID,
		"username": userDetail.Username,
		"role":     userDetail.Role,
	}

	return helpers.SuccessResponse(c, responseData)
}

// CreateKey는 SSH 키 쌍을 생성하거나 재생성합니다.
func CreateKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	key, err := services.GenerateSSHKeyPair(userID)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패: %v", err)
		return helpers.InternalServerErrorResponse(c, "Failed to generate key pair")
	}

	log.Printf("✅ SSH 키 생성 성공 (사용자 ID: %d)", userID)
	return helpers.SuccessWithMessageResponse(c, "SSH 키가 성공적으로 생성되었습니다", key)
}

// GetKey는 사용자의 SSH 키 쌍을 조회합니다.
func GetKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	key, err := services.GetKeyByUserID(userID)
	if err != nil {
		log.Printf("❌ SSH 키 조회 실패 (사용자 ID: %d): %v", userID, err)
		return helpers.NotFoundResponse(c, err.Error())
	}

	log.Printf("✅ SSH 키 조회 성공 (사용자 ID: %d)", userID)
	return helpers.SuccessResponse(c, key)
}

// DeleteKey는 사용자의 SSH 키 쌍을 삭제합니다.
func DeleteKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	if err := services.DeleteKeyByUserID(userID); err != nil {
		log.Printf("❌ SSH 키 삭제 실패 (사용자 ID: %d): %v", userID, err)
		return helpers.NotFoundResponse(c, err.Error())
	}

	log.Printf("✅ SSH 키 삭제 성공 (사용자 ID: %d)", userID)
	return helpers.SuccessWithMessageResponse(c, "SSH 키가 성공적으로 삭제되었습니다", nil)
}
