package handler

import (
	"log"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetAllUsers는 모든 사용자 목록을 반환합니다 (관리자용).
func GetAllUsers(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	users, err := service.GetAllUsers()
	if err != nil {
		log.Printf("❌ 사용자 목록 조회 실패 (관리자 ID: %d): %v", adminUserID, err)
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, users)
}

// GetUserDetail는 특정 사용자의 상세 정보를 반환합니다 (관리자용).
func GetUserDetail(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	// URL 파라미터에서 사용자 ID 추출
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	userDetail, err := service.GetUserDetailWithKey(uint(userID))
	if err != nil {
		log.Printf("❌ 사용자 상세 조회 실패 (관리자 ID: %d, 대상 ID: %d): %v", adminUserID, userID, err)
		return NotFoundResponse(c, err.Error())
	}

	return SuccessResponse(c, userDetail)
}

// UpdateUserRole은 사용자 권한을 변경합니다 (관리자용).
func UpdateUserRole(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	// 대상 사용자 ID 추출
	userIDParam := c.Param("id")
	targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 요청 바디 파싱
	var req model.UserRoleUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	// 권한 변경 실행
	err = service.UpdateUserRole(adminUserID, uint(targetUserID), req.Role)
	if err != nil {
		log.Printf("❌ 사용자 권한 변경 실패 (관리자 ID: %d, 대상 ID: %d): %v", adminUserID, targetUserID, err)
		return BadRequestResponse(c, err.Error())
	}

	// 변경된 사용자 정보 조회
	userDetail, err := service.GetUserDetailWithKey(uint(targetUserID))
	if err != nil {
		return InternalServerErrorResponse(c, "사용자 정보 조회 실패")
	}

	log.Printf("✅ 사용자 권한 변경 성공 (관리자 ID: %d, 대상 ID: %d → %s)", adminUserID, targetUserID, req.Role)
	return SuccessWithMessageResponse(c, "사용자 권한이 변경되었습니다", userDetail)
}

// DeleteUser는 사용자를 삭제합니다 (관리자용).
func DeleteUser(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	// 대상 사용자 ID 추출
	userIDParam := c.Param("id")
	targetUserID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 자신을 삭제하려는지 확인
	if adminUserID == uint(targetUserID) {
		return BadRequestResponse(c, "자신의 계정은 삭제할 수 없습니다")
	}

	// 사용자 삭제 실행
	err = service.DeleteUser(adminUserID, uint(targetUserID))
	if err != nil {
		log.Printf("❌ 사용자 삭제 실패 (관리자 ID: %d, 대상 ID: %d): %v", adminUserID, targetUserID, err)
		return BadRequestResponse(c, err.Error())
	}

	log.Printf("✅ 사용자 삭제 성공 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)
	return SuccessWithMessageResponse(c, "사용자가 삭제되었습니다", nil)
}

// CreateDepartment는 새로운 부서를 생성합니다.
func CreateDepartment(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	var req model.DepartmentCreateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	department, err := service.CreateDepartment(req)
	if err != nil {
		log.Printf("❌ 부서 생성 실패 (관리자 ID: %d): %v", adminUserID, err)
		return BadRequestResponse(c, err.Error())
	}

	log.Printf("✅ 부서 생성 성공 (관리자 ID: %d): %s", adminUserID, department.Name)
	return CreatedResponse(c, "부서가 성공적으로 생성되었습니다", department)
}

// UpdateDepartment는 부서 정보를 수정합니다.
func UpdateDepartment(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	var req model.DepartmentUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	department, err := service.UpdateDepartment(uint(deptID), req)
	if err != nil {
		log.Printf("❌ 부서 수정 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		return BadRequestResponse(c, err.Error())
	}

	log.Printf("✅ 부서 수정 성공 (관리자 ID: %d, 부서 ID: %d)", adminUserID, deptID)
	return SuccessWithMessageResponse(c, "부서 정보가 수정되었습니다", department)
}

// DeleteDepartment는 부서를 삭제합니다.
func DeleteDepartment(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	err = service.DeleteDepartment(uint(deptID))
	if err != nil {
		log.Printf("❌ 부서 삭제 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		return BadRequestResponse(c, err.Error())
	}

	log.Printf("✅ 부서 삭제 성공 (관리자 ID: %d, 부서 ID: %d)", adminUserID, deptID)
	return SuccessWithMessageResponse(c, "부서가 삭제되었습니다", nil)
}

// GetDepartmentUsers는 특정 부서의 사용자 목록을 조회합니다.
func GetDepartmentUsers(c echo.Context) error {
	adminUserID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	users, err := service.GetDepartmentUsers(uint(deptID))
	if err != nil {
		log.Printf("❌ 부서 사용자 조회 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, users)
}
