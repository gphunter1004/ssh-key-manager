package middleware

import (
	"fmt"
	"log"
	"net/http"
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
			log.Printf("❌ 토큰에서 사용자 ID 추출 실패: %v", err)
			return c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    model.ErrInvalidToken,
					Message: "유효하지 않은 토큰입니다",
				},
			})
		}

		// 안전한 DB 접근
		db, err := model.GetDB()
		if err != nil {
			log.Printf("❌ 데이터베이스 접근 실패: %v", err)
			return c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    model.ErrDatabaseError,
					Message: "데이터베이스 연결 오류가 발생했습니다",
				},
			})
		}

		// 사용자 권한 확인
		var user model.User
		if err := db.Select("role").First(&user, userID).Error; err != nil {
			log.Printf("❌ 사용자 조회 실패 (ID: %d): %v", userID, err)
			return c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    model.ErrUserNotFound,
					Message: "사용자를 찾을 수 없습니다",
				},
			})
		}

		if user.Role != model.RoleAdmin {
			log.Printf("❌ 관리자 권한 없음 (사용자 ID: %d, 권한: %s)", userID, user.Role)
			return c.JSON(http.StatusForbidden, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    model.ErrPermissionDenied,
					Message: "관리자 권한이 필요합니다",
				},
			})
		}

		return next(c)
	}
}

// UserIDFromToken은 JWT 토큰에서 사용자 ID를 안전하게 추출합니다.
func UserIDFromToken(c echo.Context) (uint, error) {
	// Context에서 user 정보 추출
	user := c.Get("user")
	if user == nil {
		return 0, fmt.Errorf("token not found in context")
	}

	// JWT Token 타입 확인
	token, ok := user.(*jwt.Token)
	if !ok {
		return 0, fmt.Errorf("invalid token type: expected *jwt.Token, got %T", user)
	}

	if token == nil {
		return 0, fmt.Errorf("token is nil")
	}

	// Claims 타입 확인
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims: expected jwt.MapClaims, got %T", token.Claims)
	}

	if claims == nil {
		return 0, fmt.Errorf("claims is nil")
	}

	// user_id 클레임 존재 확인
	userIDClaim, exists := claims["user_id"]
	if !exists {
		return 0, fmt.Errorf("user_id not found in token claims")
	}

	if userIDClaim == nil {
		return 0, fmt.Errorf("user_id claim is nil")
	}

	// 안전한 타입 변환
	var userID uint
	switch v := userIDClaim.(type) {
	case float64:
		// float64 범위 체크
		if v < 0 {
			return 0, fmt.Errorf("user_id cannot be negative: %f", v)
		}
		if v > float64(^uint(0)) {
			return 0, fmt.Errorf("user_id out of uint range: %f", v)
		}
		userID = uint(v)
	case int:
		if v < 0 {
			return 0, fmt.Errorf("user_id cannot be negative: %d", v)
		}
		userID = uint(v)
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("user_id cannot be negative: %d", v)
		}
		if v > int64(^uint(0)) {
			return 0, fmt.Errorf("user_id out of uint range: %d", v)
		}
		userID = uint(v)
	case string:
		if v == "" {
			return 0, fmt.Errorf("user_id string is empty")
		}
		id, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid user_id string format: %v", err)
		}
		if id > uint64(^uint(0)) {
			return 0, fmt.Errorf("user_id string out of uint range: %d", id)
		}
		userID = uint(id)
	default:
		return 0, fmt.Errorf("unsupported user_id type: %T (value: %v)", v, v)
	}

	// userID가 0인 경우 체크
	if userID == 0 {
		return 0, fmt.Errorf("user_id cannot be zero")
	}

	// 토큰 만료 시간 확인
	if exp, ok := claims["exp"]; ok {
		switch expValue := exp.(type) {
		case float64:
			if time.Now().Unix() > int64(expValue) {
				return 0, fmt.Errorf("token has expired")
			}
		case int64:
			if time.Now().Unix() > expValue {
				return 0, fmt.Errorf("token has expired")
			}
		default:
			log.Printf("⚠️ 예상하지 못한 exp 타입: %T", exp)
		}
	}

	// 토큰 발급 시간 확인 (미래 토큰 방지)
	if iat, ok := claims["iat"]; ok {
		switch iatValue := iat.(type) {
		case float64:
			if time.Now().Unix() < int64(iatValue) {
				return 0, fmt.Errorf("token issued in the future")
			}
		case int64:
			if time.Now().Unix() < iatValue {
				return 0, fmt.Errorf("token issued in the future")
			}
		}
	}

	return userID, nil
}

// ValidateTokenClaims는 JWT 클레임의 유효성을 검증합니다.
func ValidateTokenClaims(claims jwt.MapClaims) error {
	if claims == nil {
		return fmt.Errorf("claims is nil")
	}

	// 필수 클레임 확인
	requiredClaims := []string{"user_id", "exp", "iat"}
	for _, claim := range requiredClaims {
		if _, exists := claims[claim]; !exists {
			return fmt.Errorf("required claim '%s' not found", claim)
		}
	}

	// Issuer 확인 (설정된 경우)
	if issuer, ok := claims["iss"]; ok {
		if issuerStr, ok := issuer.(string); ok {
			if issuerStr != "ssh-key-manager" {
				return fmt.Errorf("invalid token issuer: %s", issuerStr)
			}
		}
	}

	return nil
}

// ExtractUserIDSafely는 안전한 사용자 ID 추출을 위한 헬퍼 함수입니다.
func ExtractUserIDSafely(c echo.Context) (uint, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("🚨 UserIDFromToken에서 panic 복구: %v", r)
		}
	}()

	return UserIDFromToken(c)
}
