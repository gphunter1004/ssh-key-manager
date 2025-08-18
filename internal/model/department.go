package model

import (
	"gorm.io/gorm"
)

// Department는 부서 정보를 저장하는 모델입니다.
type Department struct {
	gorm.Model
	Code        string `gorm:"unique;not null;index" json:"code"`
	Name        string `gorm:"not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `gorm:"index" json:"parent_id"`
	Level       int    `gorm:"not null;default:1" json:"level"`
	IsActive    bool   `gorm:"not null;default:true" json:"is_active"`

	// 관계
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Users    []User       `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}
