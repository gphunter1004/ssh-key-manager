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

	// 보안 기능 추가
	IsActive         bool       `gorm:"not null;default:true" json:"is_active"`       // 활성/비활성 상태
	IsLocked         bool       `gorm:"not null;default:false" json:"is_locked"`      // 계정 잠금 상태
	FailedLoginCount int        `gorm:"not null;default:0" json:"failed_login_count"` // 로그인 실패 횟수
	LockedAt         *time.Time `gorm:"index" json:"locked_at,omitempty"`             // 잠금 시간
	LastLoginAt      *time.Time `gorm:"index" json:"last_login_at,omitempty"`         // 마지막 로그인 시간

	// 부서 관련 (기본 기능 유지)
	DepartmentID *uint `gorm:"index" json:"department_id"`

	// 관계 (핵심 기능)
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	SSHKeys    []SSHKey    `gorm:"foreignKey:UserID" json:"ssh_keys,omitempty"`
	Servers    []Server    `gorm:"foreignKey:UserID" json:"servers,omitempty"`
}

// IsAccountAccessible 계정이 접근 가능한 상태인지 확인합니다.
func (u *User) IsAccountAccessible() bool {
	return u.IsActive && !u.IsLocked
}

// ShouldLockAccount 계정을 잠궈야 하는지 확인합니다.
func (u *User) ShouldLockAccount() bool {
	return u.FailedLoginCount >= 5
}

// LockAccount 계정을 잠급니다.
func (u *User) LockAccount() {
	u.IsLocked = true
	now := time.Now()
	u.LockedAt = &now
}

// UnlockAccount 계정 잠금을 해제합니다.
func (u *User) UnlockAccount() {
	u.IsLocked = false
	u.FailedLoginCount = 0
	u.LockedAt = nil
}

// IncrementFailedLogin 로그인 실패 횟수를 증가시킵니다.
func (u *User) IncrementFailedLogin() {
	u.FailedLoginCount++
	if u.ShouldLockAccount() {
		u.LockAccount()
	}
}

// ResetFailedLogin 로그인 성공 시 실패 횟수를 초기화합니다.
func (u *User) ResetFailedLogin() {
	u.FailedLoginCount = 0
}

// UpdateLastLogin 마지막 로그인 시간을 업데이트합니다.
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}
