// services/user_service.go 수정 - Role 필드가 제대로 조회되도록 수정

package services

import (
	"errors"
	"log"
	"ssh-key-manager/models"
	"ssh-key-manager/types"
	"ssh-key-manager/utils"
	"strings"

	"gorm.io/gorm"
)

// GetAllUsers는 모든 사용자 목록을 반환합니다. (관리자 전용)
func GetAllUsers() ([]types.UserInfo, error) {
	log.Printf("👥 모든 사용자 목록 조회 중...")

	var users []models.User
	// role 필드도 포함하여 조회
	result := models.DB.Select("id, username, role, created_at, updated_at").Find(&users)
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
	var userInfos []types.UserInfo
	for _, user := range users {
		userInfo := types.ToUserInfo(user, keyMap[user.ID])
		userInfos = append(userInfos, userInfo)
	}

	log.Printf("✅ 사용자 목록 조회 완료 (총 %d명)", len(userInfos))
	return userInfos, nil
}

// GetUserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보를 반환합니다.
func GetUserDetailWithKey(userID uint) (*types.UserDetailWithKey, error) {
	log.Printf("🔍 사용자 상세 정보 조회 중 (ID: %d)", userID)

	var user models.User
	// role 필드도 포함하여 조회
	result := models.DB.Select("id, username, role, created_at, updated_at").First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		log.Printf("❌ 사용자 조회 실패: %v", result.Error)
		return nil, result.Error
	}

	hasSSHKey := false
	var sshKeyResponse *types.SSHKeyResponse

	// SSH 키 정보 조회
	var sshKey models.SSHKey
	keyResult := models.DB.Where("user_id = ?", userID).First(&sshKey)
	if keyResult.Error == nil {
		// SSH 키가 있는 경우
		hasSSHKey = true

		// SSH 키 핑거프린트 생성
		fingerprint, err := generateSSHKeyFingerprint(sshKey.PublicKey)
		if err != nil {
			log.Printf("⚠️ 핑거프린트 생성 실패: %v", err)
		}

		response := types.ToSSHKeyResponse(sshKey, fingerprint)
		sshKeyResponse = &response
	} else if !errors.Is(keyResult.Error, gorm.ErrRecordNotFound) {
		// SSH 키 조회 중 다른 오류 발생
		log.Printf("⚠️ SSH 키 조회 중 오류: %v", keyResult.Error)
	}

	userDetail := types.ToUserDetailWithKey(user, hasSSHKey, sshKeyResponse)

	log.Printf("✅ 사용자 상세 정보 조회 완료 (사용자: %s, 권한: %s, SSH 키: %t)",
		user.Username, string(user.Role), hasSSHKey)
	return &userDetail, nil
}

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(userID uint, updateData types.UserProfileUpdate) (*types.UserInfo, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 중 (ID: %d)", userID)

	var user models.User
	// role 필드도 포함하여 조회
	result := models.DB.Select("id, username, role, password, created_at, updated_at").First(&user, userID)
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

		// 업데이트된 사용자 정보 다시 조회 (role 포함)
		models.DB.Select("id, username, role, created_at, updated_at").First(&user, userID)
	}

	// SSH 키 존재 여부 확인
	var keyCount int64
	models.DB.Model(&models.SSHKey{}).Where("user_id = ?", userID).Count(&keyCount)

	userInfo := types.ToUserInfo(user, keyCount > 0)

	log.Printf("✅ 사용자 프로필 업데이트 완료 (사용자: %s, 권한: %s)",
		user.Username, string(user.Role))
	return &userInfo, nil
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
func GetUserStats() (*types.UserStats, error) {
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

	stats := &types.UserStats{
		TotalUsers:         totalUsers,
		UsersWithKeys:      usersWithKeys,
		UsersWithoutKeys:   totalUsers - usersWithKeys,
		KeyCoveragePercent: float64(usersWithKeys) / float64(totalUsers) * 100,
	}

	log.Printf("✅ 사용자 통계 조회 완료 (전체: %d명, 키 보유: %d명)", totalUsers, usersWithKeys)
	return stats, nil
}

// CreateAdminUser는 초기 관리자 계정을 생성합니다.
func CreateAdminUser(username, password string) error {
	log.Printf("👑 초기 관리자 계정 생성 시도: %s", username)

	// 이미 관리자가 있는지 확인
	var adminCount int64
	if err := models.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		log.Printf("⚠️ 관리자 계정이 이미 존재합니다. 건너뜀")
		return nil
	}

	// 해당 사용자명이 이미 존재하는지 확인
	var existingUser models.User
	result := models.DB.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		// 사용자가 존재하면 관리자로 승격
		log.Printf("🔄 기존 사용자를 관리자로 승격: %s", username)
		if err := models.DB.Model(&existingUser).Update("role", models.RoleAdmin).Error; err != nil {
			return err
		}
		log.Printf("✅ 사용자 %s가 관리자로 승격되었습니다", username)
		return nil
	}

	// 새로운 관리자 계정 생성
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return err
	}

	admin := models.User{
		Username: username,
		Password: hashedPassword,
		Role:     models.RoleAdmin,
	}

	if err := models.DB.Create(&admin).Error; err != nil {
		log.Printf("❌ 관리자 계정 생성 실패: %v", err)
		return err
	}

	log.Printf("✅ 초기 관리자 계정 생성 완료: %s (ID: %d)", username, admin.ID)
	log.Printf("🔑 관리자 비밀번호: %s", password)
	log.Printf("⚠️ 보안을 위해 비밀번호를 변경하세요!")

	return nil
}

