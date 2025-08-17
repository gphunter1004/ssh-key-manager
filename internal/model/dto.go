package model

// ========== 인증 관련 DTO ==========

// AuthRequest는 인증 요청 구조체입니다.
type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ========== 서버 관련 DTO ==========

// ServerCreateRequest는 서버 생성 요청 구조체입니다.
type ServerCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port"`
	Username    string `json:"username" binding:"required"`
	Description string `json:"description"`
}

// ServerUpdateRequest는 서버 수정 요청 구조체입니다.
type ServerUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
	Username    string `json:"username,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ========== 키 배포 관련 DTO ==========

// KeyDeploymentRequest는 키 배포 요청 구조체입니다.
type KeyDeploymentRequest struct {
	ServerIDs []uint `json:"server_ids" binding:"required"`
}

// DeploymentResult는 배포 결과 구조체입니다.
type DeploymentResult struct {
	ServerID     uint   `json:"server_id"`
	ServerName   string `json:"server_name"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

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

// ========== 부서 관련 DTO ==========

// DepartmentCreateRequest는 부서 생성 요청 구조체입니다.
type DepartmentCreateRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
}

// DepartmentUpdateRequest는 부서 수정 요청 구조체입니다.
type DepartmentUpdateRequest struct {
	Code        string `json:"code,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// ========== 표준 응답 구조체 ==========

// APIResponse는 표준 API 응답 구조체입니다.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
