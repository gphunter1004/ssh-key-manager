package service

import (
	"log"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"
	"strings"

	"gorm.io/gorm"
)

// RegisterUser는 새로운 사용자를 등록합니다.
func RegisterUser(username, password string) (*model.User, error) {
	log.Printf("👤 새 사용자 등록 시도: %s", username)

	// 1. 입력값 검증
	if err := validateRegistrationInput(username, password); err != nil {
		return nil, err
	}

	// 2. 사용자명 중복 확인
	if err := checkUsernameAvailability(username); err != nil {
		return nil, err
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
	user, err := createNewUser(username, hashedPassword)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ 사용자 등록 완료: %s (ID: %d)", username, user.ID)
	user.Password = "" // 응답에서 비밀번호 제거
	return user, nil
}

// AuthenticateUser는 사용자를 인증하고 JWT를 반환합니다.
func AuthenticateUser(username, password string) (string, *model.User, error) {
	log.Printf("🔐 사용자 인증 시도: %s", username)

	// 1. 입력값 검증
	if err := validateAuthenticationInput(username, password); err != nil {
		return "", nil, err
	}

	// 2. 사용자 조회
	user, err := getUserByUsername(username)
	if err != nil {
		return "", nil, err
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

// RefreshUserToken은 사용자의 JWT 토큰을 갱신합니다.
func RefreshUserToken(userID uint) (string, error) {
	// 1. 사용자 존재 확인
	if err := validateUserExists(userID); err != nil {
		return "", err
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

// GetUserByID는 사용자 ID로 사용자를 조회합니다.
func GetUserByID(userID uint) (*model.User, error) {
	if userID == 0 {
		return nil, model.NewBusinessError(
			model.ErrInvalidInput, 
			"유효하지 않은 사용자 ID입니다",
		)
	}

	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var user model.User
	if err := db.Select("id, username, role, created_at, updated_at").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	return &user, nil
}

// IsUserAdmin은 사용자가 관리자인지 확인합니다.
func IsUserAdmin(userID uint) bool {
	if userID == 0 {
		return false
	}

	db, err := model.GetDB()
	if err != nil {
		log.Printf("❌ DB 접근 실패 (IsUserAdmin): %v", err)
		return false
	}

	var user model.User
	if err := db.Select("role").First(&user, userID).Error; err != nil {
		log.Printf("❌ 사용자 권한 확인 실패 (ID: %d): %v", userID, err)
		return false
	}
	
	return user.Role == model.RoleAdmin
}

// ChangePassword는 사용자의 비밀번호를 변경합니다.
func ChangePassword(userID uint, currentPassword, newPassword string) error {
	log.Printf("🔑 비밀번호 변경 시도 (사용자 ID: %d)", userID)

	// 1. 입력값 검증
	if err := validatePasswordChangeInput(currentPassword, newPassword); err != nil {
		return err
	}

	// 2. 현재 사용자 조회
	user, err := getUserWithPassword(userID)
	if err != nil {
		return err
	}

	// 3. 현재 비밀번호 확인
	if !util.CheckPasswordHash(currentPassword, user.Password) {
		return model.NewBusinessError(
			model.ErrInvalidCredentials, 
			"현재 비밀번호가 올바르지 않습니다",
		)
	}

	// 4. 새 비밀번호 해시
	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return model.NewBusinessError(
			model.ErrInternalServer, 
			"새 비밀번호 처리 중 오류가 발생했습니다",
		)
	}

	// 5. 비밀번호 업데이트
	if err := updateUserPassword(userID, hashedPassword); err != nil {
		return err
	}

	log.Printf("✅ 비밀번호 변경 완료 (사용자 ID: %d)", userID)
	return nil
}

// ========== 내부 헬퍼 함수들 (비즈니스 로직) ==========

// validateRegistrationInput은 회원가입 입력값을 검증합니다.
func validateRegistrationInput(username, password string) error {
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

// checkUsernameAvailability는 사용자명 중복을 확인합니다.
func checkUsernameAvailability(username string) error {
	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"사용자명 중복 확인 중 오류가 발생했습니다",
		)
	}

	if count > 0 {
		return model.NewBusinessError(
			model.ErrUserAlreadyExists, 
			"이미 사용 중인 사용자명입니다",
		)
	}

	return nil
}

// createNewUser는 새로운 사용자를 생성합니다.
func createNewUser(username, hashedPassword string) (*model.User, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	user := model.User{
		Username: strings.TrimSpace(username),
		Password: hashedPassword,
		Role:     model.RoleUser,
	}

	if err := db.Create(&user).Error; err != nil {
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

	return &user, nil
}

// validateAuthenticationInput은 로그인 입력값을 검증합니다.
func validateAuthenticationInput(username, password string) error {
	username = strings.TrimSpace(username)
	
	if username == "" || password == "" {
		return model.NewBusinessError(
			model.ErrRequiredField, 
			"사용자명과 비밀번호를 입력해주세요",
		)
	}

	return nil
}

// getUserByUsername은 사용자명으로 사용자를 조회합니다.
func getUserByUsername(username string) (*model.User, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewInvalidCredentialsError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	return &user, nil
}

// validateUserExists는 사용자 존재를 확인합니다.
func validateUserExists(userID uint) error {
	if userID == 0 {
		return model.NewBusinessError(
			model.ErrInvalidInput, 
			"유효하지 않은 사용자 ID입니다",
		)
	}

	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var count int64
	if err := db.Model(&model.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"사용자 확인 중 오류가 발생했습니다",
		)
	}

	if count == 0 {
		return model.NewUserNotFoundError()
	}

	return nil
}

// validatePasswordChangeInput은 비밀번호 변경 입력값을 검증합니다.
func validatePasswordChangeInput(currentPassword, newPassword string) error {
	if currentPassword == "" {
		return model.NewBusinessError(
			model.ErrRequiredField, 
			"현재 비밀번호를 입력해주세요",
		)
	}

	if newPassword == "" {
		return model.NewBusinessError(
			model.ErrRequiredField, 
			"새 비밀번호를 입력해주세요",
		)
	}

	if len(newPassword) < 4 {
		return model.NewBusinessError(
			model.ErrWeakPassword, 
			"새 비밀번호는 최소 4자 이상이어야 합니다",
		)
	}

	if len(newPassword) > 100 {
		return model.NewBusinessError(
			model.ErrWeakPassword, 
			"새 비밀번호는 최대 100자까지 가능합니다",
		)
	}

	if currentPassword == newPassword {
		return model.NewBusinessError(
			model.ErrInvalidInput, 
			"새 비밀번호는 현재 비밀번호와 달라야 합니다",
		)
	}

	return nil
}

// getUserWithPassword는 비밀번호를 포함한 사용자 정보를 조회합니다.
func getUserWithPassword(userID uint) (*model.User, error) {
	db, err := model.GetDB()
	if err != nil {
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	var user model.User
	if err := db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, model.NewUserNotFoundError()
		}
		return nil, model.NewBusinessError(
			model.ErrDatabaseError, 
			"사용자 조회 중 오류가 발생했습니다",
		)
	}

	return &user, nil
}

// updateUserPassword는 사용자의 비밀번호를 업데이트합니다.
func updateUserPassword(userID uint, hashedPassword string) error {
	db, err := model.GetDB()
	if err != nil {
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"데이터베이스 연결 오류가 발생했습니다",
		)
	}

	if err := db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		log.Printf("❌ 비밀번호 업데이트 실패 (사용자 ID: %d): %v", userID, err)
		return model.NewBusinessError(
			model.ErrDatabaseError, 
			"비밀번호 변경 중 오류가 발생했습니다",
		)
	}

	return nil
}
