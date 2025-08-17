package database

import (
	"log"
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/util"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Initialize는 데이터베이스를 초기화합니다.
func Initialize(cfg *config.Config) error {
	// 데이터베이스 연결
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		return err
	}

	// 연결 테스트
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	if err := sqlDB.Ping(); err != nil {
		return err
	}

	// 전역 DB 설정
	model.SetDB(db)

	// 마이그레이션 실행
	if err := runMigrations(db); err != nil {
		return err
	}

	// 초기 관리자 계정 생성
	if err := createInitialAdmin(db, cfg); err != nil {
		log.Printf("⚠️ 초기 관리자 계정 생성 실패: %v", err)
	}

	return nil
}

// runMigrations는 데이터베이스 마이그레이션을 실행합니다.
func runMigrations(db *gorm.DB) error {
	log.Printf("📦 데이터베이스 마이그레이션 시작...")

	models := []interface{}{
		&model.User{},
		&model.SSHKey{},
		&model.Server{},
		&model.ServerKeyDeployment{},
		&model.Department{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
		log.Printf("   - %T 마이그레이션 완료", m)
	}

	log.Printf("✅ 데이터베이스 마이그레이션 완료")
	return nil
}

// createInitialAdmin은 초기 관리자 계정을 생성합니다.
func createInitialAdmin(db *gorm.DB, cfg *config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Printf("📋 관리자 계정 설정이 없습니다. 건너뜀")
		return nil
	}

	// 관리자 존재 여부 확인
	var adminCount int64
	if err := db.Model(&model.User{}).Where("role = ?", model.RoleAdmin).Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		log.Printf("⚠️ 관리자 계정이 이미 존재합니다. 건너뜀")
		return nil
	}

	// 기존 사용자 확인
	var existingUser model.User
	result := db.Where("username = ?", cfg.AdminUsername).First(&existingUser)
	if result.Error == nil {
		// 기존 사용자를 관리자로 승격
		if err := db.Model(&existingUser).Update("role", model.RoleAdmin).Error; err != nil {
			return err
		}
		log.Printf("✅ 기존 사용자 %s를 관리자로 승격", cfg.AdminUsername)
		return nil
	}

	// 새 관리자 계정 생성
	hashedPassword, err := util.HashPassword(cfg.AdminPassword)
	if err != nil {
		return err
	}

	admin := model.User{
		Username: cfg.AdminUsername,
		Password: hashedPassword,
		Role:     model.RoleAdmin,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	log.Printf("✅ 초기 관리자 계정 생성 완료: %s", cfg.AdminUsername)
	return nil
}
