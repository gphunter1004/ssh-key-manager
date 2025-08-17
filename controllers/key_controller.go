package controllers

import (
	"fmt"
	"log"
	"net/http"
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

// ValidateToken은 토큰의 유효성을 검사합니다 (새로운 엔드포인트)
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

// CreateKey creates or regenerates an SSH key pair for the authenticated user.
func CreateKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	key, err := services.GenerateSSHKeyPair(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to generate key pair"})
	}

	return c.JSON(http.StatusOK, key)
}

// GetKey retrieves the SSH key pair for the authenticated user.
func GetKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	key, err := services.GetKeyByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, key)
}

// DeleteKey deletes the SSH key pair for the authenticated user.
func DeleteKey(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
	}

	if err := services.DeleteKeyByUserID(userID); err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Key deleted successfully"})
}
