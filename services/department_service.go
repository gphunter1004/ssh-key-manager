package services

import (
	"errors"
	"log"
	"ssh-key-manager/models"
	"ssh-key-manager/types"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateDepartment은 새로운 부서를 생성합니다.
func CreateDepartment(req types.DepartmentCreateRequest) (*types.DepartmentResponse, error) {
	log.Printf("🏢 새 부서 생성 시도: %s (%s)", req.Name, req.Code)

	// 레벨 계산
	level := 1
	if req.ParentID != nil {
		var parentDept models.Department
		if err := models.DB.First(&parentDept, *req.ParentID).Error; err != nil {
			return nil, errors.New("상위 부서를 찾을 수 없습니다")
		}
		level = parentDept.Level + 1
	}

	// 부서 생성
	department := models.Department{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ParentID:    req.ParentID,
		Level:       level,
		IsActive:    true,
	}

	if err := models.DB.Create(&department).Error; err != nil {
		log.Printf("❌ 부서 생성 실패: %v", err)
		return nil, errors.New("부서 생성 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 생성 완료: %s (ID: %d)", req.Name, department.ID)

	// 응답 데이터 생성
	response := types.ToDepartmentResponse(department)
	return &response, nil
}

// GetAllDepartments는 모든 부서 목록을 조회합니다.
func GetAllDepartments(includeInactive bool) ([]types.DepartmentResponse, error) {
	log.Printf("🏢 부서 목록 조회 (비활성 포함: %t)", includeInactive)

	var departments []models.Department
	query := models.DB.Preload("Parent").Preload("Children")

	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("level ASC, code ASC").Find(&departments).Error; err != nil {
		log.Printf("❌ 부서 목록 조회 실패: %v", err)
		return nil, err
	}

	// 각 부서의 사용자 수 계산
	var responses []types.DepartmentResponse
	for _, dept := range departments {
		var userCount int64
		models.DB.Model(&models.User{}).Where("department_id = ?", dept.ID).Count(&userCount)

		response := types.ToDepartmentResponse(dept)
		response.UserCount = int(userCount)
		responses = append(responses, response)
	}

	log.Printf("✅ 부서 목록 조회 완료 (총 %d개)", len(responses))
	return responses, nil
}

// GetDepartmentTree는 부서 트리 구조를 조회합니다.
func GetDepartmentTree() ([]types.DepartmentTreeResponse, error) {
	log.Printf("🌳 부서 트리 구조 조회")

	var departments []models.Department
	if err := models.DB.Where("is_active = ?", true).Order("level ASC, code ASC").Find(&departments).Error; err != nil {
		return nil, err
	}

	// 부서별 사용자 수 계산
	userCounts := make(map[uint]int)
	for _, dept := range departments {
		var count int64
		models.DB.Model(&models.User{}).Where("department_id = ?", dept.ID).Count(&count)
		userCounts[dept.ID] = int(count)
	}

	// 트리 구조 생성
	tree := buildDepartmentTree(departments, userCounts, nil)

	log.Printf("✅ 부서 트리 구조 조회 완료")
	return tree, nil
}

// buildDepartmentTree는 재귀적으로 부서 트리를 구성합니다.
func buildDepartmentTree(departments []models.Department, userCounts map[uint]int, parentID *uint) []types.DepartmentTreeResponse {
	var tree []types.DepartmentTreeResponse

	for _, dept := range departments {
		if (parentID == nil && dept.ParentID == nil) || (parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {
			node := types.DepartmentTreeResponse{
				ID:        dept.ID,
				Code:      dept.Code,
				Name:      dept.Name,
				Level:     dept.Level,
				IsActive:  dept.IsActive,
				UserCount: userCounts[dept.ID],
				Children:  buildDepartmentTree(departments, userCounts, &dept.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// GetDepartmentByID는 특정 부서의 상세 정보를 조회합니다.
func GetDepartmentByID(deptID uint) (*types.DepartmentResponse, error) {
	log.Printf("🔍 부서 상세 정보 조회: ID %d", deptID)

	var department models.Department
	if err := models.DB.Preload("Parent").Preload("Children").First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("부서를 찾을 수 없습니다")
		}
		return nil, err
	}

	// 사용자 수 계산
	var userCount int64
	models.DB.Model(&models.User{}).Where("department_id = ?", deptID).Count(&userCount)

	response := types.ToDepartmentResponse(department)
	response.UserCount = int(userCount)

	log.Printf("✅ 부서 상세 정보 조회 완료: %s", department.Name)
	return &response, nil
}

// UpdateDepartment는 부서 정보를 수정합니다.
func UpdateDepartment(deptID uint, req types.DepartmentUpdateRequest) (*types.DepartmentResponse, error) {
	log.Printf("✏️ 부서 정보 수정: ID %d", deptID)

	var department models.Department
	if err := models.DB.First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("부서를 찾을 수 없습니다")
		}
		return nil, err
	}

	// 업데이트할 필드 확인
	updates := make(map[string]interface{})

	if req.Code != "" && req.Code != department.Code {
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
			var parentDept models.Department
			if err := models.DB.First(&parentDept, *req.ParentID).Error; err != nil {
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
		if err := models.DB.Model(&department).Updates(updates).Error; err != nil {
			log.Printf("❌ 부서 정보 수정 실패: %v", err)
			return nil, errors.New("부서 정보 수정 중 오류가 발생했습니다")
		}

		// 하위 부서들의 레벨 업데이트 (상위 부서가 변경된 경우)
		if _, hasParentChange := updates["parent_id"]; hasParentChange {
			updateChildDepartmentLevels(deptID)
		}
	}

	// 업데이트된 정보 다시 조회
	return GetDepartmentByID(deptID)
}

// updateChildDepartmentLevels는 하위 부서들의 레벨을 업데이트합니다.
func updateChildDepartmentLevels(parentID uint) error {
	var parent models.Department
	if err := models.DB.First(&parent, parentID).Error; err != nil {
		return err
	}

	var children []models.Department
	if err := models.DB.Where("parent_id = ?", parentID).Find(&children).Error; err != nil {
		return err
	}

	newLevel := parent.Level + 1
	for _, child := range children {
		models.DB.Model(&child).Update("level", newLevel)
		// 재귀적으로 하위 부서들도 업데이트
		updateChildDepartmentLevels(child.ID)
	}

	return nil
}

// DeleteDepartment는 부서를 삭제합니다.
func DeleteDepartment(deptID uint) error {
	log.Printf("🗑️ 부서 삭제: ID %d", deptID)

	var department models.Department
	if err := models.DB.First(&department, deptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("부서를 찾을 수 없습니다")
		}
		return err
	}

	// 소속 사용자 확인
	var userCount int64
	models.DB.Model(&models.User{}).Where("department_id = ?", deptID).Count(&userCount)
	if userCount > 0 {
		return errors.New("소속 사용자가 있는 부서는 삭제할 수 없습니다")
	}

	// 하위 부서 확인
	var childCount int64
	models.DB.Model(&models.Department{}).Where("parent_id = ?", deptID).Count(&childCount)
	if childCount > 0 {
		return errors.New("하위 부서가 있는 부서는 삭제할 수 없습니다")
	}

	// 부서 삭제
	if err := models.DB.Delete(&department).Error; err != nil {
		log.Printf("❌ 부서 삭제 실패: %v", err)
		return errors.New("부서 삭제 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 삭제 완료: %s", department.Name)
	return nil
}

// UpdateUserDepartment는 사용자의 부서를 변경합니다.
func UpdateUserDepartment(userID uint, req types.UserDepartmentUpdateRequest, changedBy uint) error {
	log.Printf("👤 사용자 부서 변경: 사용자 ID %d -> 부서 ID %d", userID, req.DepartmentID)

	// 사용자 존재 확인
	var user models.User
	if err := models.DB.Preload("Department").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("사용자를 찾을 수 없습니다")
		}
		return err
	}

	// 새 부서 존재 확인
	var newDept models.Department
	if err := models.DB.First(&newDept, req.DepartmentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("부서를 찾을 수 없습니다")
		}
		return err
	}

	// 이미 같은 부서인 경우
	if user.DepartmentID != nil && *user.DepartmentID == req.DepartmentID {
		return errors.New("이미 해당 부서에 소속되어 있습니다")
	}

	// 트랜잭션 시작
	tx := models.DB.Begin()

	// 이력 저장
	history := models.DepartmentHistory{
		UserID:      userID,
		NewDeptID:   req.DepartmentID,
		NewDeptCode: newDept.Code,
		NewDeptName: newDept.Name,
		ChangeDate:  time.Now(),
		ChangedBy:   changedBy,
		Reason:      req.Reason,
	}

	if user.Department != nil {
		history.PreviousDeptID = &user.Department.ID
		history.PreviousDeptCode = &user.Department.Code
		history.PreviousDeptName = &user.Department.Name
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return errors.New("부서 변경 이력 저장 실패")
	}

	// 사용자 부서 정보 업데이트
	updates := map[string]interface{}{
		"department_id": req.DepartmentID,
	}
	if req.Position != "" {
		updates["position"] = req.Position
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		return errors.New("사용자 부서 정보 업데이트 실패")
	}

	tx.Commit()

	log.Printf("✅ 사용자 부서 변경 완료: %s -> %s",
		func() string {
			if user.Department != nil {
				return user.Department.Name
			}
			return "미배정"
		}(),
		newDept.Name)

	return nil
}

// GetDepartmentUsers는 특정 부서의 사용자 목록을 조회합니다.
func GetDepartmentUsers(deptID uint) ([]types.UserWithDepartmentResponse, error) {
	log.Printf("👥 부서 사용자 목록 조회: 부서 ID %d", deptID)

	var users []models.User
	if err := models.DB.Where("department_id = ?", deptID).Preload("Department").Find(&users).Error; err != nil {
		return nil, err
	}

	// SSH 키 존재 여부 확인
	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	keyMap := make(map[uint]bool)
	if len(userIDs) > 0 {
		var keyUsers []struct {
			UserID uint
		}
		models.DB.Model(&models.SSHKey{}).Select("user_id").Where("user_id IN ?", userIDs).Find(&keyUsers)
		for _, ku := range keyUsers {
			keyMap[ku.UserID] = true
		}
	}

	// 응답 데이터 구성
	var responses []types.UserWithDepartmentResponse
	for _, user := range users {
		response := types.ToUserWithDepartmentResponse(user, keyMap[user.ID])
		responses = append(responses, response)
	}

	log.Printf("✅ 부서 사용자 목록 조회 완료: %d명", len(responses))
	return responses, nil
}

// GetUserDepartmentHistory는 사용자의 부서 변경 이력을 조회합니다.
func GetUserDepartmentHistory(userID uint) ([]types.DepartmentHistoryResponse, error) {
	log.Printf("📋 사용자 부서 변경 이력 조회: 사용자 ID %d", userID)

	var histories []models.DepartmentHistory
	if err := models.DB.Where("user_id = ?", userID).
		Preload("PreviousDept").
		Preload("NewDept").
		Preload("ChangedByUser").
		Order("change_date DESC").
		Find(&histories).Error; err != nil {
		return nil, err
	}

	// 응답 데이터 구성
	var responses []types.DepartmentHistoryResponse
	for _, history := range histories {
		response := types.ToDepartmentHistoryResponse(history)
		responses = append(responses, response)
	}

	log.Printf("✅ 부서 변경 이력 조회 완료: %d건", len(responses))
	return responses, nil
}
