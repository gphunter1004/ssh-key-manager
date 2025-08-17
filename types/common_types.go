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

// === 부서 관련 요청/응답 구조체 ===

// DepartmentCreateRequest는 부서 생성 요청 구조체입니다.
type DepartmentCreateRequest struct {
	Code        string `json:"code" binding:"required"` // 부서 코드
	Name        string `json:"name" binding:"required"` // 부서명
	Description string `json:"description"`             // 부서 설명
	ParentID    *uint  `json:"parent_id"`               // 상위 부서 ID
}

// DepartmentUpdateRequest는 부서 수정 요청 구조체입니다.
type DepartmentUpdateRequest struct {
	Code        string `json:"code,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ParentID    *uint  `json:"parent_id,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// DepartmentResponse는 부서 정보 응답 구조체입니다.
type DepartmentResponse struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    *uint     `json:"parent_id"`
	Level       int       `json:"level"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 관계 데이터
	Parent    *DepartmentSimple  `json:"parent,omitempty"`
	Children  []DepartmentSimple `json:"children,omitempty"`
	UserCount int                `json:"user_count"`
}

// DepartmentSimple은 간단한 부서 정보입니다.
type DepartmentSimple struct {
	ID    uint   `json:"id"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// DepartmentTreeResponse는 부서 트리 구조 응답입니다.
type DepartmentTreeResponse struct {
	ID        uint                     `json:"id"`
	Code      string                   `json:"code"`
	Name      string                   `json:"name"`
	Level     int                      `json:"level"`
	IsActive  bool                     `json:"is_active"`
	UserCount int                      `json:"user_count"`
	Children  []DepartmentTreeResponse `json:"children,omitempty"`
}

// === 사용자 부서 관련 ===

// UserDepartmentUpdateRequest는 사용자 부서 변경 요청입니다.
type UserDepartmentUpdateRequest struct {
	DepartmentID uint   `json:"department_id" binding:"required"`
	Position     string `json:"position"`
	Reason       string `json:"reason"`
}

// UserWithDepartmentResponse는 부서 정보를 포함한 사용자 응답입니다.
type UserWithDepartmentResponse struct {
	ID         uint       `json:"id"`
	Username   string     `json:"username"`
	Role       string     `json:"role"`
	EmployeeID string     `json:"employee_id"`
	Position   string     `json:"position"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	JoinDate   *time.Time `json:"join_date"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// 부서 정보
	Department *DepartmentSimple `json:"department,omitempty"`
	HasSSHKey  bool              `json:"has_ssh_key"`
}

// DepartmentHistoryResponse는 부서 변경 이력 응답입니다.
type DepartmentHistoryResponse struct {
	ID           uint              `json:"id"`
	UserID       uint              `json:"user_id"`
	PreviousDept *DepartmentSimple `json:"previous_department,omitempty"`
	NewDept      DepartmentSimple  `json:"new_department"`
	ChangeDate   time.Time         `json:"change_date"`
	ChangedBy    UserSimple        `json:"changed_by"`
	Reason       string            `json:"reason"`
}

// === 사용자 관리 관련 ===

// UserProfileUpdate는 사용자 프로필 업데이트용 구조체입니다.
type UserProfileUpdate struct {
	Username    string     `json:"username,omitempty"`
	NewPassword string     `json:"new_password,omitempty"`
	EmployeeID  string     `json:"employee_id,omitempty"`
	Position    string     `json:"position,omitempty"`
	Email       string     `json:"email,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	JoinDate    *time.Time `json:"join_date,omitempty"`
}

// UserInfo는 사용자 기본 정보를 담는 구조체입니다.
type UserInfo struct {
	ID         uint              `json:"id"`
	Username   string            `json:"username"`
	Role       string            `json:"role"`
	EmployeeID string            `json:"employee_id"`
	Position   string            `json:"position"`
	Email      string            `json:"email"`
	Phone      string            `json:"phone"`
	JoinDate   *time.Time        `json:"join_date"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	HasSSHKey  bool              `json:"has_ssh_key"`
	Department *DepartmentSimple `json:"department,omitempty"`
}

// UserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보입니다.
type UserDetailWithKey struct {
	ID         uint              `json:"id"`
	Username   string            `json:"username"`
	Role       string            `json:"role"`
	EmployeeID string            `json:"employee_id"`
	Position   string            `json:"position"`
	Email      string            `json:"email"`
	Phone      string            `json:"phone"`
	JoinDate   *time.Time        `json:"join_date"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	HasSSHKey  bool              `json:"has_ssh_key"`
	Department *DepartmentSimple `json:"department,omitempty"`
	SSHKey     *SSHKeyResponse   `json:"ssh_key,omitempty"`
}

// UserResponse는 API용 사용자 정보 응답 구조체입니다.
type UserResponse struct {
	ID         uint              `json:"id"`
	Username   string            `json:"username"`
	Role       string            `json:"role"`
	EmployeeID string            `json:"employee_id"`
	Position   string            `json:"position"`
	Email      string            `json:"email"`
	Phone      string            `json:"phone"`
	JoinDate   *time.Time        `json:"join_date"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Department *DepartmentSimple `json:"department,omitempty"`
}

// UserSimple은 간단한 사용자 정보입니다.
type UserSimple struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Position string `json:"position,omitempty"`
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
	TotalDepartments int64 `json:"total_departments"`
}

