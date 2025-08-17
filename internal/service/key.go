package service

import (
	"errors"
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"

	"gorm.io/gorm"
)

// GenerateSSHKeyPair는 사용자의 SSH 키 쌍을 생성합니다.
func GenerateSSHKeyPair(userID uint) (*model.SSHKey, error) {
	log.Printf("🔐 SSH 키 쌍 생성 시작 (사용자 ID: %d)", userID)

	// 사용자 확인
	var user model.User
	if err := model.DB.Select("id, username").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, err
	}

	// RSA 키 쌍 생성 (4096 비트)
	privateKey, err := util.GeneratePrivateKey(4096)
	if err != nil {
		log.Printf("❌ RSA 키 생성 실패: %v", err)
		return nil, errors.New("SSH 키 생성에 실패했습니다")
	}

	// PEM 형식 개인키
	pemKey := util.EncodePrivateKeyToPEM(privateKey)

	// SSH 공개키 (authorized_keys 형식)
	publicKey, err := util.GeneratePublicKeyWithComment(privateKey, user.Username)
	if err != nil {
		log.Printf("❌ 공개키 생성 실패: %v", err)
		return nil, errors.New("SSH 공개키 생성에 실패했습니다")
	}

	// PPK 형식 개인키 (PuTTY용)
	ppkKey := util.GenerateSimplePPK(privateKey, user.Username)

	// 데이터베이스에 저장 (기존 키가 있으면 업데이트)
	sshKey := &model.SSHKey{
		UserID:     userID,
		Algorithm:  "RSA",
		Bits:       4096,
		PrivateKey: string(pemKey),
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}

	// Upsert (있으면 업데이트, 없으면 생성)
	err = model.DB.Where(model.SSHKey{UserID: userID}).Assign(model.SSHKey{
		Algorithm:  "RSA",
		Bits:       4096,
		PrivateKey: string(pemKey),
		PublicKey:  string(publicKey),
		PEM:        string(pemKey),
		PPK:        string(ppkKey),
	}).FirstOrCreate(sshKey).Error

	if err != nil {
		log.Printf("❌ SSH 키 저장 실패: %v", err)
		return nil, errors.New("SSH 키 저장에 실패했습니다")
	}

	log.Printf("✅ SSH 키 생성 완료 (사용자 ID: %d)", userID)
	return sshKey, nil
}

// GetUserSSHKey는 사용자의 SSH 키를 조회합니다.
func GetUserSSHKey(userID uint) (*model.SSHKey, error) {
	var sshKey model.SSHKey
	if err := model.DB.Where("user_id = ?", userID).First(&sshKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("SSH 키를 찾을 수 없습니다")
		}
		return nil, err
	}

	return &sshKey, nil
}

// DeleteUserSSHKey는 사용자의 SSH 키를 삭제합니다.
func DeleteUserSSHKey(userID uint) error {
	result := model.DB.Where("user_id = ?", userID).Delete(&model.SSHKey{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("삭제할 SSH 키를 찾을 수 없습니다")
	}

	log.Printf("✅ SSH 키 삭제 완료 (사용자 ID: %d)", userID)
	return nil
}

// HasUserSSHKey는 사용자가 SSH 키를 가지고 있는지 확인합니다.
func HasUserSSHKey(userID uint) bool {
	var count int64
	model.DB.Model(&model.SSHKey{}).Where("user_id = ?", userID).Count(&count)
	return count > 0
}
