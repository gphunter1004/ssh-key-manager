package service

import (
	"log"
	"ssh-key-manager/internal/dto"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// UserService 사용자 관리 서비스
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 사용자 서비스 생성자 (직접 의존성 주입)
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetUserByID ID로 사용자를 조회합니다.
func (us *UserService) GetUserByID(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 사용자 ID입니다",
		)
	}

	user, err := us.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	user.Password = ""
	return user, nil
}

// UpdateUserProfile 사용자 프로필을 업데이트합니다 (비밀번호 변경 보안 강화).
func (us *UserService) UpdateUserProfile(userID uint, req dto.UserUpdateRequest) (*model.User, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 (ID: %d)", userID)

	// 사용자 조회 (FindByID로 존재 확인과 조회를 한 번에)
	user, err := us.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 계정 상태 확인
	if !user.IsAccountAccessible() {
		return nil, model.NewBusinessError(
			model.ErrAccountInactive,
			"비활성화되거나 잠긴 계정은 프로필을 수정할 수 없습니다",
		)
	}

	updates := make(map[string]interface{})

	// 사용자명 변경 처리
	if req.Username != "" && req.Username != user.Username {
		username := strings.TrimSpace(req.Username)
		if username != "" {
			// 중복 확인 (FindByUsername으로 존재 여부 확인)
			existingUser, err := us.userRepo.FindByUsername(username)
			if err != nil && err != gorm.ErrRecordNotFound {
				return nil, model.NewBusinessError(
					model.ErrDatabaseError,
					"사용자명 중복 확인 중 오류가 발생했습니다",
				)
			}
			if existingUser != nil {
				return nil, model.NewBusinessError(
					model.ErrUserAlreadyExists,
					"이미 사용 중인 사용자명입니다",
				)
			}
			updates["username"] = username
		}
	}

	// 비밀번호 변경 처리 (보안 강화)
	if req.NewPassword != "" {
		// 현재 비밀번호 확인 (필수)
		if req.CurrentPassword == "" {
			return nil, model.NewBusinessError(
				model.ErrRequiredField,
				"비밀번호 변경 시 현재 비밀번호를 입력해주세요",
			)
		}

		// 현재 비밀번호 검증
		if !util.CheckPasswordHash(req.CurrentPassword, user.Password) {
			log.Printf("❌ 잘못된 현재 비밀번호로 비밀번호 변경 시도 (사용자 ID: %d)", userID)
			return nil, model.NewBusinessError(
				model.ErrInvalidCredentials,
				"현재 비밀번호가 올바르지 않습니다",
			)
		}

		// 새 비밀번호 유효성 검사
		if len(req.NewPassword) < 4 {
			return nil, model.NewBusinessError(
				model.ErrWeakPassword,
				"비밀번호는 최소 4자 이상이어야 합니다",
			)
		}

		// 현재 비밀번호와 동일한지 확인
		if util.CheckPasswordHash(req.NewPassword, user.Password) {
			return nil, model.NewBusinessError(
				model.ErrInvalidInput,
				"새 비밀번호는 현재 비밀번호와 달라야 합니다",
			)
		}

		// 새 비밀번호 해시
		hashedPassword, err := util.HashPassword(req.NewPassword)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrInternalServer,
				"비밀번호 처리 중 오류가 발생했습니다",
			)
		}
		updates["password"] = hashedPassword
		log.Printf("✅ 비밀번호 변경 성공 (사용자 ID: %d)", userID)
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := us.userRepo.Update(userID, updates); err != nil {
			if strings.Contains(err.Error(), "duplicate") {
				return nil, model.NewBusinessError(
					model.ErrUserAlreadyExists,
					"이미 사용 중인 사용자명입니다",
				)
			}
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"프로필 업데이트 중 오류가 발생했습니다",
			)
		}

		// 업데이트된 사용자 정보 다시 조회
		user, err = us.userRepo.FindByID(userID)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"업데이트된 사용자 정보 조회 실패",
			)
		}
	}

	user.Password = ""
	log.Printf("✅ 사용자 프로필 업데이트 완료: %s", user.Username)
	return user, nil
}

// GetAllUsers 모든 사용자 목록을 반환합니다.
func (us *UserService) GetAllUsers() ([]model.User, error) {
	log.Printf("👥 모든 사용자 목록 조회")

	users, err := us.userRepo.FindAll()
	if err != nil {
		log.Printf("❌ 사용자 목록 조회 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 목록 조회 중 오류가 발생했습니다",
		)
	}

	// 모든 사용자의 비밀번호 필드 제거
	for i := range users {
		users[i].Password = ""
	}

	log.Printf("✅ 사용자 목록 조회 완료 (총 %d명)", len(users))
	return users, nil
}

