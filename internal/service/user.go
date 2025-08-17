package service

import (
	"errors"
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// UpdateUserProfile은 사용자 프로필을 업데이트합니다.
func UpdateUserProfile(userID uint, req model.UserUpdateRequest) (*model.User, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 (ID: %d)", userID)

	var user model.User
	if err := model.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, err
	}

	updates := make(map[string]interface{})

	// 사용자명 업데이트
	if req.Username != "" && req.Username != user.Username {
		username := strings.TrimSpace(req.Username)
		if username != "" {
			// 중복 확인
			var existingUser model.User
			if err := model.DB.Where("username = ? AND id != ?", username, userID).First(&existingUser).Error; err == nil {
				return nil, errors.New("이미 사용 중인 사용자명입니다")
			}
			updates["username"] = username
		}
	}

	// 비밀번호 업데이트
	if req.NewPassword != "" {
		if len(req.NewPassword) < 4 {
			return nil, errors.New("비밀번호는 최소 4자 이상이어야 합니다")
		}
		hashedPassword, err := util.HashPassword(req.NewPassword)
		if err != nil {
			return nil, errors.New("비밀번호 처리 중 오류가 발생했습니다")
		}
		updates["password"] = hashedPassword
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := model.DB.Model(&user).Updates(updates).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return nil, errors.New("이미 사용 중인 사용자명입니다")
			}
			return nil, errors.New("프로필 업데이트 중 오류가 발생했습니다")
		}
		
		// 업데이트된 정보 다시 조회
		model.DB.First(&user, userID)
	}

	// 비밀번호 필드 제거
	user.Password = ""
	log.Printf("✅ 사용자 프로필 업데이트 완료: %s", user.Username)
	return &user, nil
}

// GetAllUsers는 모든 사용자 목록을 반환합니다.
func GetAllUsers() ([]model.User, error) {
	log.Printf("👥 모든 사용자 목록 조회")

	var users []model.User
	if err := model.DB.Select("id, username, role, created_at, updated_at").Find(&users).Error; err != nil {
		log.Printf("❌ 사용자 목록 조회 실패: %v", err)
		return nil, err
	}

	log.Printf("✅ 사용자 목록 조회 완료 (총 %d명)", len(users))
	return users, nil
}

// GetUserDetailWithKey는 SSH 키 정보를 포함한 사용자 상세 정보를 반환합니다.
func GetUserDetailWithKey(userID uint) (map[string]interface{}, error) {
	log.Printf("🔍 사용자 상세 정보 조회 중 (ID: %d)", userID)

	var user model.User
	if err := model.DB.Select("id, username, role, created_at, updated_at").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("사용자를 찾을 수 없습니다")
		}
		return nil, err
	}

	// SSH 키 존재 여부 및 정보 확인
	hasSSHKey := HasUserSSHKey(userID)
	
	responseData := map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"role":       user.Role,
		"has_ssh_key": hasSSHKey,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	// SSH 키 상세 정보 포함 (있는 경우)
	if hasSSHKey {
		sshKey, err := GetUserSSHKey(userID)
		if err == nil {
			responseData["ssh_key"] = map[string]interface{}{
				"id":        sshKey.ID,
				"algorithm": sshKey.Algorithm,
				"bits":      sshKey.Bits,
				"created_at": sshKey.CreatedAt,
				"updated_at": sshKey.UpdatedAt,
			}
		}
	}

	log.Printf("✅ 사용자 상세 정보 조회 완료 (사용자: %s, SSH 키: %t)", user.Username, hasSSHKey)
	return responseData, nil
}

// UpdateUserRole은 사용자의 권한을 변경합니다 (관리자만 가능).
func UpdateUserRole(adminUserID, targetUserID uint, newRole string) error {
	log.Printf("👑 사용자 권한 변경 시도 (관리자 ID: %d, 대상 ID: %d, 새 권한: %s)", adminUserID, targetUserID, newRole)

	// 관리자 권한 확인
	if !IsUserAdmin(adminUserID) {
		return errors.New("관리자 권한이 필요합니다")
	}

	// 권한 값 검증
	if newRole != string(model.RoleUser) && newRole != string(model.RoleAdmin) {
		return errors.New("유효하지 않은 권한입니다. 'user' 또는 'admin'만 가능합니다")
	}

	// 대상 사용자 조회
	var targetUser model.User
	if err := model.DB.Select("id, username, role").First(&targetUser, targetUserID).Error; err != nil {
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
	if targetUser.Role == model.RoleAdmin && newRole == string(model.RoleUser) {
		var adminCount int64
		model.DB.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("최소 1명의 관리자가 필요합니다")
		}
	}

	// 권한 변경
	oldRole := string(targetUser.Role)
	if err := model.DB.Model(&targetUser).Update("role", model.UserRole(newRole)).Error; err != nil {
		log.Printf("❌ 권한 변경 실패: %v", err)
		return errors.New("권한 변경 중 오류가 발생했습니다")
	}

	log.Printf("✅ 사용자 권한 변경 완료: %s (%s → %s)", targetUser.Username, oldRole, newRole)
	return nil
}

// DeleteUser는 사용자를 삭제합니다 (관리자만 가능).
func DeleteUser(adminUserID, targetUserID uint) error {
	log.Printf("🗑️ 사용자 삭제 시도 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인
	if !IsUserAdmin(adminUserID) {
		return errors.New("관리자 권한이 필요합니다")
	}

	// 사용자 존재 확인
	var user model.User
	if err := model.DB.Select("id, username, role").First(&user, targetUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("사용자를 찾을 수 없습니다")
		}
		return err
	}

	// 마지막 관리자 보호
	if user.Role == model.RoleAdmin {
		var adminCount int64
		model.DB.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			return errors.New("최소 1명의 관리자가 필요합니다")
		}
	}

	// 트랜잭션으로 관련 데이터 함께 삭제
	tx := model.DB.Begin()

	// SSH 키 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&model.SSHKey{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 서버 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&model.Server{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 배포 기록 삭제
	if err := tx.Where("user_id = ?", targetUserID).Delete(&model.ServerKeyDeployment{}).Error; err != nil {
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