// UpdateUserRole은 사용자의 권한을 변경합니다 (관리자만 가능).
func UpdateUserRole(adminUserID, targetUserID uint, newRole string) error {
	log.Printf("👑 사용자 권한 변경 시도 (관리자 ID: %d, 대상 ID: %d, 새 권한: %s)", adminUserID, targetUserID, newRole)

	// 관리자 권한 확인
	if !IsUserAdmin(adminUserID) {
		return errors.New("관리자 권한이 필요합니다")
	}

	// 권한 값 검증
	if newRole != string(models.RoleUser) && newRole != string(models.RoleAdmin) {
		return errors.New("유효하지 않은 권한입니다. 'user' 또는 'admin'만 가능합니다")
	}

	// 대상 사용자 조회 (role 포함)
	var targetUser models.User
	if err := models.DB.Select("id, username, role").First(&targetUser, targetUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("사용자를 찾을 수 없습니다")
		}
		return err
	}

	// 자신의 권한은 변경할 수 없음 (안전장치)
	if adminUserID == targetUserID {
		return errors.New("자신의 권한은 변경할 수 없습니다")
	}

	// 마지막 관리자 보호 (최소 1명의 관리자 유지)
	if targetUser.Role == models.RoleAdmin && newRole == string(models.RoleUser) {
		var adminCount int64
		models.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("최소 1명의 관리자가 필요합니다")
		}
	}

	// 권한 변경
	oldRole := string(targetUser.Role)
	if err := models.DB.Model(&targetUser).Update("role", models.UserRole(newRole)).Error; err != nil {
		log.Printf("❌ 권한 변경 실패: %v", err)
		return errors.New("권한 변경 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 권한 변경 완료: %s (%s → %s)", targetUser.Username, oldRole, newRole)
	return nil
}

// IsUserAdmin은 사용자가 관리자인지 확인합니다.
func IsUserAdmin(userID uint) bool {
	var user models.User
	if err := models.DB.Select("role").First(&user, userID).Error; err != nil {
		return false
	}
	return user.Role == models.RoleAdmin
}

// GetUserRole은 사용자의 권한을 반환합니다.
func GetUserRole(userID uint) (models.UserRole, error) {
	var user models.User
	if err := models.DB.Select("role").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("사용자를 찾을 수 없습니다")
		}
		return "", err
	}
	return user.Role, nil
}

// GetAdminStats는 관리자용 통계 정보를 반환합니다.
func GetAdminStats() (*types.AdminStats, error) {
	log.Printf("📊 관리자 통계 조회 중...")

	var totalUsers, adminUsers, regularUsers int64
	var totalServers, totalSSHKeys, totalDeployments int64

	// 사용자 통계
	models.DB.Model(&models.User{}).Count(&totalUsers)
	models.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminUsers)
	models.DB.Model(&models.User{}).Where("role = ?", models.RoleUser).Count(&regularUsers)

	// 서버 통계
	models.DB.Model(&models.Server{}).Count(&totalServers)

	// SSH 키 통계
	models.DB.Model(&models.SSHKey{}).Count(&totalSSHKeys)

	// 배포 통계
	models.DB.Model(&models.ServerKeyDeployment{}).Count(&totalDeployments)

	stats := &types.AdminStats{
		TotalUsers:       totalUsers,
		AdminUsers:       adminUsers,
		RegularUsers:     regularUsers,
		TotalServers:     totalServers,
		TotalSSHKeys:     totalSSHKeys,
		TotalDeployments: totalDeployments,
	}

	log.Printf("✅ 관리자 통계 조회 완료")
	return stats, nil
}

// DeleteUser는 사용자를 삭제합니다 (관리자만 가능).
func DeleteUser(targetUserID uint) error {
	log.Printf("🗑️ 사용자 삭제 시도 (ID: %d)", targetUserID)

	// 사용자 존재 확인 (role 포함)
	var user models.User
	if err := models.DB.Select("id, username, role").First(&user, targetUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("사용자를 찾을 수 없습니다")
		}
		return err
	}

	// 마지막 관리자 보호
	if user.Role == models.RoleAdmin {
		var adminCount int64
		models.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("최소 1명의 관리자가 필요합니다")
		}
	}

	// 사용자와 관련된 모든 데이터 삭제 (CASCADE로 자동 삭제되지만 명시적으로 처리)
	tx := models.DB.Begin()

	// SSH 키 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&models.SSHKey{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 서버 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&models.Server{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 배포 기록 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&models.ServerKeyDeployment{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 사용자 삭제
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	log.Printf("✅ 사용자 삭제 완료: %s (ID: %d, 권한: %s)", user.Username, targetUserID, string(user.Role))
	return nil
}
