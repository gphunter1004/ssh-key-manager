package main

import (
	"fmt"
	"log"
	"net/http"
	"ssh-key-manager/config"
	"ssh-key-manager/models"
	"ssh-key-manager/routes"
	"ssh-key-manager/services"
	"ssh-key-manager/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🚀 SSH Key Manager 서버 시작")

	// 1. 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("❌ 설정 로드 실패: %v", err)
	}
	log.Printf("✅ 설정 로드 완료")

	// 2. 데이터베이스 초기화
	if err := initializeDatabase(cfg); err != nil {
		log.Fatalf("❌ 데이터베이스 초기화 실패: %v", err)
	}
	log.Printf("✅ 데이터베이스 초기화 완료")

	// 3. 초기 관리자 계정 생성
	if err := createAdminUser(cfg); err != nil {
		log.Printf("⚠️ 초기 관리자 계정 생성 실패: %v", err)
	}

	// 4. SSH 설정 검증
	validateSSHSettings(cfg)

	// 5. 웹 서버 시작
	if err := startWebServer(cfg); err != nil {
		log.Fatalf("❌ 서버 시작 실패: %v", err)
	}
}

// initializeDatabase는 데이터베이스 연결 및 마이그레이션을 수행합니다.
func initializeDatabase(cfg *config.Config) error {
	log.Printf("🔌 데이터베이스 연결 중...")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 마이그레이션 실행
	log.Printf("📦 데이터베이스 마이그레이션 실행 중...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.SSHKey{},
		&models.Server{},
		&models.ServerKeyDeployment{},
	); err != nil {
		return fmt.Errorf("마이그레이션 실패: %w", err)
	}

	// 전역 DB 설정
	models.SetDB(db)

	return nil
}

// createAdminUser는 초기 관리자 계정을 생성합니다.
func createAdminUser(cfg *config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		log.Printf("📋 관리자 계정 설정이 없습니다. 건너뜀")
		return nil
	}

	log.Printf("👑 초기 관리자 계정 생성 중...")
	return services.CreateAdminUser(cfg.AdminUsername, cfg.AdminPassword)
}

// validateSSHSettings는 SSH 설정을 검증합니다.
func validateSSHSettings(cfg *config.Config) {
	if !cfg.AutoInstallKeys {
		log.Printf("📋 SSH 자동 설치 기능 비활성화됨")
		return
	}

	log.Printf("🔧 SSH 자동 설치 기능 검증 중...")
	if err := utils.ValidateSSHConfig(cfg.SSHUser, cfg.SSHHomePath); err != nil {
		log.Printf("⚠️ SSH 자동 설치 설정 오류: %v", err)
		log.Printf("💡 .env 파일에서 SSH_USER와 SSH_HOME_PATH를 확인하거나 AUTO_INSTALL_KEYS=false로 설정하세요")
	} else {
		log.Printf("✅ SSH 자동 설치 설정 검증 완료")
		log.Printf("   - SSH 사용자: %s", cfg.SSHUser)
		log.Printf("   - SSH 홈 경로: %s", cfg.SSHHomePath)
	}
}

// startWebServer는 웹 서버를 시작합니다.
func startWebServer(cfg *config.Config) error {
	e := echo.New()

	// 라우트 설정
	log.Printf("🛣️ 라우트 설정 중...")
	if err := routes.SetupRoutes(e); err != nil {
		return fmt.Errorf("라우트 설정 실패: %w", err)
	}

	// 서버 시작
	serverAddr := ":" + cfg.ServerPort
	log.Printf("🌐 서버 시작: http://localhost%s", serverAddr)
	log.Printf("📁 정적 파일 서빙: /public")

	// 엔드포인트 목록 출력
	//routes.LogRoutes()

	// 서버 실행
	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("서버 실행 실패: %w", err)
	}

	return nil
}
