package service

import (
	"log"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"ssh-key-manager/internal/util"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ServerService 서버 관리 서비스
type ServerService struct {
	serverRepo *repository.ServerRepository
	keyRepo    *repository.SSHKeyRepository
	deployRepo *repository.DeploymentRepository
}

// NewServerService 서버 서비스 생성자
func NewServerService(serverRepo *repository.ServerRepository,
	keyRepo *repository.SSHKeyRepository,
	deployRepo *repository.DeploymentRepository) *ServerService {
	return &ServerService{serverRepo: serverRepo, keyRepo: keyRepo, deployRepo: deployRepo}
}

// CreateServer 새로운 서버를 등록합니다.
func (ss *ServerService) CreateServer(userID uint, req dto.ServerCreateRequest) (*model.Server, error) {
	log.Printf("🖥️ 새 서버 등록 시도: %s (%s)", req.Name, req.Host)

	// 입력값 검증
	if err := ss.validateServerCreateRequest(req); err != nil {
		return nil, err
	}

	// 중복 확인 (동일 사용자가 같은 호스트+포트 조합으로 등록했는지)
	duplicate, err := ss.serverRepo.ExistsByUserAndHost(userID, req.Host, req.Port)
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"서버 중복 확인 중 오류가 발생했습니다",
		)
	}
	if duplicate {
		return nil, model.NewBusinessError(
			model.ErrServerExists,
			"이미 등록된 서버입니다",
		)
	}

	targetDir := strings.TrimSpace(req.TargetDirectory)
	if targetDir == "" {
		targetDir = "~/.ssh" // 기본값 설정
	}

	server := &model.Server{
		UserID:      userID,
		Name:        strings.TrimSpace(req.Name),
		Host:        strings.TrimSpace(req.Host),
		Port:        req.Port,
		Username:    strings.TrimSpace(req.Username),
		Description: strings.TrimSpace(req.Description),
		Status:      "active",
	}

	if err := ss.serverRepo.Create(server); err != nil {
		log.Printf("❌ 서버 등록 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"서버 등록 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 서버 등록 완료: %s (ID: %d)", req.Name, server.ID)
	return server, nil
}

// GetUserServers 사용자의 모든 서버 목록을 반환합니다.
func (ss *ServerService) GetUserServers(userID uint) ([]model.Server, error) {
	log.Printf("🖥️ 사용자 서버 목록 조회 중 (사용자 ID: %d)", userID)

	servers, err := ss.serverRepo.FindByUserID(userID)
	if err != nil {
		log.Printf("❌ 서버 목록 조회 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"서버 목록 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 서버 목록 조회 완료 (총 %d개)", len(servers))
	return servers, nil
}

// GetServerByID 특정 서버 정보를 조회합니다.
func (ss *ServerService) GetServerByID(userID, serverID uint) (*model.Server, error) {
	log.Printf("🔍 서버 상세 정보 조회 중 (서버 ID: %d)", serverID)

	server, err := ss.serverRepo.FindByID(serverID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrServerNotFound,
				"서버를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"서버 조회 중 오류가 발생했습니다",
		)
	}

	// 서버 소유권 확인
	if server.UserID != userID {
		return nil, model.NewBusinessError(
			model.ErrPermissionDenied,
			"해당 서버에 접근할 권한이 없습니다",
		)
	}

	log.Printf("✅ 서버 상세 정보 조회 완료: %s", server.Name)
	return server, nil
}

// UpdateServer 서버 정보를 업데이트합니다.
func (ss *ServerService) UpdateServer(userID, serverID uint, req dto.ServerUpdateRequest) (*model.Server, error) {
	log.Printf("✏️ 서버 정보 업데이트 중 (서버 ID: %d)", serverID)

	// 서버 존재 및 소유권 확인
	server, err := ss.GetServerByID(userID, serverID)
	if err != nil {
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
			return nil, model.NewBusinessError(
				model.ErrInvalidInput,
				"상태는 'active' 또는 'inactive'만 가능합니다",
			)
		}
		updates["status"] = req.Status
	}
	if req.TargetDirectory != "" && req.TargetDirectory != server.TargetDirectory {
		updates["target_directory"] = strings.TrimSpace(req.TargetDirectory)
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := ss.serverRepo.Update(serverID, updates); err != nil {
			log.Printf("❌ 서버 업데이트 실패: %v", err)
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"서버 업데이트 중 오류가 발생했습니다",
			)
		}

		// 업데이트된 서버 정보 다시 조회
		server, err = ss.serverRepo.FindByID(serverID)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"업데이트된 서버 정보 조회 실패",
			)
		}
	}

	log.Printf("✅ 서버 정보 업데이트 완료: %s", server.Name)
	return server, nil
}

