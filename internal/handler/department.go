package handler

import (
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetDepartments는 부서 목록을 조회합니다.
func GetDepartments(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"

	departments, err := service.C().Department.GetAllDepartments(includeInactive)
	if err != nil {
		return InternalServerErrorResponse(c, "부서 목록 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, departments)
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
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrDepartmentNotFound:
				return NotFoundResponse(c, "부서를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "부서 조회 중 오류가 발생했습니다")
			}
		}
		return NotFoundResponse(c, "부서를 찾을 수 없습니다")
	}

	// 부서 사용자 수 조회
	users, err := service.C().Department.GetDepartmentUsers(uint(deptID))
	if err != nil {
		return InternalServerErrorResponse(c, "부서 사용자 조회 실패")
	}

	// 응답 데이터 구성 (단순화)
	responseData := map[string]interface{}{
		"department": department,
		"user_count": len(users),
		"users":      users,
	}

	return SuccessResponse(c, responseData)
}
