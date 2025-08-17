package handler

import (
	"ssh-key-manager/internal/service"

	"github.com/labstack/echo/v4"
)

// GetDepartments는 부서 목록을 조회합니다.
func GetDepartments(c echo.Context) error {
	includeInactive := c.QueryParam("include_inactive") == "true"

	departments, err := service.GetAllDepartments(includeInactive)
	if err != nil {
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, departments)
}

// GetDepartmentTree는 부서 트리 구조를 조회합니다.
func GetDepartmentTree(c echo.Context) error {
	tree, err := service.GetDepartmentTree()
	if err != nil {
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, tree)
}

// GetDepartment는 특정 부서의 상세 정보를 조회합니다.
func GetDepartment(c echo.Context) error {
	// TODO: URL 파라미터에서 부서 ID 추출 후 구현
	return InternalServerErrorResponse(c, "아직 구현되지 않았습니다")
}
