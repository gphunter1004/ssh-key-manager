package types

import (
	"ssh-key-manager/models"
	"time"
)

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

// UserCreateRequest는 사용자 생성 요청 구조체입니다.
type UserCreateRequest struct {
	Username   string     `json:"username" binding:"required"`
	Password   string     `json:"password" binding:"required"`
	Role       string     `json:"role,omitempty"`
	EmployeeID string     `json:"employee_id,omitempty"`
	Position   string     `json:"position,omitempty"`
	Email      string     `json:"email,omitempty"`
	Phone      string     `json:"phone,omitempty"`
	JoinDate   *time.Time `json:"join_date,omitempty"`
}

// UserSearchRequest는 사용자 검색 요청 구조체입니다.
type UserSearchRequest struct {
	Query      string `json:"query" query:"q"`
	Role       string `json:"role" query:"role"`
	Department string `json:"department" query:"department"`
	HasSSHKey  *bool  `json:"has_ssh_key" query:"has_ssh_key"`
}

// UserListRequest는 사용자 목록 요청 구조체입니다.
type UserListRequest struct {
	PaginationRequest
	SortRequest
	UserSearchRequest
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
