package services

import (
	"errors"
	"log"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserInfo는 사용자 기본 정보를 담는 구조체입니다.
type UserInfo struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HasSSHKey bool      `json:"has_ssh_key"`
}

// UserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보입니다.
type UserDetailWithKey struct {
	ID        uint            `json:"id"`
	Username  string          `json:"username"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	HasSSHKey bool            `json:"has_ssh_key"`
	SSHKey    *SSHKeyResponse `json:"ssh_key,omitempty"`
}

// SSHKeyResponse는 API 응답용 SSH 키 정보입니다.
type SSHKeyResponse struct {
	ID          uint      `json:"id"`
	Algorithm   string    `json:"algorithm"`
	Bits        int       `json:"bits"`
	PublicKey   string    `json:"public_key"`
	PEM         string    `json:"pem"`
	PPK         string    `json:"ppk"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Fingerprint string    `json:"fingerprint,omitempty"`
}

// UserProfileUpdate는 사용자 프로필 업데이트용 구조체입니다.
type UserProfileUpdate struct {
	Username    string `json:"username,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
}

// GetAllUsers는 모든 사용자 목록을 반환합니다.
func GetAllUsers() ([]UserInfo, error) {
	log.Printf("👥 모든 사용자 목록 조회 중...")

	var users []models.User
	result := models.DB.Select("id, username, created_at, updated_at").Find(&users)
	if result.Error != nil {
		log.Printf("❌ 사용자 목록 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	// SSH 키 존재 여부 확인을 위한 맵
	var userIDs []uint
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	// 각 사용자의 SSH 키 존재 여부 확인
	var keyUsers []struct {
		UserID uint
	}
	models.DB.Model(&models.SSHKey{}).Select("user_id").Where("user_id IN ?", userIDs).Find(&keyUsers)

	keyMap := make(map[uint]bool)
	for _, ku := range keyUsers {
		keyMap[ku.UserID] = true
	}

	// 응답 데이터 구성
	var userInfos []UserInfo
	for _, user := range users {
		userInfo := UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			HasSSHKey: keyMap[user.ID],
		}
		userInfos = append(userInfos, userInfo)
	}

	log.Printf("✅ 사용자 목록 조회 완료 (총 %d명)", len(userInfos))
	return userInfos, nil
}

// GetUserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보를 반환합니다.
func GetUserDetailWithKey(userID uint) (*UserDetailWithKey, error) {
	log.Printf("🔍 사용자 상세 정보 조회 중 (ID: %d)", userID)

	var user models.User
	result := models.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		log.Printf("❌ 사용자 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	userDetail := &UserDetailWithKey{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		HasSSHKey: false,
	}

	// SSH 키 정보 조회
	var sshKey models.SSHKey
	keyResult := models.DB.Where("user_id = ?", userID).First(&sshKey)
	if keyResult.Error == nil {
		// SSH 키가 있는 경우
		userDetail.HasSSHKey = true

		// SSH 키 핑거프린트 생성
		fingerprint, err := generateSSHKeyFingerprint(sshKey.PublicKey)
		if err != nil {
			log.Printf("⚠️ 핑거프린트 생성 실패: %v", err)
		}

		userDetail.SSHKey = &SSHKeyResponse{
			ID:          sshKey.ID,
			Algorithm:   sshKey.Algorithm,
			Bits:        sshKey.Bits,
			PublicKey:   sshKey.PublicKey,
			PEM:         sshKey.PEM,
			PPK:         sshKey.PPK,
			CreatedAt:   sshKey.CreatedAt,
			UpdatedAt:   sshKey.UpdatedAt,
			Fingerprint: fingerprint,
		}
	} else if !errors.Is(keyResult.Error, gorm.ErrRecordNotFound) {
		// SSH 키 조회 중 다른 오류 발생
		log.Printf("⚠️ SSH 키 조회 중 오류: %v", keyResult.Error)
	}

	log.Printf("✅ 사용자 상세 정보 조회 완료 (사용자: %s, SSH 키: %t)", user.Username, userDetail.HasSSHKey)
	return userDetail, nil
}

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(userID uint, updateData UserProfileUpdate) (*UserInfo, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 중 (ID: %d)", userID)

	var user models.User
	result := models.DB.First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, result.Error
	}

	// 업데이트할 필드 확인 및 검증
	updates := make(map[string]interface{})

	// 사용자명 업데이트
	if updateData.Username != "" && updateData.Username != user.Username {
		username := strings.TrimSpace(updateData.Username)
		if username == "" {
			return nil, errors.New("사용자명을 입력해주세요")
		}

		// 중복 사용자명 확인
		var existingUser models.User
		if err := models.DB.Where("username = ? AND id != ?", username, userID).First(&existingUser).Error; err == nil {
			return nil, errors.New("이미 사용 중인 사용자명입니다")
		}

		updates["username"] = username
		log.Printf("   - 사용자명 변경: %s -> %s", user.Username, username)
	}

	// 비밀번호 업데이트
	if updateData.NewPassword != "" {
		if len(updateData.NewPassword) < 4 {
			return nil, errors.New("비밀번호는 최소 4자 이상이어야 합니다")
		}

		hashedPassword, err := utils.HashPassword(updateData.NewPassword)
		if err != nil {
			log.Printf("❌ 비밀번호 해싱 실패: %v", err)
			return nil, errors.New("비밀번호 처리 중 오류가 발생했습니다")
		}

		updates["password"] = hashedPassword
		log.Printf("   - 비밀번호 변경됨")
	}

	// 업데이트할 내용이 있는 경우에만 실행
	if len(updates) > 0 {
		if err := models.DB.Model(&user).Updates(updates).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				return nil, errors.New("이미 사용 중인 사용자명입니다")
			}
			log.Printf("❌ 프로필 업데이트 실패: %v", err)
			return nil, errors.New("프로필 업데이트 중 오류가 발생했습니다")
		}

		// 업데이트된 사용자 정보 다시 조회
		models.DB.First(&user, userID)
	}

	// SSH 키 존재 여부 확인
	var keyCount int64
	models.DB.Model(&models.SSHKey{}).Where("user_id = ?", userID).Count(&keyCount)

	userInfo := &UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		HasSSHKey: keyCount > 0,
	}

	log.Printf("✅ 사용자 프로필 업데이트 완료 (사용자: %s)", user.Username)
	return userInfo, nil
}

// generateSSHKeyFingerprint는 SSH 공개키의 핑거프린트를 생성합니다.
func generateSSHKeyFingerprint(publicKey string) (string, error) {
	// 간단한 핑거프린트 생성 (실제로는 더 복잡한 알고리즘 사용)
	// 여기서는 공개키의 앞부분과 뒷부분을 조합하여 간단한 식별자 생성
	lines := strings.Split(strings.TrimSpace(publicKey), " ")
	if len(lines) < 2 {
		return "", errors.New("유효하지 않은 공개키 형식")
	}

	keyData := lines[1] // base64 인코딩된 키 데이터
	if len(keyData) < 16 {
		return "", errors.New("키 데이터가 너무 짧습니다")
	}

	// 간단한 핑거프린트: 처음 8자 + 마지막 8자를 콜론으로 구분
	fingerprint := ""
	start := keyData[:8]
	end := keyData[len(keyData)-8:]

	// 콜론으로 구분된 형식으로 변환
	for i, c := range start + end {
		if i > 0 && i%2 == 0 {
			fingerprint += ":"
		}
		fingerprint += string(c)
	}

	return fingerprint, nil
}

// GetUserStats는 사용자 통계 정보를 반환합니다.
func GetUserStats() (map[string]interface{}, error) {
	log.Printf("📊 사용자 통계 조회 중...")

	var totalUsers int64
	var usersWithKeys int64

	// 전체 사용자 수
	if err := models.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return nil, err
	}

	// SSH 키를 가진 사용자 수
	if err := models.DB.Model(&models.User{}).
		Joins("JOIN ssh_keys ON users.id = ssh_keys.user_id").
		Count(&usersWithKeys).Error; err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_users":          totalUsers,
		"users_with_keys":      usersWithKeys,
		"users_without_keys":   totalUsers - usersWithKeys,
		"key_coverage_percent": float64(usersWithKeys) / float64(totalUsers) * 100,
	}

	log.Printf("✅ 사용자 통계 조회 완료 (전체: %d명, 키 보유: %d명)", totalUsers, usersWithKeys)
	return stats, nil
}
