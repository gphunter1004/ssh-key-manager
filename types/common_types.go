package types

import (
	"ssh-key-manager/models"
	"time"
)

// === 공통 요청/응답 구조체 ===

// AuthRequest는 인증 관련 요청을 위한 공통 구조체입니다.
type AuthRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

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

// === 사용자 관리 관련 ===

// UserProfileUpdate는 사용자 프로필 업데이트용 구조체입니다.
type UserProfileUpdate struct {
	Username    string `json:"username,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// UserInfo는 사용자 기본 정보를 담는 구조체입니다.
type UserInfo struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`      // 권한 추가
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HasSSHKey bool      `json:"has_ssh_key"`
}

// UserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보입니다.
type UserDetailWithKey struct {
	ID        uint            `json:"id"`
	Username  string          `json:"username"`
	Role      string          `json:"role"`      // 권한 추가
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	HasSSHKey bool            `json:"has_ssh_key"`
	SSHKey    *SSHKeyResponse `json:"ssh_key,omitempty"`
}

// UserResponse는 API용 사용자 정보 응답 구조체입니다.
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`      // 권한 추가
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// === 권한 관리 관련 ===

// UserRoleUpdateRequest는 사용자 권한 변경 요청 구조체입니다.
type UserRoleUpdateRequest struct {
	Role string `json:"role" binding:"required"` // "user" 또는 "admin"
}

// === SSH 키 관련 ===

// SSHKeyResponse는 API 응답용 SSH 키 정보입니다.
type SSHKeyResponse struct {
	ID          uint      `json:"id"`
	Algorithm   string    `json:"algorithm"`
	Bits        int       `json:"bits"`
	PublicKey   string    `json:"public_key"`
	PEM         string    `json:"pem"`
	PPK         string    `json:"ppk"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fingerprint string    `json:"fingerprint,omitempty"`
}

// === 공통 응답 구조체 ===

// APIResponse는 표준 API 응답 구조체입니다.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ListResponse는 목록 조회 응답을 위한 공통 구조체입니다.
type ListResponse struct {
	Items interface{} `json:"items"`
	Count int         `json:"count"`
	Page  int         `json:"page,omitempty"`
	Limit int         `json:"limit,omitempty"`
	Total int         `json:"total,omitempty"`
}

// DeploymentSummary는 배포 결과 요약 정보입니다.
type DeploymentSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
}

// === 통계 관련 ===

// UserStats는 사용자 통계 정보입니다.
type UserStats struct {
	TotalUsers         int64   `json:"total_users"`
	UsersWithKeys      int64   `json:"users_with_keys"`
	UsersWithoutKeys   int64   `json:"users_without_keys"`
	KeyCoveragePercent float64 `json:"key_coverage_percent"`
}

// AdminStats는 관리자용 통계 정보입니다.
type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	AdminUsers       int64 `json:"admin_users"`
	RegularUsers     int64 `json:"regular_users"`
	TotalServers     int64 `json:"total_servers"`
	TotalSSHKeys     int64 `json:"total_ssh_keys"`
	TotalDeployments int64 `json:"total_deployments"`
}

// === 페이징 관련 ===

// PaginationRequest는 페이징 요청을 위한 공통 구조체입니다.
type PaginationRequest struct {
	Page  int `json:"page" query:"page" form:"page"`
	Limit int `json:"limit" query:"limit" form:"limit"`
}

// SortRequest는 정렬 요청을 위한 구조체입니다.
type SortRequest struct {
	SortBy    string `json:"sort_by" query:"sort_by" form:"sort_by"`
	SortOrder string `json:"sort_order" query:"sort_order" form:"sort_order"` // asc, desc
}

// SearchRequest는 검색 요청을 위한 구조체입니다.
type SearchRequest struct {
	Query  string `json:"query" query:"q" form:"q"`
	Fields string `json:"fields" query:"fields" form:"fields"` // 검색할 필드 목록
}

// === 공통 헬퍼 메서드 ===

// GetDefaultPagination은 기본 페이징 설정을 반환합니다.
func (p *PaginationRequest) GetDefaultPagination() PaginationRequest {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	return *p
}

// GetOffset은 데이터베이스 쿼리용 오프셋을 계산합니다.
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// GetDefaultSort는 기본 정렬 설정을 반환합니다.
func (s *SortRequest) GetDefaultSort(defaultField string) SortRequest {
	if s.SortBy == "" {
		s.SortBy = defaultField
	}
	if s.SortOrder == "" {
		s.SortOrder = "desc"
	}
	return *s
}

// === 에러 응답 표준화 ===

// ErrorResponse는 에러 응답을 위한 표준 구조체입니다.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationError는 입력값 검증 에러를 위한 구조체입니다.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse는 검증 에러 응답입니다.
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Errors []ValidationError `json:"validation_errors"`
}

// === 연결 테스트 관련 ===

// ConnectionTestResult는 서버 연결 테스트 결과입니다.
type ConnectionTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
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

// === 변환 헬퍼 함수들 ===

// ToUserInfo는 모델을 UserInfo로 변환합니다.
func ToUserInfo(user models.User, hasSSHKey bool) UserInfo {
	return UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Role:      string(user.Role), // Role 추가
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		HasSSHKey: hasSSHKey,
	}
}

// ToUserDetailWithKey는 모델을 UserDetailWithKey로 변환합니다.
func ToUserDetailWithKey(user models.User, hasSSHKey bool, sshKey *SSHKeyResponse) UserDetailWithKey {
	return UserDetailWithKey{
		ID:        user.ID,
		Username:  user.Username,
		Role:      string(user.Role), // Role 추가
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		HasSSHKey: hasSSHKey,
		SSHKey:    sshKey,
	}
}

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

// ToSSHKeyResponse는 모델을 SSHKeyResponse로 변환합니다.
func ToSSHKeyResponse(sshKey models.SSHKey, fingerprint string) SSHKeyResponse {
	return SSHKeyResponse{
		ID:          sshKey.ID,
		Algorithm:   sshKey.Algorithm,
		Bits:        sshKey.Bits,
		PublicKey:   sshKey.PublicKey,
		PEM:         sshKey.PEM,
		PPK:         sshKey.PPK,
		CreatedAt:   sshKey.CreatedAt,
		UpdatedAt:   sshKey.UpdatedAt,
		Fingerprint: fingerprint,
	}
}