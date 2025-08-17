package service

import (
	"errors"
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// RegisterUser는 새로운 사용자를 등록합니다.
func RegisterUser(username, password string) (*model.User, error) {
	log.Printf("👤 새 사용자 등록 시도: %s", username)

	// 입력값 검증
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("사용자명을 입력해주세요")
	}
	if len(password) < 4 {
		return nil, errors.New("비밀번호는 최소 4자 이상이어야 합니다")
	}

	// 비밀번호 해시
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return nil, errors.New("비밀번호 처리 중 오류가 발생했습니다")
	}

	// 사용자 생성
	user := model.User{
		Username: username,
		Password: hashedPassword,
		Role:     model.RoleUser,
	}

	if err := model.DB.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique") {
			return nil, errors.New("이미 사용 중인 사용자명입니다")
		}
		log.Printf("❌ 사용자 등록 실패: %v", err)
		return nil, errors.New("사용자 등록 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 등록 완료: %s (ID: %d)", username, user.ID)
	
	// 비밀번호 필드 제거 후 반환
	user.Password = ""
	return &user, nil
}

// AuthenticateUser는 사용자를 인증하고 JWT를 반환합니다.
func AuthenticateUser(username, password string) (string, *model.User, error) {
	log.Printf("🔐 사용자 인증 시도: %s", username)

	// 입력값 검증
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", nil, errors.New("사용자명과 비밀번호를 입력해주세요")
	}

	// 사용자 조회
	var user model.User
	if err := model.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("사용자명 또는 비밀번호가 올바르지 않습니다")
		}
		return "", nil, err
	}

	// 비밀번호 확인
	if !util.CheckPasswordHash(password, user.Password) {
		return "", nil, errors.New("사용자명 또는 비밀번호가 올바르지 않습니다")
	}

	// JWT 토큰 생성
	token, err := util.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("❌ JWT 생성 실패: %v", err)
		return "", nil, errors.New("인증 토큰 생성 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 인증 완료: %s (ID: %d)", username, user.ID)
	
	// 비밀번호 필드 제거
	user.Password = ""
	return token, &user, nil
}

// RefreshUserToken은 사용자의 JWT 토큰을 갱신합니다.
func RefreshUserToken(userID uint) (string, error) {
	// 사용자 존재 확인
	var user model.User
	if err := model.DB.Select("id").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("사용자를 찾을 수 없습니다")
		}
		return "", err
	}

	// 새 토큰 생성
	token, err := util.GenerateJWT(userID)
	if err != nil {
		log.Printf("❌ 토큰 갱신 실패 (사용자 ID: %d): %v", userID, err)
		return "", errors.New("토큰 갱신 중 오류가 발생했습니다")
	}

	log.Printf("✅ 토큰 갱신 완료 (사용자 ID: %d)", userID)
	return token, nil
}

// GetUserByID는 사용자 ID로 사용자를 조회합니다.
func GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	if err := model.DB.Select("id, username, role, created_at, updated_at").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, err
	}

	return &user, nil
}

// IsUserAdmin은 사용자가 관리자인지 확인합니다.
func IsUserAdmin(userID uint) bool {
	var user model.User
	if err := model.DB.Select("role").First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == model.RoleAdmin
}