// DeleteServer 서버를 삭제합니다.
func (ss *ServerService) DeleteServer(userID, serverID uint) error {
	log.Printf("🗑️ 서버 삭제 중 (서버 ID: %d)", serverID)

	// 서버 존재 및 소유권 확인
	server, err := ss.GetServerByID(userID, serverID)
	if err != nil {
		return err
	}

	// 트랜잭션으로 관련 데이터 함께 삭제
	err = ss.serverRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		// 관련된 배포 기록 삭제
		if err := ss.deployRepo.DeleteByServerID(serverID); err != nil {
			return err
		}

		// 서버 삭제
		return ss.serverRepo.Delete(serverID)
	})

	if err != nil {
		log.Printf("❌ 서버 삭제 실패: %v", err)
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"서버 삭제 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 서버 삭제 완료: %s", server.Name)
	return nil
}

// TestServerConnection 서버 연결을 테스트합니다.
func (ss *ServerService) TestServerConnection(userID, serverID uint) (map[string]interface{}, error) {
	log.Printf("🔍 서버 연결 테스트 중 (서버 ID: %d)", serverID)

	// 서버 정보 조회
	server, err := ss.GetServerByID(userID, serverID)
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

// DeployKeyToServers SSH 키를 선택된 서버들에 배포합니다.
func (ss *ServerService) DeployKeyToServers(userID uint, req dto.KeyDeploymentRequest) ([]dto.DeploymentResult, error) {
	log.Printf("🚀 SSH 키 배포 시작 (사용자 ID: %d, 서버 수: %d)", userID, len(req.ServerIDs))

	// 사용자의 SSH 키 조회
	sshKey, err := ss.keyRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrSSHKeyNotFound,
				"SSH 키를 찾을 수 없습니다. 먼저 키를 생성해주세요",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 조회 중 오류가 발생했습니다",
		)
	}

	// 선택된 서버들 조회 및 소유권 확인
	var servers []model.Server
	for _, serverID := range req.ServerIDs {
		server, err := ss.GetServerByID(userID, serverID)
		if err != nil {
			continue // 접근 권한이 없는 서버는 건너뜀
		}
		servers = append(servers, *server)
	}

	if len(servers) == 0 {
		return nil, model.NewBusinessError(
			model.ErrServerNotFound,
			"배포할 수 있는 서버를 찾을 수 없습니다",
		)
	}

	var results []dto.DeploymentResult

	// 각 서버에 키 배포
	for _, server := range servers {
		log.Printf("📡 서버에 키 배포 중: %s (%s:%d)", server.Name, server.Host, server.Port)

		result := dto.DeploymentResult{
			ServerID:   server.ID,
			ServerName: server.Name,
		}

		// 배포 기록 생성
		deployment := &model.ServerKeyDeployment{
			ServerID: server.ID,
			SSHKeyID: sshKey.ID,
			UserID:   userID,
			Status:   "pending",
		}
		ss.deployRepo.Create(deployment)

		// 실제 키 배포 실행
		err := util.DeploySSHKeyToServer(sshKey.PublicKey, server.Host, server.Port, server.Username, server.TargetDirectory)

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

		// 배포 기록 업데이트 (간단히 Update 호출로 대체)
		// deployment 업데이트 로직은 Repository에 Update 메서드 추가 필요
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

// GetDeploymentHistory 배포 기록을 반환합니다.
func (ss *ServerService) GetDeploymentHistory(userID uint) ([]map[string]interface{}, error) {
	log.Printf("📋 배포 기록 조회 중 (사용자 ID: %d)", userID)

	deployments, err := ss.deployRepo.FindByUserID(userID)
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"배포 기록 조회 중 오류가 발생했습니다",
		)
	}

	var history []map[string]interface{}
	for _, deployment := range deployments {
		record := map[string]interface{}{
			"id":         deployment.ID,
			"server_id":  deployment.ServerID,
			"ssh_key_id": deployment.SSHKeyID,
			"status":     deployment.Status,
			"created_at": deployment.CreatedAt,
		}

		if deployment.Server.ID != 0 {
			record["server"] = map[string]interface{}{
				"name": deployment.Server.Name,
				"host": deployment.Server.Host,
				"port": deployment.Server.Port,
			}
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

// ========== 내부 헬퍼 함수들 ==========

// validateServerCreateRequest 서버 생성 요청을 검증합니다.
func (ss *ServerService) validateServerCreateRequest(req dto.ServerCreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"서버 이름을 입력해주세요",
		)
	}
	if strings.TrimSpace(req.Host) == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"서버 호스트를 입력해주세요",
		)
	}
	if strings.TrimSpace(req.Username) == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"SSH 사용자명을 입력해주세요",
		)
	}
	if req.Port <= 0 {
		req.Port = 22 // 기본 SSH 포트
	}
	return nil
}
