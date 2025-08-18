package dto

// ========== 인증 관련 DTO ==========

// AuthRequest는 인증 요청 구조체입니다.
type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
