package model

// ErrorCode는 표준 에러 코드를 정의합니다.
type ErrorCode string

const (
	// 인증 관련 에러 (AUTH_xxx)
	ErrInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrTokenExpired       ErrorCode = "AUTH_TOKEN_EXPIRED"
	ErrInvalidToken       ErrorCode = "AUTH_INVALID_TOKEN"
	ErrPermissionDenied   ErrorCode = "AUTH_PERMISSION_DENIED"
	ErrInvalidJWT         ErrorCode = "AUTH_INVALID_JWT"

	// 사용자 관련 에러 (USER_xxx)
	ErrUserNotFound      ErrorCode = "USER_NOT_FOUND"
	ErrUserAlreadyExists ErrorCode = "USER_ALREADY_EXISTS"
	ErrInvalidUsername   ErrorCode = "USER_INVALID_USERNAME"
	ErrWeakPassword      ErrorCode = "USER_WEAK_PASSWORD"
	ErrCannotDeleteSelf  ErrorCode = "USER_CANNOT_DELETE_SELF"
	ErrLastAdmin         ErrorCode = "USER_LAST_ADMIN"

	// 서버 관련 에러 (SERVER_xxx)
	ErrServerNotFound   ErrorCode = "SERVER_NOT_FOUND"
	ErrServerExists     ErrorCode = "SERVER_ALREADY_EXISTS"
	ErrConnectionFailed ErrorCode = "SERVER_CONNECTION_FAILED"
	ErrInvalidServerID  ErrorCode = "SERVER_INVALID_ID"
	ErrServerNotOwned   ErrorCode = "SERVER_NOT_OWNED"

	// SSH 키 관련 에러 (SSH_xxx)
	ErrSSHKeyNotFound   ErrorCode = "SSH_KEY_NOT_FOUND"
	ErrSSHKeyGeneration ErrorCode = "SSH_KEY_GENERATION_FAILED"
	ErrSSHKeyDeployment ErrorCode = "SSH_KEY_DEPLOYMENT_FAILED"
	ErrSSHKeyExists     ErrorCode = "SSH_KEY_ALREADY_EXISTS"
	ErrInvalidSSHKey    ErrorCode = "SSH_INVALID_KEY_FORMAT"

	// 부서 관련 에러 (DEPT_xxx)
	ErrDepartmentNotFound ErrorCode = "DEPT_NOT_FOUND"
	ErrDepartmentExists   ErrorCode = "DEPT_ALREADY_EXISTS"
	ErrDepartmentHasUsers ErrorCode = "DEPT_HAS_USERS"
	ErrDepartmentHasChild ErrorCode = "DEPT_HAS_CHILDREN"
	ErrInvalidDeptID      ErrorCode = "DEPT_INVALID_ID"
	ErrInvalidParentDept  ErrorCode = "DEPT_INVALID_PARENT"

	// 입력 검증 에러 (VALIDATION_xxx)
	ErrValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrInvalidInput     ErrorCode = "VALIDATION_INVALID_INPUT"
	ErrRequiredField    ErrorCode = "VALIDATION_REQUIRED_FIELD"
	ErrInvalidFormat    ErrorCode = "VALIDATION_INVALID_FORMAT"
	ErrInvalidRange     ErrorCode = "VALIDATION_INVALID_RANGE"

	// 시스템 에러 (SYS_xxx)
	ErrInternalServer  ErrorCode = "SYS_INTERNAL_ERROR"
	ErrDatabaseError   ErrorCode = "SYS_DATABASE_ERROR"
	ErrConfigError     ErrorCode = "SYS_CONFIG_ERROR"
	ErrFileSystemError ErrorCode = "SYS_FILESYSTEM_ERROR"
)

// APIError는 표준 API 에러 응답 구조체입니다.
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// BusinessError는 비즈니스 로직 에러를 나타냅니다.
type BusinessError struct {
	Code    ErrorCode
	Message string
	Details string
}

// Error는 error 인터페이스를 구현합니다.
func (e BusinessError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewBusinessError는 새로운 비즈니스 에러를 생성합니다.
func NewBusinessError(code ErrorCode, message string, details ...string) *BusinessError {
	err := &BusinessError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// 자주 사용되는 에러 생성 함수들
func NewUserNotFoundError() *BusinessError {
	return NewBusinessError(ErrUserNotFound, "사용자를 찾을 수 없습니다")
}

func NewInvalidCredentialsError() *BusinessError {
	return NewBusinessError(ErrInvalidCredentials, "사용자명 또는 비밀번호가 올바르지 않습니다")
}

func NewPermissionDeniedError() *BusinessError {
	return NewBusinessError(ErrPermissionDenied, "권한이 없습니다")
}

func NewSSHKeyNotFoundError() *BusinessError {
	return NewBusinessError(ErrSSHKeyNotFound, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
}

func NewServerNotFoundError() *BusinessError {
	return NewBusinessError(ErrServerNotFound, "서버를 찾을 수 없습니다")
}

func NewDepartmentNotFoundError() *BusinessError {
	return NewBusinessError(ErrDepartmentNotFound, "부서를 찾을 수 없습니다")
}
