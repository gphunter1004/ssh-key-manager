package services

import (
	"errors"
	"log"
	"ssh-key-manager/models"
	"ssh-key-manager/types"
	"ssh-key-manager/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateServer는 새로운 서버를 등록합니다.
func CreateServer(userID uint, req types.ServerCreateRequest) (*types.ServerResponse, error) {
	log.Printf("🖥️ 새 서버 등록 시도: %s (%s)", req.Name, req.Host)

	// 입력값 검증
	if strings.TrimSpace(req.Name) == "" {
		return nil, errors.New("서버 이름을 입력해주세요")
	}
	if strings.TrimSpace(req.Host) == "" {
		return nil, errors.New("서버 호스트를 입력해주세요")
	}
	if strings.TrimSpace(req.Username) == "" {
		return nil, errors.New("SSH 사용자명을 입력해주세요")
	}
	if req.Port <= 0 {
		req.Port = 22 // 기본 SSH 포트
	}

	// 중복 확인 (동일 사용자가 같은 호스트+포트 조합으로 등록했는지)
	var existingServer models.Server
	err := models.DB.Where("user_id = ? AND host = ? AND port = ?", userID, req.Host, req.Port).First(&existingServer).Error
	if err == nil {
		return nil, errors.New("이미 등록된 서버입니다")
	}

	server := models.Server{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Host:        strings.TrimSpace(req.Host),
		Port:        req.Port,
		Username:    strings.TrimSpace(req.Username),
		Description: strings.TrimSpace(req.Description),
		Status:      "active",
	}

	result := models.DB.Create(&server)
	if result.Error != nil {
		log.Printf("❌ 서버 등록 실패: %v", result.Error)
		return nil, errors.New("서버 등록 중 오류가 발생했습니다")
	}

	log.Printf("✅ 서버 등록 완료: %s (ID: %d)", req.Name, server.ID)

	// types.ToServerResponse 사용
	serverResponse := types.ToServerResponse(server)
	return &serverResponse, nil
}

// GetUserServers는 사용자의 모든 서버 목록을 반환합니다.
func GetUserServers(userID uint) ([]types.ServerResponse, error) {
	log.Printf("🖥️ 사용자 서버 목록 조회 중 (사용자 ID: %d)", userID)

	var servers []models.Server
	result := models.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&servers)
	if result.Error != nil {
		log.Printf("❌ 서버 목록 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	// 응답 데이터 구성 - types.ToServerResponse 사용
	var serverResponses []types.ServerResponse
	for _, server := range servers {
		serverResponse := types.ToServerResponse(server)
		serverResponses = append(serverResponses, serverResponse)
	}

	log.Printf("✅ 서버 목록 조회 완료 (총 %d개)", len(serverResponses))
	return serverResponses, nil
}

// GetServerByID는 특정 서버 정보를 조회합니다.
func GetServerByID(userID, serverID uint) (*types.ServerResponse, error) {
	log.Printf("🔍 서버 상세 정보 조회 중 (서버 ID: %d)", serverID)

	var server models.Server
	result := models.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("서버를 찾을 수 없습니다")
		}
		log.Printf("❌ 서버 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	log.Printf("✅ 서버 상세 정보 조회 완료: %s", server.Name)

	// types.ToServerResponse 사용
	serverResponse := types.ToServerResponse(server)
	return &serverResponse, nil
}

// UpdateServer는 서버 정보를 업데이트합니다.
func UpdateServer(userID, serverID uint, req types.ServerUpdateRequest) (*types.ServerResponse, error) {
	log.Printf("✏️ 서버 정보 업데이트 중 (서버 ID: %d)", serverID)

	var server models.Server
	result := models.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("서버를 찾을 수 없습니다")
		}
		return nil, result.Error
	}

	// 업데이트할 필드 확인 및 검증
	updates := make(map[string]interface{})

	if req.Name != "" && req.Name != server.Name {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Host != "" && req.Host != server.Host {
		updates["host"] = strings.TrimSpace(req.Host)
	}
	if req.Port > 0 && req.Port != server.Port {
		updates["port"] = req.Port
	}
	if req.Username != "" && req.Username != server.Username {
		updates["username"] = strings.TrimSpace(req.Username)
	}
	if req.Description != server.Description {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if req.Status != "" && req.Status != server.Status {
		if req.Status != "active" && req.Status != "inactive" {
			return nil, errors.New("상태는 'active' 또는 'inactive'만 가능합니다")
		}
		updates["status"] = req.Status
	}

	// 업데이트할 내용이 있는 경우에만 실행
	if len(updates) > 0 {
		if err := models.DB.Model(&server).Updates(updates).Error; err != nil {
			log.Printf("❌ 서버 업데이트 실패: %v", err)
			return nil, errors.New("서버 업데이트 중 오류가 발생했습니다")
		}

		// 업데이트된 서버 정보 다시 조회
		models.DB.First(&server, serverID)
	}

	log.Printf("✅ 서버 정보 업데이트 완료: %s", server.Name)

	// types.ToServerResponse 사용
	serverResponse := types.ToServerResponse(server)
	return &serverResponse, nil
}

// DeleteServer는 서버를 삭제합니다.
func DeleteServer(userID, serverID uint) error {
	log.Printf("🗑️ 서버 삭제 중 (서버 ID: %d)", serverID)

	// 서버 존재 여부 확인
	var server models.Server
	result := models.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("서버를 찾을 수 없습니다")
		}
		return result.Error
	}

	// 관련된 배포 기록도 함께 삭제 (CASCADE)
	if err := models.DB.Where("server_id = ?", serverID).Delete(&models.ServerKeyDeployment{}).Error; err != nil {
		log.Printf("⚠️ 배포 기록 삭제 실패: %v", err)
	}

	// 서버 삭제
	if err := models.DB.Delete(&server).Error; err != nil {
		log.Printf("❌ 서버 삭제 실패: %v", err)
		return errors.New("서버 삭제 중 오류가 발생했습니다")
	}

	log.Printf("✅ 서버 삭제 완료: %s", server.Name)
	return nil
}

