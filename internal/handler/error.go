package handler

import (
	"net/http"
	"ssh-key-manager/internal/model"

	"github.com/labstack/echo/v4"
)

// StandardErrorResponse는 표준 에러 응답을 생성합니다.
func StandardErrorResponse(c echo.Context, statusCode int, errorCode model.ErrorCode, message string, details ...string) error {
	apiError := &model.APIError{
		Code:    errorCode,
		Message: message,
	}

	if len(details) > 0 {
		apiError.Details = details[0]
	}

	return c.JSON(statusCode, model.APIResponse{
		Success: false,
		Error:   apiError,
	})
}

// BusinessErrorResponse는 BusinessError를 API 응답으로 변환합니다.
func BusinessErrorResponse(c echo.Context, statusCode int, err *model.BusinessError) error {
	return StandardErrorResponse(c, statusCode, err.Code, err.Message, err.Details)
}

// ========== 인증 관련 에러 응답 ==========

// InvalidCredentialsResponse는 인증 실패 응답을 생성합니다.
func InvalidCredentialsResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusUnauthorized,
		model.ErrInvalidCredentials, "사용자명 또는 비밀번호가 올바르지 않습니다")
}

// InvalidTokenResponse는 유효하지 않은 토큰 응답을 생성합니다.
func InvalidTokenResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusUnauthorized,
		model.ErrInvalidToken, "유효하지 않은 토큰입니다")
}

// PermissionDeniedResponse는 권한 거부 응답을 생성합니다.
func PermissionDeniedResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusForbidden,
		model.ErrPermissionDenied, "권한이 없습니다")
}

// AdminRequiredResponse는 관리자 권한 필요 응답을 생성합니다.
func AdminRequiredResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusForbidden,
		model.ErrPermissionDenied, "관리자 권한이 필요합니다")
}

// ========== 사용자 관련 에러 응답 ==========

// UserNotFoundResponse는 사용자 없음 응답을 생성합니다.
func UserNotFoundResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusNotFound,
		model.ErrUserNotFound, "사용자를 찾을 수 없습니다")
}

// UserAlreadyExistsResponse는 사용자 중복 응답을 생성합니다.
func UserAlreadyExistsResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusConflict,
		model.ErrUserAlreadyExists, "이미 사용 중인 사용자명입니다")
}

// WeakPasswordResponse는 약한 비밀번호 응답을 생성합니다.
func WeakPasswordResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrWeakPassword, "비밀번호는 최소 4자 이상이어야 합니다")
}

// CannotDeleteSelfResponse는 자신 삭제 불가 응답을 생성합니다.
func CannotDeleteSelfResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrCannotDeleteSelf, "자신의 계정은 삭제할 수 없습니다")
}

// LastAdminResponse는 마지막 관리자 보호 응답을 생성합니다.
func LastAdminResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrLastAdmin, "최소 1명의 관리자가 필요합니다")
}

// ========== 서버 관련 에러 응답 ==========

// ServerNotFoundResponse는 서버 없음 응답을 생성합니다.
func ServerNotFoundResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusNotFound,
		model.ErrServerNotFound, "서버를 찾을 수 없습니다")
}

// ServerAlreadyExistsResponse는 서버 중복 응답을 생성합니다.
func ServerAlreadyExistsResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusConflict,
		model.ErrServerExists, "이미 등록된 서버입니다")
}

// ServerConnectionFailedResponse는 서버 연결 실패 응답을 생성합니다.
func ServerConnectionFailedResponse(c echo.Context, details string) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrConnectionFailed, "서버 연결에 실패했습니다", details)
}

// InvalidServerIDResponse는 유효하지 않은 서버 ID 응답을 생성합니다.
func InvalidServerIDResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrInvalidServerID, "유효하지 않은 서버 ID입니다")
}

// ========== SSH 키 관련 에러 응답 ==========

// SSHKeyNotFoundResponse는 SSH 키 없음 응답을 생성합니다.
func SSHKeyNotFoundResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusNotFound,
		model.ErrSSHKeyNotFound, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
}

