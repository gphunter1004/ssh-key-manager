package service

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"strings"

	"gorm.io/gorm"
)

// DepartmentService 부서 관리 서비스
type DepartmentService struct {
	repos *repository.Repositories
}

// NewDepartmentService 부서 서비스 생성자
func NewDepartmentService(repos *repository.Repositories) *DepartmentService {
	return &DepartmentService{repos: repos}
}

// CreateDepartment 새로운 부서를 생성합니다.
func (ds *DepartmentService) CreateDepartment(req model.DepartmentCreateRequest) (*model.Department, error) {
	log.Printf("🏢 새 부서 생성 시도: %s (%s)", req.Name, req.Code)

	// 입력값 검증
	if err := ds.validateDepartmentCreateRequest(req); err != nil {
		return nil, err
	}

	// 레벨 계산
	level := 1
	if req.ParentID != nil {
		parentDept, err := ds.repos.Department.FindByID(*req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, model.NewBusinessError(
					model.ErrDepartmentNotFound,
					"상위 부서를 찾을 수 없습니다",
				)
			}
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"상위 부서 조회 중 오류가 발생했습니다",
			)
		}
		level = parentDept.Level + 1
	}

	// 부서 코드 중복 확인
	exists, err := ds.repos.Department.ExistsByCode(strings.TrimSpace(req.Code))
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 코드 중복 확인 중 오류가 발생했습니다",
		)
	}
	if exists {
		return nil, model.NewBusinessError(
			model.ErrDepartmentExists,
			"이미 사용 중인 부서 코드입니다",
		)
	}

	// 부서 생성
	department := &model.Department{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		ParentID:    req.ParentID,
		Level:       level,
		IsActive:    true,
	}

	if err := ds.repos.Department.Create(department); err != nil {
		log.Printf("❌ 부서 생성 실패: %v", err)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, model.NewBusinessError(
				model.ErrDepartmentExists,
				"이미 사용 중인 부서 코드입니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 생성 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 부서 생성 완료: %s (ID: %d)", req.Name, department.ID)
	return department, nil
}

// GetAllDepartments 모든 부서 목록을 조회합니다.
func (ds *DepartmentService) GetAllDepartments(includeInactive bool) ([]model.Department, error) {
	log.Printf("🏢 부서 목록 조회 (비활성 포함: %t)", includeInactive)

	departments, err := ds.repos.Department.FindAll(includeInactive)
	if err != nil {
		log.Printf("❌ 부서 목록 조회 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 목록 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 부서 목록 조회 완료 (총 %d개)", len(departments))
	return departments, nil
}

// GetDepartmentTree 부서 트리 구조를 조회합니다.
func (ds *DepartmentService) GetDepartmentTree() ([]map[string]interface{}, error) {
	log.Printf("🌳 부서 트리 구조 조회")

	departments, err := ds.repos.Department.FindAll(false) // 활성 부서만
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 목록 조회 중 오류가 발생했습니다",
		)
	}

	// 부서별 사용자 수 계산
	userCounts := make(map[uint]int64)
	for _, dept := range departments {
		count, err := ds.repos.Department.CountUsers(dept.ID)
		if err != nil {
			count = 0 // 에러 시 0으로 설정
		}
		userCounts[dept.ID] = count
	}

	// 트리 구조 생성
	tree := ds.buildDepartmentTree(departments, userCounts, nil)

	log.Printf("✅ 부서 트리 구조 조회 완료")
	return tree, nil
}

// GetDepartmentByID 특정 부서의 상세 정보를 조회합니다.
func (ds *DepartmentService) GetDepartmentByID(deptID uint) (*model.Department, error) {
	log.Printf("🔍 부서 상세 정보 조회: ID %d", deptID)

	department, err := ds.repos.Department.FindByID(deptID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrDepartmentNotFound,
				"부서를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 부서 상세 정보 조회 완료: %s", department.Name)
	return department, nil
}

