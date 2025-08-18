package model

import (
	"gorm.io/gorm"
)

// SSHKey는 SSH 키 정보를 저장하는 모델입니다.
type SSHKey struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Algorithm  string `gorm:"not null;default:'RSA'" json:"algorithm"`
	Bits       int    `gorm:"not null;default:4096" json:"bits"`
	PrivateKey string `gorm:"type:text;not null" json:"private_key"`
	PublicKey  string `gorm:"type:text;not null" json:"public_key"`
	PEM        string `gorm:"type:text;not null" json:"pem"`
	PPK        string `gorm:"type:text;not null" json:"ppk"`

	// 관계
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}