// SSHKeyGenerationFailedResponse는 SSH 키 생성 실패 응답을 생성합니다.
func SSHKeyGenerationFailedResponse(c echo.Context, details string) error {
	return StandardErrorResponse(c, http.StatusInternalServerError,
		model.ErrSSHKeyGeneration, "SSH 키 생성에 실패했습니다", details)
}

// SSHKeyDeploymentFailedResponse는 SSH 키 배포 실패 응답을 생성합니다.
func SSHKeyDeploymentFailedResponse(c echo.Context, details string) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrSSHKeyDeployment, "SSH 키 배포에 실패했습니다", details)
}

// ========== 부서 관련 에러 응답 ==========

// DepartmentNotFoundResponse는 부서 없음 응답을 생성합니다.
func DepartmentNotFoundResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusNotFound,
		model.ErrDepartmentNotFound, "부서를 찾을 수 없습니다")
}

// DepartmentAlreadyExistsResponse는 부서 중복 응답을 생성합니다.
func DepartmentAlreadyExistsResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusConflict,
		model.ErrDepartmentExists, "이미 사용 중인 부서 코드입니다")
}

// DepartmentHasUsersResponse는 부서에 사용자 있음 응답을 생성합니다.
func DepartmentHasUsersResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrDepartmentHasUsers, "소속 사용자가 있는 부서는 삭제할 수 없습니다")
}

// DepartmentHasChildrenResponse는 부서에 하위 부서 있음 응답을 생성합니다.
func DepartmentHasChildrenResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrDepartmentHasChild, "하위 부서가 있는 부서는 삭제할 수 없습니다")
}

// InvalidDepartmentIDResponse는 유효하지 않은 부서 ID 응답을 생성합니다.
func InvalidDepartmentIDResponse(c echo.Context) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrInvalidDeptID, "유효하지 않은 부서 ID입니다")
}

// ========== 검증 관련 에러 응답 ==========

// ValidationFailedResponse는 입력 검증 실패 응답을 생성합니다.
func ValidationFailedResponse(c echo.Context, message string) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrValidationFailed, message)
}

// InvalidInputResponse는 유효하지 않은 입력 응답을 생성합니다.
func InvalidInputResponse(c echo.Context, message string) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrInvalidInput, message)
}

// RequiredFieldResponse는 필수 필드 누락 응답을 생성합니다.
func RequiredFieldResponse(c echo.Context, fieldName string) error {
	return StandardErrorResponse(c, http.StatusBadRequest,
		model.ErrRequiredField, fieldName+"을(를) 입력해주세요")
}

// ========== 시스템 에러 응답 ==========

// InternalServerErrorResponse는 서버 내부 에러 응답을 생성합니다.
func InternalServerErrorResponse(c echo.Context, details string) error {
	return StandardErrorResponse(c, http.StatusInternalServerError,
		model.ErrInternalServer, "서버 내부 오류가 발생했습니다", details)
}

// DatabaseErrorResponse는 데이터베이스 에러 응답을 생성합니다.
func DatabaseErrorResponse(c echo.Context, details string) error {
	return StandardErrorResponse(c, http.StatusInternalServerError,
		model.ErrDatabaseError, "데이터베이스 오류가 발생했습니다", details)
}

// ========== 일반적인 HTTP 상태 코드 응답 (표준 함수명) ==========

// BadRequestResponse는 잘못된 요청 응답을 생성합니다.
func BadRequestResponse(c echo.Context, message string) error {
	return ValidationFailedResponse(c, message)
}

// UnauthorizedResponse는 인증 실패 응답을 생성합니다.
func UnauthorizedResponse(c echo.Context, message string) error {
	return InvalidTokenResponse(c)
}

// NotFoundResponse는 리소스 없음 응답을 생성합니다.
func NotFoundResponse(c echo.Context, message string) error {
	return StandardErrorResponse(c, http.StatusNotFound,
		model.ErrValidationFailed, message)
}

// ForbiddenResponse는 권한 거부 응답을 생성합니다.
func ForbiddenResponse(c echo.Context, message string) error {
	return PermissionDeniedResponse(c)
}

// ConflictResponse는 충돌 응답을 생성합니다.
func ConflictResponse(c echo.Context, message string) error {
	return StandardErrorResponse(c, http.StatusConflict,
		model.ErrValidationFailed, message)
}
