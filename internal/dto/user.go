package dto

// ========== 사용자 관련 DTO ==========

// UserUpdateRequest는 사용자 정보 수정 요청 구조체입니다.
type UserUpdateRequest struct {
	Username    string `json:"username,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// UserRoleUpdateRequest는 사용자 권한 변경 요청 구조체입니다.
type UserRoleUpdateRequest struct {
	Role string `json:"role" binding:"required"` // "user" 또는 "admin"
}
