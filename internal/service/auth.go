package service

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/repository"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// AuthService 인증 관리 서비스
type AuthService struct {
	userRepo *repository.UserRepository
}

// NewAuthService 인증 서비스 생성자
func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// RegisterUser 새로운 사용자를 등록합니다.
func (as *AuthService) RegisterUser(username, password string) (*model.User, error) {
	log.Printf("👤 새 사용자 등록 시도: %s", username)

	// 1. 입력값 검증
	if err := as.validateRegistrationInput(username, password); err != nil {
		return nil, err
	}

	// 2. 사용자명 중복 확인 (FindByUsername으로 존재 여부와 조회를 한 번에 처리)
	existingUser, err := as.userRepo.FindByUsername(username)
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

	// 3. 비밀번호 해시
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrInternalServer,
			"비밀번호 처리 중 오류가 발생했습니다",
		)
	}

	// 4. 사용자 생성 (기본적으로 활성 상태)
	user := &model.User{
		Username:         strings.TrimSpace(username),
		Password:         hashedPassword,
		Role:             model.RoleUser,
		IsActive:         true,  // 기본 활성 상태
		IsLocked:         false, // 기본 잠금 해제 상태
		FailedLoginCount: 0,     // 실패 횟수 초기화
	}

	if err := as.userRepo.Create(user); err != nil {
		if strings.Contains(err.Error(), "duplicate") ||
			strings.Contains(err.Error(), "unique") {
			return nil, model.NewBusinessError(
				model.ErrUserAlreadyExists,
				"이미 사용 중인 사용자명입니다",
			)
		}
		log.Printf("❌ 사용자 등록 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 등록 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 사용자 등록 완료: %s (ID: %d)", username, user.ID)
	user.Password = "" // 응답에서 비밀번호 제거
	return user, nil
}

// AuthenticateUser 사용자를 인증하고 JWT를 반환합니다 (보안 기능 강화).
func (as *AuthService) AuthenticateUser(username, password string) (string, *model.User, error) {
	log.Printf("🔐 사용자 인증 시도: %s", username)

	// 1. 입력값 검증
	if err := as.validateAuthenticationInput(username, password); err != nil {
		return "", nil, err
	}

	// 2. 사용자 조회
	user, err := as.userRepo.FindByUsername(username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("❌ 존재하지 않는 사용자 로그인 시도: %s", username)
			return "", nil, model.NewInvalidCredentialsError()
		}
		return "", nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 3. 계정 상태 확인 (활성 상태 + 잠금 상태)
	if !user.IsAccountAccessible() {
		if !user.IsActive {
			log.Printf("❌ 비활성 계정 로그인 시도: %s (ID: %d)", username, user.ID)
			return "", nil, model.NewBusinessError(
				model.ErrAccountInactive,
				"비활성화된 계정입니다. 관리자에게 문의하세요",
			)
		}
		if user.IsLocked {
			log.Printf("❌ 잠긴 계정 로그인 시도: %s (ID: %d, 실패횟수: %d)", username, user.ID, user.FailedLoginCount)
			return "", nil, model.NewBusinessError(
				model.ErrAccountLocked,
				"계정이 잠겼습니다. 관리자에게 문의하세요",
			)
		}
	}

	// 4. 비밀번호 확인
	if !util.CheckPasswordHash(password, user.Password) {
		log.Printf("❌ 잘못된 비밀번호 로그인 시도: %s (ID: %d, 현재 실패횟수: %d)", username, user.ID, user.FailedLoginCount)

		// 로그인 실패 횟수 증가 및 계정 잠금 처리
		user.IncrementFailedLogin()
		if err := as.updateUserSecurityFields(user); err != nil {
			log.Printf("❌ 사용자 보안 필드 업데이트 실패: %v", err)
		}

		if user.IsLocked {
			log.Printf("🔒 계정 자동 잠금: %s (ID: %d, 실패횟수: %d)", username, user.ID, user.FailedLoginCount)
			return "", nil, model.NewBusinessError(
				model.ErrAccountLocked,
				"로그인 실패 횟수 초과로 계정이 잠겼습니다. 관리자에게 문의하세요",
			)
		}

		return "", nil, model.NewInvalidCredentialsError()
	}

	// 5. 로그인 성공 처리
	user.ResetFailedLogin()
	user.UpdateLastLogin()
	if err := as.updateUserSecurityFields(user); err != nil {
		log.Printf("❌ 로그인 성공 정보 업데이트 실패: %v", err)
	}

	// 6. JWT 토큰 생성
	token, err := util.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("❌ JWT 생성 실패: %v", err)
		return "", nil, model.NewBusinessError(
			model.ErrInternalServer,
			"인증 토큰 생성 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 사용자 인증 완료: %s (ID: %d)", username, user.ID)

	user.Password = "" // 응답에서 비밀번호 제거
	return token, user, nil
}

