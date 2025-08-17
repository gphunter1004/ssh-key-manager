package models

import (
	"time"

	"gorm.io/gorm"
)

// Department는 부서 정보를 저장하는 모델입니다.
type Department struct {
	gorm.Model
	Code        string `gorm:"unique;not null;index"` // 부서 코드 (예: IT001, HR002)
	Name        string `gorm:"not null"`              // 부서명 (예: IT개발팀, 인사팀)
	Description string `gorm:"type:text"`             // 부서 설명
	ParentID    *uint  `gorm:"index"`                 // 상위 부서 ID (NULL 가능)
	Level       int    `gorm:"not null;default:1"`    // 부서 레벨 (1: 최상위, 2: 2단계 등)
	IsActive    bool   `gorm:"not null;default:true"` // 활성 상태

	// 관계 정의
	Parent   *Department  `gorm:"foreignKey:ParentID"`     // 상위 부서
	Children []Department `gorm:"foreignKey:ParentID"`     // 하위 부서들
	Users    []User       `gorm:"foreignKey:DepartmentID"` // 부서 소속 사용자들
}

// TableName은 테이블명을 명시적으로 지정합니다.
func (Department) TableName() string {
	return "departments"
}

// DepartmentHistory는 부서 변경 이력을 저장하는 모델입니다.
type DepartmentHistory struct {
	gorm.Model
	UserID           uint      `gorm:"not null;index"`    // 사용자 ID
	PreviousDeptID   *uint     `gorm:"index"`             // 이전 부서 ID
	PreviousDeptCode *string   `gorm:"size:20"`           // 이전 부서 코드
	PreviousDeptName *string   `gorm:"size:100"`          // 이전 부서명
	NewDeptID        uint      `gorm:"not null;index"`    // 새 부서 ID
	NewDeptCode      string    `gorm:"not null;size:20"`  // 새 부서 코드
	NewDeptName      string    `gorm:"not null;size:100"` // 새 부서명
	ChangeDate       time.Time `gorm:"not null"`          // 변경 일시
	ChangedBy        uint      `gorm:"not null;index"`    // 변경한 사람 ID
	Reason           string    `gorm:"type:text"`         // 변경 사유

	// 관계 정의
	User          User        `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PreviousDept  *Department `gorm:"foreignKey:PreviousDeptID"`
	NewDept       Department  `gorm:"foreignKey:NewDeptID"`
	ChangedByUser User        `gorm:"foreignKey:ChangedBy"`
}

// TableName은 테이블명을 명시적으로 지정합니다.
func (DepartmentHistory) TableName() string {
	return "department_histories"
}
