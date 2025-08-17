package main

import (
	"fmt"
	"log"
	"net/http"
	"ssh-key-manager/config"
	"ssh-key-manager/database"
	"ssh-key-manager/models"
	"ssh-key-manager/routes"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	log.Printf("🚀 SSH Key Manager 서버 시작")

	// 1. 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("❌ 설정 로드 실패: %v", err)
	}
	log.Printf("✅ 설정 로드 완료")

	// 2. 마이그레이션 실행 (서버 시작 전)
	if err := runMigrations(cfg); err != nil {
		log.Fatalf("❌ 마이그레이션 실패: %v", err)
	}

	// 3. 데이터베이스 연결
	if err := connectDatabase(cfg); err != nil {
		log.Fatalf("❌ 데이터베이스 연결 실패: %v", err)
	}
	log.Printf("✅ 데이터베이스 연결 완료")

	// 4. 웹 서버 시작
	if err := startWebServer(cfg); err != nil {
		log.Fatalf("❌ 서버 시작 실패: %v", err)
	}
}

// runMigrations는 마이그레이션을 실행합니다.
func runMigrations(cfg *config.Config) error {
	log.Printf("📦 데이터베이스 마이그레이션 실행 중...")

	migrationMgr, err := database.NewMigrationManager(cfg)
	if err != nil {
		return fmt.Errorf("마이그레이션 관리자 생성 실패: %w", err)
	}
	defer migrationMgr.Close()

	if err := migrationMgr.RunMigrations(); err != nil {
		return fmt.Errorf("마이그레이션 실행 실패: %w", err)
	}

	log.Printf("✅ 마이그레이션 완료")
	return nil
}

// connectDatabase는 데이터베이스에 연결합니다.
func connectDatabase(cfg *config.Config) error {
	log.Printf("🔌 데이터베이스 연결 중...")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	// 연결 확인
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("데이터베이스 인스턴스 접근 실패: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("데이터베이스 연결 테스트 실패: %w", err)
	}

	// 전역 DB 설정
	models.SetDB(db)

	return nil
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

	// 서버 실행
	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("서버 실행 실패: %w", err)
	}

	return nil
}
