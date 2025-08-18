package service

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"ssh-key-manager/internal/util"

	"gorm.io/gorm"
)

// KeyService SSH 키 관리 서비스
type KeyService struct {
	keyRepo    *repository.SSHKeyRepository
	deployRepo *repository.DeploymentRepository
}

// NewKeyService 키 서비스 생성자 (직접 의존성 주입)
func NewKeyService(keyRepo *repository.SSHKeyRepository, deployRepo *repository.DeploymentRepository) *KeyService {
	return &KeyService{
		keyRepo:    keyRepo,
		deployRepo: deployRepo,
	}
}

// GenerateSSHKeyPair SSH 키 쌍 생성
func (ks *KeyService) GenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔐 SSH 키 쌍 생성 시작 (사용자 ID: %d)", userID)

	// 1. 사용자 정보 조회
	user, err := C().User.userRepo.FindByID(userID)
	if err != nil {
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

	// 2. 키 생성
	keyPair, err := util.GenerateSSHKeyPair(4096, user.Username)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrSSHKeyGeneration,
			"SSH 키 생성에 실패했습니다",
			err.Error(),
		)
	}

	// 3. 키 모델 생성 (핵심 필드만)
	sshKey := &model.SSHKey{
		UserID:     userID,
		PrivateKey: string(keyPair.PrivateKeyPEM),
		PublicKey:  string(keyPair.PublicKeySSH),
		PPK:        string(keyPair.PPKKey),
	}

	// 4. 데이터베이스에 저장 (기존 키 교체)
	if err := ks.keyRepo.ReplaceUserKey(userID, sshKey); err != nil {
		log.Printf("❌ SSH 키 저장 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 저장에 실패했습니다",
			err.Error(),
		)
	}

	log.Printf("✅ SSH 키 생성 완료 (사용자 ID: %d)", userID)
	return sshKey, nil
}

// GetUserSSHKey 사용자의 SSH 키 조회
func (ks *KeyService) GetUserSSHKey(userID uint) (*model.SSHKey, error) {
	if userID == 0 {
		return nil, model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 사용자 ID입니다",
		)
	}

	key, err := ks.keyRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewSSHKeyNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 조회 중 오류가 발생했습니다",
		)
	}

	return key, nil
}

// DeleteUserSSHKey 사용자의 SSH 키 삭제
func (ks *KeyService) DeleteUserSSHKey(userID uint) error {
	log.Printf("🗑️ SSH 키 삭제 시작 (사용자 ID: %d)", userID)

	// 트랜잭션으로 키와 관련 데이터 삭제
	//err = ks.repos.TxManager.WithTransaction(func(tx *gorm.DB) error {
	err := ks.keyRepo.GetDB().Transaction(func(tx *gorm.DB) error {

		// 1. SSH 키 조회
		key, err := ks.keyRepo.FindByUserID(userID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// 2. 배포 기록 삭제 (키가 존재하는 경우)
		if err == nil && key != nil {
			if err := ks.deployRepo.DeleteBySSHKeyID(key.ID); err != nil {
				return err
			}
		}

		// 3. SSH 키 삭제
		return ks.keyRepo.DeleteByUserID(userID)
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

	log.Printf("✅ SSH 키 삭제 완료 (사용자 ID: %d)", userID)
	return nil
}

// HasUserSSHKey 사용자가 SSH 키를 가지고 있는지 확인
func (ks *KeyService) HasUserSSHKey(userID uint) bool {
	if userID == 0 {
		return false
	}

	exists, err := ks.keyRepo.ExistsByUserID(userID)
	if err != nil {
		log.Printf("❌ SSH 키 존재 확인 실패 (사용자 ID: %d): %v", userID, err)
		return false
	}

	return exists
}

// RegenerateSSHKeyPair 기존 SSH 키 재생성
func (ks *KeyService) RegenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔄 SSH 키 재생성 시작 (사용자 ID: %d)", userID)

	// 기존 키가 있어도 ReplaceUserKey에서 처리하므로 별도 삭제 불필요
	return ks.GenerateSSHKeyPair(userID)
}

// GetSSHKeyStatistics SSH 키 통계 조회
func (ks *KeyService) GetSSHKeyStatistics() (map[string]interface{}, error) {
	stats, err := ks.keyRepo.GetStatistics()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"키 통계 조회 중 오류가 발생했습니다",
		)
	}
	return stats, nil
}
