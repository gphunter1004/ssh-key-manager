package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"ssh-key-manager/internal/model"

	"github.com/labstack/echo/v4"
)

// CustomHTTPErrorHandler는 글로벌 에러 핸들러입니다.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	var (
		code   = http.StatusInternalServerError
		apiErr = &model.APIError{
			Code:    model.ErrInternalServer,
			Message: "서버 내부 오류가 발생했습니다",
		}
	)

	// 응답 헤더가 이미 전송되었는지 확인
	if c.Response().Committed {
		return
	}

	// BusinessError 처리 (우선순위 높음)
	if be, ok := err.(*model.BusinessError); ok {
		code = mapBusinessErrorToHTTPStatus(be.Code)
		apiErr.Code = be.Code
		apiErr.Message = be.Message
		apiErr.Details = be.Details
		
		log.Printf("비즈니스 에러: %s (코드: %s)", be.Message, string(be.Code))
	} else if he, ok := err.(*echo.HTTPError); ok {
		// Echo HTTPError 처리
		code = he.Code
		message := fmt.Sprintf("%v", he.Message)
		
		// HTTP 상태별 에러 코드 매핑
		switch code {
		case http.StatusNotFound:
			apiErr.Code = model.ErrValidationFailed
			apiErr.Message = "요청하신 리소스를 찾을 수 없습니다"
		case http.StatusMethodNotAllowed:
			apiErr.Code = model.ErrValidationFailed
			apiErr.Message = "허용되지 않은 HTTP 메서드입니다"
		case http.StatusBadRequest:
			apiErr.Code = model.ErrValidationFailed
			apiErr.Message = "잘못된 요청입니다"
		case http.StatusUnauthorized:
			apiErr.Code = model.ErrInvalidToken
			apiErr.Message = "인증이 필요합니다"
		case http.StatusForbidden:
			apiErr.Code = model.ErrPermissionDenied
			apiErr.Message = "권한이 없습니다"
		case http.StatusInternalServerError:
			apiErr.Code = model.ErrInternalServer
			apiErr.Message = "서버 내부 오류가 발생했습니다"
		default:
			apiErr.Code = model.ErrInternalServer
			apiErr.Message = message
		}
		
		log.Printf("HTTP 에러: %d %s", code, message)
	} else {
		// 일반 에러 처리
		apiErr.Code = model.ErrInternalServer
		apiErr.Message = "서버 내부 오류가 발생했습니다"
		apiErr.Details = err.Error()
		
		log.Printf("일반 에러: %v", err)
	}

	// 서버 에러는 상세 로깅
	if code >= 500 {
		log.Printf("🚨 서버 에러 발생: %v", err)
		// 운영환경에서는 상세 에러 메시지 숨김
		if !isDebugMode() {
			apiErr.Details = ""
		}
	}

	// 표준 에러 응답
	if err := c.JSON(code, model.APIResponse{
		Success: false,
		Error:   apiErr,
	}); err != nil {
		log.Printf("에러 응답 전송 실패: %v", err)
	}
}

// mapBusinessErrorToHTTPStatus는 비즈니스 에러를 HTTP 상태 코드로 매핑합니다.
func mapBusinessErrorToHTTPStatus(code model.ErrorCode) int {
	switch code {
	// 인증 에러 (401)
	case model.ErrInvalidCredentials, model.ErrTokenExpired, model.ErrInvalidToken, model.ErrInvalidJWT:
		return http.StatusUnauthorized

	// 권한 에러 (403)
	case model.ErrPermissionDenied:
		return http.StatusForbidden

	// 리소스 없음 (404)
	case model.ErrUserNotFound, model.ErrServerNotFound, model.ErrSSHKeyNotFound, model.ErrDepartmentNotFound:
		return http.StatusNotFound

	// 중복/충돌 (409)
	case model.ErrUserAlreadyExists, model.ErrServerExists, model.ErrDepartmentExists, model.ErrSSHKeyExists:
		return http.StatusConflict

	// 검증 실패 (400)
	case model.ErrValidationFailed, model.ErrInvalidInput, model.ErrWeakPassword, model.ErrRequiredField,
		 model.ErrInvalidFormat, model.ErrInvalidRange, model.ErrInvalidUsername, model.ErrInvalidServerID,
		 model.ErrInvalidDeptID, model.ErrCannotDeleteSelf, model.ErrLastAdmin, model.ErrDepartmentHasUsers,
		 model.ErrDepartmentHasChild, model.ErrInvalidParentDept, model.ErrServerNotOwned, model.ErrInvalidSSHKey:
		return http.StatusBadRequest

	// 서버 연결 실패 (502)
	case model.ErrConnectionFailed:
		return http.StatusBadGateway

	// SSH 키 생성/배포 실패 (500)
	case model.ErrSSHKeyGeneration, model.ErrSSHKeyDeployment:
		return http.StatusInternalServerError

	// 시스템 에러 (500)
	case model.ErrInternalServer, model.ErrDatabaseError, model.ErrConfigError, model.ErrFileSystemError:
		return http.StatusInternalServerError

	// 기본값 (400)
	default:
		return http.StatusBadRequest
	}
}

// isDebugMode는 디버그 모드 여부를 확인합니다.
func isDebugMode() bool {
	return os.Getenv("DEBUG") == "true" || os.Getenv("ENV") == "development"
}

// RecoverMiddleware는 panic을 복구하고 에러로 변환합니다.
func RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("panic: %v", r)
					}
					
					log.Printf("🚨 Panic 발생: %v", err)
					
					// panic을 에러로 변환하여 글로벌 에러 핸들러가 처리하도록 함
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}
