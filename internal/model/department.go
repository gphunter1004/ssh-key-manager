package model

import (
	"gorm.io/gorm"
)

// Department는 부서 정보를 저장하는 모델입니다 (단순 구조).
type Department struct {
	gorm.Model
	Code        string `gorm:"unique;not null;index" json:"code"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	IsActive    bool   `gorm:"not null;default:true" json:"is_active"`

	// 관계 (기본)
	Users []User `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}
