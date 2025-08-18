package handler

import (
	"log"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateServer는 새로운 서버를 등록합니다.
func CreateServer(c echo.Context) error {
	userID, _ := GetUserID(c)

	var req dto.ServerCreateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	server, err := service.C().Server.CreateServer(userID, req)
	if err != nil {
		LogError("서버 등록", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrServerExists:
				return ConflictResponse(c, "이미 등록된 서버입니다")
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			case model.ErrRequiredField:
				return BadRequestResponse(c, be.Message)
			default:
				return InternalServerErrorResponse(c, "서버 등록 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "서버 등록 중 오류가 발생했습니다")
	}

	LogSuccess("서버 등록", userID, server.Name)
	return CreatedResponse(c, "서버가 성공적으로 등록되었습니다", server)
}

// GetServers는 사용자의 서버 목록을 반환합니다.
func GetServers(c echo.Context) error {
	userID, _ := GetUserID(c)

	servers, err := service.C().Server.GetUserServers(userID)
	if err != nil {
		log.Printf("❌ 서버 목록 조회 실패 (사용자 ID: %d): %v", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrUserNotFound:
				return NotFoundResponse(c, "사용자를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "서버 목록 조회 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "서버 목록 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, servers)
}

// GetServer는 특정 서버 정보를 반환합니다.
func GetServer(c echo.Context) error {
	userID, _ := GetUserID(c)

	serverID, err := ParseServerIDParam(c)
	if err != nil {
		return BadRequestResponse(c, err.Error())
	}

	server, err := service.C().Server.GetServerByID(userID, serverID)
	if err != nil {
		LogError("서버 조회", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrServerNotFound:
				return NotFoundResponse(c, "서버를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "해당 서버에 접근할 권한이 없습니다")
			default:
				return InternalServerErrorResponse(c, "서버 조회 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "서버 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, server)
}

// UpdateServer는 서버 정보를 업데이트합니다.
func UpdateServer(c echo.Context) error {
	userID, _ := GetUserID(c)

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	var req dto.ServerUpdateRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	server, err := service.C().Server.UpdateServer(userID, uint(serverID), req)
	if err != nil {
		log.Printf("❌ 서버 수정 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrServerNotFound:
				return NotFoundResponse(c, "서버를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "해당 서버에 접근할 권한이 없습니다")
			case model.ErrInvalidInput:
				return BadRequestResponse(c, be.Message)
			default:
				return InternalServerErrorResponse(c, "서버 수정 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "서버 수정 중 오류가 발생했습니다")
	}

	return SuccessWithMessageResponse(c, "서버 정보가 업데이트되었습니다", server)
}

// DeleteServer는 서버를 삭제합니다.
func DeleteServer(c echo.Context) error {
	userID, _ := GetUserID(c)

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	err = service.C().Server.DeleteServer(userID, uint(serverID))
	if err != nil {
		log.Printf("❌ 서버 삭제 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrServerNotFound:
				return NotFoundResponse(c, "서버를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "해당 서버에 접근할 권한이 없습니다")
			default:
				return InternalServerErrorResponse(c, "서버 삭제 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "서버 삭제 중 오류가 발생했습니다")
	}

	log.Printf("✅ 서버 삭제 성공 (사용자 ID: %d, 서버 ID: %d)", userID, serverID)
	return SuccessWithMessageResponse(c, "서버가 삭제되었습니다", nil)
}

// TestServerConnection은 서버 연결을 테스트합니다.
func TestServerConnection(c echo.Context) error {
	userID, _ := GetUserID(c)

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	result, err := service.C().Server.TestServerConnection(userID, uint(serverID))
	if err != nil {
		log.Printf("❌ 서버 연결 테스트 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrServerNotFound:
				return NotFoundResponse(c, "서버를 찾을 수 없습니다")
			case model.ErrPermissionDenied:
				return ForbiddenResponse(c, "해당 서버에 접근할 권한이 없습니다")
			case model.ErrConnectionFailed:
				return BadRequestResponse(c, "서버 연결에 실패했습니다")
			default:
				return InternalServerErrorResponse(c, "연결 테스트 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "연결 테스트 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, result)
}

// DeployKeyToServers는 SSH 키를 서버에 배포합니다.
func DeployKeyToServers(c echo.Context) error {
	userID, _ := GetUserID(c)

	var req dto.KeyDeploymentRequest
	if err := ValidateJSONRequest(c, &req); err != nil {
		return err
	}

	if len(req.ServerIDs) == 0 {
		return BadRequestResponse(c, "배포할 서버를 선택해주세요")
	}

	results, err := service.C().Server.DeployKeyToServers(userID, req)
	if err != nil {
		log.Printf("❌ 키 배포 실패 (사용자 ID: %d): %v", userID, err)
		if be, ok := err.(*model.BusinessError); ok {
			switch be.Code {
			case model.ErrSSHKeyNotFound:
				return BadRequestResponse(c, "SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
			case model.ErrServerNotFound:
				return NotFoundResponse(c, "배포할 수 있는 서버를 찾을 수 없습니다")
			default:
				return InternalServerErrorResponse(c, "키 배포 중 오류가 발생했습니다")
			}
		}
		return InternalServerErrorResponse(c, "키 배포 중 오류가 발생했습니다")
	}

	// 성공/실패 카운트
	successCount := 0
	failedCount := 0
	for _, result := range results {
		if result.Status == "success" {
			successCount++
		} else {
			failedCount++
		}
	}

	summary := map[string]interface{}{
		"total":   len(results),
		"success": successCount,
		"failed":  failedCount,
	}

	responseData := map[string]interface{}{
		"results": results,
		"summary": summary,
	}

	log.Printf("✅ 키 배포 완료 (사용자 ID: %d): 성공 %d/%d", userID, successCount, len(results))
	return SuccessWithMessageResponse(c, "키 배포가 완료되었습니다", responseData)
}

// GetDeploymentHistory는 배포 기록을 반환합니다.
func GetDeploymentHistory(c echo.Context) error {
	userID, _ := GetUserID(c)

	history, err := service.C().Server.GetDeploymentHistory(userID)
	if err != nil {
		log.Printf("❌ 배포 기록 조회 실패 (사용자 ID: %d): %v", userID, err)
		return InternalServerErrorResponse(c, "배포 기록 조회 중 오류가 발생했습니다")
	}

	return SuccessResponse(c, history)
}
