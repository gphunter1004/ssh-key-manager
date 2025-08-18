package repository

import (
	"ssh-key-manager/internal/model"
)

// ServerRepositoryImpl Server Repository 구현체
type ServerRepositoryImpl struct {
	*BaseRepository
}

// NewServerRepository Server Repository 생성자
func NewServerRepository() (ServerRepository, error) {
	base, err := NewBaseRepository()
	if err != nil {
		return nil, err
	}
	return &ServerRepositoryImpl{BaseRepository: base}, nil
}

// Create 서버 생성
func (sr *ServerRepositoryImpl) Create(server *model.Server) error {
	return sr.db.Create(server).Error
}

// FindByID ID로 서버 조회
func (sr *ServerRepositoryImpl) FindByID(id uint) (*model.Server, error) {
	var server model.Server
	err := sr.db.First(&server, id).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// FindByUserID 사용자 ID로 서버 목록 조회
func (sr *ServerRepositoryImpl) FindByUserID(userID uint) ([]model.Server, error) {
	var servers []model.Server
	err := sr.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&servers).Error
	return servers, err
}

// Update 서버 정보 업데이트
func (sr *ServerRepositoryImpl) Update(id uint, updates map[string]interface{}) error {
	return sr.db.Model(&model.Server{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 서버 삭제
func (sr *ServerRepositoryImpl) Delete(id uint) error {
	return sr.db.Delete(&model.Server{}, id).Error
}

// ExistsByUserAndHost 사용자와 호스트로 서버 존재 확인
func (sr *ServerRepositoryImpl) ExistsByUserAndHost(userID uint, host string, port int) (bool, error) {
	var count int64
	err := sr.db.Model(&model.Server{}).
		Where("user_id = ? AND host = ? AND port = ?", userID, host, port).
		Count(&count).Error
	return count > 0, err
}
