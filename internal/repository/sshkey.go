package repository

import (
	"ssh-key-manager/internal/model"

	"gorm.io/gorm"
)

// SSHKeyRepository SSH Key Repository 구현체 (인터페이스 제거)
type SSHKeyRepository struct {
	*BaseRepository
}

// NewSSHKeyRepository SSH Key Repository 생성자
func NewSSHKeyRepository() (*SSHKeyRepository, error) {
	base, err := NewBaseRepository()
	if err != nil {
		return nil, err
	}
	return &SSHKeyRepository{BaseRepository: base}, nil
}

// Create SSH 키 생성
func (sr *SSHKeyRepository) Create(key *model.SSHKey) error {
	return sr.db.Create(key).Error
}

// FindByUserID 사용자 ID로 SSH 키 조회
func (sr *SSHKeyRepository) FindByUserID(userID uint) (*model.SSHKey, error) {
	var key model.SSHKey
	err := sr.db.Where("user_id = ?", userID).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// DeleteByUserID 사용자 ID로 SSH 키 삭제
func (sr *SSHKeyRepository) DeleteByUserID(userID uint) error {
	return sr.db.Where("user_id = ?", userID).Delete(&model.SSHKey{}).Error
}

// ExistsByUserID 사용자 ID로 SSH 키 존재 확인
func (sr *SSHKeyRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := sr.db.Model(&model.SSHKey{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

// ReplaceUserKey 사용자의 SSH 키를 교체 (트랜잭션)
func (sr *SSHKeyRepository) ReplaceUserKey(userID uint, key *model.SSHKey) error {
	return sr.db.Transaction(func(tx *gorm.DB) error {
		// 기존 키 삭제
		if err := tx.Where("user_id = ?", userID).Delete(&model.SSHKey{}).Error; err != nil {
			return err
		}

		// 새 키 생성
		return tx.Create(key).Error
	})
}

// GetStatistics SSH 키 통계 조회 (단순화)
func (sr *SSHKeyRepository) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 전체 키 수만 조회
	var totalKeys int64
	if err := sr.db.Model(&model.SSHKey{}).Count(&totalKeys).Error; err != nil {
		return nil, err
	}
	stats["total_keys"] = totalKeys

	return stats, nil
}
