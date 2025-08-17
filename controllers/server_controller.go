package controllers

import (
	"fmt"
	"ssh-key-manager/helpers"
	"ssh-key-manager/services"
	"ssh-key-manager/types"
	"ssh-key-manager/utils"

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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var req types.ServerCreateRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}

	var server interface{}
	err = utils.LogOperation("서버 등록", func() error {
		utils.LogServiceCall("ServerService", "CreateServer", userID, req.Name, req.Host)
		var createErr error
		server, createErr = services.CreateServer(userID, req)
		return createErr
	})

	if err != nil {
		utils.LogUserAction(userID, "등록", "서버", false, err.Error())
		return utils.HandleServiceError(c, err, "서버 등록")
	}

	utils.LogUserAction(userID, "등록", "서버", true, req.Name)
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	// 페이징 및 검색 파라미터 처리
	page, limit := utils.ExtractPaginationParams(c)
	sortBy, sortOrder := utils.ExtractSortParams(c, "created_at")
	searchQuery, _ := utils.ExtractSearchParams(c)

	utils.LogServiceCall("ServerService", "GetUserServers", userID, page, limit)
	servers, err := services.GetUserServers(userID)
	if err != nil {
		utils.LogUserAction(userID, "조회", "서버 목록", false, err.Error())
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "조회", "서버 목록", true, fmt.Sprintf("총 %d개", len(servers)))

	// TODO: 실제로는 서비스에서 페이징과 검색을 처리해야 함
	metadata := map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"sort_by":    sortBy,
		"sort_order": sortOrder,
		"search":     searchQuery,
	}

	return helpers.APIResponseWithMetadata(c, servers, metadata)
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	utils.LogServiceCall("ServerService", "GetServerByID", userID, serverID)
	server, err := services.GetServerByID(userID, serverID)
	if err != nil {
		utils.LogUserAction(userID, "조회", "서버", false, err.Error())
		return helpers.NotFoundResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "조회", "서버", true, fmt.Sprintf("ID: %d", serverID))
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	var req types.ServerUpdateRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}

	var server interface{}
	err = utils.LogOperation("서버 수정", func() error {
		utils.LogServiceCall("ServerService", "UpdateServer", userID, serverID)
		var updateErr error
		server, updateErr = services.UpdateServer(userID, serverID, req)
		return updateErr
	})

	if err != nil {
		utils.LogUserAction(userID, "수정", "서버", false, err.Error())
		return utils.HandleServiceError(c, err, "서버 수정")
	}

	utils.LogUserAction(userID, "수정", "서버", true, fmt.Sprintf("ID: %d", serverID))
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	err = utils.LogOperation("서버 삭제", func() error {
		utils.LogServiceCall("ServerService", "DeleteServer", userID, serverID)
		return services.DeleteServer(userID, serverID)
	})

	if err != nil {
		utils.LogUserAction(userID, "삭제", "서버", false, err.Error())
		return helpers.NotFoundResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "삭제", "서버", true, fmt.Sprintf("ID: %d", serverID))
	utils.LogSecurityEvent("서버 삭제", userID, fmt.Sprintf("서버 ID: %d", serverID), "low")
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	var req types.KeyDeploymentRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}

	if len(req.ServerIDs) == 0 {
		return helpers.BadRequestResponse(c, "배포할 서버를 선택해주세요")
	}

	var results interface{}
	err = utils.LogOperation("SSH 키 배포", func() error {
		utils.LogServiceCall("ServerService", "DeployKeyToServers", userID, len(req.ServerIDs))
		var deployErr error
		results, deployErr = services.DeployKeyToServers(userID, req)
		return deployErr
	})

	if err != nil {
		utils.LogUserAction(userID, "배포", "SSH 키", false, err.Error())
		return utils.HandleServiceError(c, err, "키 배포")
	}

	// 성공/실패 카운트
	resultsSlice := results.([]types.DeploymentResult)
	successCount := 0
	failedCount := 0
	for _, result := range resultsSlice {
		if result.Status == "success" {
			successCount++
		} else {
			failedCount++
		}
	}

	summary := types.DeploymentSummary{
		Total:   len(resultsSlice),
		Success: successCount,
		Failed:  failedCount,
	}

	responseData := map[string]interface{}{
		"results": results,
		"summary": summary,
	}

	utils.LogUserAction(userID, "배포", "SSH 키", true,
		fmt.Sprintf("총 %d개 서버 (성공: %d, 실패: %d)", len(req.ServerIDs), successCount, failedCount))

	if failedCount > 0 {
		utils.LogSecurityEvent("키 배포 부분 실패", userID,
			fmt.Sprintf("성공: %d, 실패: %d", successCount, failedCount), "medium")
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	utils.LogServiceCall("ServerService", "GetDeploymentHistory", userID)
	history, err := services.GetDeploymentHistory(userID)
	if err != nil {
		utils.LogUserAction(userID, "조회", "배포 이력", false, err.Error())
		return helpers.InternalServerErrorResponse(c, err.Error())
	}

	utils.LogUserAction(userID, "조회", "배포 이력", true, fmt.Sprintf("총 %d건", len(history)))
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
	userID, err := utils.UserIDFromToken(c)
	if err != nil {
		return helpers.UnauthorizedResponse(c, "Invalid token")
	}

	serverID, err := utils.ParseUintParam(c, "id")
	if err != nil {
		return helpers.BadRequestResponse(c, err.Error())
	}

	// 서버 정보 조회
	utils.LogServiceCall("ServerService", "GetServerByID", userID, serverID)
	server, err := services.GetServerByID(userID, serverID)
	if err != nil {
		utils.LogUserAction(userID, "테스트", "서버 연결", false, err.Error())
		return helpers.NotFoundResponse(c, err.Error())
	}

	// 연결 테스트 실행
	var result types.ConnectionTestResult
	err = utils.LogOperation("서버 연결 테스트", func() error {
		utils.LogServiceCall("ServerService", "TestRemoteServerConnection", userID, server.Host)
		testErr := utils.TestRemoteServerConnection(server.Host, server.Port, server.Username)
		if testErr != nil {
			result = types.ConnectionTestResult{
				Success: false,
				Message: "연결 테스트 실패",
				Error:   testErr.Error(),
			}
			return testErr
		}

		result = types.ConnectionTestResult{
			Success: true,
			Message: "연결 테스트 성공",
		}
		return nil
	})

	if err != nil {
		utils.LogUserAction(userID, "테스트", "서버 연결", false, fmt.Sprintf("%s:%d", server.Host, server.Port))
	} else {
		utils.LogUserAction(userID, "테스트", "서버 연결", true, fmt.Sprintf("%s:%d", server.Host, server.Port))
	}

	return helpers.SuccessResponse(c, result)
}
