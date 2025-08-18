package service

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"

	"gorm.io/gorm"
)

// GenerateSSHKeyPair는 사용자의 SSH 키 쌍을 생성합니다.
func GenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔐 SSH 키 쌍 생성 시작 (사용자 ID: %d)", userID)

	// 1. 비즈니스 규칙 검증
	if err := validateKeyGeneration(userID); err != nil {
		return nil, err
	}

	// 2. 사용자 정보 조회
	user, err := getUserForKeyGeneration(userID)
	if err != nil {
		return nil, err
	}

	// 3. 키 생성 (crypto 유틸리티 사용)
	keyPair, err := util.GenerateSSHKeyPair(4096, user.Username)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrSSHKeyGeneration,
			"SSH 키 생성에 실패했습니다",
			err.Error(),
		)
	}

	// 4. 데이터베이스에 저장
	sshKey, err := saveSSHKeyPair(userID, keyPair)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ SSH 키 생성 완료 (사용자 ID: %d)", userID)
	return sshKey, nil
}

// GetUserSSHKey는 사용자의 SSH 키를 조회합니다.
func GetUserSSHKey(userID uint) (*model.SSHKey, error) {
	// 1. 비즈니스 규칙 검증
	if userID == 0 {
		return nil, model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 사용자 ID입니다",
		)
	}

	// 2. 데이터베이스에서 조회
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var sshKey model.SSHKey
	if err := db.Where("user_id = ?", userID).First(&sshKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewSSHKeyNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 조회 중 오류가 발생했습니다",
		)
	}

	return &sshKey, nil
}

// DeleteUserSSHKey는 사용자의 SSH 키를 삭제합니다.
func DeleteUserSSHKey(userID uint) error {
	log.Printf("🗑️ SSH 키 삭제 시작 (사용자 ID: %d)", userID)

	// 1. 비즈니스 규칙 검증
	if err := validateKeyDeletion(userID); err != nil {
		return err
	}

	// 2. 관련 데이터 정리 및 삭제
	if err := deleteSSHKeyWithCleanup(userID); err != nil {
		return err
	}

	log.Printf("✅ SSH 키 삭제 완료 (사용자 ID: %d)", userID)
	return nil
}

// HasUserSSHKey는 사용자가 SSH 키를 가지고 있는지 확인합니다.
func HasUserSSHKey(userID uint) bool {
	if userID == 0 {
		return false
	}

	db, err := model.GetDB()
	if err != nil {
		log.Printf("❌ DB 접근 실패 (HasUserSSHKey): %v", err)
		return false
	}

	var count int64
	if err := db.Model(&model.SSHKey{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		log.Printf("❌ SSH 키 존재 확인 실패 (사용자 ID: %d): %v", userID, err)
		return false
	}

	return count > 0
}

// RegenerateSSHKeyPair는 기존 SSH 키를 새로 생성합니다.
func RegenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔄 SSH 키 재생성 시작 (사용자 ID: %d)", userID)

	// 1. 기존 키 삭제
	if HasUserSSHKey(userID) {
		if err := DeleteUserSSHKey(userID); err != nil {
			return nil, err
		}
	}

	// 2. 새 키 생성
	return GenerateSSHKeyPair(userID)
}

// GetSSHKeyStatistics는 SSH 키 통계를 조회합니다.
func GetSSHKeyStatistics() (map[string]interface{}, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	stats := make(map[string]interface{})

	// 전체 키 수
	var totalKeys int64
	if err := db.Model(&model.SSHKey{}).Count(&totalKeys).Error; err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"키 통계 조회 중 오류가 발생했습니다",
		)
	}
	stats["total_keys"] = totalKeys

	// 알고리즘별 통계
	var algorithmStats []struct {
		Algorithm string
		Count     int64
	}
	if err := db.Model(&model.SSHKey{}).
		Select("algorithm, count(*) as count").
		Group("algorithm").
		Scan(&algorithmStats).Error; err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"알고리즘 통계 조회 중 오류가 발생했습니다",
		)
	}
	stats["by_algorithm"] = algorithmStats

	// 키 크기별 통계
	var bitsStats []struct {
		Bits  int
		Count int64
	}
	if err := db.Model(&model.SSHKey{}).
		Select("bits, count(*) as count").
		Group("bits").
		Scan(&bitsStats).Error; err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"키 크기 통계 조회 중 오류가 발생했습니다",
		)
	}
	stats["by_bits"] = bitsStats

	return stats, nil
}

