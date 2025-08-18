package handler

import (
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetDepartments는 부서 목록을 조회합니다.
func GetDepartments(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"

	departments, err := service.C().Department.GetAllDepartments(includeInactive)
	if err != nil {
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, departments)
}

// GetDepartmentTree는 부서 트리 구조를 조회합니다.
func GetDepartmentTree(c echo.Context) error {
	tree, err := service.C().Department.GetDepartmentTree()
	if err != nil {
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, tree)
}

// GetDepartment는 특정 부서의 상세 정보를 조회합니다.
func GetDepartment(c echo.Context) error {
	// URL 파라미터에서 부서 ID 추출
	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	// 부서 상세 정보 조회
	department, err := service.C().Department.GetDepartmentByID(uint(deptID))
	if err != nil {
		return NotFoundResponse(c, err.Error())
	}

	// 부서 사용자 수 조회
	users, err := service.C().Department.GetDepartmentUsers(uint(deptID))
	if err != nil {
		return InternalServerErrorResponse(c, "부서 사용자 조회 실패")
	}

	// 응답 데이터 구성
	responseData := map[string]interface{}{
		"department": department,
		"user_count": len(users),
		"users":      users,
	}

	return SuccessResponse(c, responseData)
}
