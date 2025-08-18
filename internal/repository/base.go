package repository

import (
	"ssh-key-manager/internal/model"

	"gorm.io/gorm"
)

// BaseRepository 모든 Repository의 기본 구조체
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository 기본 Repository 생성자
func NewBaseRepository() (*BaseRepository, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, err
	}
	return &BaseRepository{db: db}, nil
}

// GetDB 데이터베이스 인스턴스 반환
func (br *BaseRepository) GetDB() *gorm.DB {
	return br.db
}