// UpdateDepartment 부서 정보를 수정합니다.
func (ds *DepartmentService) UpdateDepartment(deptID uint, req model.DepartmentUpdateRequest) (*model.Department, error) {
	log.Printf("✏️ 부서 정보 수정: ID %d", deptID)

	department, err := ds.repos.Department.FindByID(deptID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrDepartmentNotFound,
				"부서를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 조회 중 오류가 발생했습니다",
		)
	}

	// 업데이트할 필드 확인
	updates := make(map[string]interface{})

	if req.Code != "" && req.Code != department.Code {
		// 코드 중복 확인
		exists, err := ds.repos.Department.ExistsByCode(strings.TrimSpace(req.Code))
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"부서 코드 중복 확인 중 오류가 발생했습니다",
			)
		}
		if exists {
			return nil, model.NewBusinessError(
				model.ErrDepartmentExists,
				"이미 사용 중인 부서 코드입니다",
			)
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
			parentDept, err := ds.repos.Department.FindByID(*req.ParentID)
			if err != nil {
				return nil, model.NewBusinessError(
					model.ErrDepartmentNotFound,
					"상위 부서를 찾을 수 없습니다",
				)
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
		if err := ds.repos.Department.Update(deptID, updates); err != nil {
			log.Printf("❌ 부서 정보 수정 실패: %v", err)
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				return nil, model.NewBusinessError(
					model.ErrDepartmentExists,
					"이미 사용 중인 부서 코드입니다",
				)
			}
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"부서 정보 수정 중 오류가 발생했습니다",
			)
		}

		// 하위 부서들의 레벨 업데이트 (상위 부서가 변경된 경우)
		if _, hasParentChange := updates["parent_id"]; hasParentChange {
			ds.updateChildDepartmentLevels(deptID)
		}

		// 업데이트된 정보 다시 조회
		department, err = ds.repos.Department.FindByID(deptID)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"업데이트된 부서 정보 조회 실패",
			)
		}
	}

	log.Printf("✅ 부서 정보 수정 완료: %s", department.Name)
	return department, nil
}

// DeleteDepartment 부서를 삭제합니다.
func (ds *DepartmentService) DeleteDepartment(deptID uint) error {
	log.Printf("🗑️ 부서 삭제: ID %d", deptID)

	department, err := ds.repos.Department.FindByID(deptID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewBusinessError(
				model.ErrDepartmentNotFound,
				"부서를 찾을 수 없습니다",
			)
		}
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 조회 중 오류가 발생했습니다",
		)
	}

	// 소속 사용자 확인
	userCount, err := ds.repos.Department.CountUsers(deptID)
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 사용자 수 확인 중 오류가 발생했습니다",
		)
	}
	if userCount > 0 {
		return model.NewBusinessError(
			model.ErrDepartmentHasUsers,
			"소속 사용자가 있는 부서는 삭제할 수 없습니다",
		)
	}

	// 하위 부서 확인
	children, err := ds.repos.Department.FindChildren(deptID)
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"하위 부서 확인 중 오류가 발생했습니다",
		)
	}
	if len(children) > 0 {
		return model.NewBusinessError(
			model.ErrDepartmentHasChild,
			"하위 부서가 있는 부서는 삭제할 수 없습니다",
		)
	}

	// 부서 삭제
	if err := ds.repos.Department.Delete(deptID); err != nil {
		log.Printf("❌ 부서 삭제 실패: %v", err)
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 삭제 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 부서 삭제 완료: %s", department.Name)
	return nil
}

// GetDepartmentUsers 특정 부서의 사용자 목록을 조회합니다.
func (ds *DepartmentService) GetDepartmentUsers(deptID uint) ([]model.User, error) {
	log.Printf("👥 부서 사용자 목록 조회: 부서 ID %d", deptID)

	// 부서 존재 확인
	_, err := ds.repos.Department.FindByID(deptID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrDepartmentNotFound,
				"부서를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 조회 중 오류가 발생했습니다",
		)
	}

	users, err := ds.repos.Department.FindUsersByDepartment(deptID)
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"부서 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 모든 사용자의 비밀번호 필드 제거
	for i := range users {
		users[i].Password = ""
	}

	log.Printf("✅ 부서 사용자 목록 조회 완료: %d명", len(users))
	return users, nil
}

// ========== 내부 헬퍼 함수들 ==========

// validateDepartmentCreateRequest 부서 생성 요청을 검증합니다.
func (ds *DepartmentService) validateDepartmentCreateRequest(req model.DepartmentCreateRequest) error {
	if strings.TrimSpace(req.Code) == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"부서 코드를 입력해주세요",
		)
	}
	if strings.TrimSpace(req.Name) == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"부서명을 입력해주세요",
		)
	}
	return nil
}

// buildDepartmentTree 재귀적으로 부서 트리를 구성합니다.
func (ds *DepartmentService) buildDepartmentTree(departments []model.Department, userCounts map[uint]int64, parentID *uint) []map[string]interface{} {
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
				"children":   ds.buildDepartmentTree(departments, userCounts, &dept.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// updateChildDepartmentLevels 하위 부서들의 레벨을 업데이트합니다.
func (ds *DepartmentService) updateChildDepartmentLevels(parentID uint) error {
	parent, err := ds.repos.Department.FindByID(parentID)
	if err != nil {
		return err
	}

	children, err := ds.repos.Department.FindChildren(parentID)
	if err != nil {
		return err
	}

	newLevel := parent.Level + 1
	for _, child := range children {
		updates := map[string]interface{}{
			"level": newLevel,
		}
		ds.repos.Department.Update(child.ID, updates)
		// 재귀적으로 하위 부서들도 업데이트
		ds.updateChildDepartmentLevels(child.ID)
	}

	return nil
}
