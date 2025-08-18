package dto

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
