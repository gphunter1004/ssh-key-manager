package models

import "gorm.io/gorm"

type SSHKey struct {
	gorm.Model
	UserID     uint   `gorm:"not null;index"`                                // 인덱스 추가로 성능 향상
	Algorithm  string `gorm:"not null;default:'RSA'"`                        // 알고리즘 (RSA, ECDSA 등)
	Bits       int    `gorm:"not null;default:4096"`                         // 키 크기
	PrivateKey string `gorm:"type:text;not null"`                            // PEM 형식 개인키
	PublicKey  string `gorm:"type:text;not null"`                            // SSH 공개키 (authorized_keys 형식)
	PEM        string `gorm:"type:text;not null"`                            // PEM 형식 개인키 (PrivateKey와 동일)
	PPK        string `gorm:"type:text;not null"`                            // PuTTY 형식 개인키
	User       User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"` // 외래키 제약조건 추가
}
