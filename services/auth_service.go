package services

import (
	"errors"
	"log"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"
	"strings"
)

// RegisterUser는 새로운 사용자를 등록합니다.
func RegisterUser(username, password string) error {
	log.Printf("👤 새 사용자 등록 시도: %s", username)

	// 입력값 검증
	if strings.TrimSpace(username) == "" {
		return errors.New("사용자명을 입력해주세요")
	}
	if len(password) < 4 {
		return errors.New("비밀번호는 최소 4자 이상이어야 합니다")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return errors.New("비밀번호 처리 중 오류가 발생했습니다")
	}

	user := models.User{
		Username: strings.TrimSpace(username),
		Password: hashedPassword,
	}

	result := models.DB.Create(&user)
	if result.Error != nil {
		// 중복 사용자명 체크
		if strings.Contains(result.Error.Error(), "duplicate") ||
			strings.Contains(result.Error.Error(), "unique") ||
			strings.Contains(result.Error.Error(), "uni_users_username") {
			log.Printf("⚠️ 중복 사용자명 시도: %s", username)
			return errors.New("이미 사용 중인 사용자명입니다")
		}
		log.Printf("❌ 사용자 등록 실패: %v", result.Error)
		return errors.New("사용자 등록 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 등록 완료: %s (ID: %d)", username, user.ID)
	return nil
}

// AuthenticateUser는 사용자를 인증하고 JWT를 반환합니다.
func AuthenticateUser(username, password string) (string, error) {
	log.Printf("🔐 사용자 인증 시도: %s", username)

	// 입력값 검증
	if strings.TrimSpace(username) == "" {
		return "", errors.New("사용자명을 입력해주세요")
	}
	if strings.TrimSpace(password) == "" {
		return "", errors.New("비밀번호를 입력해주세요")
	}

	var user models.User
	result := models.DB.Where("username = ?", strings.TrimSpace(username)).First(&user)
	if result.Error != nil {
		log.Printf("⚠️ 존재하지 않는 사용자: %s", username)
		return "", errors.New("사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		log.Printf("⚠️ 잘못된 비밀번호 시도: %s", username)
		return "", errors.New("사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("❌ JWT 생성 실패: %v", err)
		return "", errors.New("인증 토큰 생성 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 인증 완료: %s (ID: %d)", username, user.ID)
	return token, nil
}