// GetUserDetailWithKey SSH 키 정보를 포함한 사용자 상세 정보를 반환합니다.
func (us *UserService) GetUserDetailWithKey(userID uint) (map[string]interface{}, error) {
	log.Printf("🔍 사용자 상세 정보 조회 중 (ID: %d)", userID)

	// 사용자 조회 (FindByID로 존재 확인과 조회를 한 번에)
	user, err := us.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	// SSH 키 존재 여부 확인 (KeyService의 개선된 HasUserSSHKey 사용)
	hasSSHKey := C().Key.HasUserSSHKey(userID)

	responseData := map[string]interface{}{
		"id":                 user.ID,
		"username":           user.Username,
		"role":               user.Role,
		"is_active":          user.IsActive,
		"is_locked":          user.IsLocked,
		"failed_login_count": user.FailedLoginCount,
		"locked_at":          user.LockedAt,
		"last_login_at":      user.LastLoginAt,
		"has_ssh_key":        hasSSHKey,
		"created_at":         user.CreatedAt,
		"updated_at":         user.UpdatedAt,
	}

	// SSH 키 상세 정보 추가 (필요한 경우에만)
	if hasSSHKey {
		sshKey, err := C().Key.GetUserSSHKey(userID)
		if err == nil {
			responseData["ssh_key"] = map[string]interface{}{
				"id":         sshKey.ID,
				"created_at": sshKey.CreatedAt,
				"updated_at": sshKey.UpdatedAt,
			}
		}
	}

	log.Printf("✅ 사용자 상세 정보 조회 완료 (사용자: %s, SSH 키: %t)", user.Username, hasSSHKey)
	return responseData, nil
}

// UpdateUserStatus 사용자의 활성/비활성 상태를 변경합니다 (관리자용).
func (us *UserService) UpdateUserStatus(adminUserID, targetUserID uint, req dto.UserStatusUpdateRequest) (*model.User, error) {
	log.Printf("👑 사용자 상태 변경 시도 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인
	admin, err := us.userRepo.FindByID(adminUserID)
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

	// 대상 사용자 조회
	targetUser, err := us.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 자기 자신의 상태 변경 방지
	if adminUserID == targetUserID {
		return nil, model.NewBusinessError(
			model.ErrCannotDeleteSelf,
			"자신의 계정 상태는 변경할 수 없습니다",
		)
	}

	updates := make(map[string]interface{})

	// 활성/비활성 상태 변경
	if req.IsActive != nil && *req.IsActive != targetUser.IsActive {
		updates["is_active"] = *req.IsActive

		// 마지막 관리자 비활성화 방지
		if targetUser.Role == model.RoleAdmin && !*req.IsActive {
			adminCount, err := us.userRepo.CountByRole(model.RoleAdmin)
			if err != nil {
				return nil, model.NewBusinessError(
					model.ErrDatabaseError,
					"관리자 수 확인 중 오류가 발생했습니다",
				)
			}
			// 활성 관리자 수 계산 (현재 대상을 제외)
			if adminCount <= 1 {
				return nil, model.NewBusinessError(
					model.ErrLastAdmin,
					"최소 1명의 활성 관리자가 필요합니다",
				)
			}
		}
	}

	// 업데이트 실행
	if len(updates) > 0 {
		if err := us.userRepo.Update(targetUserID, updates); err != nil {
			log.Printf("❌ 사용자 상태 변경 실패: %v", err)
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"사용자 상태 변경 중 오류가 발생했습니다",
			)
		}

		// 업데이트된 사용자 정보 다시 조회
		targetUser, err = us.userRepo.FindByID(targetUserID)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrDatabaseError,
				"업데이트된 사용자 정보 조회 실패",
			)
		}
	}

	statusText := "비활성화"
	if targetUser.IsActive {
		statusText = "활성화"
	}

	log.Printf("✅ 사용자 상태 변경 완료: %s를 %s (관리자: %s)", targetUser.Username, statusText, admin.Username)
	targetUser.Password = ""
	return targetUser, nil
}

