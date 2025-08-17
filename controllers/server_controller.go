package controllers

import (
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/types"
	"ssh-key-manager/utils"
	"strconv"

	"github.com/labstack/echo/v4"
)

// CreateServer godoc
// @Summary Create a new server
// @Description Register a new remote server for SSH key deployment
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   server  body   types.ServerCreateRequest  true  "Server Info"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /servers [post]
func CreateServer(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var req types.ServerCreateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	server, err := services.CreateServer(userID, req)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.CreatedResponse(c, "서버가 성공적으로 등록되었습니다", server)
}

// GetServers godoc
// @Summary Get user's servers
// @Description Get all servers registered by the current user
// @Tags servers
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /servers [get]
func GetServers(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 페이징 및 검색 파라미터 처리
	var pagination types.PaginationRequest
	var search types.SearchRequest
	
	if err := c.Bind(&pagination); err == nil {
		pagination = pagination.GetDefaultPagination()
	}
	c.Bind(&search)

	servers, err := services.GetUserServers(userID)
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	// TODO: 실제로는 서비스에서 페이징과 검색을 처리해야 함
	return helpers.ListResponse(c, servers, len(servers))
}

// GetServer godoc
// @Summary Get server details
// @Description Get detailed information of a specific server
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "Server ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /servers/{id} [get]
func GetServer(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	server, err := services.GetServerByID(userID, uint(serverID))
	if err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	return helpers.SuccessResponse(c, server)
}

// UpdateServer godoc
// @Summary Update server
// @Description Update server information
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   id      path   int                           true  "Server ID"
// @Param   server  body   types.ServerUpdateRequest  true  "Server Update Info"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /servers/{id} [put]
func UpdateServer(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	var req types.ServerUpdateRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	server, err := services.UpdateServer(userID, uint(serverID), req)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "서버 정보가 업데이트되었습니다", server)
}

// DeleteServer godoc
// @Summary Delete server
// @Description Delete a server from the user's server list
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "Server ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /servers/{id} [delete]
func DeleteServer(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	if err := services.DeleteServer(userID, uint(serverID)); err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	return helpers.SuccessWithMessageResponse(c, "서버가 삭제되었습니다", nil)
}

// DeployKeyToServers godoc
// @Summary Deploy SSH key to servers
// @Description Deploy the user's SSH key to selected servers
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   deployment  body   types.KeyDeploymentRequest  true  "Deployment Info"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /servers/deploy [post]
func DeployKeyToServers(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var req types.KeyDeploymentRequest
	if err := c.Bind(&req); err != nil {
		return helpers.BadRequestResponse(c, "Invalid request body")
	}

	if len(req.ServerIDs) == 0 {
		return helpers.BadRequestResponse(c, "배포할 서버를 선택해주세요")
	}

	results, err := services.DeployKeyToServers(userID, req)
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
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

	summary := types.DeploymentSummary{
		Total:   len(results),
		Success: successCount,
		Failed:  failedCount,
	}

	responseData := map[string]interface{}{
		"results": results,
		"summary": summary,
	}

	return helpers.SuccessWithMessageResponse(c, "키 배포가 완료되었습니다", responseData)
}

// GetDeploymentHistory godoc
// @Summary Get deployment history
// @Description Get SSH key deployment history for the current user
// @Tags servers
// @Accept  json
// @Produce  json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /servers/deployments [get]
func GetDeploymentHistory(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	history, err := services.GetDeploymentHistory(userID)
	if err != nil {
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	return helpers.ListResponse(c, history, len(history))
}

// TestServerConnection godoc
// @Summary Test server connection
// @Description Test SSH connection to a specific server
// @Tags servers
// @Accept  json
// @Produce  json
// @Param   id   path      int  true  "Server ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /servers/{id}/test [post]
func TestServerConnection(c echo.Context) error {
	userID, err := userIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverIDParam := c.Param("id")
	serverID, err := strconv.ParseUint(serverIDParam, 10, 32)
	if err != nil {
		return helpers.BadRequestResponse(c, "유효하지 않은 서버 ID입니다")
	}

	// 서버 정보 조회
	server, err := services.GetServerByID(userID, uint(serverID))
	if err != nil {
		return helpers.NotFoundResponse(c, err.Error())
	}

	// 연결 테스트 실행
	if err := utils.TestRemoteServerConnection(server.Host, server.Port, server.Username); err != nil {
		result := types.ConnectionTestResult{
			Success: false,
			Message: "연결 테스트 실패",
			Error:   err.Error(),
		}
		return helpers.SuccessResponse(c, result)
	}

	result := types.ConnectionTestResult{
		Success: true,
		Message: "연결 테스트 성공",
	}
	return helpers.SuccessResponse(c, result)
}