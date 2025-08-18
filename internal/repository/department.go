package repository

import (
	"ssh-key-manager/internal/model"
)

// DepartmentRepositoryImpl Department Repository 구현체
type DepartmentRepositoryImpl struct {
	*BaseRepository
}

// NewDepartmentRepository Department Repository 생성자
func NewDepartmentRepository() (DepartmentRepository, error) {
	base, err := NewBaseRepository()
	if err != nil {
		return nil, err
	}
	return &DepartmentRepositoryImpl{BaseRepository: base}, nil
}

// Create 부서 생성
func (dr *DepartmentRepositoryImpl) Create(dept *model.Department) error {
	return dr.db.Create(dept).Error
}

// FindByID ID로 부서 조회
func (dr *DepartmentRepositoryImpl) FindByID(id uint) (*model.Department, error) {
	var dept model.Department
	err := dr.db.Preload("Parent").Preload("Children").First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindAll 모든 부서 조회
func (dr *DepartmentRepositoryImpl) FindAll(includeInactive bool) ([]model.Department, error) {
	var departments []model.Department
	query := dr.db.Preload("Parent").Preload("Children")

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	err := query.Order("level ASC, code ASC").Find(&departments).Error
	return departments, err
}

// Update 부서 정보 업데이트
func (dr *DepartmentRepositoryImpl) Update(id uint, updates map[string]interface{}) error {
	return dr.db.Model(&model.Department{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 부서 삭제
func (dr *DepartmentRepositoryImpl) Delete(id uint) error {
	return dr.db.Delete(&model.Department{}, id).Error
}

// ExistsByCode 부서 코드로 존재 확인
func (dr *DepartmentRepositoryImpl) ExistsByCode(code string) (bool, error) {
	var count int64
	err := dr.db.Model(&model.Department{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

// FindChildren 하위 부서 조회
func (dr *DepartmentRepositoryImpl) FindChildren(parentID uint) ([]model.Department, error) {
	var children []model.Department
	err := dr.db.Where("parent_id = ?", parentID).Find(&children).Error
	return children, err
}

// CountUsers 부서의 사용자 수 조회
func (dr *DepartmentRepositoryImpl) CountUsers(deptID uint) (int64, error) {
	var count int64
	err := dr.db.Model(&model.User{}).Where("department_id = ?", deptID).Count(&count).Error
	return count, err
}

// FindUsersByDepartment 부서의 사용자 목록 조회
func (dr *DepartmentRepositoryImpl) FindUsersByDepartment(deptID uint) ([]model.User, error) {
	var users []model.User
	err := dr.db.Where("department_id = ?", deptID).
		Preload("Department").
		Select("id, username, role, department_id, employee_id, position, email, phone, join_date, created_at, updated_at").
		Find(&users).Error
	return users, err
}
