package types

import (
	"ssh-key-manager/models"
	"time"
)

// === 서버 관리 관련 ===

// ServerCreateRequest는 서버 생성 요청 구조체입니다.
type ServerCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port"`
	Username    string `json:"username" binding:"required"`
	Description string `json:"description"`
}

// ServerUpdateRequest는 서버 업데이트 요청 구조체입니다.
type ServerUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Host        string `json:"host,omitempty"`
	Port        int    `json:"port,omitempty"`
	Username    string `json:"username,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ServerResponse는 API용 서버 정보 응답 구조체입니다.
type ServerResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Username    string    `json:"username"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServerListRequest는 서버 목록 요청 구조체입니다.
type ServerListRequest struct {
	PaginationRequest
	SortRequest
	SearchRequest
	Status string `json:"status" query:"status"`
}

// === 키 배포 관련 ===

// KeyDeploymentRequest는 키 배포 요청 구조체입니다.
type KeyDeploymentRequest struct {
	ServerIDs []uint `json:"server_ids" binding:"required"`
}

// DeploymentResult는 키 배포 결과를 담는 구조체입니다.
type DeploymentResult struct {
	ServerID     uint   `json:"server_id"`
	ServerName   string `json:"server_name"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// DeploymentSummary는 배포 결과 요약 정보입니다.
type DeploymentSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// KeyDeploymentHistoryRequest는 배포 이력 조회 요청입니다.
type KeyDeploymentHistoryRequest struct {
	PaginationRequest
	SortRequest
	ServerID *uint      `json:"server_id" query:"server_id"`
	Status   string     `json:"status" query:"status"`
	DateFrom *time.Time `json:"date_from" query:"date_from"`
	DateTo   *time.Time `json:"date_to" query:"date_to"`
}

// === 연결 테스트 관련 ===

// ConnectionTestResult는 서버 연결 테스트 결과입니다.
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// ServerInfoResult는 서버 정보 조회 결과입니다.
type ServerInfoResult struct {
	OS           string `json:"os"`
	Kernel       string `json:"kernel"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
	Uptime       string `json:"uptime"`
}

// === 배치 배포 관련 ===

// ServerDeployTarget은 배포 대상 서버 정보입니다.
type ServerDeployTarget struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
}

// ServerDeployResult는 서버별 배포 결과입니다.
type ServerDeployResult struct {
	Index        int           `json:"index"`
	ServerName   string        `json:"server_name"`
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	Username     string        `json:"username"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Duration     time.Duration `json:"duration"`
}

// BatchDeploymentRequest는 배치 배포 요청 구조체입니다.
type BatchDeploymentRequest struct {
	Servers   []ServerDeployTarget `json:"servers" binding:"required"`
	PublicKey string               `json:"public_key" binding:"required"`
	Options   DeploymentOptions    `json:"options,omitempty"`
}

// DeploymentOptions는 배포 옵션 구조체입니다.
type DeploymentOptions struct {
	Timeout       int  `json:"timeout"`        // 초 단위
	CreateBackup  bool `json:"create_backup"`  // 백업 생성 여부
	OverwriteKeys bool `json:"overwrite_keys"` // 기존 키 덮어쓰기 여부
}

// === 변환 헬퍼 함수들 ===

// ToServerResponse는 모델을 ServerResponse로 변환합니다.
func ToServerResponse(server models.Server) ServerResponse {
	return ServerResponse{
		ID:          server.ID,
		Name:        server.Name,
		Host:        server.Host,
		Port:        server.Port,
		Username:    server.Username,
		Description: server.Description,
		Status:      server.Status,
		CreatedAt:   server.CreatedAt,
		UpdatedAt:   server.UpdatedAt,
	}
}
