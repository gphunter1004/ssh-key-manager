package model

import (
	"fmt"

	"gorm.io/gorm"
)

var DB *gorm.DB

// SetDB는 전역 DB 인스턴스를 설정합니다.
func SetDB(db *gorm.DB) {
	DB = db
}

// SafeDB는 안전한 DB 인스턴스를 반환합니다.
func SafeDB() *gorm.DB {
	if DB == nil {
		panic("Database not initialized. Call SetDB() first.")
	}
	return DB
}

// IsDBInitialized는 DB가 초기화되었는지 확인합니다.
func IsDBInitialized() bool {
	return DB != nil
}

// GetDB는 DB 인스턴스를 안전하게 반환합니다 (에러 반환 버전).
func GetDB() (*gorm.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return DB, nil
}
