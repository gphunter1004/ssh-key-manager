package dto

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
