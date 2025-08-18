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

	// 2. 사용자명 중복 확인
	exists, err := as.userRepo.ExistsByUsername(username)
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

	// 3. 비밀번호 해시
	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		log.Printf("❌ 비밀번호 해싱 실패: %v", err)
		return nil, model.NewBusinessError(
			model.ErrInternalServer,
			"비밀번호 처리 중 오류가 발생했습니다",
		)
	}

	// 4. 사용자 생성
	user := &model.User{
		Username: strings.TrimSpace(username),
		Password: hashedPassword,
		Role:     model.RoleUser,
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

// AuthenticateUser 사용자를 인증하고 JWT를 반환합니다.
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
			return "", nil, model.NewInvalidCredentialsError()
		}
		return "", nil, model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	// 3. 비밀번호 확인
	if !util.CheckPasswordHash(password, user.Password) {
		return "", nil, model.NewInvalidCredentialsError()
	}

	// 4. JWT 토큰 생성
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
	// 1. 사용자 존재 확인
	exists, err := as.userRepo.ExistsByID(userID)
	if err != nil {
		return "", model.NewBusinessError(
			model.ErrDatabaseError,
			"사용자 확인 중 오류가 발생했습니다",
		)
	}
	if !exists {
		return "", model.NewUserNotFoundError()
	}

	// 2. 새 토큰 생성
	token, err := util.GenerateJWT(userID)
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

// IsUserAdmin 사용자가 관리자인지 확인합니다.
func (as *AuthService) IsUserAdmin(userID uint) bool {
	if userID == 0 {
		return false
	}

	user, err := as.userRepo.FindByID(userID)
	if err != nil {
		log.Printf("❌ 사용자 권한 확인 실패 (ID: %d): %v", userID, err)
		return false
	}

	return user.Role == model.RoleAdmin
}

// ========== 내부 헬퍼 함수들 ==========

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
