package types

import (
	"time"
)

// === 인증 관련 타입 ===

// AuthRequest는 인증 관련 요청을 위한 공통 구조체입니다.
type AuthRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// LoginResponse는 로그인 성공 응답 구조체입니다.
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
	UserID   uint   `json:"user_id"`
}

// TokenValidationResponse는 토큰 검증 응답 구조체입니다.
type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// RefreshTokenRequest는 토큰 갱신 요청 구조체입니다.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse는 토큰 갱신 응답 구조체입니다.
type RefreshTokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// PasswordChangeRequest는 비밀번호 변경 요청 구조체입니다.
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
}

// === 권한 관리 관련 ===

// UserRoleUpdateRequest는 사용자 권한 변경 요청 구조체입니다.
type UserRoleUpdateRequest struct {
	Role string `json:"role" binding:"required"` // "user" 또는 "admin"
}

// PermissionCheckRequest는 권한 확인 요청 구조체입니다.
type PermissionCheckRequest struct {
	Action   string `json:"action" binding:"required"`   // 수행하려는 작업
	Resource string `json:"resource" binding:"required"` // 대상 리소스
}

// PermissionCheckResponse는 권한 확인 응답 구조체입니다.
type PermissionCheckResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}
