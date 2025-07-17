package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"ssh-key-manager/config"
	"ssh-key-manager/controllers"
	"ssh-key-manager/models"
	"ssh-key-manager/utils"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// 2. 데이터베이스 연결 및 마이그레이션
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	log.Printf("🔌 데이터베이스 연결 중...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 데이터베이스 연결 실패: %v", err)
	}

	// 자동 마이그레이션 실행
	log.Printf("📦 데이터베이스 마이그레이션 실행 중...")
	if err := db.AutoMigrate(&models.User{}, &models.SSHKey{}); err != nil {
		log.Fatalf("❌ 마이그레이션 실패: %v", err)
	}
	log.Printf("✅ 데이터베이스 연결 및 마이그레이션 완료")

	// 전역 DB 설정
	models.SetDB(db)

	// SSH 자동 설치 설정 검증 (활성화된 경우에만)
	if cfg.AutoInstallKeys {
		log.Printf("🔧 SSH 자동 설치 기능 검증 중...")
		if err := utils.ValidateSSHConfig(cfg.SSHUser, cfg.SSHHomePath); err != nil {
			log.Printf("⚠️ SSH 자동 설치 설정 오류: %v", err)
			log.Printf("💡 .env 파일에서 SSH_USER와 SSH_HOME_PATH를 확인하거나 AUTO_INSTALL_KEYS=false로 설정하세요")
		} else {
			log.Printf("✅ SSH 자동 설치 설정 검증 완료")
			log.Printf("   - SSH 사용자: %s", cfg.SSHUser)
			log.Printf("   - SSH 홈 경로: %s", cfg.SSHHomePath)
		}
	} else {
		log.Printf("📋 SSH 자동 설치 기능 비활성화됨")
	}

	// 3. Echo 인스턴스 생성 및 미들웨어 설정
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 4. 정적 파일 제공 (프론트엔드 UI)
	e.Static("/", "public")
	log.Printf("📁 정적 파일 서빙: /public")

	// 5. API 라우팅 설정
	api := e.Group("/api")

	// 공개 엔드포인트 (인증 불필요)
	api.POST("/register", controllers.Register)
	api.POST("/login", controllers.Login)

	// 보호된 엔드포인트 (JWT 인증 필요)
	protected := api.Group("/keys")

	// JWT 시크릿 확인
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET 환경변수가 설정되지 않았습니다")
	}

	// JWT 미들웨어 설정
	jwtConfig := echojwt.Config{
		SigningKey:  []byte(jwtSecret),
		ContextKey:  "user",
		TokenLookup: "header:Authorization:Bearer ",
	}
	protected.Use(echojwt.WithConfig(jwtConfig))

	// SSH 키 관리 API
	protected.POST("", controllers.CreateKey)   // 키 생성/재생성
	protected.GET("", controllers.GetKey)       // 키 조회
	protected.DELETE("", controllers.DeleteKey) // 키 삭제

	// 사용자 관리 API
	users := api.Group("/users")
	users.Use(echojwt.WithConfig(jwtConfig)) // 모든 사용자 API는 인증 필요

	users.GET("", controllers.GetUsers)             // 사용자 목록
	users.GET("/me", controllers.GetCurrentUser)    // 현재 사용자 정보
	users.PUT("/me", controllers.UpdateUserProfile) // 프로필 업데이트
	users.GET("/:id", controllers.GetUserDetail)    // 특정 사용자 상세

	// 6. 서버 시작
	serverAddr := ":" + cfg.ServerPort
	log.Printf("🌐 서버 시작: http://localhost%s", serverAddr)
	log.Printf("📋 사용 가능한 엔드포인트:")
	log.Printf("   - POST /api/register     : 사용자 등록")
	log.Printf("   - POST /api/login        : 사용자 로그인")
	log.Printf("   - POST /api/keys         : SSH 키 생성/재생성")
	log.Printf("   - GET  /api/keys         : SSH 키 조회")
	log.Printf("   - DELETE /api/keys       : SSH 키 삭제")
	log.Printf("   - GET  /api/users        : 사용자 목록")
	log.Printf("   - GET  /api/users/me     : 현재 사용자 정보")
	log.Printf("   - PUT  /api/users/me     : 프로필 업데이트")
	log.Printf("   - GET  /api/users/{id}   : 특정 사용자 상세")

	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ 서버 시작 실패: %v", err)
	}
}
