// internal/handler/admin.go
package handler

import (
	"log"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// GetAllUsers는 모든 사용자 목록을 반환합니다 (관리자용).
func GetAllUsers(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	users, err := service.C().User.GetAllUsers()
	if err != nil {
		log.Printf("❌ 사용자 목록 조회 실패 (관리자 ID: %d): %v", adminUserID, err)
		return InternalServerErrorResponse(c, "사용자 목록 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, users)
}

// GetUserDetail는 특정 사용자의 상세 정보를 반환합니다 (관리자용).
func GetUserDetail(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	// URL 파라미터에서 사용자 ID 추출
	userIDParam := c.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 사용자 ID입니다")
	}

	userDetail, err := service.C().User.GetUserDetailWithKey(uint(userID))
	if err != nil {
		log.Printf("❌ 사용자 상세 조회 실패 (관리자 ID: %d, 대상 ID: %d): %v", adminUserID, userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "사용자 조회 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "사용자 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, userDetail)
}

// UpdateUserRole은 사용자 권한을 변경합니다 (관리자용).
func UpdateUserRole(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	targetUserID, err := ParseUserIDParam(c)
	if err != nil {
		return BadRequestResponse(c, err.Error())
	}

	var req dto.UserRoleUpdateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	err = service.C().User.UpdateUserRole(adminUserID, targetUserID, req.Role)
	if err != nil {
		LogAdminError("사용자 권한 변경", adminUserID, targetUserID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "권한이 없습니다")
			case model.ErrCannotDeleteSelf:
				return BadRequestResponse(c, "자신의 권한은 변경할 수 없습니다")
			case model.ErrLastAdmin:
				return BadRequestResponse(c, "최소 1명의 관리자가 필요합니다")
			case model.ErrInvalidInput:
				return BadRequestResponse(c, be.Message)
			default:
				return InternalServerErrorResponse(c, "권한 변경 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "권한 변경 중 오류가 발생했습니다")
	}

	userDetail, err := service.C().User.GetUserDetailWithKey(targetUserID)
	if err != nil {
		return InternalServerErrorResponse(c, "사용자 정보 조회 실패")
	}

	LogAdminAction("사용자 권한 변경", adminUserID, targetUserID, req.Role)
	return SuccessWithMessageResponse(c, "사용자 권한이 변경되었습니다", userDetail)
}

// DeleteUser는 사용자를 삭제합니다 (관리자용).
func DeleteUser(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	targetUserID, err := ParseUserIDParam(c)
	if err != nil {
		return BadRequestResponse(c, err.Error())
	}

	if adminUserID == targetUserID {
		return BadRequestResponse(c, "자신의 계정은 삭제할 수 없습니다")
	}

	err = service.C().User.DeleteUser(adminUserID, targetUserID)
	if err != nil {
		LogAdminError("사용자 삭제", adminUserID, targetUserID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "권한이 없습니다")
			case model.ErrCannotDeleteSelf:
				return BadRequestResponse(c, "자신의 계정은 삭제할 수 없습니다")
			case model.ErrLastAdmin:
				return BadRequestResponse(c, "최소 1명의 관리자가 필요합니다")
			default:
				return InternalServerErrorResponse(c, "사용자 삭제 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "사용자 삭제 중 오류가 발생했습니다")
	}

	LogAdminAction("사용자 삭제", adminUserID, targetUserID)
	return SuccessWithMessageResponse(c, "사용자가 삭제되었습니다", nil)
}

// CreateDepartment는 새로운 부서를 생성합니다.
func CreateDepartment(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	var req dto.DepartmentCreateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	department, err := service.C().Department.CreateDepartment(req)
	if err != nil {
		log.Printf("❌ 부서 생성 실패 (관리자 ID: %d): %v", adminUserID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrDepartmentExists:
				return ConflictResponse(c, "이미 사용 중인 부서 코드입니다")
			case model.ErrDepartmentNotFound:
				return NotFoundResponse(c, "상위 부서를 찾을 수 없습니다")
			case model.ErrRequiredField:
				return BadRequestResponse(c, be.Message)
			default:
				return InternalServerErrorResponse(c, "부서 생성 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "부서 생성 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 생성 성공 (관리자 ID: %d): %s", adminUserID, department.Name)
	return CreatedResponse(c, "부서가 성공적으로 생성되었습니다", department)
}

// UpdateDepartment는 부서 정보를 수정합니다.
func UpdateDepartment(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	var req dto.DepartmentUpdateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	department, err := service.C().Department.UpdateDepartment(uint(deptID), req)
	if err != nil {
		log.Printf("❌ 부서 수정 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrDepartmentNotFound:
				return NotFoundResponse(c, "부서를 찾을 수 없습니다")
			case model.ErrDepartmentExists:
				return ConflictResponse(c, "이미 사용 중인 부서 코드입니다")
			default:
				return InternalServerErrorResponse(c, "부서 수정 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "부서 수정 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 수정 성공 (관리자 ID: %d, 부서 ID: %d)", adminUserID, deptID)
	return SuccessWithMessageResponse(c, "부서 정보가 수정되었습니다", department)
}

// DeleteDepartment는 부서를 삭제합니다.
func DeleteDepartment(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	err = service.C().Department.DeleteDepartment(uint(deptID))
	if err != nil {
		log.Printf("❌ 부서 삭제 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrDepartmentNotFound:
				return NotFoundResponse(c, "부서를 찾을 수 없습니다")
			case model.ErrDepartmentHasUsers:
				return BadRequestResponse(c, "소속 사용자가 있는 부서는 삭제할 수 없습니다")
			case model.ErrDepartmentHasChild:
				return BadRequestResponse(c, "하위 부서가 있는 부서는 삭제할 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "부서 삭제 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "부서 삭제 중 오류가 발생했습니다")
	}

	log.Printf("✅ 부서 삭제 성공 (관리자 ID: %d, 부서 ID: %d)", adminUserID, deptID)
	return SuccessWithMessageResponse(c, "부서가 삭제되었습니다", nil)
}

// GetDepartmentUsers는 특정 부서의 사용자 목록을 조회합니다.
func GetDepartmentUsers(c echo.Context) error {
	adminUserID, _ := GetUserID(c)

	deptIDParam := c.Param("id")
	deptID, err := strconv.ParseUint(deptIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 부서 ID입니다")
	}

	users, err := service.C().Department.GetDepartmentUsers(uint(deptID))
	if err != nil {
		log.Printf("❌ 부서 사용자 조회 실패 (관리자 ID: %d, 부서 ID: %d): %v", adminUserID, deptID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrDepartmentNotFound:
				return NotFoundResponse(c, "부서를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "부서 사용자 조회 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "부서 사용자 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, users)
}
