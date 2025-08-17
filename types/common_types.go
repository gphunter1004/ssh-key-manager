package types

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

// FilterRequest는 필터링 요청을 위한 구통 구조체입니다.
type FilterRequest struct {
	Filters map[string]interface{} `json:"filters" query:"filters"`
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

// ValidateSort는 정렬 요청의 유효성을 검사합니다.
func (s *SortRequest) ValidateSort(allowedFields []string) bool {
	if s.SortBy == "" {
		return true
	}

	for _, field := range allowedFields {
		if field == s.SortBy {
			return true
		}
	}
	return false
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

// BusinessError는 비즈니스 로직 에러를 위한 구조체입니다.
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// === 상태 및 메타데이터 ===

// HealthCheckResponse는 헬스체크 응답 구조체입니다.
type HealthCheckResponse struct {
	Status      string                   `json:"status"` // OK, WARNING, ERROR
	Version     string                   `json:"version"`
	Timestamp   string                   `json:"timestamp"`
	Services    map[string]ServiceHealth `json:"services"`
	Uptime      string                   `json:"uptime"`
	Environment string                   `json:"environment"`
}

// ServiceHealth는 개별 서비스의 상태입니다.
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// VersionInfo는 버전 정보 구조체입니다.
type VersionInfo struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	GoVersion string `json:"go_version"`
}

// === 배치 작업 관련 ===

// BatchRequest는 배치 요청 공통 구조체입니다.
type BatchRequest struct {
	IDs    []uint      `json:"ids" binding:"required"`
	Action string      `json:"action" binding:"required"`
	Data   interface{} `json:"data,omitempty"`
}

// BatchResponse는 배치 응답 공통 구조체입니다.
type BatchResponse struct {
	Total   int                    `json:"total"`
	Success int                    `json:"success"`
	Failed  int                    `json:"failed"`
	Results []BatchItemResult      `json:"results"`
	Summary map[string]interface{} `json:"summary,omitempty"`
}

// BatchItemResult는 배치 작업의 개별 항목 결과입니다.
type BatchItemResult struct {
	ID      uint   `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// === 파일 업로드/다운로드 관련 ===

// FileUploadResponse는 파일 업로드 응답 구조체입니다.
type FileUploadResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url,omitempty"`
	ID       string `json:"id,omitempty"`
}

// FileDownloadRequest는 파일 다운로드 요청 구조체입니다.
type FileDownloadRequest struct {
	FileID   string `json:"file_id" query:"file_id"`
	Format   string `json:"format" query:"format"`
	Filename string `json:"filename" query:"filename"`
}

// === 알림 관련 ===

// NotificationRequest는 알림 요청 구조체입니다.
type NotificationRequest struct {
	Type     string                 `json:"type"` // INFO, WARNING, ERROR, SUCCESS
	Title    string                 `json:"title"`
	Message  string                 `json:"message"`
	UserIDs  []uint                 `json:"user_ids,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationResponse는 알림 응답 구조체입니다.
type NotificationResponse struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Read      bool   `json:"read"`
	CreatedAt string `json:"created_at"`
}

// === 작업 진행 상황 관련 ===

// ProgressResponse는 작업 진행 상황 응답입니다.
type ProgressResponse struct {
	TaskID      string      `json:"task_id"`
	Status      string      `json:"status"`   // PENDING, RUNNING, COMPLETED, FAILED
	Progress    float64     `json:"progress"` // 0.0 - 1.0
	Message     string      `json:"message"`
	StartedAt   string      `json:"started_at"`
	CompletedAt string      `json:"completed_at,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// TaskRequest는 비동기 작업 요청 구조체입니다.
type TaskRequest struct {
	Type       string                 `json:"type" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority,omitempty"`
}

// === 감사 로그 관련 ===

// AuditLogEntry는 감사 로그 항목입니다.
type AuditLogEntry struct {
	ID         uint                   `json:"id"`
	UserID     uint                   `json:"user_id"`
	Username   string                 `json:"username"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resource_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Timestamp  string                 `json:"timestamp"`
	Result     string                 `json:"result"` // SUCCESS, FAILED
}