// DeployKeyToServers는 SSH 키를 선택된 서버들에 배포합니다.
func DeployKeyToServers(userID uint, req types.KeyDeploymentRequest) ([]types.DeploymentResult, error) {
	log.Printf("🚀 SSH 키 배포 시작 (사용자 ID: %d, 서버 수: %d)", userID, len(req.ServerIDs))

	// 사용자의 SSH 키 조회
	sshKey, err := GetKeyByUserID(userID)
	if err != nil {
		return nil, errors.New("SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
	}

	// 선택된 서버들 조회
	var servers []models.Server
	result := models.DB.Where("id IN ? AND user_id = ?", req.ServerIDs, userID).Find(&servers)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(servers) == 0 {
		return nil, errors.New("선택된 서버를 찾을 수 없습니다")
	}

	var deploymentResults []types.DeploymentResult

	// 각 서버에 키 배포
	for _, server := range servers {
		log.Printf("📡 서버에 키 배포 중: %s (%s:%d)", server.Name, server.Host, server.Port)

		result := types.DeploymentResult{
			ServerID:   server.ID,
			ServerName: server.Name,
		}

		// 배포 기록 생성
		deployment := models.ServerKeyDeployment{
			ServerID: server.ID,
			SSHKeyID: sshKey.ID,
			UserID:   userID,
			Status:   "pending",
		}
		models.DB.Create(&deployment)

		// 실제 키 배포 실행
		err := utils.DeploySSHKeyToRemoteServer(
			sshKey.PublicKey,
			server.Host,
			server.Port,
			server.Username,
		)

		if err != nil {
			// 배포 실패
			result.Status = "failed"
			result.ErrorMessage = err.Error()
			deployment.Status = "failed"
			deployment.ErrorMsg = err.Error()
			log.Printf("❌ 키 배포 실패 [%s]: %v", server.Name, err)
		} else {
			// 배포 성공
			result.Status = "success"
			deployment.Status = "success"
			now := gorm.DeletedAt{Time: time.Now(), Valid: true}
			deployment.DeployedAt = &now
			log.Printf("✅ 키 배포 성공: %s", server.Name)
		}

		// 배포 기록 업데이트
		models.DB.Save(&deployment)
		deploymentResults = append(deploymentResults, result)
	}

	successCount := 0
	for _, result := range deploymentResults {
		if result.Status == "success" {
			successCount++
		}
	}

	log.Printf("🎯 키 배포 완료: 성공 %d/%d", successCount, len(deploymentResults))
	return deploymentResults, nil
}

// GetDeploymentHistory는 키 배포 기록을 조회합니다.
func GetDeploymentHistory(userID uint) ([]map[string]interface{}, error) {
	log.Printf("📋 배포 기록 조회 중 (사용자 ID: %d)", userID)

	var deployments []models.ServerKeyDeployment
	result := models.DB.Where("user_id = ?", userID).
		Preload("Server").
		Preload("SSHKey").
		Order("created_at DESC").
		Find(&deployments)

	if result.Error != nil {
		log.Printf("❌ 배포 기록 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	var history []map[string]interface{}
	for _, deployment := range deployments {
		record := map[string]interface{}{
			"id":         deployment.ID,
			"server_id":  deployment.ServerID,
			"ssh_key_id": deployment.SSHKeyID,
			"status":     deployment.Status,
			"created_at": deployment.CreatedAt,
			"server": map[string]interface{}{
				"name": deployment.Server.Name,
				"host": deployment.Server.Host,
				"port": deployment.Server.Port,
			},
		}

		if deployment.DeployedAt != nil && deployment.DeployedAt.Valid {
			record["deployed_at"] = deployment.DeployedAt.Time
		}

		if deployment.ErrorMsg != "" {
			record["error_message"] = deployment.ErrorMsg
		}

		history = append(history, record)
	}

	log.Printf("✅ 배포 기록 조회 완료 (총 %d건)", len(history))
	return history, nil
}
