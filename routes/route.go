// routes/route.go - 간소화된 라우터 설정

package routes

import (
	"fmt"
	"os"
	"ssh-key-manager/controllers"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// SetupRoutes는 모든 라우팅을 설정합니다.
func SetupRoutes(e *echo.Echo) error {
	// 미들웨어 설정
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// 정적 파일 제공 (프론트엔드 UI)
	e.Static("/", "public")

	// JWT 설정
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return fmt.Errorf("JWT_SECRET 환경변수가 설정되지 않았습니다")
	}

	jwtConfig := echojwt.Config{
		SigningKey:  []byte(jwtSecret),
		ContextKey:  "user",
		TokenLookup: "header:Authorization:Bearer ",
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(401, map[string]interface{}{
				"success": false,
				"error":   "invalid or expired jwt",
				"details": err.Error(),
			})
		},
	}

	// API 그룹 설정
	setupPublicRoutes(e)
	setupAuthenticatedRoutes(e, jwtConfig)
	setupAdminRoutes(e, jwtConfig)

	return nil
}

// setupPublicRoutes는 인증이 필요없는 공개 라우트를 설정합니다.
func setupPublicRoutes(e *echo.Echo) {
	api := e.Group("/api")

	// 인증 관련 (공개)
	api.POST("/register", controllers.Register)
	api.POST("/login", controllers.Login)

	// 헬스체크
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "ok",
			"service": "SSH Key Manager",
		})
	})
}

// setupAuthenticatedRoutes는 인증이 필요한 라우트를 설정합니다.
func setupAuthenticatedRoutes(e *echo.Echo, jwtConfig echojwt.Config) {
	api := e.Group("/api")

	// 인증된 사용자 전용 라우트
	auth := api.Group("")
	auth.Use(echojwt.WithConfig(jwtConfig))

	// 토큰 검증 엔드포인트
	auth.GET("/validate", controllers.ValidateToken)
	auth.POST("/refresh", controllers.RefreshToken)

	// SSH 키 관리 API
	keys := auth.Group("/keys")
	keys.POST("", controllers.CreateKey)   // 키 생성/재생성
	keys.GET("", controllers.GetKey)       // 키 조회
	keys.DELETE("", controllers.DeleteKey) // 키 삭제

	// 개별 사용자 관리 API (본인만 접근 가능)
	users := auth.Group("/users")
	users.GET("/me", controllers.GetCurrentUser)    // 현재 사용자 정보
	users.PUT("/me", controllers.UpdateUserProfile) // 프로필 업데이트

	// 서버 관리 API
	servers := auth.Group("/servers")
	servers.POST("", controllers.CreateServer)                    // 서버 등록
	servers.GET("", controllers.GetServers)                       // 서버 목록
	servers.GET("/:id", controllers.GetServer)                    // 서버 상세
	servers.PUT("/:id", controllers.UpdateServer)                 // 서버 수정
	servers.DELETE("/:id", controllers.DeleteServer)              // 서버 삭제
	servers.POST("/:id/test", controllers.TestServerConnection)   // 서버 연결 테스트
	servers.POST("/deploy", controllers.DeployKeyToServers)       // 키 배포
	servers.GET("/deployments", controllers.GetDeploymentHistory) // 배포 기록
}

// setupAdminRoutes는 관리자 전용 라우트를 설정합니다.
func setupAdminRoutes(e *echo.Echo, jwtConfig echojwt.Config) {
	api := e.Group("/api")

	// 관리자 전용 API
	admin := api.Group("/admin")
	admin.Use(echojwt.WithConfig(jwtConfig)) // JWT 인증 필요
	admin.Use(controllers.AdminRequired)     // 관리자 권한 필요

	admin.GET("/users", controllers.GetAllUsersAdmin)        // 모든 사용자 (관리자용)
	admin.GET("/users/:id", controllers.GetUserDetail)       // 특정 사용자 상세 (관리자용)
	admin.PUT("/users/:id/role", controllers.UpdateUserRole) // 사용자 권한 변경
	admin.DELETE("/users/:id", controllers.DeleteUser)       // 사용자 삭제

	// 사용자 목록 조회도 관리자 전용으로 이동
	admin.GET("/users-list", controllers.GetUsers) // 기본 사용자 목록 (관리자용)
}