// UnlockUserAccount 사용자 계정 잠금을 해제합니다 (관리자용).
func (us *UserService) UnlockUserAccount(adminUserID, targetUserID uint) (*model.User, error) {
	log.Printf("🔓 계정 잠금 해제 시도 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인
	admin, err := us.userRepo.FindByID(adminUserID)
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

	// 대상 사용자 조회
	targetUser, err := us.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 이미 잠금 해제된 계정인지 확인
	if !targetUser.IsLocked {
		return nil, model.NewBusinessError(
			model.ErrInvalidInput,
			"이미 잠금 해제된 계정입니다",
		)
	}

	// 계정 잠금 해제
	targetUser.UnlockAccount()

	updates := map[string]interface{}{
		"is_locked":          targetUser.IsLocked,
		"failed_login_count": targetUser.FailedLoginCount,
		"locked_at":          targetUser.LockedAt,
	}

	if err := us.userRepo.Update(targetUserID, updates); err != nil {
		log.Printf("❌ 계정 잠금 해제 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"계정 잠금 해제 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 계정 잠금 해제 완료: %s (관리자: %s)", targetUser.Username, admin.Username)
	targetUser.Password = ""
	return targetUser, nil
}

// UpdateUserRole 사용자의 권한을 변경합니다 (관리자만 가능).
func (us *UserService) UpdateUserRole(adminUserID, targetUserID uint, newRole string) error {
	log.Printf("👑 사용자 권한 변경 시도 (관리자 ID: %d, 대상 ID: %d, 새 권한: %s)", adminUserID, targetUserID, newRole)

	// 관리자 권한 확인 (FindByID로 존재 확인과 권한 확인을 한 번에)
	admin, err := us.userRepo.FindByID(adminUserID)
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

	// 새 권한 유효성 검사
	if newRole != string(model.RoleUser) && newRole != string(model.RoleAdmin) {
		return model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 권한입니다. 'user' 또는 'admin'만 가능합니다",
		)
	}

	// 대상 사용자 조회 (FindByID로 존재 확인과 조회를 한 번에)
	targetUser, err := us.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewUserNotFoundError()
		}
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 자기 자신의 권한 변경 방지
	if adminUserID == targetUserID {
		return model.NewBusinessError(
			model.ErrCannotDeleteSelf,
			"자신의 권한은 변경할 수 없습니다",
		)
	}

	// 마지막 관리자 확인 (관리자 → 일반 사용자로 변경하는 경우)
	if targetUser.Role == model.RoleAdmin && newRole == string(model.RoleUser) {
		adminCount, err := us.userRepo.CountByRole(model.RoleAdmin)
		if err != nil {
			return model.NewBusinessError(
				model.ErrDatabaseError,
				"관리자 수 확인 중 오류가 발생했습니다",
			)
		}
		if adminCount <= 1 {
			return model.NewBusinessError(
				model.ErrLastAdmin,
				"최소 1명의 관리자가 필요합니다",
			)
		}
	}

	// 권한 업데이트
	oldRole := string(targetUser.Role)
	updates := map[string]interface{}{
		"role": model.UserRole(newRole),
	}
	if err := us.userRepo.Update(targetUserID, updates); err != nil {
		log.Printf("❌ 권한 변경 실패: %v", err)
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"권한 변경 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 사용자 권한 변경 완료: %s (%s → %s)", targetUser.Username, oldRole, newRole)
	return nil
}

// DeleteUser 사용자를 삭제합니다 (관리자만 가능).
func (us *UserService) DeleteUser(adminUserID, targetUserID uint) error {
	log.Printf("🗑️ 사용자 삭제 시도 (관리자 ID: %d, 대상 ID: %d)", adminUserID, targetUserID)

	// 관리자 권한 확인 (FindByID로 존재 확인과 권한 확인을 한 번에)
	admin, err := us.userRepo.FindByID(adminUserID)
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

	// 대상 사용자 조회 (FindByID로 존재 확인과 조회를 한 번에)
	user, err := us.userRepo.FindByID(targetUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return model.NewUserNotFoundError()
		}
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"대상 사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 자기 자신 삭제 방지
	if adminUserID == targetUserID {
		return model.NewBusinessError(
			model.ErrCannotDeleteSelf,
			"자신의 계정은 삭제할 수 없습니다",
		)
	}

	// 마지막 관리자 삭제 방지
	if user.Role == model.RoleAdmin {
		adminCount, err := us.userRepo.CountByRole(model.RoleAdmin)
		if err != nil {
			return model.NewBusinessError(
				model.ErrDatabaseError,
				"관리자 수 확인 중 오류가 발생했습니다",
			)
		}
		if adminCount <= 1 {
			return model.NewBusinessError(
				model.ErrLastAdmin,
				"최소 1명의 관리자가 필요합니다",
			)
		}
	}

	// 트랜잭션으로 관련 데이터 함께 삭제 (다른 서비스 사용)
	err = us.userRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		// SSH 키 삭제
		C().Key.DeleteUserSSHKey(targetUserID)

		// 서버 삭제 (사용자 소유 서버들)
		servers, err := C().Server.GetUserServers(targetUserID)
		if err == nil {
			for _, server := range servers {
				C().Server.DeleteServer(targetUserID, server.ID)
			}
		}

		// 사용자 삭제
		return us.userRepo.Delete(targetUserID)
	})

	if err != nil {
		log.Printf("❌ 사용자 삭제 실패: %v", err)
		return model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 삭제 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 사용자 삭제 완료: %s (ID: %d, 권한: %s)", user.Username, targetUserID, string(user.Role))
	return nil
}
