package middleware

import (
	"fmt"
	"log"
	"net/http"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// RequireAuth는 JWT 검증 후 사용자 ID를 Context에 저장하는 미들웨어입니다.
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, err := UserIDFromToken(c)
			if err != nil {
				log.Printf("❌ 토큰 검증 실패: %v", err)
				return c.JSON(http.StatusUnauthorized, dto.APIResponse{
					Success: false,
					Error: &model.APIError{
						Code:    model.ErrInvalidToken,
						Message: "유효하지 않은 토큰입니다",
					},
				})
			}

			// Context에 사용자 ID 저장 (핸들러에서 사용)
			c.Set("userID", userID)
			return next(c)
		}
	}
}

// RequireAdmin는 JWT 검증 + 관리자 권한 확인을 통합한 미들웨어입니다.
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. JWT에서 사용자 ID 추출
			userID, err := UserIDFromToken(c)
			if err != nil {
				log.Printf("❌ 관리자 토큰 검증 실패: %v", err)
				return c.JSON(http.StatusUnauthorized, dto.APIResponse{
					Success: false,
					Error: &model.APIError{
						Code:    model.ErrInvalidToken,
						Message: "유효하지 않은 토큰입니다",
					},
				})
			}

			// 2. 데이터베이스에서 사용자 권한 확인
			db, err := model.GetDB()
			if err != nil {
				log.Printf("❌ 데이터베이스 접근 실패: %v", err)
				return c.JSON(http.StatusInternalServerError, dto.APIResponse{
					Success: false,
					Error: &model.APIError{
						Code:    model.ErrDatabaseError,
						Message: "데이터베이스 연결 오류가 발생했습니다",
					},
				})
			}

			// 3. 사용자 권한 확인 (role만 조회로 최적화)
			var user model.User
			if err := db.Select("role").First(&user, userID).Error; err != nil {
				log.Printf("❌ 관리자 권한 확인 실패 (ID: %d): %v", userID, err)
				return c.JSON(http.StatusForbidden, dto.APIResponse{
					Success: false,
					Error: &model.APIError{
						Code:    model.ErrUserNotFound,
						Message: "사용자를 찾을 수 없습니다",
					},
				})
			}

			if user.Role != model.RoleAdmin {
				log.Printf("❌ 관리자 권한 없음 (사용자 ID: %d, 권한: %s)", userID, user.Role)
				return c.JSON(http.StatusForbidden, dto.APIResponse{
					Success: false,
					Error: &model.APIError{
						Code:    model.ErrPermissionDenied,
						Message: "관리자 권한이 필요합니다",
					},
				})
			}

			// 4. Context에 사용자 ID 저장 (핸들러에서 사용)
			c.Set("userID", userID)

			return next(c)
		}
	}
}

// GetUserID는 Context에서 사용자 ID를 안전하게 추출합니다.
func GetUserID(c echo.Context) (uint, error) {
	userID, exists := c.Get("userID").(uint)
	if !exists {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
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
	if !ok || token == nil {
		return 0, fmt.Errorf("invalid token type")
	}

	// Claims 타입 확인
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims == nil {
		return 0, fmt.Errorf("invalid token claims")
	}

	// user_id 클레임 추출
	userIDClaim, exists := claims["user_id"]
	if !exists || userIDClaim == nil {
		return 0, fmt.Errorf("user_id not found in token")
	}

	// 간단한 타입 변환 (가장 일반적인 케이스만 처리)
	var userID uint
	switch v := userIDClaim.(type) {
	case float64:
		if v <= 0 || v > 4294967295 { // uint32 최대값
			return 0, fmt.Errorf("invalid user_id value: %f", v)
		}
		userID = uint(v)
	case string:
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil || id == 0 {
			return 0, fmt.Errorf("invalid user_id string: %s", v)
		}
		userID = uint(id)
	default:
		return 0, fmt.Errorf("unsupported user_id type: %T", v)
	}

	// 기본 토큰 만료 확인만 (JWT 라이브러리가 대부분 처리)
	if exp, ok := claims["exp"]; ok {
		if expFloat, ok := exp.(float64); ok {
			if time.Now().Unix() > int64(expFloat) {
				return 0, fmt.Errorf("token has expired")
			}
		}
	}

	return userID, nil
}
