package controllers

import (
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/types"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateDepartment는 새로운 부서를 생성합니다.
func CreateDepartment(c echo.Context) error {
	var req types.DepartmentCreateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	department, err := services.CreateDepartment(req)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.CreatedResponse(c, "부서가 성공적으로 생성되었습니다", department)
}

// GetDepartments는 부서 목록을 조회합니다.
func GetDepartments(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"

	departments, err := services.GetAllDepartments(includeInactive)
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.ListResponse(c, departments, len(departments))
}

// GetDepartmentTree는 부서 트리 구조를 조회합니다.
func GetDepartmentTree(c echo.Context) error {
	tree, err := services.GetDepartmentTree()
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, tree)
}

// GetDepartment는 특정 부서의 상세 정보를 조회합니다.
func GetDepartment(c echo.Context) error {
	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	department, err := services.GetDepartmentByID(uint(deptID))
	if err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, department)
}

// UpdateDepartment는 부서 정보를 수정합니다.
func UpdateDepartment(c echo.Context) error {
	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	var req types.DepartmentUpdateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	department, err := services.UpdateDepartment(uint(deptID), req)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "부서 정보가 수정되었습니다", department)
}

// DeleteDepartment는 부서를 삭제합니다.
func DeleteDepartment(c echo.Context) error {
	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	if err := services.DeleteDepartment(uint(deptID)); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "부서가 삭제되었습니다", nil)
}

// GetDepartmentUsers는 특정 부서의 사용자 목록을 조회합니다.
func GetDepartmentUsers(c echo.Context) error {
	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	users, err := services.GetDepartmentUsers(uint(deptID))
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.ListResponse(c, users, len(users))
}

// UpdateUserDepartment는 사용자의 부서를 변경합니다.
func UpdateUserDepartment(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	// 현재 로그인한 사용자 ID 추출
	changedBy, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var req types.UserDepartmentUpdateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	if err := services.UpdateUserDepartment(uint(userID), req, changedBy); err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "사용자 부서가 변경되었습니다", nil)
}

// GetUserDepartmentHistory는 사용자의 부서 변경 이력을 조회합니다.
func GetUserDepartmentHistory(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	histories, err := services.GetUserDepartmentHistory(uint(userID))
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.ListResponse(c, histories, len(histories))
}
