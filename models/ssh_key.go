// models/ssh_key.go
package models

import "gorm.io/gorm"

type SSHKey struct {
	gorm.Model
	UserID     uint   `gorm:"not null"`
	Algorithm  string `gorm:"not null;default:'RSA'"`
	Bits       int    `gorm:"not null;default:4096"`
	PrivateKey string `gorm:"type:text;not null"`
	PublicKey  string `gorm:"type:text;not null"`
	PEM        string `gorm:"type:text;not null"`
	PPK        string `gorm:"type:text;not null"`
	User       User
}
