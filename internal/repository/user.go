package repository

import (
	"ssh-key-manager/internal/model"
)

// UserRepository User Repository 구현체 (인터페이스 제거)
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository User Repository 생성자
func NewUserRepository() (*UserRepository, error) {
	base, err := NewBaseRepository()
	if err != nil {
		return nil, err
	}
	return &UserRepository{BaseRepository: base}, nil
}

// Create 사용자 생성
func (ur *UserRepository) Create(user *model.User) error {
	return ur.db.Create(user).Error
}

// FindByID ID로 사용자 조회
func (ur *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := ur.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 사용자명으로 사용자 조회
func (ur *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := ur.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 사용자 정보 업데이트
func (ur *UserRepository) Update(id uint, updates map[string]interface{}) error {
	return ur.db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 사용자 삭제
func (ur *UserRepository) Delete(id uint) error {
	return ur.db.Delete(&model.User{}, id).Error
}

// FindAll 모든 사용자 조회
func (ur *UserRepository) FindAll() ([]model.User, error) {
	var users []model.User
	err := ur.db.Select("id, username, role, created_at, updated_at").Find(&users).Error
	return users, err
}

// ExistsByID ID로 사용자 존재 확인
func (ur *UserRepository) ExistsByID(id uint) (bool, error) {
	var count int64
	err := ur.db.Model(&model.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername 사용자명으로 사용자 존재 확인
func (ur *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := ur.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// CountByRole 역할별 사용자 수 조회
func (ur *UserRepository) CountByRole(role model.UserRole) (int64, error) {
	var count int64
	err := ur.db.Model(&model.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}
