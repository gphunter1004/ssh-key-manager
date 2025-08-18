package model

import (
	"gorm.io/gorm"
)

// Server는 원격 서버 정보를 저장하는 모델입니다.
type Server struct {
	gorm.Model
	UserID      uint   `gorm:"not null;index" json:"user_id"`
	Name        string `gorm:"not null" json:"name"`
	Host        string `gorm:"not null" json:"host"`
	Port        int    `gorm:"not null;default:22" json:"port"`
	Username    string `gorm:"not null" json:"username"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"not null;default:'active'" json:"status"`

	// 관계
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// ServerKeyDeployment는 서버별 키 배포 기록을 저장하는 모델입니다.
type ServerKeyDeployment struct {
	gorm.Model
	ServerID   uint            `gorm:"not null;index" json:"server_id"`
	SSHKeyID   uint            `gorm:"not null;index" json:"ssh_key_id"`
	UserID     uint            `gorm:"not null;index" json:"user_id"`
	Status     string          `gorm:"not null;default:'pending'" json:"status"`
	DeployedAt *gorm.DeletedAt `gorm:"index" json:"deployed_at"`
	ErrorMsg   string          `gorm:"type:text" json:"error_msg"`

	// 관계
	Server Server `gorm:"foreignKey:ServerID;constraint:OnDelete:CASCADE" json:"server,omitempty"`
	SSHKey SSHKey `gorm:"foreignKey:SSHKeyID;constraint:OnDelete:CASCADE" json:"ssh_key,omitempty"`
	User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}
