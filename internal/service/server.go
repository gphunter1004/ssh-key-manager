package service

import (
	"errors"
	"fmt"
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CreateServer는 새로운 서버를 생성합니다.
func CreateServer(userID uint, req model.ServerCreateRequest) (*model.Server, error) {
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
	var existingServer model.Server
	err := model.DB.Where("user_id = ? AND host = ? AND port = ?", userID, req.Host, req.Port).First(&existingServer).Error
	if err == nil {
		return nil, errors.New("이미 등록된 서버입니다")
	}

	server := model.Server{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Host:        strings.TrimSpace(req.Host),
		Port:        req.Port,
		Username:    strings.TrimSpace(req.Username),
		Description: strings.TrimSpace(req.Description),
		Status:      "active",
	}

	if err := model.DB.Create(&server).Error; err != nil {
		log.Printf("❌ 서버 등록 실패: %v", err)
		return nil, errors.New("서버 등록 중 오류가 발생했습니다")
	}

	log.Printf("✅ 서버 등록 완료: %s (ID: %d)", req.Name, server.ID)
	return &server, nil
}

// GetUserServers는 사용자의 모든 서버 목록을 반환합니다.
func GetUserServers(userID uint) ([]model.Server, error) {
	log.Printf("🖥️ 사용자 서버 목록 조회 중 (사용자 ID: %d)", userID)

	var servers []model.Server
	if err := model.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&servers).Error; err != nil {
		log.Printf("❌ 서버 목록 조회 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ 서버 목록 조회 완료 (총 %d개)", len(servers))
	return servers, nil
}

// GetServerByID는 특정 서버 정보를 조회합니다.
func GetServerByID(userID, serverID uint) (*model.Server, error) {
	log.Printf("🔍 서버 상세 정보 조회 중 (서버 ID: %d)", serverID)

	var server model.Server
	if err := model.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("서버를 찾을 수 없습니다")
		}
		return nil, err
	}

	log.Printf("✅ 서버 상세 정보 조회 완료: %s", server.Name)
	return &server, nil
}

// UpdateServer는 서버 정보를 업데이트합니다.
func UpdateServer(userID, serverID uint, req model.ServerUpdateRequest) (*model.Server, error) {
	log.Printf("✏️ 서버 정보 업데이트 중 (서버 ID: %d)", serverID)

	var server model.Server
	if err := model.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("서버를 찾을 수 없습니다")
		}
		return nil, err
	}

	// 업데이트할 필드 확인
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

	// 업데이트 실행
	if len(updates) > 0 {
		if err := model.DB.Model(&server).Updates(updates).Error; err != nil {
			log.Printf("❌ 서버 업데이트 실패: %v", err)
			return nil, errors.New("서버 업데이트 중 오류가 발생했습니다")
		}

		// 업데이트된 서버 정보 다시 조회
		model.DB.First(&server, serverID)
	}

	log.Printf("✅ 서버 정보 업데이트 완료: %s", server.Name)
	return &server, nil
}

// DeleteServer는 서버를 삭제합니다.
func DeleteServer(userID, serverID uint) error {
	log.Printf("🗑️ 서버 삭제 중 (서버 ID: %d)", serverID)

	// 서버 존재 여부 확인
	var server model.Server
	if err := model.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("서버를 찾을 수 없습니다")
		}
		return err
	}

	// 트랜잭션으로 관련 데이터 함께 삭제
	tx := model.DB.Begin()

	// 관련된 배포 기록 삭제
	if err := tx.Where("server_id = ?", serverID).Delete(&model.ServerKeyDeployment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 서버 삭제
	if err := tx.Delete(&server).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	log.Printf("✅ 서버 삭제 완료: %s", server.Name)
	return nil
}

// TestServerConnection은 서버 연결을 테스트합니다.
func TestServerConnection(userID, serverID uint) (map[string]interface{}, error) {
	log.Printf("🔍 서버 연결 테스트 중 (서버 ID: %d)", serverID)

	// 서버 정보 조회
	server, err := GetServerByID(userID, serverID)
	if err != nil {
		return nil, err
	}

	// 연결 테스트 실행
	result := map[string]interface{}{
		"server_id":   server.ID,
		"server_name": server.Name,
		"host":        server.Host,
		"port":        server.Port,
		"username":    server.Username,
	}

	// SSH 연결 테스트
	err = util.TestSSHConnection(server.Host, server.Port, server.Username)
	if err != nil {
		result["success"] = false
		result["message"] = "연결 테스트 실패"
		result["error"] = err.Error()
		log.Printf("❌ 서버 연결 테스트 실패 [%s]: %v", server.Name, err)
	} else {
		result["success"] = true
		result["message"] = "연결 테스트 성공"
		log.Printf("✅ 서버 연결 테스트 성공: %s", server.Name)
	}

	return result, nil
}

// DeployKeyToServers는 SSH 키를 선택된 서버들에 배포합니다.
func DeployKeyToServers(userID uint, req model.KeyDeploymentRequest) ([]model.DeploymentResult, error) {
	log.Printf("🚀 SSH 키 배포 시작 (사용자 ID: %d, 서버 수: %d)", userID, len(req.ServerIDs))

	// 사용자의 SSH 키 조회
	sshKey, err := GetUserSSHKey(userID)
	if err != nil {
		return nil, errors.New("SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
	}

	// 선택된 서버들 조회
	var servers []model.Server
	if err := model.DB.Where("id IN ? AND user_id = ?", req.ServerIDs, userID).Find(&servers).Error; err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, errors.New("선택된 서버를 찾을 수 없습니다")
	}

	var results []model.DeploymentResult

	// 각 서버에 키 배포
	for _, server := range servers {
		log.Printf("📡 서버에 키 배포 중: %s (%s:%d)", server.Name, server.Host, server.Port)

		result := model.DeploymentResult{
			ServerID:   server.ID,
			ServerName: server.Name,
		}

		// 배포 기록 생성
		deployment := model.ServerKeyDeployment{
			ServerID: server.ID,
			SSHKeyID: sshKey.ID,
			UserID:   userID,
			Status:   "pending",
		}
		model.DB.Create(&deployment)

		// 실제 키 배포 실행
		err := util.DeploySSHKeyToServer(sshKey.PublicKey, server.Host, server.Port, server.Username)

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
			now := time.Now()
			deployment.DeployedAt = &gorm.DeletedAt{Time: now, Valid: true}
			log.Printf("✅ 키 배포 성공: %s", server.Name)
		}

		// 배포 기록 업데이트
		model.DB.Save(&deployment)
		results = append(results, result)
	}

	successCount := 0
	for _, result := range results {
		if result.Status == "success" {
			successCount++
		}
	}

	log.Printf("🎯 키 배포 완료: 성공 %d/%d", successCount, len(results))
	return results, nil
}

// GetDeploymentHistory는 키 배포 기록을 조회합니다.
func GetDeploymentHistory(userID uint) ([]map[string]interface{}, error) {
	log.Printf("📋 배포 기록 조회 중 (사용자 ID: %d)", userID)

	var deployments []model.ServerKeyDeployment
	if err := model.DB.Where("user_id = ?", userID).
		Preload("Server").
		Preload("SSHKey").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		return nil, err
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
