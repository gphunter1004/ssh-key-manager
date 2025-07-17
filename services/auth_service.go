package services

import (
	"errors"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"
	"strings"
)

// RegisterUser는 새로운 사용자를 등록합니다.
func RegisterUser(username, password string) error {
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	user := models.User{
		Username: username,
		Password: hashedPassword,
	}

	result := models.DB.Create(&user)

	// 오류 확인 및 사용자 친화적 메시지 반환
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "uni_users_username") {
			return errors.New("이미 사용 중인 이름입니다.")
		}
		return result.Error
	}

	return nil
}

// AuthenticateUser는 사용자를 인증하고 JWT를 반환합니다.
func AuthenticateUser(username, password string) (string, error) {
	var user models.User
	result := models.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return "", errors.New("user not found")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