// ========== 내부 헬퍼 함수들 (비즈니스 로직) ==========

// validateKeyGeneration은 키 생성 전 비즈니스 규칙을 검증합니다.
func validateKeyGeneration(userID uint) error {
	if userID == 0 {
		return model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 사용자 ID입니다",
		)
	}

	// 사용자 존재 확인
	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var count int64
	if err := db.Model(&model.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 확인 중 오류가 발생했습니다",
		)
	}

	if count == 0 {
		return model.NewBusinessError(
			model.ErrUserNotFound,
			"사용자를 찾을 수 없습니다",
		)
	}

	return nil
}

// getUserForKeyGeneration은 키 생성을 위한 사용자 정보를 조회합니다.
func getUserForKeyGeneration(userID uint) (*model.User, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var user model.User
	if err := db.Select("id, username").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	return &user, nil
}

// saveSSHKeyPair는 생성된 키 쌍을 데이터베이스에 저장합니다.
func saveSSHKeyPair(userID uint, keyPair *util.SSHKeyPair) (*model.SSHKey, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	sshKey := &model.SSHKey{
		UserID:     userID,
		Algorithm:  keyPair.Algorithm,
		Bits:       keyPair.Bits,
		PrivateKey: string(keyPair.PrivateKeyPEM),
		PublicKey:  string(keyPair.PublicKeySSH),
		PEM:        string(keyPair.PrivateKeyPEM),
		PPK:        string(keyPair.PPKKey),
	}

	// 트랜잭션으로 안전하게 저장 (기존 키가 있으면 교체)
	err = db.Transaction(func(tx *gorm.DB) error {
		// 기존 키 삭제
		if err := tx.Where("user_id = ?", userID).Delete(&model.SSHKey{}).Error; err != nil {
			return err
		}

		// 새 키 생성
		if err := tx.Create(sshKey).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ SSH 키 저장 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 저장에 실패했습니다",
			err.Error(),
		)
	}

	return sshKey, nil
}

// validateKeyDeletion은 키 삭제 전 비즈니스 규칙을 검증합니다.
func validateKeyDeletion(userID uint) error {
	if userID == 0 {
		return model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 사용자 ID입니다",
		)
	}

	// 사용자 존재 확인
	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var count int64
	if err := db.Model(&model.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 확인 중 오류가 발생했습니다",
		)
	}

	if count == 0 {
		return model.NewBusinessError(
			model.ErrUserNotFound,
			"사용자를 찾을 수 없습니다",
		)
	}

	return nil
}

// deleteSSHKeyWithCleanup은 키 삭제와 관련 데이터 정리를 수행합니다.
func deleteSSHKeyWithCleanup(userID uint) error {
	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	// 트랜잭션으로 안전하게 삭제
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. 배포 기록에서 해당 키와 관련된 레코드 삭제
		var sshKey model.SSHKey
		if err := tx.Where("user_id = ?", userID).First(&sshKey).Error; err == nil {
			// 배포 기록 삭제
			if err := tx.Where("ssh_key_id = ?", sshKey.ID).Delete(&model.ServerKeyDeployment{}).Error; err != nil {
				return err
			}
		}

		// 2. SSH 키 삭제
		if err := tx.Where("user_id = ?", userID).Delete(&model.SSHKey{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewSSHKeyNotFoundError()
		}
		log.Printf("❌ SSH 키 삭제 실패 (사용자 ID: %d): %v", userID, err)
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 삭제 중 오류가 발생했습니다",
			err.Error(),
		)
	}

	return nil
}
