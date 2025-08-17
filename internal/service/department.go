package service

import (
	"errors"
	"log"
	"ssh-key-manager/internal/model"
	"strings"

	"gorm.io/gorm"
)

// CreateDepartment은 새로운 부서를 생성합니다.
func CreateDepartment(req model.DepartmentCreateRequest) (*model.Department, error) {
	log.Printf("🏢 새 부서 생성 시도: %s (%s)", req.Name, req.Code)

	// 입력값 검증
	if strings.TrimSpace(req.Code) == "" {
		return nil, errors.New("부서 코드를 입력해주세요")
	}
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("부서명을 입력해주세요")
	}

	// 레벨 계산
	level := 1
	if req.ParentID != nil {
		var parentDept model.Department
		if err := model.DB.First(&parentDept, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("상위 부서를 찾을 수 없습니다")
			}
			return nil, err
		}
		level = parentDept.Level + 1
	}

	// 부서 코드 중복 확인
	var existingDept model.Department
	if err := model.DB.Where("code = ?", strings.TrimSpace(req.Code)).First(&existingDept).Error; err == nil {
		return nil, errors.New("이미 사용 중인 부서 코드입니다")
	}

	// 부서 생성
	department := model.Department{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ParentID:    req.ParentID,
		Level:       level,
		IsActive:    true,
	}

	if err := model.DB.Create(&department).Error; err != nil {
		log.Printf("❌ 부서 생성 실패: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, errors.New("이미 사용 중인 부서 코드입니다")
		}
		return nil, errors.New("부서 생성 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 생성 완료: %s (ID: %d)", req.Name, department.ID)
	return &department, nil
}

