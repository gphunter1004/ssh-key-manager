package main

import (
	"log"
	"os/exec"
	"ssh-key-manager/internal/config"
	"ssh-key-manager/internal/database"
	"ssh-key-manager/internal/middleware"
	"ssh-key-manager/internal/model"
	"ssh-key-manager/internal/router"
	"ssh-key-manager/internal/util"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func main() {
	log.Printf("🚀 SSH Key Manager 서버 시작")

	// 0. 시스템 의존성 체크
	if err := checkSystemDependencies(); err != nil {
		log.Fatalf("❌ 시스템 의존성 체크 실패: %v", err)
	}
	log.Printf("✅ 시스템 의존성 체크 완료")

	// 1. 설정 로드
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 설정 로드 실패: %v", err)
	}
	log.Printf("✅ 설정 로드 완료")

	// 1-1. JWT 시크릿 초기화 (데이터베이스 초기화 전에 실행)
	util.InitializeJWTSecret(cfg.JWTSecret)
	log.Printf("✅ JWT 시크릿 초기화 완료")

	// 2. 데이터베이스 초기화
	if err := database.Initialize(cfg); err != nil {
		log.Fatalf("❌ 데이터베이스 초기화 실패: %v", err)
	}
	log.Printf("✅ 데이터베이스 초기화 완료")

	// 2-1. 데이터베이스 연결 상태 확인
	if !model.IsDBInitialized() {
		log.Fatalf("❌ 데이터베이스가 초기화되지 않았습니다")
	}
	log.Printf("✅ 데이터베이스 연결 상태 확인 완료")

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

// checkSystemDependencies는 시스템 의존성을 체크합니다.
func checkSystemDependencies() error {
	log.Printf("🔍 시스템 의존성 체크 중...")

	// SSH 클라이언트 존재 확인
	if _, err := exec.LookPath("ssh"); err != nil {
		log.Printf("❌ SSH 클라이언트를 찾을 수 없습니다: %v", err)
		log.Printf("💡 해결 방법:")
		log.Printf("   - Ubuntu/Debian: sudo apt-get install openssh-client")
		log.Printf("   - CentOS/RHEL: sudo yum install openssh-clients")
		log.Printf("   - macOS: SSH는 기본 설치됨")
		return err
	}
	log.Printf("   ✓ SSH 클라이언트 확인됨")

	// SSH-KEYGEN 존재 확인 (키 생성용)
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		log.Printf("⚠️ ssh-keygen을 찾을 수 없습니다: %v", err)
		log.Printf("💡 일부 고급 기능이 제한될 수 있습니다")
	} else {
		log.Printf("   ✓ ssh-keygen 확인됨")
	}

	return nil
}
