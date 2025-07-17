package main

import (
	"fmt"
	"log"
	"net/http" // 추가
	"os"
	"ssh-key-manager/config"
	"ssh-key-manager/controllers"
	"ssh-key-manager/models"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// ... (1~2단계 설정 코드는 이전과 동일) ...
	// 1. 설정 로드
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// 2. 데이터베이스 연결 및 마이그레이션
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&models.User{}, &models.SSHKey{})
	models.SetDB(db)

	// 3. Echo 인스턴스 생성
	e := echo.New()

	// 4. 미들웨어 설정
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS()) // ✅ 1. CORS 미들웨어 추가

	// 5. 정적 파일 제공 (UI)
	e.Static("/", "public") // ✅ 2. 'public' 디렉토리의 파일을 제공

	// 6. API 라우팅 설정
	api := e.Group("/api") // API 경로를 /api 그룹으로 분리
	{
		// 공개 엔드포인트
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)

		// 보호된 엔드포인트
		r := api.Group("/keys")
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			log.Fatal("JWT_SECRET must be set in .env file")
		}

		jwtConfig := echojwt.Config{
			SigningKey:  []byte(jwtSecret),
			ContextKey:  "user",
			TokenLookup: "header:Authorization:Bearer ",
		}
		r.Use(echojwt.WithConfig(jwtConfig))

		// 키 관련 API
		r.POST("", controllers.CreateKey)
		r.GET("", controllers.GetKey)
		r.DELETE("", controllers.DeleteKey)
	}

	// 7. 서버 시작
	serverAddr := ":" + cfg.ServerPort
	fmt.Printf("Starting server on http://localhost%s\n", serverAddr)
	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
