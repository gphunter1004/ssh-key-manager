package services

import (
	"errors"
	"fmt"
	"log"
	"ssh-key-manager/config"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"

	"gorm.io/gorm"
)

// GenerateSSHKeyPair는 로그인된 사용자의 SSH 키 쌍을 생성하고 DB에 저장합니다.
func GenerateSSHKeyPair(userID uint) (*models.SSHKey, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// 현재 사용자 정보 조회 (코멘트용)
	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("사용자를 찾을 수 없습니다")
	}

	// 기존 키가 있으면 서버에서 제거
	if cfg.AutoInstallKeys {
		if existingKey, err := GetKeyByUserID(userID); err == nil {
			log.Printf("🔄 기존 키를 서버에서 제거 중...")
			if removeErr := utils.RemovePublicKeyFromServer(existingKey.PublicKey, cfg.SSHUser, cfg.SSHHomePath); removeErr != nil {
				log.Printf("⚠️ 기존 키 제거 실패 (계속 진행): %v", removeErr)
			}
		}
	}

	log.Printf("🚀 SSH 키 쌍 생성 시작 (사용자: %s)", user.Username)

	// 1. RSA 키 쌍 생성 (개인키 생성하면 공개키도 함께 생성됨)
	privateKey, err := utils.GeneratePrivateKey(cfg.KeyBits)
	if err != nil {
		return nil, err
	}
	log.Printf("   ✅ RSA 키 쌍 생성 완료 (개인키 + 공개키)")

	// 2. 개인키를 PEM 형식으로 인코딩 (OpenSSH, Linux/macOS 용)
	pemKey := utils.EncodePrivateKeyToPEM(privateKey)
	log.Printf("   📄 PEM 형식 개인키 생성 완료")

	// 3. 개인키에서 공개키 추출하여 SSH 형식으로 변환 (authorized_keys 용)
	publicKey, err := utils.GeneratePublicKeyWithUserComment(privateKey, user.Username)
	if err != nil {
		return nil, err
	}
	log.Printf("   🔑 SSH 공개키 생성 완료 (코멘트: %s)", user.Username)

	// 4. 개인키를 PPK 형식으로 변환 (PuTTY 용)
	ppkKey, err := utils.EncodePrivateKeyToPPKWithUser(privateKey, user.Username)
	if err != nil {
		log.Printf("⚠️ PPK 생성 실패, 기본 방법으로 재시도: %v", err)
		// PPK 생성 실패 시 기본 방법으로 재시도
		ppkKey, err = utils.EncodePrivateKeyToPPK(privateKey)
		if err != nil {
			return nil, errors.New("PPK 키 생성에 실패했습니다. 시스템에 puttygen이 설치되어 있는지 확인하세요")
		}
	}
	log.Printf("   🔧 PPK 형식 개인키 생성 완료")

	// 5. 로컬 서버에 공개키 자동 설치 (설정이 활성화된 경우)
	var installationStatus string
	if cfg.AutoInstallKeys {
		log.Printf("🔧 로컬 서버 자동 설치 시작...")

		// SSH 설정 검증
		if err := utils.ValidateSSHConfig(cfg.SSHUser, cfg.SSHHomePath); err != nil {
			log.Printf("⚠️ SSH 설정 검증 실패: %v", err)
			installationStatus = fmt.Sprintf("자동 설치 실패: %v", err)
		} else {
			// 공개키 설치
			if err := utils.InstallPublicKeyToServer(string(publicKey), cfg.SSHUser, cfg.SSHHomePath); err != nil {
				log.Printf("⚠️ 공개키 자동 설치 실패: %v", err)
				installationStatus = fmt.Sprintf("자동 설치 실패: %v", err)
			} else {
				log.Printf("✅ 로컬 서버에 공개키 자동 설치 완료")
				installationStatus = "로컬 서버에 자동 설치 완료"
			}
		}
	} else {
		log.Printf("📋 자동 설치 비활성화됨 (AUTO_INSTALL_KEYS=false)")
		installationStatus = "자동 설치 비활성화됨"
	}

	// 6. DB에 저장 또는 업데이트 (Upsert)
	sshKey := &models.SSHKey{
		UserID:     userID,
		Algorithm:  "RSA",
		Bits:       cfg.KeyBits,
		PrivateKey: string(pemKey),
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}

	// 기존 키가 있으면 업데이트, 없으면 새로 생성
	err = models.DB.Where(models.SSHKey{UserID: userID}).Assign(models.SSHKey{
		Algorithm:  "RSA",
		Bits:       cfg.KeyBits,
		PrivateKey: string(pemKey),
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}).FirstOrCreate(sshKey).Error

	if err != nil {
		log.Printf("❌ 키 저장 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ SSH 키 쌍 생성 및 저장 완료")
	log.Printf("   - 사용자: %s", user.Username)
	log.Printf("   - 알고리즘: RSA")
	log.Printf("   - 키 크기: %d bits", cfg.KeyBits)
	log.Printf("   - 개인키 형식: PEM (OpenSSH용), PPK (PuTTY용)")
	log.Printf("   - 공개키 코멘트: %s", user.Username)
	log.Printf("   - 자동 설치 상태: %s", installationStatus)
	log.Printf("📋 생성된 키 쌍:")
	log.Printf("   🔒 개인키: 클라이언트에서 사용 (절대 공유하지 마세요!)")
	log.Printf("   🔓 공개키: 서버의 ~/.ssh/authorized_keys에 추가")

	return sshKey, nil
}

// GetKeyByUserID는 사용자의 키를 조회합니다.
func GetKeyByUserID(userID uint) (*models.SSHKey, error) {
	var key models.SSHKey
	result := models.DB.Where("user_id = ?", userID).First(&key)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("키를 찾을 수 없습니다. 먼저 키를 생성해주세요")
		}
		return nil, result.Error
	}

	log.Printf("🔍 키 조회 완료 (사용자 ID: %d)", userID)
	return &key, nil
}

// DeleteKeyByUserID는 사용자의 키를 삭제합니다.
func DeleteKeyByUserID(userID uint) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("⚠️ 설정 로드 실패, 자동 제거 건너뜀: %v", err)
	}

	// 삭제 전에 서버에서 공개키 제거 (설정이 활성화된 경우)
	if cfg != nil && cfg.AutoInstallKeys {
		if existingKey, err := GetKeyByUserID(userID); err == nil {
			log.Printf("🗑️ 서버에서 SSH 공개키 자동 제거 중...")

			if err := utils.ValidateSSHConfig(cfg.SSHUser, cfg.SSHHomePath); err != nil {
				log.Printf("⚠️ SSH 설정 검증 실패: %v", err)
			} else {
				if err := utils.RemovePublicKeyFromServer(existingKey.PublicKey, cfg.SSHUser, cfg.SSHHomePath); err != nil {
					log.Printf("⚠️ 서버에서 공개키 제거 실패: %v", err)
				} else {
					log.Printf("✅ 서버에서 공개키 자동 제거 완료")
				}
			}
		}
	}

	// DB에서 키 삭제
	result := models.DB.Where("user_id = ?", userID).Delete(&models.SSHKey{})
	if result.Error != nil {
		log.Printf("❌ 키 삭제 실패: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("삭제할 키를 찾을 수 없습니다")
	}

	log.Printf("🗑️ 키 삭제 완료 (사용자 ID: %d)", userID)
	return nil
}
