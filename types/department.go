package types

import (
	"ssh-key-manager/models"
	"time"
)

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

// DepartmentListRequest는 부서 목록 조회 요청입니다.
type DepartmentListRequest struct {
	PaginationRequest
	SortRequest
	IncludeInactive bool   `json:"include_inactive" query:"include_inactive"`
	ParentID        *uint  `json:"parent_id" query:"parent_id"`
	Level           *int   `json:"level" query:"level"`
	SearchQuery     string `json:"q" query:"q"`
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

// DepartmentMoveRequest는 부서 이동 요청입니다.
type DepartmentMoveRequest struct {
	NewParentID *uint  `json:"new_parent_id"`
	Reason      string `json:"reason"`
}

// DepartmentBulkUpdateRequest는 부서 일괄 업데이트 요청입니다.
type DepartmentBulkUpdateRequest struct {
	DepartmentIDs []uint                  `json:"department_ids" binding:"required"`
	Updates       DepartmentUpdateRequest `json:"updates"`
}

// === 변환 헬퍼 함수들 ===

// ToDepartmentResponse는 모델을 DepartmentResponse로 변환합니다.
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