// GetAllDepartments는 모든 부서 목록을 조회합니다.
func GetAllDepartments(includeInactive bool) ([]model.Department, error) {
	log.Printf("🏢 부서 목록 조회 (비활성 포함: %t)", includeInactive)

	var departments []model.Department
	query := model.DB.Preload("Parent").Preload("Children")

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("level ASC, code ASC").Find(&departments).Error; err != nil {
		log.Printf("❌ 부서 목록 조회 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ 부서 목록 조회 완료 (총 %d개)", len(departments))
	return departments, nil
}

// GetDepartmentTree는 부서 트리 구조를 조회합니다.
func GetDepartmentTree() ([]map[string]interface{}, error) {
	log.Printf("🌳 부서 트리 구조 조회")

	var departments []model.Department
	if err := model.DB.Where("is_active = ?", true).Order("level ASC, code ASC").Find(&departments).Error; err != nil {
		return nil, err
	}

	// 부서별 사용자 수 계산
	userCounts := make(map[uint]int)
	for _, dept := range departments {
		var count int64
		model.DB.Model(&model.User{}).Where("department_id = ?", dept.ID).Count(&count)
		userCounts[dept.ID] = int(count)
	}

	// 트리 구조 생성
	tree := buildDepartmentTree(departments, userCounts, nil)

	log.Printf("✅ 부서 트리 구조 조회 완료")
	return tree, nil
}

// buildDepartmentTree는 재귀적으로 부서 트리를 구성합니다.
func buildDepartmentTree(departments []model.Department, userCounts map[uint]int, parentID *uint) []map[string]interface{} {
	var tree []map[string]interface{}

	for _, dept := range departments {
		if (parentID == nil && dept.ParentID == nil) || (parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {
			node := map[string]interface{}{
				"id":         dept.ID,
				"code":       dept.Code,
				"name":       dept.Name,
				"level":      dept.Level,
				"is_active":  dept.IsActive,
				"user_count": userCounts[dept.ID],
				"children":   buildDepartmentTree(departments, userCounts, &dept.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// GetDepartmentByID는 특정 부서의 상세 정보를 조회합니다.
func GetDepartmentByID(deptID uint) (*model.Department, error) {
	log.Printf("🔍 부서 상세 정보 조회: ID %d", deptID)

	var department model.Department
	if err := model.DB.Preload("Parent").Preload("Children").First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("부서를 찾을 수 없습니다")
		}
		return nil, err
	}

	log.Printf("✅ 부서 상세 정보 조회 완료: %s", department.Name)
	return &department, nil
}

// UpdateDepartment는 부서 정보를 수정합니다.
func UpdateDepartment(deptID uint, req model.DepartmentUpdateRequest) (*model.Department, error) {
	log.Printf("✏️ 부서 정보 수정: ID %d", deptID)

	var department model.Department
	if err := model.DB.First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("부서를 찾을 수 없습니다")
		}
		return nil, err
	}

	// 업데이트할 필드 확인
	updates := make(map[string]interface{})

	if req.Code != "" && req.Code != department.Code {
		// 코드 중복 확인
		var existingDept model.Department
		if err := model.DB.Where("code = ? AND id != ?", strings.TrimSpace(req.Code), deptID).First(&existingDept).Error; err == nil {
			return nil, errors.New("이미 사용 중인 부서 코드입니다")
		}
		updates["code"] = strings.TrimSpace(req.Code)
	}

	if req.Name != "" && req.Name != department.Name {
		updates["name"] = strings.TrimSpace(req.Name)
	}

	if req.Description != department.Description {
		updates["description"] = strings.TrimSpace(req.Description)
	}

	if req.ParentID != nil && (department.ParentID == nil || *req.ParentID != *department.ParentID) {
		// 레벨 재계산
		if *req.ParentID == 0 {
			updates["parent_id"] = nil
			updates["level"] = 1
		} else {
			var parentDept model.Department
			if err := model.DB.First(&parentDept, *req.ParentID).Error; err != nil {
				return nil, errors.New("상위 부서를 찾을 수 없습니다")
			}
			updates["parent_id"] = *req.ParentID
			updates["level"] = parentDept.Level + 1
		}
	}

	if req.IsActive != nil && *req.IsActive != department.IsActive {
		updates["is_active"] = *req.IsActive
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := model.DB.Model(&department).Updates(updates).Error; err != nil {
			log.Printf("❌ 부서 정보 수정 실패: %v", err)
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				return nil, errors.New("이미 사용 중인 부서 코드입니다")
			}
			return nil, errors.New("부서 정보 수정 중 오류가 발생했습니다")
		}

		// 하위 부서들의 레벨 업데이트 (상위 부서가 변경된 경우)
		if _, hasParentChange := updates["parent_id"]; hasParentChange {
			updateChildDepartmentLevels(deptID)
		}

		// 업데이트된 정보 다시 조회
		model.DB.Preload("Parent").Preload("Children").First(&department, deptID)
	}

	log.Printf("✅ 부서 정보 수정 완료: %s", department.Name)
	return &department, nil
}

// updateChildDepartmentLevels는 하위 부서들의 레벨을 업데이트합니다.
func updateChildDepartmentLevels(parentID uint) error {
	var parent model.Department
	if err := model.DB.First(&parent, parentID).Error; err != nil {
		return err
	}

	var children []model.Department
	if err := model.DB.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	newLevel := parent.Level + 1
	for _, child := range children {
		model.DB.Model(&child).Update("level", newLevel)
		// 재귀적으로 하위 부서들도 업데이트
		updateChildDepartmentLevels(child.ID)
	}

	return nil
}

// DeleteDepartment는 부서를 삭제합니다.
func DeleteDepartment(deptID uint) error {
	log.Printf("🗑️ 부서 삭제: ID %d", deptID)

	var department model.Department
	if err := model.DB.First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("부서를 찾을 수 없습니다")
		}
		return err
	}

	// 소속 사용자 확인
	var userCount int64
	model.DB.Model(&model.User{}).Where("department_id = ?", deptID).Count(&userCount)
	if userCount > 0 {
		return errors.New("소속 사용자가 있는 부서는 삭제할 수 없습니다")
	}

	// 하위 부서 확인
	var childCount int64
	model.DB.Model(&model.Department{}).Where("parent_id = ?", deptID).Count(&childCount)
	if childCount > 0 {
		return errors.New("하위 부서가 있는 부서는 삭제할 수 없습니다")
	}

	// 부서 삭제
	if err := model.DB.Delete(&department).Error; err != nil {
		log.Printf("❌ 부서 삭제 실패: %v", err)
		return errors.New("부서 삭제 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 삭제 완료: %s", department.Name)
	return nil
}

// GetDepartmentUsers는 특정 부서의 사용자 목록을 조회합니다.
func GetDepartmentUsers(deptID uint) ([]model.User, error) {
	log.Printf("👥 부서 사용자 목록 조회: 부서 ID %d", deptID)

	// 부서 존재 확인
	var department model.Department
	if err := model.DB.First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("부서를 찾을 수 없습니다")
		}
		return nil, err
	}

	var users []model.User
	if err := model.DB.Where("department_id = ?", deptID).
		Preload("Department").
		Select("id, username, role, department_id, employee_id, position, email, phone, join_date, created_at, updated_at").
		Find(&users).Error; err != nil {
		return nil, err
	}

	log.Printf("✅ 부서 사용자 목록 조회 완료: %d명", len(users))
	return users, nil
}
