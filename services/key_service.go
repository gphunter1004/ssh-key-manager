package services

import (
	"errors"
	"ssh-key-manager/config"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"

	"gorm.io/gorm"
)

// GenerateSSHKeyPair는 SSH 키 쌍을 생성하고 DB에 저장합니다.
func GenerateSSHKeyPair(userID uint) (*models.SSHKey, error) {
	cfg, err := config.LoadConfig() // 키 비트 수 등 설정 로드
	if err != nil {
		return nil, err
	}

	// 1. 개인키 생성
	privateKey, err := utils.GeneratePrivateKey(cfg.KeyBits)
	if err != nil {
		return nil, err
	}

	// 2. 키 형식 변환
	pemKey := utils.EncodePrivateKeyToPEM(privateKey)
	publicKey, err := utils.GeneratePublicKey(privateKey)
	if err != nil {
		return nil, err
	}
	ppkKey, err := utils.EncodePrivateKeyToPPK(privateKey)
	if err != nil {
		return nil, err
	}

	// 3. DB에 저장 또는 업데이트
	sshKey := &models.SSHKey{
		UserID:     userID,
		PrivateKey: string(pemKey), // 편의상 PEM을 PrivateKey 필드에 저장
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}

	// 기존 키가 있으면 업데이트, 없으면 새로 생성 (Upsert)
	err = models.DB.Where(models.SSHKey{UserID: userID}).Assign(models.SSHKey{
		PrivateKey: string(pemKey),
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}).FirstOrCreate(sshKey).Error

	if err != nil {
		return nil, err
	}

	return sshKey, nil
}

// GetKeyByUserID는 사용자의 키를 조회합니다.
func GetKeyByUserID(userID uint) (*models.SSHKey, error) {
	var key models.SSHKey
	result := models.DB.Where("user_id = ?", userID).First(&key)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("no key found for this user, please create one first")
		}
		return nil, result.Error
	}
	return &key, nil
}

// DeleteKeyByUserID는 사용자의 키를 삭제합니다.
func DeleteKeyByUserID(userID uint) error {
	result := models.DB.Where("user_id = ?", userID).Delete(&models.SSHKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no key found to delete")
	}
	return nil
}
