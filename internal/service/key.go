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

// GenerateSSHKeyPair SSH 키 쌍 생성 (일반 사용자용)
func (ks *KeyService) GenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔐 SSH 키 쌍 생성 시작 (사용자 ID: %d)", userID)

	// 사용자 정보 조회 (existsByID 대신 FindByID 사용)
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

	return ks.generateKeyForUser(user)
}

// GenerateSSHKeyPairByAdmin 관리자가 다른 사용자의 SSH 키 쌍 생성
func (ks *KeyService) GenerateSSHKeyPairByAdmin(adminUserID, targetUserID uint) (*model.SSHKey, error) {
	log.Printf("👑 관리자 SSH 키 쌍 생성 시작 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인 (FindByID 사용)
	admin, err := C().User.userRepo.FindByID(adminUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"관리자 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"관리자 조회 중 오류가 발생했습니다",
		)
	}

	if admin.Role != model.RoleAdmin {
		return nil, model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

	// 대상 사용자 정보 조회 (FindByID 사용)
	targetUser, err := C().User.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"대상 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 관리자 권한 확인 완료: %s가 %s의 키 생성", admin.Username, targetUser.Username)
	return ks.generateKeyForUser(targetUser)
}

// GetUserSSHKey 사용자의 SSH 키 조회 (일반 사용자용)
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

// GetUserSSHKeyByAdmin 관리자가 다른 사용자의 SSH 키 조회
func (ks *KeyService) GetUserSSHKeyByAdmin(adminUserID, targetUserID uint) (*model.SSHKey, error) {
	log.Printf("👑 관리자 SSH 키 조회 시작 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인 (FindByID 사용)
	admin, err := C().User.userRepo.FindByID(adminUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"관리자 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"관리자 조회 중 오류가 발생했습니다",
		)
	}

	if admin.Role != model.RoleAdmin {
		return nil, model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

	// 대상 사용자 존재 확인 (FindByID 사용)
	targetUser, err := C().User.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"대상 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// SSH 키 조회
	key, err := ks.keyRepo.FindByUserID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrSSHKeyNotFound,
				"해당 사용자의 SSH 키를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 관리자 SSH 키 조회 완료: %s가 %s의 키 조회", admin.Username, targetUser.Username)
	return key, nil
}

// DeleteUserSSHKey 사용자의 SSH 키 삭제 (일반 사용자용)
func (ks *KeyService) DeleteUserSSHKey(userID uint) error {
	log.Printf("🗑️ SSH 키 삭제 시작 (사용자 ID: %d)", userID)

	return ks.deleteKeyForUser(userID)
}

// DeleteUserSSHKeyByAdmin 관리자가 다른 사용자의 SSH 키 삭제
func (ks *KeyService) DeleteUserSSHKeyByAdmin(adminUserID, targetUserID uint) error {
	log.Printf("👑 관리자 SSH 키 삭제 시작 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인 (FindByID 사용)
	admin, err := C().User.userRepo.FindByID(adminUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewBusinessError(
				model.ErrUserNotFound,
				"관리자 사용자를 찾을 수 없습니다",
			)
		}
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"관리자 조회 중 오류가 발생했습니다",
		)
	}

	if admin.Role != model.RoleAdmin {
		return model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

	// 대상 사용자 존재 확인 (FindByID 사용)
	targetUser, err := C().User.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewBusinessError(
				model.ErrUserNotFound,
				"대상 사용자를 찾을 수 없습니다",
			)
		}
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 관리자 권한 확인 완료: %s가 %s의 키 삭제", admin.Username, targetUser.Username)
	return ks.deleteKeyForUser(targetUserID)
}

// HasUserSSHKey 사용자가 SSH 키를 가지고 있는지 확인
func (ks *KeyService) HasUserSSHKey(userID uint) bool {
	if userID == 0 {
		return false
	}

	// existsByUserID 대신 FindByUserID를 사용하여 한 번의 호출로 처리
	_, err := ks.keyRepo.FindByUserID(userID)
	return err == nil
}

// RegenerateSSHKeyPair 기존 SSH 키 재생성 (일반 사용자용)
func (ks *KeyService) RegenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔄 SSH 키 재생성 시작 (사용자 ID: %d)", userID)

	// 사용자 정보 조회 (FindByID 사용)
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

	return ks.generateKeyForUser(user)
}

// RegenerateSSHKeyPairByAdmin 관리자가 다른 사용자의 SSH 키 재생성
func (ks *KeyService) RegenerateSSHKeyPairByAdmin(adminUserID, targetUserID uint) (*model.SSHKey, error) {
	log.Printf("👑 관리자 SSH 키 재생성 시작 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인 (FindByID 사용)
	admin, err := C().User.userRepo.FindByID(adminUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"관리자 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"관리자 조회 중 오류가 발생했습니다",
		)
	}

	if admin.Role != model.RoleAdmin {
		return nil, model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

	// 대상 사용자 정보 조회 (FindByID 사용)
	targetUser, err := C().User.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewBusinessError(
				model.ErrUserNotFound,
				"대상 사용자를 찾을 수 없습니다",
			)
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 관리자 권한 확인 완료: %s가 %s의 키 재생성", admin.Username, targetUser.Username)
	return ks.generateKeyForUser(targetUser)
}

// ========== 내부 헬퍼 함수들 ==========

// generateKeyForUser 특정 사용자의 SSH 키를 생성하는 공통 로직
func (ks *KeyService) generateKeyForUser(user *model.User) (*model.SSHKey, error) {
	// 키 생성
	keyPair, err := util.GenerateSSHKeyPair(4096, user.Username)
	if err != nil {
		log.Printf("❌ SSH 키 생성 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrSSHKeyGeneration,
			"SSH 키 생성에 실패했습니다",
			err.Error(),
		)
	}

	// 키 모델 생성
	sshKey := &model.SSHKey{
		UserID:     user.ID,
		PrivateKey: string(keyPair.PrivateKeyPEM),
		PublicKey:  string(keyPair.PublicKeySSH),
		PPK:        string(keyPair.PPKKey),
	}

	// 데이터베이스에 저장 (기존 키 교체)
	if err := ks.keyRepo.ReplaceUserKey(user.ID, sshKey); err != nil {
		log.Printf("❌ SSH 키 저장 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"SSH 키 저장에 실패했습니다",
			err.Error(),
		)
	}

	log.Printf("✅ SSH 키 생성 완료 (사용자: %s, ID: %d)", user.Username, user.ID)
	return sshKey, nil
}

// deleteKeyForUser 특정 사용자의 SSH 키를 삭제하는 공통 로직
func (ks *KeyService) deleteKeyForUser(userID uint) error {
	// 트랜잭션으로 키와 관련 데이터 삭제
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
