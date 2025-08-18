package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole은 사용자 권한을 정의합니다.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User는 사용자 정보를 저장하는 모델입니다.
type User struct {
	gorm.Model
	Username string   `gorm:"unique;not null" json:"username"`
	Password string   `gorm:"not null" json:"-"`
	Role     UserRole `gorm:"not null;default:'user'" json:"role"`

	// 부서 관련 필드
	DepartmentID *uint      `gorm:"index" json:"department_id"`
	EmployeeID   string     `gorm:"unique;size:20" json:"employee_id"`
	Position     string     `gorm:"size:50" json:"position"`
	JoinDate     *time.Time `json:"join_date"`
	Email        string     `gorm:"unique;size:100" json:"email"`
	Phone        string     `gorm:"size:20" json:"phone"`

	// 관계
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	SSHKeys    []SSHKey    `gorm:"foreignKey:UserID" json:"ssh_keys,omitempty"`
	Servers    []Server    `gorm:"foreignKey:UserID" json:"servers,omitempty"`
}