// RefreshUserToken 사용자의 JWT 토큰을 갱신합니다.
func (as *AuthService) RefreshUserToken(userID uint) (string, error) {
	// 1. 사용자 존재 및 상태 확인 (FindByID로 한 번에 처리)
	user, err := as.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", model.NewUserNotFoundError()
		}
		return "", model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 확인 중 오류가 발생했습니다",
		)
	}

	// 2. 계정 상태 확인
	if !user.IsAccountAccessible() {
		if !user.IsActive {
			return "", model.NewBusinessError(
				model.ErrAccountInactive,
				"비활성화된 계정입니다",
			)
		}
		if user.IsLocked {
			return "", model.NewBusinessError(
				model.ErrAccountLocked,
				"잠긴 계정입니다",
			)
		}
	}

	// 3. 새 토큰 생성
	token, err := util.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("❌ 토큰 갱신 실패 (사용자 ID: %d): %v", userID, err)
		return "", model.NewBusinessError(
			model.ErrInternalServer,
			"토큰 갱신 중 오류가 발생했습니다",
		)
	}

	log.Printf("✅ 토큰 갱신 완료 (사용자 ID: %d)", userID)
	return token, nil
}

// ========== 내부 헬퍼 함수들 ==========

// updateUserSecurityFields 사용자의 보안 관련 필드를 업데이트합니다.
func (as *AuthService) updateUserSecurityFields(user *model.User) error {
	updates := map[string]interface{}{
		"failed_login_count": user.FailedLoginCount,
		"is_locked":          user.IsLocked,
		"last_login_at":      user.LastLoginAt,
		"locked_at":          user.LockedAt,
	}
	return as.userRepo.Update(user.ID, updates)
}

// validateRegistrationInput 회원가입 입력값을 검증합니다.
func (as *AuthService) validateRegistrationInput(username, password string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"사용자명을 입력해주세요",
		)
	}

	if len(username) < 2 {
		return model.NewBusinessError(
			model.ErrInvalidUsername,
			"사용자명은 최소 2자 이상이어야 합니다",
		)
	}

	if len(username) > 50 {
		return model.NewBusinessError(
			model.ErrInvalidUsername,
			"사용자명은 최대 50자까지 가능합니다",
		)
	}

	if len(password) < 4 {
		return model.NewBusinessError(
			model.ErrWeakPassword,
			"비밀번호는 최소 4자 이상이어야 합니다",
		)
	}

	if len(password) > 100 {
		return model.NewBusinessError(
			model.ErrWeakPassword,
			"비밀번호는 최대 100자까지 가능합니다",
		)
	}

	return nil
}

// validateAuthenticationInput 로그인 입력값을 검증합니다.
func (as *AuthService) validateAuthenticationInput(username, password string) error {
	username = strings.TrimSpace(username)

	if username == "" || password == "" {
		return model.NewBusinessError(
			model.ErrRequiredField,
			"사용자명과 비밀번호를 입력해주세요",
		)
	}

	return nil
}
