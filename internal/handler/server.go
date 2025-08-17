package handler

import (
	"log"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/service"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateServer는 새로운 서버를 등록합니다.
func CreateServer(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	var req model.ServerCreateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	server, err := service.CreateServer(userID, req)
	if err != nil {
		log.Printf("❌ 서버 등록 실패 (사용자 ID: %d): %v", userID, err)
		return BadRequestResponse(c, err.Error())
	}

	log.Printf("✅ 서버 등록 성공 (사용자 ID: %d): %s", userID, server.Name)
	return CreatedResponse(c, "서버가 성공적으로 등록되었습니다", server)
}

// GetServers는 사용자의 서버 목록을 반환합니다.
func GetServers(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	servers, err := service.GetUserServers(userID)
	if err != nil {
		log.Printf("❌ 서버 목록 조회 실패 (사용자 ID: %d): %v", userID, err)
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, servers)
}

// GetServer는 특정 서버 정보를 반환합니다.
func GetServer(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	server, err := service.GetServerByID(userID, uint(serverID))
	if err != nil {
		log.Printf("❌ 서버 조회 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		return NotFoundResponse(c, err.Error())
	}

	return SuccessResponse(c, server)
}

// UpdateServer는 서버 정보를 업데이트합니다.
func UpdateServer(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	var req model.ServerUpdateRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	server, err := service.UpdateServer(userID, uint(serverID), req)
	if err != nil {
		log.Printf("❌ 서버 수정 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		return BadRequestResponse(c, err.Error())
	}

	return SuccessWithMessageResponse(c, "서버 정보가 업데이트되었습니다", server)
}

// DeleteServer는 서버를 삭제합니다.
func DeleteServer(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	err = service.DeleteServer(userID, uint(serverID))
	if err != nil {
		log.Printf("❌ 서버 삭제 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		return NotFoundResponse(c, err.Error())
	}

	log.Printf("✅ 서버 삭제 성공 (사용자 ID: %d, 서버 ID: %d)", userID, serverID)
	return SuccessWithMessageResponse(c, "서버가 삭제되었습니다", nil)
}

// TestServerConnection은 서버 연결을 테스트합니다.
func TestServerConnection(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	result, err := service.TestServerConnection(userID, uint(serverID))
	if err != nil {
		log.Printf("❌ 서버 연결 테스트 실패 (사용자 ID: %d, 서버 ID: %d): %v", userID, serverID, err)
		return BadRequestResponse(c, err.Error())
	}

	return SuccessResponse(c, result)
}

// DeployKeyToServers는 SSH 키를 서버에 배포합니다.
func DeployKeyToServers(c echo.Context) error {
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	var req model.KeyDeploymentRequest
	if err := c.Bind(&req); err != nil {
		return BadRequestResponse(c, "잘못된 요청 형식입니다")
	}

	if len(req.ServerIDs) == 0 {
		return BadRequestResponse(c, "배포할 서버를 선택해주세요")
	}

	results, err := service.DeployKeyToServers(userID, req)
	if err != nil {
		log.Printf("❌ 키 배포 실패 (사용자 ID: %d): %v", userID, err)
		return BadRequestResponse(c, err.Error())
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
	userID, err := middleware.UserIDFromToken(c)
	if err != nil {
		return UnauthorizedResponse(c, "Invalid token")
	}

	history, err := service.GetDeploymentHistory(userID)
	if err != nil {
		log.Printf("❌ 배포 기록 조회 실패 (사용자 ID: %d): %v", userID, err)
		return InternalServerErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, history)
}
