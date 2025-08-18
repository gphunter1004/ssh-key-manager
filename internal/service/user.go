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

// UpdateUserProfile 사용자 프로필을 업데이트합니다.
func (us *UserService) UpdateUserProfile(userID uint, req dto.UserUpdateRequest) (*model.User, error) {
	log.Printf("✏️ 사용자 프로필 업데이트 (ID: %d)", userID)

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

	updates := make(map[string]interface{})

	if req.Username != "" && req.Username != user.Username {
		username := strings.TrimSpace(req.Username)
		if username != "" {
			exists, err := us.userRepo.ExistsByUsername(username)
			if err != nil {
				return nil, model.NewBusinessError(
					model.ErrDatabaseError,
					"사용자명 중복 확인 중 오류가 발생했습니다",
				)
			}
			if exists {
				return nil, model.NewBusinessError(
					model.ErrUserAlreadyExists,
					"이미 사용 중인 사용자명입니다",
				)
			}
			updates["username"] = username
		}
	}

	if req.NewPassword != "" {
		if len(req.NewPassword) < 4 {
			return nil, model.NewBusinessError(
				model.ErrWeakPassword,
				"비밀번호는 최소 4자 이상이어야 합니다",
			)
		}
		hashedPassword, err := util.HashPassword(req.NewPassword)
		if err != nil {
			return nil, model.NewBusinessError(
				model.ErrInternalServer,
				"비밀번호 처리 중 오류가 발생했습니다",
			)
		}
		updates["password"] = hashedPassword
	}

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

	for i := range users {
		users[i].Password = ""
	}

	log.Printf("✅ 사용자 목록 조회 완료 (총 %d명)", len(users))
	return users, nil
}

// GetUserDetailWithKey SSH 키 정보를 포함한 사용자 상세 정보를 반환합니다.
func (us *UserService) GetUserDetailWithKey(userID uint) (map[string]interface{}, error) {
	log.Printf("🔍 사용자 상세 정보 조회 중 (ID: %d)", userID)

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

	// SSH 키 존재 여부 확인 (KeyService 사용)
	hasSSHKey := C().Key.HasUserSSHKey(userID)

	responseData := map[string]interface{}{
		"id":          user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"has_ssh_key": hasSSHKey,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,
	}

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

// UpdateUserRole 사용자의 권한을 변경합니다 (관리자만 가능).
func (us *UserService) UpdateUserRole(adminUserID, targetUserID uint, newRole string) error {
	log.Printf("👑 사용자 권한 변경 시도 (관리자 ID: %d, 대상 ID: %d, 새 권한: %s)", adminUserID, targetUserID, newRole)

	admin, err := us.userRepo.FindByID(adminUserID)
	if err != nil {
		return model.NewBusinessError(
			model.ErrUserNotFound,
			"관리자 사용자를 찾을 수 없습니다",
		)
	}
	if admin.Role != model.RoleAdmin {
		return model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

	if newRole != string(model.RoleUser) && newRole != string(model.RoleAdmin) {
		return model.NewBusinessError(
			model.ErrInvalidInput,
			"유효하지 않은 권한입니다. 'user' 또는 'admin'만 가능합니다",
		)
	}

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

	if adminUserID == targetUserID {
		return model.NewBusinessError(
			model.ErrCannotDeleteSelf,
			"자신의 권한은 변경할 수 없습니다",
		)
	}

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

	admin, err := us.userRepo.FindByID(adminUserID)
	if err != nil {
		return model.NewBusinessError(
			model.ErrUserNotFound,
			"관리자 사용자를 찾을 수 없습니다",
		)
	}
	if admin.Role != model.RoleAdmin {
		return model.NewBusinessError(
			model.ErrPermissionDenied,
			"관리자 권한이 필요합니다",
		)
	}

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

	if adminUserID == targetUserID {
		return model.NewBusinessError(
			model.ErrCannotDeleteSelf,
			"자신의 계정은 삭제할 수 없습니다",
		)
	}

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
