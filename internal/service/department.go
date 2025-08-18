package service

import (
	"log"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"strings"

	"gorm.io/gorm"
)

// DepartmentService 부서 관리 서비스 (단순 CRUD)
type DepartmentService struct {
	deptRepo *repository.DepartmentRepository
}

// NewDepartmentService 부서 서비스 생성자
func NewDepartmentService(deptRepo *repository.DepartmentRepository) *DepartmentService {
	return &DepartmentService{deptRepo: deptRepo}
}

// CreateDepartment 새로운 부서를 생성합니다.
func (ds *DepartmentService) CreateDepartment(req dto.DepartmentCreateRequest) (*model.Department, error) {
	log.Printf("🏢 새 부서 생성 시도: %s (%s)", req.Name, req.Code)

	// 입력값 검증
	if err := ds.validateDepartmentCreateRequest(req); err != nil {
		return nil, err
	}

	// 부서 코드 중복 확인
	exists, err := ds.deptRepo.ExistsByCode(strings.TrimSpace(req.Code))
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

	// 부서 생성 (단순 구조)
	department := &model.Department{
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}

	if err := ds.deptRepo.Create(department); err != nil {
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

	departments, err := ds.deptRepo.FindAll(includeInactive)
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

// GetDepartmentByID 특정 부서의 상세 정보를 조회합니다.
func (ds *DepartmentService) GetDepartmentByID(deptID uint) (*model.Department, error) {
	log.Printf("🔍 부서 상세 정보 조회: ID %d", deptID)

	department, err := ds.deptRepo.FindByID(deptID)
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
func (ds *DepartmentService) UpdateDepartment(deptID uint, req dto.DepartmentUpdateRequest) (*model.Department, error) {
	log.Printf("✏️ 부서 정보 수정: ID %d", deptID)

	department, err := ds.deptRepo.FindByID(deptID)
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
		exists, err := ds.deptRepo.ExistsByCode(strings.TrimSpace(req.Code))
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

	if req.IsActive != nil && *req.IsActive != department.IsActive {
		updates["is_active"] = *req.IsActive
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := ds.deptRepo.Update(deptID, updates); err != nil {
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

		// 업데이트된 정보 다시 조회
		department, err = ds.deptRepo.FindByID(deptID)
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

	department, err := ds.deptRepo.FindByID(deptID)
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
	userCount, err := ds.deptRepo.CountUsers(deptID)
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

	// 부서 삭제
	if err := ds.deptRepo.Delete(deptID); err != nil {
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
	_, err := ds.deptRepo.FindByID(deptID)
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

	users, err := ds.deptRepo.FindUsersByDepartment(deptID)
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
func (ds *DepartmentService) validateDepartmentCreateRequest(req dto.DepartmentCreateRequest) error {
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