// DepartmentStats는 부서별 통계 정보입니다.
type DepartmentStats struct {
	TotalDepartments   int64 `json:"total_departments"`
	ActiveDepartments  int64 `json:"active_departments"`
	TotalUsers         int64 `json:"total_users"`
	UsersWithDept      int64 `json:"users_with_department"`
	UsersWithoutDept   int64 `json:"users_without_department"`
	MaxDepartmentLevel int   `json:"max_department_level"`
}

// DepartmentUserStats는 부서별 사용자 통계입니다.
type DepartmentUserStats struct {
	Department  DepartmentSimple `json:"department"`
	UserCount   int              `json:"user_count"`
	AdminCount  int              `json:"admin_count"`
	SSHKeyCount int              `json:"ssh_key_count"`
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
	userInfo := UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Role:       string(user.Role),
		EmployeeID: user.EmployeeID,
		Position:   user.Position,
		Email:      user.Email,
		Phone:      user.Phone,
		JoinDate:   user.JoinDate,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		HasSSHKey:  hasSSHKey,
	}

	// 부서 정보 추가
	if user.Department != nil {
		userInfo.Department = &DepartmentSimple{
			ID:    user.Department.ID,
			Code:  user.Department.Code,
			Name:  user.Department.Name,
			Level: user.Department.Level,
		}
	}

	return userInfo
}

// ToUserDetailWithKey는 모델을 UserDetailWithKey로 변환합니다.
func ToUserDetailWithKey(user models.User, hasSSHKey bool, sshKey *SSHKeyResponse) UserDetailWithKey {
	userDetail := UserDetailWithKey{
		ID:         user.ID,
		Username:   user.Username,
		Role:       string(user.Role),
		EmployeeID: user.EmployeeID,
		Position:   user.Position,
		Email:      user.Email,
		Phone:      user.Phone,
		JoinDate:   user.JoinDate,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		HasSSHKey:  hasSSHKey,
		SSHKey:     sshKey,
	}

	// 부서 정보 추가
	if user.Department != nil {
		userDetail.Department = &DepartmentSimple{
			ID:    user.Department.ID,
			Code:  user.Department.Code,
			Name:  user.Department.Name,
			Level: user.Department.Level,
		}
	}

	return userDetail
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

func ToDepartmentResponse(dept models.Department) DepartmentResponse {
	response := DepartmentResponse{
		ID:          dept.ID,
		Code:        dept.Code,
		Name:        dept.Name,
		Description: dept.Description,
		ParentID:    dept.ParentID,
		Level:       dept.Level,
		IsActive:    dept.IsActive,
		CreatedAt:   dept.CreatedAt,
		UpdatedAt:   dept.UpdatedAt,
	}

	// 상위 부서 정보
	if dept.Parent != nil {
		response.Parent = &DepartmentSimple{
			ID:    dept.Parent.ID,
			Code:  dept.Parent.Code,
			Name:  dept.Parent.Name,
			Level: dept.Parent.Level,
		}
	}

	// 하위 부서 정보
	if len(dept.Children) > 0 {
		for _, child := range dept.Children {
			response.Children = append(response.Children, DepartmentSimple{
				ID:    child.ID,
				Code:  child.Code,
				Name:  child.Name,
				Level: child.Level,
			})
		}
	}

	return response
}

// ToUserWithDepartmentResponse는 모델을 UserWithDepartmentResponse로 변환합니다.
func ToUserWithDepartmentResponse(user models.User, hasSSHKey bool) UserWithDepartmentResponse {
	response := UserWithDepartmentResponse{
		ID:         user.ID,
		Username:   user.Username,
		Role:       string(user.Role),
		EmployeeID: user.EmployeeID,
		Position:   user.Position,
		Email:      user.Email,
		Phone:      user.Phone,
		JoinDate:   user.JoinDate,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		HasSSHKey:  hasSSHKey,
	}

	// 부서 정보
	if user.Department != nil {
		response.Department = &DepartmentSimple{
			ID:    user.Department.ID,
			Code:  user.Department.Code,
			Name:  user.Department.Name,
			Level: user.Department.Level,
		}
	}

	return response
}

// ToDepartmentHistoryResponse는 모델을 DepartmentHistoryResponse로 변환합니다.
func ToDepartmentHistoryResponse(history models.DepartmentHistory) DepartmentHistoryResponse {
	response := DepartmentHistoryResponse{
		ID:         history.ID,
		UserID:     history.UserID,
		ChangeDate: history.ChangeDate,
		Reason:     history.Reason,
		NewDept: DepartmentSimple{
			ID:    history.NewDept.ID,
			Code:  history.NewDept.Code,
			Name:  history.NewDept.Name,
			Level: history.NewDept.Level,
		},
		ChangedBy: UserSimple{
			ID:       history.ChangedByUser.ID,
			Username: history.ChangedByUser.Username,
			Position: history.ChangedByUser.Position,
		},
	}

	// 이전 부서 정보 (있는 경우)
	if history.PreviousDept != nil {
		response.PreviousDept = &DepartmentSimple{
			ID:    history.PreviousDept.ID,
			Code:  history.PreviousDept.Code,
			Name:  history.PreviousDept.Name,
			Level: history.PreviousDept.Level,
		}
	}

	return response
}
