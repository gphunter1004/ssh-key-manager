package main

import (
	"log"
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/database"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/router"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Printf("🚀 SSH Key Manager 서버 시작")

	// 1. 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 설정 로드 실패: %v", err)
	}
	log.Printf("✅ 설정 로드 완료")

	// 2. 데이터베이스 초기화
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("❌ 데이터베이스 초기화 실패: %v", err)
	}
	log.Printf("✅ 데이터베이스 초기화 완료")

	// 3. Echo 인스턴스 생성
	e := echo.New()

	// 4. 글로벌 에러 핸들러 설정 (가장 먼저!)
	e.HTTPErrorHandler = middleware.CustomHTTPErrorHandler

	// 5. 미들웨어 설정
	e.Use(echomiddleware.Logger())
	e.Use(middleware.RecoverMiddleware()) // 사용자 정의 panic 복구
	e.Use(echomiddleware.CORS())

	// 6. 정적 파일 서빙
	e.Static("/", "web/static")

	// 7. 라우터 설정
	router.Setup(e, cfg)

	// 8. 서버 시작
	serverAddr := ":" + cfg.ServerPort
	log.Printf("🌐 서버 시작: http://localhost%s", serverAddr)

	if err := e.Start(serverAddr); err != nil {
		log.Fatalf("❌ 서버 시작 실패: %v", err)
	}
}
