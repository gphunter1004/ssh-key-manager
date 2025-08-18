package dto

// ========== 사용자 관련 DTO ==========

// UserUpdateRequest는 사용자 정보 수정 요청 구조체입니다.
type UserUpdateRequest struct {
	Username        string `json:"username,omitempty"`
	CurrentPassword string `json:"current_password,omitempty"` // 현재 비밀번호 (비밀번호 변경 시 필수)
	NewPassword     string `json:"new_password,omitempty"`     // 새 비밀번호
}

// UserRoleUpdateRequest는 사용자 권한 변경 요청 구조체입니다.
type UserRoleUpdateRequest struct {
	Role string `json:"role" binding:"required"` // "user" 또는 "admin"
}

// UserStatusUpdateRequest는 사용자 상태 변경 요청 구조체입니다 (관리자용).
type UserStatusUpdateRequest struct {
	IsActive *bool `json:"is_active"` // 활성/비활성 상태
}

// UserUnlockRequest는 계정 잠금 해제 요청 구조체입니다 (관리자용).
type UserUnlockRequest struct {
	Force bool `json:"force,omitempty"` // 강제 잠금 해제 여부
}
